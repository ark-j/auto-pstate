package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/ark-j/auto-pstate/internal"
)

func main() {
	internal.SetLogger()

	// prechecks
	// internal.IsRoot()
	// internal.IsPState()
	// createDaemonDir()

	// daemon start
	// epp := internal.NewEPP(internal.AutoMode)
	// srv := internal.NewServer(epp)
	// defer srv.Close()
	// go srv.Start()
	// done := make(chan os.Signal, 1)
	// signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	// <-done
	w, err := internal.NewWatcher()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	for e := range w.ChargeEvent {
		if e {
			fmt.Println("charging")
		} else {
			fmt.Println("on battery")
		}
	}
}

func createDaemonDir() {
	if err := os.Mkdir("/run/auto-epp", 0o644); err != nil {
		if os.IsExist(err) {
			return
		}
		slog.Error(err.Error())
		os.Exit(1)
	}
}
