package internal

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
)

const SockPath = "/var/run/epp.sock"

type Server struct {
	mux     *http.ServeMux
	srv     *http.Server
	listner net.Listener
}

func NewServer() *Server {
	var err error
	srv := &Server{}
	srv.listner, err = net.Listen("unix", SockPath)
	if err != nil {
		slog.Error(fmt.Sprintf("unable to start unix socket listner: %v", err))
		os.Exit(1)
	}
	srv.mux = http.NewServeMux()
	srv.srv = &http.Server{
		Handler: srv.mux,
	}
	return srv
}

// Start starts the server recommended to launch it in goroutine
func (srv *Server) Start() {
	srv.routes()
	if err := srv.srv.Serve(srv.listner); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func (srv *Server) Close() {
	srv.srv.Close()
	srv.listner.Close()
}

func (srv *Server) routes() {}

func (srv *Server) ManualHandler(w http.ResponseWriter, r *http.Request) {}

func (srv *Server) AutoHandler(w http.ResponseWriter, r *http.Request) {}

func (srv *Server) TimerHandler(w http.ResponseWriter, r *http.Request) {}
