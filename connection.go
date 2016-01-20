package doubanfm

import (
	"errors"
	"io"
	"net"
	"net/http"
	"time"
)

const (
	AppName    = "radio_android"
	AppVersion = "100"
)

const (
	LoginUrl         = "http://www.douban.com/j/app/login"
	PeopleUrl        = "http://www.douban.com/j/app/radio/people"
	ChannelsUrl      = "http://www.douban.com/j/app/radio/channels"
	LoginChannelsUrl = "http://douban.fm/j/explore/get_login_chls"
)

var (
	defaultConn struct {
		*http.Client
		timeout time.Duration
	}
)

const (
	Bypass = "b" // bypass current song (no longer play), refresh playlist
	End    = "e" // end of current song
	New    = "n" // new channel, refresh playlist
	Last   = "p" // last song, refresh playlist
	Skip   = "s" // skip current song, refresh playlist
	Unlike = "u" // unlike current song, refresh playlist
	Like   = "r" // like current song, refresh playlist
)

func init() {
	defaultConn.timeout = 5000 * time.Millisecond
	defaultConn.Client = &http.Client{
		Transport: &http.Transport{
			Dial:  (&net.Dialer{Timeout: defaultConn.timeout}).Dial,
			Proxy: http.ProxyFromEnvironment,
		},
	}
}

func routine(req *http.Request) (*http.Response, error) {
	timeout := false
retry:
	timer := time.AfterFunc(defaultConn.timeout, func() {
		defaultConn.Client.Transport.(*http.Transport).CancelRequest(req)
		timeout = true
	})
	resp, err := defaultConn.Do(req)
	if timer != nil {
		timer.Stop()
	}
	if err == io.EOF && !timeout {
		goto retry
	}
	if timeout {
		err = errors.New("request time out.")
	}
	return resp, err
}
