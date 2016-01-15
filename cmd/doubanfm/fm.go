package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

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
	OpChann  = "0"
	PROMPT   = "DoubanFM> "
)

type DoubanFM struct {
	Channels  map[string]*doubanfm.Channel // channel list
	Songs     []doubanfm.Song              // playlist
	Song      doubanfm.Song                // current song
	Channel   *doubanfm.Channel            // current channel
	Paused    bool                         // play/pause status
	Loop      bool                         // loop status
	User      *doubanfm.User               // login user
	player    *Player                      // gstreamer player
	channlist []string
}

func NewDoubanFM() (*DoubanFM, error) {
	player, err := newPlayer()
	if err != nil {
		return nil, err
	}
	dfm := &DoubanFM{
		Channels:  make(map[string]*doubanfm.Channel),
		player:    player,
		channlist: make([]string, 1, 100),
	}
	dfm.channlist[0] = "-3"
	dfm.Channels["-3"] = &doubanfm.Channel{Id: "-3", Name: "红星兆赫"}
	player.init(dfm.onMessage)
	return dfm, nil
}

func (this *DoubanFM) Empty() bool {
	return len(this.Songs) == 0
}

func (this *DoubanFM) Next() (song *doubanfm.Song) {
	if this.Empty() {
		return nil
	}
	this.Song = this.Songs[0]
	this.Songs = this.Songs[1:]
	return &this.Song
}

func (this *DoubanFM) onMessage(bus *gst.Bus, msg *gst.Message) {
	switch msg.GetType() {
	case gst.MESSAGE_EOS:
		if this.Loop {
			this.playSong(&this.Song)
		} else {
			this.GetSongs(doubanfm.End)
			if this.Empty() {
				this.GetSongs(doubanfm.Last)
			}
			this.playSong(this.Next())
		}
	case gst.MESSAGE_ERROR:
		s, param := msg.GetStructure()
		fmt.Println("\n[gstreamer]", msg.GetType(), s, param)
		fmt.Print(PROMPT)
		this.player.Stop()
	}
}

func (this *DoubanFM) playSong(song *doubanfm.Song) {
	if song == nil {
		return
	}
	fmt.Printf("\rPLAYING>> %s - %s\n", song.Title, song.Artist)
	this.player.Play(song.Url)
}

func (this *DoubanFM) GetChannels() {
	chls, err := doubanfm.Channels()
	if err != nil {
		fmt.Println(err)
	}
	for i, _ := range chls {
		if _, ok := this.Channels[chls[i].Id.String()]; !ok {
			this.Channels[chls[i].Id.String()] = &chls[i]
			this.channlist = append(this.channlist, chls[i].Id.String())
		}
	}
}

func (this *DoubanFM) GetMyChannels() {
	if this.User == nil {
		return
	}
	favs, recs, err := doubanfm.MyChannels(this.User.Id)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("   favorite channels:", favs)
	fmt.Println("recommended channels:", recs)
	for _, fav := range favs {
		if _, ok := this.Channels[fav.Id.String()]; !ok {
			ch := toChannel(fav)
			this.Channels[fav.Id.String()] = &ch
			this.channlist = append(this.channlist, fav.Id.String())
		} else {
			this.Channels[fav.Id.String()].Fav = true
		}
	}
	for _, rec := range recs {
		if _, ok := this.Channels[rec.Id.String()]; !ok {
			ch := toChannel(rec)
			this.Channels[rec.Id.String()] = &ch
			this.channlist = append(this.channlist, rec.Id.String())
		}
	}
}

func toChannel(chl doubanfm.MyChannel) doubanfm.Channel {
	return doubanfm.Channel{
		Id:   chl.Id,
		Name: chl.Name,
	}
}

func (this *DoubanFM) GetSongs(types string) {
	if this.Channel == nil {
		fmt.Println("\r>>>>>>>>> Error in fetching songs: nil channel")
		return
	}
	songs, err := doubanfm.Songs(types, this.Channel.Id.String(), this.Song.Sid, this.User)
	if err != nil {
		fmt.Println("\r>>>>>>>>> Error in fetching songs:", err)
		return
	}

	if len(songs) > 0 {
		this.Songs = songs
	}
}

func (this *DoubanFM) Login(uid string) error {
	term := newTerm("Douban Id: ")
	defer term.Restore()
	var err error
	if uid == "" {
		uid, err = term.ReadLine()
		if err != nil {
			return err
		}
	}
	uid = strings.TrimSpace(uid)
	if len(uid) == 0 {
		return fmt.Errorf("empty id")
	}
	pwd, err := term.ReadPassword("Password: ")
	if err != nil {
		return err
	}
	this.User, err = doubanfm.Login(uid, string(pwd))
	return err
}

func (this *DoubanFM) SetDefaultChannel() {
	if this.User != nil {
		this.Channel = this.Channels["-3"]
	} else {
		this.Channel = this.Channels["0"]
	}
}

func (this *DoubanFM) printChannel() {
	fmt.Println(this.Channel.String())
}

func (this *DoubanFM) printChannels() {
	b := &bytes.Buffer{}
	for j, id := range this.channlist {
		cur := "-"
		fav := ""
		if id == this.Channel.Id.String() {
			cur = "+"
		}
		if this.Channels[id].Fav {
			fav = "*"
		}
		fmt.Fprintf(b, "%2d %s [%s]\r\t\t%s %s\n", j+1, cur, this.Channels[id].Id, this.Channels[id].Name, fav)
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
		fmt.Println("\r>>>>>>>>> Not logon")
		return
	}
	fmt.Printf("\r    Id:\t%s\n  Name:\t%s\n Token:\t%s\nExpire:\t%s\n",
		this.User.Id, this.User.Name, this.User.Token, parseTime(this.User.Expire))
}

func parseTime(ut string) string {
	if sec, err := strconv.ParseInt(ut, 10, 64); err == nil {
		t := time.Unix(sec, 0)
		return t.String()
	}
	return ut
}
