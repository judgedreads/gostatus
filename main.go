package main

// #cgo LDFLAGS: -lX11
// #include <X11/Xlib.h>
import "C"

import (
	"fmt"
	"time"
	"io/ioutil"
)

var dpy = C.XOpenDisplay(nil)
var done = make(chan int)
var batPerc []byte
var localTime string
var utcTime string

func bat() {
	for {
		perc, err := ioutil.ReadFile("/sys/class/power_supply/BAT0/capacity")
		if err != nil {
			panic("Cannot read battery status.")
		}
		batPerc = perc[:len(perc)-1]
		done <- 1
		fmt.Printf("bat update")
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
	if dpy == nil {
		panic("Cannot open display.")
	}
	go bat()
	go clock()
	for {
		_ = <-done
		out := fmt.Sprintf(" \u26A1%s%% | %s (%s UTC)", batPerc, localTime, utcTime)
		C.XStoreName(dpy, C.XDefaultRootWindow(dpy), C.CString(out))
		C.XSync(dpy, 1)
	}
}
