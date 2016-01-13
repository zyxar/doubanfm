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
	Loop     bool
	User     *doubanfm.User // login user
	opChan   chan string
	gst      *gstreamer
}

func NewDoubanFM() (*DoubanFM, error) {
	gst, err := newGstreamer()
	if err != nil {
		return nil, err
	}
	db := &DoubanFM{
		opChan: make(chan string, 1),
		gst:    gst,
	}
	gst.init(db.onMessage)
	return db, nil
}

func (db *DoubanFM) Exec(op string) {
	select {
	case db.opChan <- op:
	default:
	}
}

func (db *DoubanFM) Empty() bool {
	return len(db.Songs) == 0
}

func (db *DoubanFM) Next() (song doubanfm.Song) {
	if db.Empty() {
		return
	}
	db.Song = db.Songs[0]
	db.Songs = db.Songs[1:]
	return db.Song
}

func (db *DoubanFM) onMessage(bus *gst.Bus, msg *gst.Message) {
	switch msg.GetType() {
	case gst.MESSAGE_EOS:
		if db.Loop {
			db.playNext(db.Song)
		} else {
			db.GetSongs(doubanfm.End)
			if db.Empty() {
				db.GetSongs(doubanfm.Last)
			}
			db.playNext(db.Next())
		}
	case gst.MESSAGE_ERROR:
		s, param := msg.GetStructure()
		log.Println("\n[gstreamer]", msg.GetType(), s, param)
		fmt.Print(PROMPT)
		db.gst.Stop()
	}
}

func (db *DoubanFM) playNext(song doubanfm.Song) {
	fmt.Printf("\rPLAYING>> %s - %s\n", song.Title, song.Artist)
	db.gst.NewSource(song.Url)
}

func (db *DoubanFM) GetChannels() {
	chls, err := doubanfm.Channels()
	if err != nil {
		log.Println(err)
	}
	var ch []Channel
	for _, chl := range chls {
		ch = append(ch, toChannel(chl))
	}
	db.Channels = ch
}

func (db *DoubanFM) GetLoginChannels() {
	if db.User == nil {
		return
	}
	favs, recs, err := doubanfm.LoginChannels(db.User.Id)
	if err != nil {
		log.Println(err)
	}

	for _, fav := range favs {
		find := false
		for i, chl := range db.Channels {
			if chl.Id == fav.Id.String() {
				db.Channels[i].Fav = true
				find = true
			}
		}
		if !find {
			db.Channels = append(db.Channels, toChannelLogin(fav))
		}
	}
	for _, rec := range recs {
		db.Channels = append(db.Channels, toChannelLogin(rec))
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

func (db *DoubanFM) GetSongs(types string) {
	chl := db.Channels[db.Channel-1].Id
	songs, err := doubanfm.Songs(types, chl, db.Song.Sid, db.User)
	if err != nil {
		log.Println(err)
		return
	}

	if len(songs) > 0 {
		db.Songs = songs
	}
}

func (db *DoubanFM) Login() {
	var id, pwd string

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Douban ID: ")
		id, _ = reader.ReadString('\n')
		id = strings.TrimSpace(id)
		if id != "" {
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

	db.User, _ = doubanfm.Login(id, pwd)
}

func (db *DoubanFM) printChannels() {
	b := &bytes.Buffer{}
	for i, chl := range db.Channels {
		cur := "-"
		fav := ""
		if i == db.Channel-1 {
			cur = "+"
		}
		if chl.Fav {
			fav = "*"
		}
		fmt.Fprintf(b, "%2d %s %s %s\n", i+1, cur, chl.Name, fav)
	}
	fmt.Println(b)
}

func (db *DoubanFM) printPlaylist() {
	b := &bytes.Buffer{}
	if db.Song.Sid != "" {
		loop := "-"
		if db.Loop {
			loop = "*"
		}
		fmt.Fprintf(b, "%s %s %s\n",
			db.Song.Title, loop, db.Song.Artist)
	}
	for _, song := range db.Songs {
		fmt.Fprintf(b, "%s - %s\n",
			song.Title, song.Artist)
	}
	fmt.Println(b)
}

func (db *DoubanFM) printSong() {
	fmt.Println(db.Song)
}

func (db *DoubanFM) printUser() {
	if db.User == nil {
		fmt.Println("Not Login")
		return
	}
	b := new(bytes.Buffer)
	fmt.Fprintf(b, "Id: %s\n", db.User.Id)
	fmt.Fprintf(b, "Email: %s\n", db.User.Email)
	fmt.Fprintf(b, "Name: %s\n", db.User.Name)
	fmt.Println(b)
}
