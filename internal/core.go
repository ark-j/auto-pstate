package internal

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
)

type EPP struct {
	watcher *Watcher
	stop    chan struct{}
	Mode    string
	isClose bool
}

func NewEPP(mode string) *EPP {
	e := &EPP{
		Mode: mode,
		stop: make(chan struct{}),
	}
	return e
}

func (e *EPP) Start() {
	var err error
	e.watcher, err = NewWatcher()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	go e.setState()
}

// WithManual will set governor and profile until service os restarted.
// It will be nop if mode is auto
func (e *EPP) WithManual(governor, profile string) {
	if e.Mode == ManualMode {
		e.setGoverner(governor)
		e.setEPP(profile)
	}
}

// actively listen to file modify event from inotify watcher
// if filemodified and it is chargining it will set proper default profile
func (e *EPP) setState() {
	e.firstBoot()
	for {
		select {
		case ok := <-e.watcher.ChargeEvent:
			switch ok {
			case true:
				e.setEPP(defaultEppStateAC)
			case false:
				e.setEPP(defaultEppStateBat)
			}
		case <-e.stop:
			if err := e.watcher.Close(); err != nil {
				slog.Error(err.Error())
			}
			return
		}
	}
}

// Sends close single stop event loop
func (e *EPP) Close() {
	if !e.isClose {
		e.stop <- struct{}{}
		e.isClose = true
		slog.Info("closing automatic power state setter")
	}
}

// execute only when laptop boots
// or service restarted
func (e *EPP) firstBoot() {
	e.setGoverner(defaultGovernor)
	switch charging() {
	case true:
		e.setEPP(defaultEppStateAC)
	case false:
		e.setEPP(defaultEppStateBat)
	}
}

// set the powersave governor if not already set
func (e *EPP) setGoverner(val string) {
	for i := 0; i < runtime.NumCPU(); i++ {
		if err := os.WriteFile(fmt.Sprintf(governorPath, i),
			[]byte(val), os.ModePerm); err != nil {
			slog.Error("while setting powersave governor", slog.Int("core", i), slog.String("err", err.Error()))
			continue
		}
	}
	slog.Info("epp governor is set", slog.String("governor", val), slog.String("mode", e.Mode))
}

// set epp value for performance and power consumption
func (e *EPP) setEPP(val string) {
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
	slog.Info("epp state is set", slog.String("profile", val), slog.String("mode", e.Mode))
}
