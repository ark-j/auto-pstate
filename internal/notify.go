package internal

import (
	"errors"
	"syscall"
	"time"
	"unsafe"
)

type Watcher struct {
	fd, wd      int
	ChargeEvent chan bool
	stop        chan struct{}
	closed      bool
}

// NewWatcher will create instance of watcher.
//
// 1. It will create Inotify syscall in nonblock mode
//
// 2. Add the /sys/class/power_supply/AC/online to watch mode for modify
func NewWatcher() (*Watcher, error) {
	w := &Watcher{ChargeEvent: make(chan bool), stop: make(chan struct{})}
	var err error
	w.fd, err = syscall.InotifyInit1(syscall.IN_CLOEXEC | syscall.IN_NONBLOCK)
	if err != nil {
		return nil, err
	}
	w.wd, err = syscall.InotifyAddWatch(w.fd, batPath, syscall.SYS_MODIFY_LDT)
	if err != nil {
		return nil, err
	}
	return w, nil
}

// Close function sends single to stop readevent loop. Closes Watch and inotify file descriptor.
func (w *Watcher) Close() error {
	if !w.closed {
		w.closed = true
		w.stop <- struct{}{}
		if _, err := syscall.InotifyRmWatch(w.fd, uint32(w.wd)); err != nil {
			return err
		}
		return syscall.Close(w.fd)
	}
	return errors.New("already closed")
}

// ReadEvents reads the modify event and send bool over ChargeEvent event channel
func (w *Watcher) ReadEvents() {
	var buf [4096]byte
	for {
		select {
		case <-w.stop:
			return
		default:
			n, err := syscall.Read(w.fd, buf[:])
			if err != nil {
				return
			}
			var offset uint32
			for offset < uint32(n) {
				ie := (*syscall.InotifyEvent)(unsafe.Pointer(&buf[offset]))
				mask := ie.Mask
				if mask&syscall.SYS_MODIFY_LDT != 0 {
					w.ChargeEvent <- charging()
				}
				offset += syscall.SizeofInotifyEvent + ie.Len
			}
		}
	}
}
