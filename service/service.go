package service

import (
	"github.com/videocoin/cloud-miners/eventbus"
)

type Service struct {
	cfg *Config
	rpc *RPCServer
	ds  *Datastore
	eb  *eventbus.EventBus
	mc  *MetricsCollector
	ms  *MetricsServer
}

func NewService(cfg *Config) (*Service, error) {
	rpcConfig := &RPCServerOptions{
		Addr:            cfg.Addr,
		DBURI:           cfg.DBURI,
		AuthTokenSecret: cfg.AuthTokenSecret,
		Logger:          cfg.Logger,
	}

	ebConfig := &eventbus.Config{
		URI:    cfg.MQURI,
		Name:   cfg.Name,
		Logger: cfg.Logger.WithField("system", "eventbus"),
	}
	eb, err := eventbus.New(ebConfig)
	if err != nil {
		return nil, err
	}

	ds, err := NewDatastore(cfg.DBURI, eb, cfg.Logger.WithField("system", "datastore"))
	if err != nil {
		return nil, err
	}

	rpc, err := NewRPCServer(rpcConfig, ds)
	if err != nil {
		return nil, err
	}

	metricsServer, err := NewMetricsServer(cfg.MetricsAddr, cfg.Logger.WithField("system", "metrics-server"))
	if err != nil {
		return nil, err
	}

	metricsCollector := NewMetricsCollector(cfg.Name, ds)

	svc := &Service{
		cfg: cfg,
		rpc: rpc,
		ds:  ds,
		eb:  eb,
		mc:  metricsCollector,
		ms:  metricsServer,
	}

	return svc, nil
}

func (s *Service) Start() error {
	go s.rpc.Start()  //nolint
	go s.eb.Start()  //nolint
	go s.mc.Start()  //nolint
	go s.ms.Start()  //nolint
	s.ds.StartBackgroundTasks()
	return nil
}

func (s *Service) Stop() error {
	err := s.ds.StopBackgroundTasks()
	if err != nil {
		return err
	}
	err = s.eb.Stop()
	if err != nil {
		return err
	}
	err = s.mc.Stop()
	if err != nil {
		return err
	}
	return nil
}
