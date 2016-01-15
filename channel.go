package doubanfm

import (
	"encoding/json"
	"errors"
)

type Channel struct {
	Id     json.Number `json:"channel_id,Number"`
	Name   string
	Intro  string
	NameEn string `json:"name_en"`
	AbbrEn string `json:"abbr_en"`
	Seq    int    `json:"seq_id"`
	Fav    bool
}

func (c Channel) String() string {
	return string(c.Id) + " - " + c.Name
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

func (c MyChannel) String() string {
	return string(c.Id) + " - " + c.Name
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
func Channels() (chls []Channel, err error) {
	resp, err := get(ChannelsUrl)
	if err != nil {
		return
	}

	var r struct {
		Channels []Channel
	}

	if err = decode(resp, &r); err != nil {
		return
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
func MyChannels(id string) (favs []MyChannel, recs []MyChannel, err error) {
	resp, err := get(LoginChannelsUrl + "?uk=" + id)
	if err != nil {
		return
	}
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

	if err = decode(resp, &r); err != nil {
		return
	}

	if !r.Status {
		err = errors.New(r.Msg)
		return
	}

	return r.Data.Res.Favs, r.Data.Res.Recs, nil
}
