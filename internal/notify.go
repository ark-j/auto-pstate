package internal

import (
	"log/slog"
	"os"
	"syscall"
	"unsafe"
)

type Watcher struct {
	fd, wd      int
	ChargeEvent chan bool
	stop        chan struct{}
	isClosed    bool
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
	go w.ReadEvents()
	return w, nil
}

// Close function sends single to stop readevent loop. Closes Watch and inotify file descriptor.
func (w *Watcher) Close() error {
	if !w.isClosed {
		w.isClosed = true
		slog.Info("closing event watcher")
		w.stop <- struct{}{}
		if _, err := syscall.InotifyRmWatch(w.fd, uint32(w.wd)); err != nil {
			return err
		}
		return syscall.Close(w.fd)
	}
	return nil
}

// ReadEvents reads the modify event and send bool over ChargeEvent event channel
// it uses loop counter toggle as technique to avoid repeate event handling due to
// the nature of inotify generating SYS_MODIFY_LDT even after read
func (w *Watcher) ReadEvents() {
	var buf [4096]byte
	f := os.NewFile(uintptr(w.fd), "")
	defer f.Close()
	next := false
	for {
		select {
		case <-w.stop:
			return
		default:
			n, err := f.Read(buf[:])
			if err != nil {
				slog.Error(err.Error())
				return
			}
			var offset uint32
			for offset < uint32(n) {
				ie := (*syscall.InotifyEvent)(unsafe.Pointer(&buf[offset]))
				mask := ie.Mask
				if mask&syscall.SYS_MODIFY_LDT != 0 && next {
					w.ChargeEvent <- charging()
				}
				next = !next
				offset += syscall.SizeofInotifyEvent + ie.Len
			}
		}
	}
}
