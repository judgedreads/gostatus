package main

import (
	"bufio"
	"os/exec"
)

type volMon struct {
	vol string
}

func (v *volMon) Run(done chan int) {
	cmd := exec.Command("volmon", "default", "Master")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	defer cmd.Wait()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		v.vol = scanner.Text()
		done <- 1
	}
}

func (v *volMon) String() string {
	return v.vol
}
