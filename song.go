package doubanfm

import (
	"bytes"
	"fmt"
	"strings"
)

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
