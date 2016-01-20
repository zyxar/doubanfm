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

	dfm, err := NewDoubanFM()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		quit(1)
	}

	var logon = func(uid string) {
		if dfm.User != nil {
			fmt.Println(dfm.User)
			return
		}
		fn := filepath.Join(homeDir, uid)
		cookieFn := filepath.Join(homeDir, "cookies.json")
		dfm.User = &doubanfm.User{}
		err := dfm.User.LoadFile(fn)
		if err == nil {
			if err = doubanfm.ReadCookieFile(cookieFn); err == nil {
				fmt.Println("\r>>>>>>>>> Token loaded.")
				// TODO: check session status.
				return
			}
			fmt.Println("\r>>>>>>>>> Token loading error:", err)
		} else {
			fmt.Println("\r>>>>>>>>> Token loading error:", err)
		}

		err = dfm.Login(uid)
		if err != nil {
			fmt.Println("\r>>>>>>>>> Access denied:", err)
			return
		}
		fmt.Println("\r>>>>>>>>> Access acquired.")
		fmt.Println(dfm.User)
		if err = dfm.User.SaveFile(fn); err != nil {
			fmt.Println("\r>>>>>>>>> Token saving error:", err)
		} else {
			if err = doubanfm.WriteCookieFile(cookieFn); err == nil {
				fmt.Println("\r>>>>>>>>> Token saved.")
			} else {
				fmt.Println("\r>>>>>>>>> Token saving error:", err)
			}
		}
		dfm.GetMyChannels()
		return
	}

	dfm.GetChannels()
	if userId != "" {
		logon(userId)
	}
	dfm.SetDefaultChannel()
	if dfm.Channel == nil {
		fmt.Println("\r>>>>>>>>> Error in fetching channels.")
		quit(1)
	}
	dfm.printChannel()
	dfm.GetSongs(doubanfm.New)
	if dfm.Empty() {
		fmt.Println("\r>>>>>>>>> Error in fetching songs.")
		quit(1)
	}
	dfm.playSong(dfm.Next())

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
			logon("")
			continue
		case OpList:
			dfm.printPlaylist()
		case OpSong:
			dfm.printSong()
		case OpExit:
			quit(0)
		case OpChann:
			dfm.printChannels()
		case OpHelp:
			help()
		default:
			if ch, err := strconv.Atoi(op); err == nil &&
				ch > 0 &&
				ch <= len(dfm.channlist) {
				dfm.Channel = dfm.Channels[dfm.channlist[ch-1]]
				dfm.printChannel()
				dfm.GetSongs(doubanfm.New)
				dfm.playSong(dfm.Next())
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
