package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"time"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/godbus/dbus"
)

var (
	done      = make(chan int)
	batPerc   []byte
	localTime string
	utcTime   string
	netDevs   string
	volume    string
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

func activeNetDevices(conn *dbus.Conn, obj dbus.BusObject) error {
	var devs []string
	resp, err := obj.GetProperty("org.freedesktop.NetworkManager.ActiveConnections")
	if err != nil {
		return err
	}
	s := resp.String()
	list := strings.Split(s[5:len(s)-1], ",")
	for _, path := range list {
		if path == "" {
			// some Split weirdness?
			continue
		}
		path = strings.Trim(path, "\" ")
		obj := conn.Object("org.freedesktop.NetworkManager", dbus.ObjectPath(path))
		resp, err := obj.GetProperty("org.freedesktop.NetworkManager.Connection.Active.Devices")
		if err != nil {
			// when a device is disactivated, it remains in the list until
			// a new one is activated
			continue
		}
		s = resp.String()
		obj = conn.Object("org.freedesktop.NetworkManager", dbus.ObjectPath(s[6:len(s)-2]))
		resp, err = obj.GetProperty("org.freedesktop.NetworkManager.Device.Interface")
		if err != nil {
			return err
		}
		devs = append(devs, strings.Trim(resp.String(), "\" "))
	}
	netDevs = strings.Join(devs, " | ")
	devs = devs[:0]
	done <- 1
	return nil
}

func net() {
	conn, err := dbus.SystemBus()
	if err != nil {
		netDevs = "Could not connect to the system bus."
		return
	}

	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, "type='signal',interface='org.freedesktop.NetworkManager',sender='org.freedesktop.NetworkManager'")

	ch := make(chan *dbus.Signal, 10)
	conn.Signal(ch)
	obj := conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")
	err = activeNetDevices(conn, obj)
	if err != nil {
		netDevs = fmt.Sprintf("%v", err)
	}
	for _ = range ch {
		err := activeNetDevices(conn, obj)
		if err != nil {
			netDevs = fmt.Sprintf("%v", err)
		}
	}
	netDevs = "Exited."
}

func vol() {
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
		volume = scanner.Text()
		done <- 1
	}
}

// TODO: if possible, write errors to status area, rather than panic
func main() {
	conn, err := xgbutil.NewConn()
	if err != nil {
		panic("Cannot open display.")
	}
	go bat()
	go clock()
	go net()
	go vol()
	for {
		_ = <-done
		// TODO: use strings.Join
		out := fmt.Sprintf(" 🔊 %s | %s | \u26A1%s%% | %s (%s UTC)", volume, netDevs, batPerc, localTime, utcTime)
		err := xprop.ChangeProp(conn, conn.RootWin(), 8, "WM_NAME", "STRING", []byte(out))
		if err != nil {
			panic("Cannot set status.")
		}
	}
}
