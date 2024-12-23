package internal

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"
)

type EPP struct {
	Watcher    *Watcher
	halt, stop chan struct{}
	Mode       string
	timer      *time.Timer
	closed     bool
}

func NewEPP(mode string) *EPP {
	w, err := NewWatcher()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	e := &EPP{
		halt:    make(chan struct{}),
		Mode:    mode,
		Watcher: w,
	}
	go e.SetState()
	return e
}

// WithTimer will set governor and profile for said amount of Duration.
// It will only be set in automode, on maunal mode it is nop
func (e *EPP) WithTimer(d time.Duration, governor, profile string) {
	if e.Mode == AutoMode {
		e.halt <- struct{}{}
		setGoverner(governor)
		setEPP(profile, ManualMode)
		t := time.NewTimer(d)
		e.timer = t
	}
}

// WithManual will set governor and profile until service os restarted.
// It will be nop if mode is auto
func (e *EPP) WithManual(governor, profile string) {
	if e.Mode == ManualMode {
		setGoverner(governor)
		setEPP(profile, ManualMode)
	}
}

// actively listen to file modify event from inotify watcher
// if filemodified and it is chargining it will set proper default profile
func (e *EPP) SetState() {
	firstBoot()
	setGoverner(defaultGovernor)
	for {
		select {
		case ok := <-e.Watcher.ChargeEvent:
			switch ok {
			case true:
				setEPP(defaultEppStateAC, AutoMode)
			case false:
				setEPP(defaultEppStateBat, AutoMode)
			}
		case <-e.halt:
			<-e.timer.C
			e.timer.Stop()
		case <-e.stop:
			if err := e.Watcher.Close(); err != nil {
				slog.Error(err.Error())
			}
		}
	}
}

// Sends close single stop event loop
func (e *EPP) Close() {
	if !e.closed {
		e.stop <- struct{}{}
		e.closed = true
	}
}

// execute only when laptop boots
// or service restarted
func firstBoot() {
	switch charging() {
	case true:
		setEPP(defaultEppStateAC, AutoMode)
	case false:
		setEPP(defaultEppStateBat, AutoMode)
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
