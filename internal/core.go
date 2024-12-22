package internal

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"
)

type EPP struct {
	watcher    *Watcher
	halt, stop chan struct{}
	mode       string
	timer      *time.Timer
}

func NewEPP(mode string) *EPP {
	w, err := NewWatcher()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	e := &EPP{
		halt:    make(chan struct{}),
		mode:    mode,
		watcher: w,
	}
	go e.SetState()
	return e
}

// EnableManual enables manual mode
// it is needed to run WithTimer method
func (e *EPP) EnableManual() {
	e.mode = manual
}

// WithTimer will set governor and profile for said amount of Duration.
// It will be nop if mode is auto
func (e *EPP) WithTimer(d time.Duration, governor, profile string) {
	if e.mode == manual {
		e.halt <- struct{}{}
		setGoverner(governor)
		setEPP(profile, manual)
		t := time.NewTimer(d)
		e.timer = t
	}
}

// WithManual will set governor and profile until service os restarted.
// It will be nop if mode is auto
func (e *EPP) WithManual(governor, profile string) {
	if e.mode == manual {
		setGoverner(governor)
		setEPP(profile, manual)
	}
}

// actively listen to file modify event from inotify watcher
// if filemodified and it is chargining it will set proper default profile
func (e *EPP) SetState() {
	firstBoot()
	setGoverner(defaultGovernor)
	for {
		select {
		case ok := <-e.watcher.ChargeEvent:
			switch ok {
			case true:
				setEPP(defaultEppStateAC, auto)
			case false:
				setEPP(defaultEppStateBat, auto)
			}
		case <-e.halt:
			<-e.timer.C
			e.timer.Stop()
		case <-e.stop:
			if err := e.watcher.Close(); err != nil {
				slog.Error(err.Error())
			}
		}
	}
}

// Sends close single stop event loop
func (e *EPP) Close() {
	e.stop <- struct{}{}
}

// execute only when laptop boots
// or service restarted
func firstBoot() {
	switch charging() {
	case true:
		setEPP(defaultEppStateAC, auto)
	case false:
		setEPP(defaultEppStateBat, auto)
	}
}

// set the powersave governor if not already set
func setGoverner(g string) {
	for i := 0; i < runtime.NumCPU(); i++ {
		if err := os.WriteFile(fmt.Sprintf(governorPath, i),
			[]byte(g), os.ModePerm); err != nil {
			slog.Error("while setting powersave governor", slog.Int("core", i), slog.String("err", err.Error()))
			continue
		}
	}
}

// set epp value for performance and power consumption
func setEPP(val string, mode string) {
	for i := 0; i < runtime.NumCPU(); i++ {
		err := os.WriteFile(fmt.Sprintf(eppPath, i), []byte(val), os.ModePerm)
		if err != nil {
			slog.Error(
				"while setting epp_value",
				slog.String("value", val),
				slog.Int("core", i),
				slog.String("err", err.Error()),
			)
			continue
		}
	}
	slog.Info("epp state is set", slog.String("profile", val), slog.String("mode", mode))
}
