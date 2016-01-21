package doubanfm

import (
	"errors"
	"fmt"
)

type Session struct {
	id       *Identity
	channels map[string]*Channel
	chanIds  []string
	songs    []Song
	channel  *Channel
	song     Song
	Loop     bool
}

func NewSession() *Session {
	session := &Session{
		channels: make(map[string]*Channel),
		chanIds:  make([]string, 1, 100),
		id:       &AnonymousId,
	}
	session.chanIds[0] = "-3"
	session.channels["-3"] = &heartChannel
	return session
}

func (this Session) Id() string {
	return this.id.String()
}

func (this Session) SongNum() int {
	return len(this.songs)
}

func (this *Session) NextSong() *Song {
	if len(this.songs) == 0 {
		return nil
	}
	if !this.Loop {
		this.song = this.songs[0]
		this.songs = this.songs[1:]
	}
	return &this.song
}

func (this *Session) FetchMyChannels() error {
	if this.id.anonymous {
		return errors.New("anonymous user")
	}
	favs, recs, err := this.id.GetMyChannels()
	if err != nil {
		return err
	}
	for _, fav := range favs {
		if _, ok := this.channels[fav.Id.String()]; !ok {
			this.channels[fav.Id.String()] = fav.Channel()
			this.chanIds = append(this.chanIds, fav.Id.String())
		} else {
			this.channels[fav.Id.String()].Fav = true
		}
	}
	for _, rec := range recs {
		if _, ok := this.channels[rec.Id.String()]; !ok {
			this.channels[rec.Id.String()] = rec.Channel()
			this.chanIds = append(this.chanIds, rec.Id.String())
		}
	}
	return nil
}

func (this *Session) FetchChannels() error {
	if channels, err := this.id.GetChannels(); err != nil {
		return err
	} else {
		for i, _ := range channels {
			if _, ok := this.channels[channels[i].Id.String()]; !ok {
				this.channels[channels[i].Id.String()] = &channels[i]
				this.chanIds = append(this.chanIds, channels[i].Id.String())
			}
		}
	}
	return nil
}

func (this *Session) FetchSongs(types string) error {
	if this.channel == nil {
		return errors.New("nil channel")
	}
	songs, err := this.id.GetSongs(types, this.channel.Id.String(), this.song.Sid)
	if err != nil {
		return err
	}
	if len(songs) == 0 {
		return errors.New("empty song list")
	}
	this.songs = songs
	return nil
}

type PasswordReadFunc func() (string, error)

func (this *Session) LoginAs(uid string, readPasswd PasswordReadFunc) error {
	passwd, err := readPasswd()
	if err != nil {
		return err
	}
	id := NewIdentity(uid)
	if err = id.Login(passwd); err != nil {
		return err
	}
	this.id = id
	return nil
}

func (this *Session) SetChannel(i int) error {
	if i > 0 && i <= len(this.chanIds) {
		prevChannel := this.channel
		this.channel = this.channels[this.chanIds[i-1]]
		if err := this.FetchSongs(New); err != nil {
			this.channel = prevChannel
			return err
		}
		return nil
	}
	return errors.New("no such channel")
}

func (this *Session) RandomChannel() *Channel {
	for _, channel := range this.channels {
		if channel.Id == "-3" && this.id.anonymous {
			continue
		}
		this.channel = channel
		break
	}
	return this.channel
}

func (this *Session) LoadFile(fn string) error {
	if id, err := NewIdentityFromFile(fn); err != nil {
		return err
	} else {
		this.id = id
	}
	return nil
}

func (this Session) SaveFile(fn string) error {
	return this.id.SaveFile(fn)
}

func (this *Session) Channel() *Channel {
	return this.channel
}

func (this *Session) Song() *Song {
	return &this.song
}

func (this *Session) Songs() []Song {
	s := make([]Song, 0, len(this.songs)+1)
	s = append(s, this.song)
	return append(s, this.songs...)
}

func (this *Session) PrintChannels() {
	for j, id := range this.chanIds {
		cur := "-"
		fav := ""
		if id == this.channel.Id.String() {
			cur = "+"
		}
		if this.channels[id].Fav {
			fav = "*"
		}
		fmt.Printf("%2d %s [%s]\r\t\t%s %s\n", j+1, cur, this.channels[id].Id, this.channels[id].Name, fav)
	}
}
