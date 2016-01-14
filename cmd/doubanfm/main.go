package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/zyxar/doubanfm"
)

func quit() {
	fmt.Println("\rBye!")
	os.Exit(0)
}

func main() {
	term := newTerm()
	defer term.Restore()

	dfm, err := NewDoubanFM()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	dfm.GetChannels()
	dfm.Channel = 1
	dfm.GetSongs(doubanfm.New)
	dfm.playSong(dfm.Next())
	var op string
	var prevOp = OpNext
	for {
		op, err = term.ReadLine()
		if err == io.EOF {
			quit()
		} else if err != nil {
			fmt.Println(err)
			continue
		}
		op = strings.ToLower(strings.TrimSpace(op))
	PREV:
		switch op {
		case OpAgain:
			op = prevOp
			goto PREV
		case OpPlay:
			dfm.Paused = !dfm.Paused
			if dfm.Paused {
				dfm.player.Pause()
			} else {
				dfm.player.Resume()
			}
		case OpLoop:
			dfm.Loop = !dfm.Loop
		case OpNext:
			if dfm.Empty() {
				dfm.GetSongs(doubanfm.Last)
			}
			dfm.playSong(dfm.Next())
		case OpSkip:
			dfm.GetSongs(doubanfm.Skip)
			dfm.playSong(dfm.Next())
		case OpTrash:
			dfm.GetSongs(doubanfm.Bypass)
			dfm.playSong(dfm.Next())
		case OpLike:
			dfm.GetSongs(doubanfm.Like)
			dfm.Song.Like = 1
		case OpUnlike:
			dfm.GetSongs(doubanfm.Unlike)
			dfm.Song.Like = 0
		case OpLogin:
			if dfm.User != nil {
				dfm.printUser()
				continue
			}
			dfm.Login()

			if dfm.User == nil {
				fmt.Println("Login Failed")
				continue
			}
			chls := []Channel{
				{Id: "-3", Name: "红星兆赫"},
			}
			chls = append(chls, dfm.Channels...)
			dfm.Channels = chls
			dfm.GetLoginChannels()
			continue
		case OpList:
			dfm.printPlaylist()
		case OpSong:
			dfm.printSong()
		case OpExit:
			quit()
		case OpHelp:
			fallthrough
		default:
			chl, err := strconv.Atoi(op)
			if err != nil {
				help()
				continue
			}
			if chl == 0 {
				dfm.printChannels()
				continue
			}
			if chl > 0 && chl <= len(dfm.Channels) {
				dfm.Channel = chl
			}
			dfm.GetSongs(doubanfm.New)
			dfm.playSong(dfm.Next())
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
