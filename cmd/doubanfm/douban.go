package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ziutek/gst"
	"github.com/zyxar/doubanfm"
)

const (
	OpAgain  = ""
	OpPlay   = "p"
	OpLoop   = "x"
	OpNext   = "n"
	OpSkip   = "s"
	OpTrash  = "t"
	OpLike   = "r"
	OpUnlike = "u"
	OpList   = "l"
	OpSong   = "c"
	OpLogin  = "z"
	OpHelp   = "h"
	OpExit   = "q"
	PROMPT   = "DoubanFM> "
)

type Channel struct {
	Id   string
	Name string
	Fav  bool
}

func (c Channel) String() string {
	return c.Id + " - " + c.Name
}

type DoubanFM struct {
	Channels []Channel       // channel list
	Songs    []doubanfm.Song // playlist
	Song     doubanfm.Song   // current song
	Channel  int             // current channel
	Paused   bool            // play/pause status
	Loop     bool            // loop status
	User     *doubanfm.User  // login user
	player   *Player
	opChan   chan string
}

func NewDoubanFM() (*DoubanFM, error) {
	player, err := newPlayer()
	if err != nil {
		return nil, err
	}
	dfm := &DoubanFM{
		opChan: make(chan string, 1),
		player: player,
	}
	player.init(dfm.onMessage)
	return dfm, nil
}

func (this *DoubanFM) Exec(op string) {
	select {
	case this.opChan <- op:
	default:
	}
}

func (this *DoubanFM) Empty() bool {
	return len(this.Songs) == 0
}

func (this *DoubanFM) Next() (song doubanfm.Song) {
	if this.Empty() {
		return
	}
	this.Song = this.Songs[0]
	this.Songs = this.Songs[1:]
	return this.Song
}

func (this *DoubanFM) onMessage(bus *gst.Bus, msg *gst.Message) {
	switch msg.GetType() {
	case gst.MESSAGE_EOS:
		if this.Loop {
			this.playNext(this.Song)
		} else {
			this.GetSongs(doubanfm.End)
			if this.Empty() {
				this.GetSongs(doubanfm.Last)
			}
			this.playNext(this.Next())
		}
	case gst.MESSAGE_ERROR:
		s, param := msg.GetStructure()
		log.Println("\n[gstreamer]", msg.GetType(), s, param)
		fmt.Print(PROMPT)
		this.player.Stop()
	}
}

func (this *DoubanFM) playNext(song doubanfm.Song) {
	fmt.Printf("\rPLAYING>> %s - %s\n", song.Title, song.Artist)
	this.player.NewSource(song.Url)
}

func (this *DoubanFM) GetChannels() {
	chls, err := doubanfm.Channels()
	if err != nil {
		log.Println(err)
	}
	var ch []Channel
	for _, chl := range chls {
		ch = append(ch, toChannel(chl))
	}
	this.Channels = ch
}

func (this *DoubanFM) GetLoginChannels() {
	if this.User == nil {
		return
	}
	favs, recs, err := doubanfm.LoginChannels(this.User.Id)
	if err != nil {
		log.Println(err)
	}

	for _, fav := range favs {
		find := false
		for i, chl := range this.Channels {
			if chl.Id == fav.Id.String() {
				this.Channels[i].Fav = true
				find = true
			}
		}
		if !find {
			this.Channels = append(this.Channels, toChannelLogin(fav))
		}
	}
	for _, rec := range recs {
		this.Channels = append(this.Channels, toChannelLogin(rec))
	}
}

func toChannel(chl doubanfm.Channel) Channel {
	return Channel{
		Id:   chl.Id.String(),
		Name: chl.Name,
	}
}

func toChannelLogin(chl doubanfm.LoginChannel) Channel {
	return Channel{
		Id:   chl.Id.String(),
		Name: chl.Name,
	}
}

func (this *DoubanFM) GetSongs(types string) {
	chl := this.Channels[this.Channel-1].Id
	songs, err := doubanfm.Songs(types, chl, this.Song.Sid, this.User)
	if err != nil {
		log.Println(err)
		return
	}

	if len(songs) > 0 {
		this.Songs = songs
	}
}

func (this *DoubanFM) Login() {
	var uid, pwd string

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Douban ID: ")
		uid, _ = reader.ReadString('\n')
		uid = strings.TrimSpace(uid)
		if uid != "" {
			break
		}
	}

	for {
		fmt.Print("Password: ")
		pwd, _ = reader.ReadString('\n')
		pwd = strings.TrimRight(pwd, "\n")
		if pwd != "" {
			break
		}
	}

	this.User, _ = doubanfm.Login(uid, pwd)
}

func (this *DoubanFM) printChannels() {
	b := &bytes.Buffer{}
	for i, chl := range this.Channels {
		cur := "-"
		fav := ""
		if i == this.Channel-1 {
			cur = "+"
		}
		if chl.Fav {
			fav = "*"
		}
		fmt.Fprintf(b, "%2d %s %s %s\n", i+1, cur, chl.Name, fav)
	}
	fmt.Println(b)
}

func (this *DoubanFM) printPlaylist() {
	b := &bytes.Buffer{}
	if this.Song.Sid != "" {
		loop := "-"
		if this.Loop {
			loop = "*"
		}
		fmt.Fprintf(b, "%s %s %s\n",
			this.Song.Title, loop, this.Song.Artist)
	}
	for _, song := range this.Songs {
		fmt.Fprintf(b, "%s - %s\n",
			song.Title, song.Artist)
	}
	fmt.Println(b)
}

func (this *DoubanFM) printSong() {
	fmt.Println(this.Song)
}

func (this *DoubanFM) printUser() {
	if this.User == nil {
		fmt.Println("Not Login")
		return
	}
	b := new(bytes.Buffer)
	fmt.Fprintf(b, "Id: %s\n", this.User.Id)
	fmt.Fprintf(b, "Email: %s\n", this.User.Email)
	fmt.Fprintf(b, "Name: %s\n", this.User.Name)
	fmt.Println(b)
}
