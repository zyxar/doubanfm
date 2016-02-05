package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

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
	OpId     = "i"
	OpHelp   = "h"
	OpExit   = "q"
	OpChann  = "0"
	PROMPT   = "DoubanFM> "
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
             i: Current Id
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

	player, err := newPlayer()
	if err != nil {
		fmt.Println("\r>>>>>>>>> Player error:", err)
		os.Exit(1)
	}

	term := newTerm(PROMPT)
	var quit = func(code int) {
		term.Restore()
		fmt.Println("\r>>>>>>>>> Bye!")
		os.Exit(code)
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1)
		for s := range c {
			fmt.Printf("\r>>>>>>>>> %v caught, exit\n", s)
			quit(1)
			break
		}
	}()

	session := doubanfm.NewSession()
	var logon = func(uid string) {
		term := newTerm("Douban Id:")
		defer term.Restore()
		var err error
		if uid == "" {
			uid, err = term.ReadLine()
			if err != nil {
				fmt.Println("\r>>>>>>>>> Access denied:", err)
				return
			}
		}
		uid = strings.TrimSpace(uid)
		if len(uid) == 0 {
			fmt.Println("\r>>>>>>>>> Access denied: empty id")
			return
		}

		tokenFn := filepath.Join(homeDir, uid)
		if err = session.LoadFile(tokenFn); err == nil {
			fmt.Println("\r>>>>>>>>> Token loaded from", tokenFn)
			fmt.Println("\r>>>>>>>>> Access acquired:")
			fmt.Println(session.Id())
			return
		}

		if err = session.LoginAs(uid, func() (string, error) {
			return term.ReadPassword("Password: ")
		}); err != nil {
			fmt.Println("\r>>>>>>>>> Access denied:", err)
		} else {
			fmt.Println("\r>>>>>>>>> Access acquired:")
			fmt.Println(session.Id())
			if err = session.SaveFile(tokenFn); err != nil {
				fmt.Println("\r>>>>>>>>> Token saving error:", err)
			} else {
				fmt.Println("\r>>>>>>>>> Token saved as", tokenFn)
			}
		}
	}

	var playSong = func(song *doubanfm.Song) {
		if song != nil {
			fmt.Printf("\rPLAYING>> %s - %s\n", song.Title, song.Artist)
			player.Play(song.Url)
		}
	}

	player.init(func(bus *gst.Bus, msg *gst.Message) {
		switch msg.GetType() {
		case gst.MESSAGE_EOS:
			if session.Loop {
				playSong(session.NextSong())
				return
			}
			session.FetchSongs(doubanfm.End)
			if session.SongNum() == 0 {
				session.FetchSongs(doubanfm.Last)
			}
			playSong(session.NextSong())
		case gst.MESSAGE_ERROR:
			s, param := msg.GetStructure()
			fmt.Println("[gstreamer]", msg.GetType(), s, param)
			player.Stop()
		}
	})

	if err = session.FetchChannels(); err != nil {
		fmt.Println("\r>>>>>>>>> Error in fetching channels:", err)
	}
	if userId != "" {
		logon(userId)
		if err = session.FetchMyChannels(); err != nil {
			fmt.Println("\r>>>>>>>>> Error in fetching channels:", err)
		}
	}

	if session.RandomChannel() == nil {
		fmt.Println("\r>>>>>>>>> No channel in list.")
		quit(1)
	} else {
		fmt.Println(session.Channel().String())
	}

	if err = session.FetchSongs(doubanfm.New); err != nil {
		fmt.Println("\r>>>>>>>>> Error in fetching songs:", err)
		quit(1)
	}
	playSong(session.NextSong())

	var op string
	var prevOp = OpNext
	var paused bool = false
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
			paused = !paused
			if paused {
				player.Pause()
			} else {
				player.Resume()
			}
		case OpLoop:
			session.Loop = !session.Loop
		case OpNext:
			if session.SongNum() == 0 {
				session.FetchSongs(doubanfm.Last)
			}
			playSong(session.NextSong())
		case OpSkip:
			session.FetchSongs(doubanfm.Skip)
			playSong(session.NextSong())
		case OpTrash:
			session.FetchSongs(doubanfm.Bypass)
			playSong(session.NextSong())
		case OpLike:
			session.FetchSongs(doubanfm.Like)
			session.Song().Like = 1
		case OpUnlike:
			session.FetchSongs(doubanfm.Unlike)
			session.Song().Like = 0
		case OpLogin:
			logon("")
			continue
		case OpId:
			fmt.Println(session.Id())
		case OpList:
			for _, song := range session.Songs() {
				fmt.Printf("%s - %s\n", song.Title, song.Artist)
			}
		case OpSong:
			fmt.Println(session.Song())
		case OpExit:
			quit(0)
		case OpChann:
			session.PrintChannels()
		case OpHelp:
			help()
		default:
			if ch, err := strconv.Atoi(op); err == nil {
				err = session.SetChannel(ch)
				if err != nil {
					fmt.Println("\r>>>>>>>>>", err)
					prevOp = OpHelp
					continue
				}
				fmt.Println(session.Channel().String())
				playSong(session.NextSong())
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
