package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/ark-j/auto-pstate/internal"
)

func main() {
	epp := internal.NewEPP(internal.AutoMode)
	srv := internal.NewServer(epp)
	defer srv.Close()
	go srv.Start()
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-done
}
