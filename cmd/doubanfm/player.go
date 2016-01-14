package main

import (
	"fmt"
	"sync"

	"github.com/ziutek/glib"
	"github.com/ziutek/gst"
)

type Player struct {
	mainloop *glib.MainLoop
	pipe     *gst.Element
	sync.Mutex
}

func newPlayer() (*Player, error) {
	pipe := gst.ElementFactoryMake("playbin", "mp3_pipe")
	if pipe == nil {
		return nil, fmt.Errorf("gstreamer error")
	}

	return &Player{
		mainloop: glib.NewMainLoop(nil),
		pipe:     pipe,
	}, nil
}

func (this *Player) init(onMessage func(*gst.Bus, *gst.Message)) {
	bus := this.pipe.GetBus()
	bus.AddSignalWatch()
	bus.Connect("message", onMessage, nil)
	go this.mainloop.Run()
}

func (this *Player) Stop() {
	this.Lock()
	this.pipe.SetState(gst.STATE_NULL)
	this.Unlock()
}

func (this *Player) Resume() {
	this.Lock()
	this.pipe.SetState(gst.STATE_PLAYING)
	this.Unlock()
}

func (this *Player) Pause() {
	this.Lock()
	this.pipe.SetState(gst.STATE_PAUSED)
	this.Unlock()
}

func (this *Player) Play(uri string) {
	this.Lock()
	this.pipe.SetState(gst.STATE_NULL)
	this.pipe.SetProperty("uri", uri)
	this.pipe.SetState(gst.STATE_PLAYING)
	this.Unlock()
}
