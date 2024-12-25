package internal

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
)

const SockPath = "/var/run/auto-epp.sock"

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

func (srv *Server) Close() {
	srv.epp.Close()
	srv.srv.Close()
	srv.listner.Close()
}

func (srv *Server) routes() {
	srv.mux.HandleFunc("/epp/manual", srv.ManualHandler)
	srv.mux.HandleFunc("/epp/auto", srv.AutoHandler)
	srv.mux.HandleFunc("/epp/timer", srv.TimerHandler)
}

// TODO: extract and refactor
func (srv *Server) ManualHandler(w http.ResponseWriter, r *http.Request) {
	srv.epp.Close()
	srv.epp.Mode = ManualMode
	var m ManualRequest
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		if err = json.NewEncoder(w).Encode(map[string]string{"msg": err.Error()}); err != nil {
			slog.Error(err.Error())
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		return
	}

	if !ListAvailable(false)[m.Governor] {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"msg": fmt.Sprintf("%s governor is not found", m.Governor)}); err != nil { //nolint
			slog.Error(err.Error())
		}
		return
	}

	if !ListAvailable(true)[m.Profile] {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"msg": fmt.Sprintf("%s profile is not found", m.Profile)}); err != nil { //nolint
			slog.Error(err.Error())
		}
		return
	}
	srv.epp.WithManual(m.Governor, m.Profile)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"msg":      "epp state saved successfully. please renable auto mode or restart reservice",
		"mode":     ManualMode,
		"governor": m.Governor,
		"profile":  m.Profile,
	}); err != nil {
		slog.Error(err.Error())
	}
}

// TODO: extract and refactor
func (srv *Server) AutoHandler(w http.ResponseWriter, r *http.Request) { //nolint
	if srv.epp.Mode == ManualMode {
		srv.epp.Mode = AutoMode
		go srv.epp.SetState()
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{"msg": "auto mode enabled"}); err != nil {
		slog.Error(err.Error())
	}
}

// TODO: extract and refactor. add meaningful response and validate duration
func (srv *Server) TimerHandler(w http.ResponseWriter, r *http.Request) {
	var m TimerRequest
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		if err = json.NewEncoder(w).Encode(map[string]string{"msg": err.Error()}); err != nil {
			slog.Error(err.Error())
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		return
	}

	if !srv.governors[m.Governor] {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"msg": fmt.Sprintf("%s governor is not found", m.Governor)}); err != nil { //nolint
			slog.Error(err.Error())
		}
		return
	}

	if !srv.profiles[m.Profile] {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"msg": fmt.Sprintf("%s profile is not found", m.Profile)}); err != nil { //nolint
			slog.Error(err.Error())
		}
		return
	}

	srv.epp.WithTimer(m.Duration, m.Governor, m.Profile)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{"msg": "timer mode is enabled"}); err != nil {
		slog.Error(err.Error())
	}
}
