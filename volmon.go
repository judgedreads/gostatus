package main

import (
	"fmt"

	"github.com/judgedreads/gostatus/alsa"
)

type volMon struct {
	vol  int
	card string
	mix  string
	err  string
}

func (v *volMon) Run(done chan int) {
	mixer, err := alsa.Open(v.card)
	if err != nil {
		v.err = err.Error()
		done <- 1
		return
	}
	defer mixer.Close()
	elem, err := mixer.Elem(v.mix)
	if err != nil {
		v.err = err.Error()
		done <- 1
		return
	}
	_, max := elem.PlaybackVolumeRange()
	vol := elem.PlaybackVolume()
	v.vol = (vol * 100) / max
	ch := elem.Subscribe()
	go func() {
		for _ = range ch {
			vol := elem.PlaybackVolume()
			v.vol = (vol * 100) / max
			done <- 1
		}
	}()
	mixer.Listen()
}

func (v *volMon) String() string {
	if v.err != "" {
		return v.err
	}
	return fmt.Sprintf("v: %d%%", v.vol)
}
