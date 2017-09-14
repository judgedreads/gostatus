package main

import "time"

type timeMon struct {
	now    time.Time
	format string
	loc    *time.Location
}

func newTimeMon(format, zone string) *timeMon {
	loc, err := time.LoadLocation(zone)
	if err != nil {
		panic(err)
	}
	return &timeMon{
		format: format,
		loc:    loc,
	}
}

func (t *timeMon) Run(done chan int) {
	for {
		t.now = time.Now()
		done <- 1
		time.Sleep(time.Second)
	}
}

func (t *timeMon) String() string {
	return t.now.In(t.loc).Format(t.format)
}
