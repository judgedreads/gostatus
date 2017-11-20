package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"
)

const (
	batfile = "/sys/class/power_supply/BAT0/capacity"
	bolt    = "\u26A1"
)

type batMon struct {
	batPerc []byte
	emsg    string
}

func (b *batMon) Run(done chan int) {
	for {
		perc, err := ioutil.ReadFile(batfile)
		if err != nil {
			log.Println("Cannot read battery status.")
			b.emsg = "N/A"
			return
		}
		b.batPerc = perc[:len(perc)-1]
		done <- 1
		time.Sleep(time.Minute)
	}
}

func (b *batMon) String() string {
	if b.emsg != "" {
		return b.emsg
	}
	return fmt.Sprintf("%s%%", b.batPerc)
}
