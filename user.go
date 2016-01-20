package doubanfm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

func init() {
	rand.Seed(int64(time.Now().Nanosecond()))
}

type User struct {
	Id     string `json:"user_id"`
	Name   string `json:"user_name"`
	Email  string
	Token  string `json:",omitempty"`
	Expire string `json:",omitempty"`
}

type Identity struct {
	User
	Cookies   map[string]*http.Cookie
	anonymous bool `json:"-"`
}

var Anonymous = Identity{
	User: User{Id: "-",
		Name: "anonymous",
	},
	Cookies:   make(map[string]*http.Cookie),
	anonymous: true,
}

func NewIdentity(uid string) *Identity {
	return &Identity{
		User:      User{Id: uid},
		Cookies:   make(map[string]*http.Cookie),
		anonymous: false,
	}
}

func (this Identity) Json() string {
	v, err := json.MarshalIndent(this, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(v)
}

func (this Identity) String() string {
	return fmt.Sprintf("\r    Id:\t%s\n  Name:\t%s\n Token:\t%s\nExpire:\t%s",
		this.Id, this.Name, this.Token, parseTime(this.Expire))
}

func (this Identity) Save(w io.Writer) error {
	return json.NewEncoder(w).Encode(this)
}

func (this Identity) SaveFile(fn string) error {
	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	return this.Save(f)
}

func (this *Identity) Load(r io.Reader) error {
	return json.NewDecoder(r).Decode(this)
}

func (this *Identity) LoadFile(fn string) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	return this.Load(f)
}

func (this *Identity) get(url string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for _, cookie := range this.Cookies {
		req.AddCookie(cookie)
	}
	r, err := routine(req)
	if err != nil {
		return nil, err
	}
	for _, cookie := range r.Cookies() {
		this.Cookies[cookie.Name] = cookie
	}
	return r.Body, err
}

func (this *Identity) Login(password string) error {
	if this.anonymous {
		return errors.New("anonymous user")
	}
	formdata := &bytes.Buffer{}
	w := multipart.NewWriter(formdata)
	w.WriteField("app_name", AppName)
	w.WriteField("version", AppVersion)
	w.WriteField("email", this.Id)
	w.WriteField("password", password)
	w.Close()
	req, err := http.NewRequest("POST", LoginUrl, formdata)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	for _, cookie := range this.Cookies {
		req.AddCookie(cookie)
	}
	resp, err := routine(req)
	if err != nil {
		return err
	}
	for _, cookie := range resp.Cookies() {
		this.Cookies[cookie.Name] = cookie
	}
	defer resp.Body.Close()
	var r struct {
		User
		dfmError
	}
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}

	if r.R != 0 {
		return &r.dfmError
	}
	this.Name = r.Name
	this.Email = r.Email
	this.Token = r.Token
	this.Expire = r.Expire
	return nil
}

func parseTime(ut string) string {
	if sec, err := strconv.ParseInt(ut, 10, 64); err == nil {
		t := time.Unix(sec, 0)
		return t.String()
	}
	return ut
}

// GET http://www.douban.com/j/app/radio/channels
//	{
//	  "channels":[
//	    {
//	      "name_en": "Personal Radio",
//	      "seq_id": 0,
//	      "abbr_en": "",
//	      "name": "私人兆赫",
//	      "channel_id": "1"
//	    }
//	  ]
//	}
func (this *Identity) GetChannels() ([]Channel, error) {
	resp, err := this.get(ChannelsUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Close()
	var r struct {
		Channels []Channel
	}
	if err = json.NewDecoder(resp).Decode(&r); err != nil {
		return nil, err
	}
	return r.Channels, nil
}

// GET http://douban.fm/j/explore/get_login_chls
//	{
//	  "status":true,
//	  "data":{
//	    "res":{
//		  "fav_chls":[
//		    {
//			  "related_artists":[
//				{
//				  "cover":"http://img3.douban.com/img/fmadmin/large/31550.jpg",
//				  "id":"18250",
//				  "name":"李健"
//			    },
//			    ...
//		      ],
//		      "creator":{
//			    "url":"http://site.douban.com/douban.fm/",
//			    "name":"豆瓣FM",
//			    "id":1
//		      },
//		      "intro":"为你推荐 李健 以及相似的艺术家",
//		      "banner":"http://img3.douban.com/img/fmadmin/chlBanner/31550.jpg",
//		      "id":28250,
//		      "name":"李健 系",
//		      "cover":"http://img3.douban.com/img/fmadmin/small/31550.jpg",
//		      "song_to_start":"",
//		      "song_num":0,
//		      "collected":"false",
//		      "hot_songs":["往日时光","花房姑娘","如果有来生"]
//		    }
//	      ],
//	      "rec_chls":[]
//	    }
//	  }
//	}
func (this *Identity) GetMyChannels(id string) ([]MyChannel, []MyChannel, error) {
	resp, err := this.get(LoginChannelsUrl + "?uk=" + id)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Close()
	var r struct {
		Data struct {
			Res struct {
				Favs []MyChannel `json:"fav_chls"`
				Recs []MyChannel `json:"rec_chls"`
			}
		}
		Status bool
		Msg    string
	}

	if err = json.NewDecoder(resp).Decode(&r); err != nil {
		return nil, nil, err
	}

	if !r.Status {
		return nil, nil, errors.New(r.Msg)
	}

	return r.Data.Res.Favs, r.Data.Res.Recs, nil
}

// GET http://www.douban.com/j/app/radio/people?app_name=radio_desktop_win&version=100&type=n&channel=1
//	{
//	  "logout":1,
//	  "r":0,
//	  "version_max":100,
//	  "is_show_quick_start":0,
//	  "song":[
//	    {
//	      "albumtitle":"Album Title",
//	      "file_ext":"mp3",
//	      "album":"\/subject\/1407573\/",
//	      "ssid":"b037",
//	      "title":"Title",
//	      "sid":"630223",
//	      "sha256":"53b19c0854e9016b39c6144ee6c0da08d5b01393b5899544080cc3d368939245",
//	      "status":0,
//	      "picture":"http:\/\/img4.douban.com\/lpic\/s1734968.jpg",
//	      "alert_msg":"",
//	      "company":"New World Music",
//	      "rating_avg":4.30244,
//	      "public_time":"1992",
//	      "singers":[
//	        {
//	          "related_site_id":0,
//	          "is_site_artist":false,
//	          "id":"4204",
//	          "name":"Kitaro"
//	        }
//	      ],
//	      "like":0,
//	      "songlists_count":11,
//	      "artist":"Kitaro",
//	      "url":"http:\/\/mr3.douban.com\/201508251840\/fd0ebea86b6626f0caa00faff3d28eb2\/view\/song\/small\/p630223_128k.mp3",
//	      "subtype":"",
//	      "length":382,
//	      "aid":"1407573",
//	      "kbps":"128"
//	    },
//	  ]
//	}
// {"warning":"user_is_ananymous","r":0,"version_max":637,"is_show_quick_start":0,"song":[]}
func (this *Identity) GetSongs(types, cid, sid string) (songs []Song, err error) {
	v := url.Values{}
	v.Add("app_name", AppName)
	v.Add("version", AppVersion)
	v.Add("type", types)
	v.Add("channel", cid)
	v.Add("sid", sid)
	v.Add("kbps", "128")
	v.Add("preventCache", strconv.FormatFloat(rand.Float64(), 'f', 16, 64))
	if this != nil && !this.anonymous {
		v.Add("user_id", this.Id)
		v.Add("token", this.Token)
		v.Add("expire", this.Expire)
	}

	resp, err := this.get(PeopleUrl + "?" + v.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	var r struct {
		Song       []Song
		Logout     int
		VMax       int `json:"version_max"`
		QuickStart int `json"is_show_quick_start"`
		dfmError
	}

	if err = json.NewDecoder(resp).Decode(&r); err != nil {
		return nil, err
	}

	if r.R != 0 {
		return nil, &r.dfmError
	}

	if len(r.Song) == 0 && r.dfmError.Error() != "" {
		return r.Song, &r.dfmError
	}

	return r.Song, nil
}
