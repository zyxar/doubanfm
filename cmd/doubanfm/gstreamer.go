package main

import (
	"fmt"

	"github.com/ziutek/glib"
	"github.com/ziutek/gst"
)

type gstreamer struct {
	mainloop *glib.MainLoop
	pipe     *gst.Element
}

func newGstreamer() (*gstreamer, error) {
	pipe := gst.ElementFactoryMake("playbin", "mp3_pipe")
	if pipe == nil {
		return nil, fmt.Errorf("gstreamer error")
	}

	return &gstreamer{
		mainloop: glib.NewMainLoop(nil),
		pipe:     pipe,
	}, nil
}

func (g *gstreamer) init(onMessage func(*gst.Bus, *gst.Message)) {
	bus := g.pipe.GetBus()
	bus.AddSignalWatch()
	bus.Connect("message", onMessage, nil)
	go g.mainloop.Run()
}

func (g *gstreamer) Stop() {
	g.pipe.SetState(gst.STATE_NULL)
}

func (g *gstreamer) Play() {
	g.pipe.SetState(gst.STATE_PLAYING)
}

func (g *gstreamer) Pause() {
	g.pipe.SetState(gst.STATE_PAUSED)
}

func (g *gstreamer) NewSource(uri string) {
	g.Stop()
	g.pipe.SetProperty("uri", uri)
	g.Play()
}
