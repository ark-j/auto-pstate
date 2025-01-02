package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ark-j/auto-pstate/internal"
)

func main() {
	internal.SetLogger()

	// prechecks
	internal.IsRoot()
	internal.IsPState()
	createDaemonDir()

	// daemon start
	epp := internal.NewEPP(internal.AutoMode)
	epp.Start()
	srv := internal.NewServer(epp)
	defer srv.Close()
	go srv.Start()
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-done
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
