package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/zyxar/doubanfm"
)

func main() {
	loop()
}

func loop() {
	term := newTerm()
	defer term.Restore()

	db, err := NewDoubanFM()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	db.GetChannels()
	db.Channel = 1
	db.GetSongs(doubanfm.New)
	db.playNext(db.Next())
	var op string
	var prevOp = OpNext
	for {
		op, err = term.ReadLine()
		if err != nil {
			panic(err)
		}
		op = strings.ToLower(strings.TrimSpace(op))
	PREV:
		switch op {
		case OpAgain:
			op = prevOp
			goto PREV
		case OpPlay:
			db.Paused = !db.Paused
			if db.Paused {
				db.gst.Pause()
			} else {
				db.gst.Play()
			}
		case OpLoop:
			db.Loop = !db.Loop
		case OpNext:
			if db.Empty() {
				db.GetSongs(doubanfm.Last)
			}
			db.playNext(db.Next())
		case OpSkip:
			db.GetSongs(doubanfm.Skip)
			db.playNext(db.Next())
		case OpTrash:
			db.GetSongs(doubanfm.Bypass)
			db.playNext(db.Next())
		case OpLike:
			db.GetSongs(doubanfm.Like)
			db.Song.Like = 1
		case OpUnlike:
			db.GetSongs(doubanfm.Unlike)
			db.Song.Like = 0
		case OpLogin:
			if db.User != nil {
				db.printUser()
				continue
			}
			db.Login()

			if db.User == nil {
				fmt.Println("Login Failed")
				continue
			}
			chls := []Channel{
				{Id: "-3", Name: "红星兆赫"},
			}
			chls = append(chls, db.Channels...)
			db.Channels = chls
			db.GetLoginChannels()
		case OpList:
			db.printPlaylist()
		case OpSong:
			db.printSong()
		case OpExit:
			fmt.Println("Bye!")
			os.Exit(0)
		case OpHelp:
			fallthrough
		default:
			chl, err := strconv.Atoi(op)
			if err != nil {
				help()
				continue
			}
			if chl == 0 {
				db.printChannels()
				continue
			}
			if chl > 0 && chl <= len(db.Channels) {
				db.Channel = chl
			}
			db.GetSongs(doubanfm.New)
			db.playNext(db.Next())
		}
		prevOp = op
	}
}

func help() {
	s := `Command list:
	p: 	Pause or play
	n: 	Next, next song
	x:	Loop, loop playback
	s:	Skip, skip current playlist
	t: 	Trash, never play
	r: 	Like
	u:	Unlike
	c:	Current playing info
	l: 	Playlist
	0: 	Channel list
	N:	Change to Channel N, N stands for channel number, see channel list
	z:	Login, Account login
	q:	Quit
	h:	Show this help
`
	fmt.Println(s)
}
