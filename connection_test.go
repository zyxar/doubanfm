package doubanfm

import (
	"fmt"
	"testing"
)

var (
	user *User
)

func TestLogin(t *testing.T) {
	var err error
	user, err = Login("ginuerzh@gmail.com", "123456")
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%#v", user)
	for _, cookie := range cookies {
		fmt.Println(cookie.String())
	}
}

func TestChannels(t *testing.T) {
	chls, err := Channels()
	if err != nil {
		t.Error(err)
		return
	}

	for _, chl := range chls {
		fmt.Println(chl.Id, chl.Name, chl.Intro)
	}
}

func TestLoginChannels(t *testing.T) {
	id := ""
	if user != nil {
		id = user.Id
	}
	favs, recs, err := LoginChannels(id)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println("favorite")
	for _, fav := range favs {
		fmt.Println(fav.Id, fav.Name)
	}
	fmt.Println("recommend")
	for _, rec := range recs {
		fmt.Println(rec.Id, rec.Name, rec.Num)
	}
}

func TestNewChannel(t *testing.T) {
	songs, err := Songs(New, "-3", "", user)
	if err != nil {
		t.Error(err)
		return
	}

	for _, song := range songs {
		fmt.Println(song.Sid, song.Album, song.Artist, song.Title, song.Length)
	}
}
