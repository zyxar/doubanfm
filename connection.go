package doubanfm

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
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
	cookies     = make(map[string]*http.Cookie) // cookies
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

func WriteCookieFile(fn string) (err error) {
	w, err := os.Create(fn)
	if err != nil {
		return
	}
	if err = json.NewEncoder(w).Encode(cookies); err != nil {
		return
	}
	err = w.Close()
	return
}

func ReadCookieFile(fn string) (err error) {
	r, err := os.Open(fn)
	if err != nil {
		return
	}
	if err = json.NewDecoder(r).Decode(&cookies); err != nil {
		return
	}
	err = r.Close()
	return
}

func get(url string) (io.Reader, error) {
	r, err := request("GET", url, "", nil)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	saveCookies(r.Cookies())

	b := &bytes.Buffer{}
	_, err = io.Copy(b, r.Body)

	return b, err
}

func post(url, bodyType string, body io.Reader) (io.Reader, error) {
	r, err := request("POST", url, bodyType, body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	saveCookies(r.Cookies())

	b := &bytes.Buffer{}
	_, err = io.Copy(b, r.Body)

	return b, err
}

func saveCookies(cks []*http.Cookie) {
	for _, ck := range cks {
		cookies[ck.Name] = ck
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
		err = errors.New("Request time out.")
	}
	return resp, err
}

func request(method, url, bodyType string, body io.Reader) (*http.Response, error) {
	r, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if bodyType != "" {
		r.Header.Set("Content-Type", bodyType)
	}
	for _, cookie := range cookies {
		r.AddCookie(cookie)
	}
	return routine(r)
}
