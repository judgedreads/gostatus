package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xprop"
)

var (
	done      = make(chan int)
	batPerc   []byte
	localTime string
	utcTime   string
)

func bat() {
	for {
		perc, err := ioutil.ReadFile("/sys/class/power_supply/BAT0/capacity")
		if err != nil {
			panic("Cannot read battery status.")
		}
		batPerc = perc[:len(perc)-1]
		done <- 1
		time.Sleep(time.Minute)
	}
}

func clock() {
	for {
		t := time.Now()
		localTime = t.Format("Mon 2 Jan 15:04:05")
		utcTime = t.UTC().Format("15:04")
		done <- 1
		time.Sleep(time.Second)
	}
}

func main() {
	conn, err := xgbutil.NewConn()
	if err != nil {
		panic("Cannot open display.")
	}
	go bat()
	go clock()
	for {
		_ = <-done
		out := fmt.Sprintf(" \u26A1%s%% | %s (%s UTC)", batPerc, localTime, utcTime)
		err := xprop.ChangeProp(conn, conn.RootWin(), 8, "WM_NAME", "STRING", []byte(out))
		if err != nil {
			panic("Cannot set status.")
		}
	}
}
