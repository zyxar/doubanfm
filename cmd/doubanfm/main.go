package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
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

	session := doubanfm.NewSession()
	var logon = func(uid string) error {
		return session.LoginAs(uid, func() (string, error) {
			term := newTerm("Douban Id:")
			defer term.Restore()
			var err error
			if uid == "" {
				uid, err = term.ReadLine()
				if err != nil {
					return "", err
				}
			}
			uid = strings.TrimSpace(uid)
			if len(uid) == 0 {
				return "", errors.New("empty id")
			}
			passwd, err := term.ReadPassword("Password: ")
			if err != nil {
				return "", err
			}
			return passwd, nil
		})
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
			session.GetSongs(doubanfm.End)
			if session.SongNum() == 0 {
				session.GetSongs(doubanfm.Last)
			}
			playSong(session.NextSong())
		case gst.MESSAGE_ERROR:
			s, param := msg.GetStructure()
			fmt.Println("[gstreamer]", msg.GetType(), s, param)
			player.Stop()
		}
	})

	session.FetchChannels()
	if userId != "" {
		if err = logon(userId); err != nil {
			fmt.Println("\r>>>>>>>>> Access denied:", err)
			quit(1)
		}
		fmt.Println("\r>>>>>>>>> Access acquired.")
		tokenFn := filepath.Join(homeDir, userId)
		if err = session.SaveFile(tokenFn); err != nil {
			fmt.Println("\r>>>>>>>>> Token saving error:", err)
		} else {
			fmt.Println("\r>>>>>>>>> Token saved as", tokenFn)
		}
	}

	if session.RandomChannel() == nil {
		fmt.Println("\r>>>>>>>>> Error in fetching channels.")
		quit(1)
	} else {
		fmt.Println(session.Channel().String())
	}

	session.GetSongs(doubanfm.New)
	if session.SongNum() == 0 {
		fmt.Println("\r>>>>>>>>> Error in fetching songs.")
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
				session.GetSongs(doubanfm.Last)
			}
			playSong(session.NextSong())
		case OpSkip:
			session.GetSongs(doubanfm.Skip)
			playSong(session.NextSong())
		case OpTrash:
			session.GetSongs(doubanfm.Bypass)
			playSong(session.NextSong())
		case OpLike:
			session.GetSongs(doubanfm.Like)
			session.Song().Like = 1
		case OpUnlike:
			session.GetSongs(doubanfm.Unlike)
			session.Song().Like = 0
		case OpLogin:
			logon("")
			continue
		case OpList:
			for _, song := range session.Songs() {
				fmt.Printf("%s %s\n", song.Title, song.Artist)
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
				if err = session.GetSongs(doubanfm.New); err != nil {
					fmt.Println("\r>>>>>>>>>", err)
				}
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
