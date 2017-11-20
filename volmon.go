package main

import (
	"fmt"

	"github.com/judgedreads/gostatus/alsa"
)

type volMon struct {
	vol   int
	muted bool
	card  string
	mix   string
	err   string
	max   int
	done  chan int
}

func (v *volMon) update(elem *alsa.Elem) {
	vol := elem.PlaybackVolume()
	v.vol = (vol * 100) / v.max
	if elem.PlaybackSwitch() == 0 {
		v.muted = true
	} else {
		v.muted = false
	}
	v.done <- 1
}

// TODO: put done channel on structs upon creation, instead of passing
// to Run?

func (v *volMon) Run(done chan int) {
	v.done = done
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
	_, v.max = elem.PlaybackVolumeRange()
	v.update(elem)
	ch := elem.Subscribe()
	go func() {
		for _ = range ch {
			v.update(elem)
		}
	}()
	mixer.Listen()
}

func (v *volMon) String() string {
	if v.err != "" {
		return v.err
	}
	if v.muted {
		return fmt.Sprintf("v: %d%% (m)", v.vol)
	}
	return fmt.Sprintf("v: %d%%", v.vol)
}
