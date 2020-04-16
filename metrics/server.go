package metrics

import (
	"github.com/labstack/echo"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type ServerConfig struct {
	Addr string
}

type Server struct {
	logger *logrus.Entry
	addr   string
	e      *echo.Echo
}

func NewServer(addr string, logger *logrus.Entry) (*Server, error) {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.DisableHTTP2 = true

	return &Server{
		logger: logger,
		addr:   addr,
		e:      e,
	}, nil
}

func (s *Server) Start() error {
	s.logger.Infof("metrics server listening on %s", s.addr)
	s.routes()
	return s.e.Start(s.addr)
}

func (s *Server) Stop() error {
	return nil
}

func (s *Server) routes() {
	s.e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
}
