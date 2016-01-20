package doubanfm

import (
	"encoding/json"
)

type Channel struct {
	Id     json.Number `json:"channel_id,Number"`
	Name   string
	Intro  string
	NameEn string `json:"name_en"`
	AbbrEn string `json:"abbr_en"`
	Seq    int    `json:"seq_id"`
	Fav    bool   `json:"-"`
}

var heartChannel = Channel{Id: "-3", Name: "红星兆赫"}

func (this Channel) String() string {
	if this.Id == "-3" {
		return "\u2661 - " + this.Name
	}
	return string(this.Id) + " - " + this.Name
}

type MyChannel struct {
	Artists   []Artist `json:"related_artists"`
	Creator   Creator
	Intro     string
	Banner    string
	Id        json.Number `json:",Number"`
	Name      string
	Cover     string
	Start     string      `json:"song_to_start"`
	Num       json.Number `json:"song_num,Number"`
	Collected string
	HotSongs  []string `json:"hot_songs"`
}

func (this MyChannel) String() string {
	return string(this.Id) + " - " + this.Name
}

func (this MyChannel) Channel() *Channel {
	return &Channel{
		Id:   this.Id,
		Name: this.Name,
	}
}

type Artist struct {
	Id    json.Number `json:",Number"`
	Name  string
	Cover string
}

type Creator struct {
	Id   json.Number `json:",Number"`
	Name string
	Url  string
}
