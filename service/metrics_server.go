package service

import (
	"github.com/labstack/echo"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type MetricsServerConfig struct {
	Addr string
}

type MetricsServer struct {
	logger *logrus.Entry
	addr   string
	e      *echo.Echo
}

func NewMetricsServer(addr string, logger *logrus.Entry) (*MetricsServer, error) {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.DisableHTTP2 = true

	return &MetricsServer{
		logger: logger,
		addr:   addr,
		e:      e,
	}, nil
}

func (s *MetricsServer) Start() error {
	s.logger.Infof("metrics server listening on %s", s.addr)
	s.routes()
	return s.e.Start(s.addr)
}

func (s *MetricsServer) Stop() error {
	return nil
}

func (s *MetricsServer) routes() {
	s.e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
}
