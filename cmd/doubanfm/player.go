package main

import (
	"fmt"

	"github.com/ziutek/glib"
	"github.com/ziutek/gst"
)

type Player struct {
	mainloop *glib.MainLoop
	pipe     *gst.Element
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
	this.pipe.SetState(gst.STATE_NULL)
}

func (this *Player) Play() {
	this.pipe.SetState(gst.STATE_PLAYING)
}

func (this *Player) Pause() {
	this.pipe.SetState(gst.STATE_PAUSED)
}

func (this *Player) NewSource(uri string) {
	this.Stop()
	this.pipe.SetProperty("uri", uri)
	this.Play()
}
