doubanfm - douban.fm client by Go
=====

用 Google Go 语言实现的 douban.fm 命令行客户端, 基本实现了 douban.fm 的协议(请查看 API.txt)。

本应用依赖于: go1, glib-2.0, gstreamer-1.0

Go binding for [glib](http://github.com/ziutek/glib)

Go binding for [gstreamer](http://github.com/ziutek/gst)

####命令用法：
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
