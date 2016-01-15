doubanfm - Douban.FM client
=====

用 Go 语言实现的 Douban.FM 命令行客户端, 基本实现了 Douban.FM 的协议(请查看 API.txt)。

本应用依赖于: go1, glib-2.0, gstreamer-1.0

Go binding for [glib](http://github.com/ziutek/glib)

Go binding for [gstreamer](http://github.com/ziutek/gst)

##命令用法
```
doubanfm> h
Command list:
	p: 	Pause or play
	n: 	Next, next song
	b:	Prev, previous song
	x:	Loop, loop playback
	s:	Skip, skip current playlist
	t: 	Trash, never play
	r: 	Like
	u:	Unlike
	c:	Current playing info
	l: 	Playlist
	0: 	Channel list
	N:	Change to Channel N, N stands for channel number, see channel list
	z:	Login, Account login
	h:	Show this help
	q:	Quit
```

##安装

`go get github.com/zyxar/doubanfm/cmd/doubanfm`

正常播放音乐需要安装 gstreamer 插件。例如在 OS X 下：`brew install gst-libav`
