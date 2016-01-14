package doubanfm

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
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
		sync.Mutex
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
	defaultConn.Mutex = sync.Mutex{}
}

// {"err":"wrong_version", "r":1}
type dfmError struct {
	R   int
	Err string
}

func (e dfmError) Error() string {
	return strconv.Itoa(e.R) + ": " + e.Err
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
	defaultConn.Lock()
	timer := time.AfterFunc(defaultConn.timeout, func() {
		defaultConn.Client.Transport.(*http.Transport).CancelRequest(req)
		timeout = true
	})
	resp, err := defaultConn.Do(req)
	if timer != nil {
		timer.Stop()
	}
	defaultConn.Unlock()
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

func decode(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}
