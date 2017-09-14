package main

import (
	"fmt"
	"strings"

	"github.com/godbus/dbus"
)

// TODO: put an error attribute on the struct, instead of co-opting
// devices
type netMon struct {
	devices string
}

func (n *netMon) Run(done chan int) {
	conn, err := dbus.SystemBus()
	if err != nil {
		n.devices = "Could not connect to the system bus."
		return
	}

	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, "type='signal',interface='org.freedesktop.NetworkManager',sender='org.freedesktop.NetworkManager'")

	ch := make(chan *dbus.Signal, 10)
	conn.Signal(ch)
	obj := conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")
	err = n.activeNetDevices(conn, obj)
	done <- 1
	if err != nil {
		n.devices = fmt.Sprintf("%v", err)
	}
	for _ = range ch {
		err := n.activeNetDevices(conn, obj)
		if err != nil {
			n.devices = fmt.Sprintf("%v", err)
		}
		done <- 1
	}
	n.devices = "Exited."
}

func (n *netMon) activeNetDevices(conn *dbus.Conn, obj dbus.BusObject) error {
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
		path = strings.TrimPrefix(resp.String(), "@ao ")
		path = strings.Trim(path, "\"[] ")
		if path == "" {
			continue
		}
		obj = conn.Object("org.freedesktop.NetworkManager", dbus.ObjectPath(path))
		resp, err = obj.GetProperty("org.freedesktop.NetworkManager.Device.Interface")
		if err != nil {
			return err
		}
		devs = append(devs, strings.Trim(resp.String(), "\" "))
	}
	n.devices = strings.Join(devs, " | ")
	devs = devs[:0]
	return nil
}

func (n *netMon) String() string {
	return n.devices
}
