package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/zyxar/doubanfm"
)

var (
	userId  string
	helpStr string
)

func init() {
	flag.StringVar(&userId, "login", "", "login id")
	helpStr = `Operation list:
             p: Pause or play
             n: Next, next song
             x: Loop, loop playback
             s: Skip, skip current playlist
             t: Trash, never play
             r: Like
             u: Unlike
             c: Current playing info
             l: Playlist
             0: Channel list
             N: Change to Channel N, N stands for channel number, see channel list
             z: Login, Account login
             q: Quit
             h: Show this help
`
}

func main() {
	flag.Parse()
	var err error
	if err = mkHomeDir(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	term := newTerm(PROMPT)
	var quit = func(code int) {
		term.Restore()
		fmt.Println("\r>>>>>>>>> Bye!")
		os.Exit(code)
	}

	session, err := NewSession()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		quit(1)
	}

	var logon = func(uid string) {
		if session.User != nil {
			fmt.Println(session.User)
			return
		}
		fn := filepath.Join(homeDir, uid)
		err := session.User.LoadFile(fn)
		if err != nil {
			fmt.Println("\r>>>>>>>>> Token loading error:", err)
		} else {
			fmt.Println("\r>>>>>>>>> Token loaded.")
		}

		err = session.Login(uid)
		if err != nil {
			fmt.Println("\r>>>>>>>>> Access denied:", err)
			return
		}
		fmt.Println("\r>>>>>>>>> Access acquired.")
		fmt.Println(session.User)
		if err = session.User.SaveFile(fn); err != nil {
			fmt.Println("\r>>>>>>>>> Token saving error:", err)
		} else {
			fmt.Println("\r>>>>>>>>> Token saved.")
		}
		session.GetMyChannels()
		return
	}

	session.GetChannels()
	if userId != "" {
		logon(userId)
	}
	session.SetDefaultChannel()
	if session.Channel == nil {
		fmt.Println("\r>>>>>>>>> Error in fetching channels.")
		quit(1)
	}
	session.printChannel()
	session.GetSongs(doubanfm.New)
	if session.Empty() {
		fmt.Println("\r>>>>>>>>> Error in fetching songs.")
		quit(1)
	}
	session.playSong(session.Next())

	var op string
	var prevOp = OpNext
	for {
		op, err = term.ReadLine()
		if err == io.EOF {
			fmt.Println()
			quit(0)
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
			session.Paused = !session.Paused
			if session.Paused {
				session.player.Pause()
			} else {
				session.player.Resume()
			}
		case OpLoop:
			session.Loop = !session.Loop
		case OpNext:
			if session.Empty() {
				session.GetSongs(doubanfm.Last)
			}
			session.playSong(session.Next())
		case OpSkip:
			session.GetSongs(doubanfm.Skip)
			session.playSong(session.Next())
		case OpTrash:
			session.GetSongs(doubanfm.Bypass)
			session.playSong(session.Next())
		case OpLike:
			session.GetSongs(doubanfm.Like)
			session.Song.Like = 1
		case OpUnlike:
			session.GetSongs(doubanfm.Unlike)
			session.Song.Like = 0
		case OpLogin:
			logon("")
			continue
		case OpList:
			session.printPlaylist()
		case OpSong:
			session.printSong()
		case OpExit:
			quit(0)
		case OpChann:
			session.printChannels()
		case OpHelp:
			help()
		default:
			if ch, err := strconv.Atoi(op); err == nil &&
				ch > 0 &&
				ch <= len(session.channlist) {
				session.Channel = session.Channels[session.channlist[ch-1]]
				session.printChannel()
				session.GetSongs(doubanfm.New)
				session.playSong(session.Next())
			} else {
				fmt.Println("\r>>>>>>>>> No such operation:", op)
				help()
				prevOp = OpHelp
				continue
			}
		}
		prevOp = op
	}
	term.Restore()
}

func help() {
	fmt.Println(helpStr)
}
