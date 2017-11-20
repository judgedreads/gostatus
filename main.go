package main

import (
	"strings"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xprop"
)

const (
	separator = " | "
	utcfmt    = "15:04Z"
	localfmt  = "Mon 2 Jan 15:04:05"
)

type plugin interface {
	Run(chan int)
	String() string
}

var plugins = []plugin{
	&volMon{card: "default", mix: "Master"},
	&netMon{},
	&batMon{},
	newTimeMon(localfmt, "Local"),
	newTimeMon(utcfmt, "UTC"),
}

// TODO: if possible, write errors to status area, rather than panic
func main() {
	conn, err := xgbutil.NewConn()
	if err != nil {
		panic("Cannot open display.")
	}
	done := make(chan int)
	for _, p := range plugins {
		go p.Run(done)
	}
	for _ = range done {
		var vals []string
		for _, p := range plugins {
			vals = append(vals, p.String())
		}
		out := strings.Join(vals, separator)
		err := xprop.ChangeProp(conn, conn.RootWin(), 8, "WM_NAME", "STRING", []byte(out))
		if err != nil {
			panic("Cannot set status.")
		}
	}
}
