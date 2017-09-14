package main

import (
	"fmt"
	"io/ioutil"
	"time"
)

const bolt = "\u26A1"

type batMon struct {
	batPerc []byte
}

func (b *batMon) Run(done chan int) {
	for {
		perc, err := ioutil.ReadFile("/sys/class/power_supply/BAT0/capacity")
		if err != nil {
			panic("Cannot read battery status.")
		}
		b.batPerc = perc[:len(perc)-1]
		done <- 1
		time.Sleep(time.Minute)
	}
}

func (b *batMon) String() string {
	return fmt.Sprintf("%s%%", b.batPerc)
}
