package internal

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

func (srv *Server) Start() {}

func (srv *Server) Close() {}

func (srv *Server) ManualHandler() {}

func (srv *Server) AutoHandler() {}

func (srv *Server) TimerHandler() {}
