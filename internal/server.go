package internal

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
)

const SockPath = "/var/run/auto-epp.sock"

type Handler func(w http.ResponseWriter, r *http.Request) error

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		if errors.Is(err, AppErr{}) {
			e := err.(AppErr)
			if err = JSON(H{"msg": e.Msg}, w, e.StatusCode); err != nil {
				slog.Error(e.Error())
			}
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type Server struct {
	mux     *http.ServeMux
	srv     *http.Server
	listner net.Listener
	epp     *EPP
}

func NewServer(epp *EPP) *Server {
	var err error
	srv := &Server{epp: epp}
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

// Close closes the epp and server. as well as socket
func (srv *Server) Close() {
	srv.epp.Close()
	srv.srv.Close()
	srv.listner.Close()
}

// routes function registers the routes
func (srv *Server) routes() {
	srv.HandleFunc("/epp/manual", srv.ManualHandler)
	srv.HandleFunc("/epp/auto", srv.AutoHandler)
	srv.HandleFunc("/epp/timer", srv.TimerHandler)
}

func (srv *Server) HandleFunc(pattern string, h Handler) {
	srv.mux.Handle(pattern, Handler(h))
}

func (srv *Server) ManualHandler(w http.ResponseWriter, r *http.Request) error {
	srv.epp.Close()
	srv.epp.Mode = ManualMode
	var m ManualRequest
	if err := Bind(r.Body, &m); err != nil {
		return AppErr{
			Msg:        "invalid body",
			StatusCode: http.StatusUnprocessableEntity,
			Err:        err,
		}
	}

	if !ListAvailable(false)[m.Governor] {
		msg := fmt.Sprintf("%s governor is not found", m.Governor)
		return AppErr{
			Msg:        msg,
			StatusCode: http.StatusUnprocessableEntity,
			Err:        errors.New(msg),
		}
	}

	if !ListAvailable(true)[m.Profile] {
		msg := fmt.Sprintf("%s profile is not found", m.Profile)
		return AppErr{
			Msg:        msg,
			StatusCode: http.StatusUnprocessableEntity,
			Err:        errors.New(msg),
		}
	}

	srv.epp.WithManual(m.Governor, m.Profile)
	if err := JSON(H{
		"msg":      "epp state saved successfully. please renable auto mode or restart reservice",
		"mode":     ManualMode,
		"governor": m.Governor,
		"profile":  m.Profile,
	}, w, http.StatusOK); err != nil {
		slog.Error(err.Error())
	}
	return nil
}

func (srv *Server) AutoHandler(w http.ResponseWriter, r *http.Request) error { //nolint
	if srv.epp.Mode == ManualMode {
		srv.epp.Mode = AutoMode
		go srv.epp.SetState()
	}
	if err := JSON(H{"msg": "auto mode enabled"}, w, http.StatusOK); err != nil {
		slog.Error(err.Error())
	}
	return nil
}

func (srv *Server) TimerHandler(w http.ResponseWriter, r *http.Request) error {
	var m TimerRequest
	if err := Bind(r.Body, &m); err != nil {
		return AppErr{
			Msg:        "invalid body",
			StatusCode: http.StatusUnprocessableEntity,
			Err:        err,
		}
	}

	if !ListAvailable(false)[m.Governor] {
		msg := fmt.Sprintf("%s governor is not found", m.Governor)
		return AppErr{
			Msg:        msg,
			StatusCode: http.StatusUnprocessableEntity,
			Err:        errors.New(msg),
		}
	}

	if !ListAvailable(true)[m.Profile] {
		msg := fmt.Sprintf("%s profile is not found", m.Profile)
		return AppErr{
			Msg:        msg,
			StatusCode: http.StatusUnprocessableEntity,
			Err:        errors.New(msg),
		}
	}

	srv.epp.WithTimer(m.Duration, m.Governor, m.Profile)
	if err := JSON(H{"msg": "timer mode is enabled"}, w, http.StatusOK); err != nil {
		slog.Error(err.Error())
	}
	return nil
}
