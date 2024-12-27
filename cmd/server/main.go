package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ark-j/auto-pstate/internal"
)

func main() {
	// prechecks
	internal.IsRoot()
	internal.IsPState()
	createDaemonDir()

	// daemon step
	epp := internal.NewEPP(internal.AutoMode)
	srv := internal.NewServer(epp)
	defer srv.Close()
	go srv.Start()
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-done
}

func createDaemonDir() {
	if err := os.Mkdir("/run/auto-pstate", 0o644); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
