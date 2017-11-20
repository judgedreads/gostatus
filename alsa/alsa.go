package alsa

/*
#cgo pkg-config: alsa
#include <stdlib.h>
#include <alsa/asoundlib.h>
extern int mixer_elem_cb(snd_mixer_elem_t *elem, unsigned int mask);
*/
import "C"
import (
	"errors"
	"fmt"
	"log"
	"unsafe"
)

const FRONT_LEFT = C.SND_MIXER_SCHN_FRONT_LEFT
const EVENT_MASK_VALUE = C.SND_CTL_EVENT_MASK_VALUE

type Mixer C.snd_mixer_t

func Open(card string) (*Mixer, error) {
	var handle *C.snd_mixer_t
	// mode arg is unused, so just pass 0
	if C.snd_mixer_open(&handle, 0) < 0 {
		// TODO: surface alsa err messages using
		// snd_strerror(err)
		return nil, errors.New("failed to open mixer")
	}
	cardC := C.CString(card)
	defer C.free(unsafe.Pointer(cardC))
	if C.snd_mixer_attach(handle, cardC) < 0 {
		(*Mixer)(handle).Close()
		return nil, fmt.Errorf("failed to attach %q", card)
	}
	if C.snd_mixer_selem_register(handle, nil, nil) < 0 {
		(*Mixer)(handle).Close()
		return nil, errors.New("failed to register mixer")
	}
	if C.snd_mixer_load(handle) < 0 {
		(*Mixer)(handle).Close()
		return nil, errors.New("failed to load mixer")
	}
	return (*Mixer)(handle), nil
}

func (m *Mixer) Close() {
	C.snd_mixer_close((*C.snd_mixer_t)(m))
}

func (m *Mixer) Elem(name string) (*Elem, error) {
	var sid *C.snd_mixer_selem_id_t
	// had trouble with alloca, so use the heap instead
	if C.snd_mixer_selem_id_malloc(&sid) != 0 {
		return nil, errors.New("failed to malloc sid")
	}
	defer C.snd_mixer_selem_id_free(sid)
	C.snd_mixer_selem_id_set_index(sid, 0)
	nameC := C.CString(name)
	defer C.free(unsafe.Pointer(nameC))
	C.snd_mixer_selem_id_set_name(sid, nameC)
	var elem *C.snd_mixer_elem_t
	elem = C.snd_mixer_find_selem((*C.snd_mixer_t)(m), sid)
	if elem == nil {
		return nil, fmt.Errorf("failed to find elem %q", name)
	}
	return (*Elem)(elem), nil
}

func (m *Mixer) Listen() {
	for {
		if C.snd_mixer_wait((*C.snd_mixer_t)(m), -1) < 0 {
			continue
		}
		if C.snd_mixer_handle_events((*C.snd_mixer_t)(m)) < 0 {
			close(elemEvents)
			log.Printf("failed to handle events")
		}
	}
}

// TODO: maintain a map of elem_name to chan
var elemEvents = make(chan int)

//export mixer_elem_cb
func mixer_elem_cb(elem *C.snd_mixer_elem_t, mask C.uint) C.int {
	if (mask & EVENT_MASK_VALUE) != 0 {
		elemEvents <- 1
	}
	return 0
}

type Elem C.snd_mixer_elem_t

func (e *Elem) PlaybackVolumeRange() (min, max int) {
	var minC, maxC C.long
	C.snd_mixer_selem_get_playback_volume_range((*C.snd_mixer_elem_t)(e), &minC, &maxC)
	min, max = int(minC), int(maxC)
	return
}

func (e *Elem) PlaybackVolume() int {
	var volC C.long
	C.snd_mixer_selem_get_playback_volume((*C.snd_mixer_elem_t)(e), FRONT_LEFT, &volC)
	return int(volC)
}

// PlaybackSwitch returns 0 if the element is muted, and 1 otherwise
func (e *Elem) PlaybackSwitch() int {
	var psC C.int
	C.snd_mixer_selem_get_playback_switch((*C.snd_mixer_elem_t)(e), FRONT_LEFT, &psC)
	return int(psC)
}

func (e *Elem) Subscribe() chan int {
	// need to cast C function pointer to *[0]byte, as that is how
	// unsupported C pointers are represented in go
	C.snd_mixer_elem_set_callback((*C.snd_mixer_elem_t)(e), (*[0]byte)(C.mixer_elem_cb))
	return elemEvents
}
