package doubanfm

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func init() {
	rand.Seed(int64(time.Now().Nanosecond()))
}

type Song struct {
	Sid        string
	Title      string
	Artist     string
	AlbumTitle string
	Album      string
	Ext        string `json:"file_ext"`
	Ssid       string
	Sha256     string
	Status     int
	Picture    string
	Alert      string `json:"alert_msg"`
	Company    string
	RatingAvg  float64 `json:"rating_avg"`
	PubTime    string  `json:"public_time"`
	Singers    []Singer
	Like       int
	ListCount  int `json:"songlists_count"`
	Url        string
	SubType    string
	Length     int
	Aid        string
	Kbps       string
}

func (this Song) String() string {
	b := new(bytes.Buffer)
	fmt.Fprintf(b, "%7s: %s\n", "Title", this.Title)
	fmt.Fprintf(b, "%7s: %s\n", "Artist", this.Artist)
	fmt.Fprintf(b, "%7s: %s\n", "Album", this.AlbumTitle)
	album := this.Album
	if !strings.HasPrefix(album, "http") {
		album = "http://www.douban.com" + album
	}
	fmt.Fprintf(b, "%7s: %s\n", "Url", this.Url)
	fmt.Fprintf(b, "%7s: %s\n", "Company", this.Company)
	fmt.Fprintf(b, "%7s: %s\n", "Public", this.PubTime)
	fmt.Fprintf(b, "%7s: %s\n", "Kbps", this.Kbps)
	fmt.Fprintf(b, "%7s: %d\n", "Like", this.Like)

	return b.String()
}

type Singer struct {
	RelatedSite int  `json:"related_site_id"`
	SiteArtist  bool `json:"is_site_artist"`
	Id          string
	Name        string
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
func Songs(types, cid, sid string, user *User) (songs []Song, err error) {
	v := url.Values{}
	v.Add("app_name", AppName)
	v.Add("version", AppVersion)
	v.Add("type", types)
	v.Add("channel", cid)
	v.Add("sid", sid)
	v.Add("kbps", "128")
	v.Add("preventCache", strconv.FormatFloat(rand.Float64(), 'f', 16, 64))
	if user != nil {
		v.Add("user_id", user.Id)
		v.Add("token", user.Token)
		v.Add("expire", user.Expire)
	}

	resp, err := get(PeopleUrl + "?" + v.Encode())
	if err != nil {
		return
	}

	var r struct {
		Song       []Song
		Logout     int
		VMax       int `json:"version_max"`
		QuickStart int `json"is_show_quick_start"`
		dfmError
	}

	if err = decode(resp, &r); err != nil {
		return
	}

	if r.R != 0 {
		return nil, &r.dfmError
	}

	return r.Song, nil
}
