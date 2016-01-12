package doubanfm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	// ip:host
	proxyUrl string
	cookies  = make(map[string]*http.Cookie) // cookies
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
	proxyUrl = os.Getenv("http_proxy")
	if strings.HasPrefix(proxyUrl, "http://") {
		proxyUrl = strings.TrimPrefix(proxyUrl, "http://")
	}
	if strings.HasSuffix(proxyUrl, "/") {
		proxyUrl = strings.TrimSuffix(proxyUrl, "/")
	}
}

// {"err":"wrong_version", "r":1}
type dbError struct {
	R   int
	Err string
}

func (e dbError) Error() string {
	return strconv.Itoa(e.R) + ": " + e.Err
}

func get(url string) (io.Reader, error) {
	//fmt.Println(url)
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

	if proxyUrl == "" {
		return http.DefaultClient.Do(r)
	}

	addr, err := net.ResolveTCPAddr("tcp", proxyUrl)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := r.WriteProxy(conn); err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(conn)
	if err != nil {
		return nil, err
	}
	return http.ReadResponse(bufio.NewReader(bytes.NewBuffer(data)), r)
}

func decode(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}
