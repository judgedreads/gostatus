# gostatus

Event-based statusline generator for dwm, which could easily be modified
to work with other wms by writing to stdout instead of the root window
title.

## Installation
```
go install
```
The executable will then be available at $GOPATH/bin/gostatus.

## Configuration
Configuration is done in the code (typically in main.go).

## Plugins
Plugins can be added by simply writing code to satisfy the plugin
interface, and then adding an instance of the plugin to the plugins
array.

### timemon
A simple clock, which is instantiated with a format string and a time
zone. The format string and time zone should satisfy the requirements
detailed at godoc.org/time.

## batmon
Reads current capacity as a percentage from the file defined in
batmon.go, once per minute by default.

## netmon
Listens to dbus events from NetworkManager, displaying a list of the
active devices.

### Dependencies
- dbus
- NetworkManager

## volmon
Listens to ALSA events for the specified card/mix combination (e.g.
default/Master), and then displays the volume as a percentage.

### Dependencies
- ALSA dev libs
