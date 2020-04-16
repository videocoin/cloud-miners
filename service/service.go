package service

import (
	"github.com/videocoin/cloud-miners/datastore"
	"github.com/videocoin/cloud-miners/eventbus"
	"github.com/videocoin/cloud-miners/metrics"
	"github.com/videocoin/cloud-miners/rpc"
)

type Service struct {
	cfg *Config
	rpc *rpc.Server
	ds  *datastore.Datastore
	eb  *eventbus.EventBus
	mc  *metrics.Collector
	ms  *metrics.Server
}

func NewService(cfg *Config) (*Service, error) {
	rpcConfig := &rpc.ServerOption{
		Logger:          cfg.Logger,
		Addr:            cfg.Addr,
		DBURI:           cfg.DBURI,
		AuthTokenSecret: cfg.AuthTokenSecret,
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

	ds, err := datastore.NewDatastore(cfg.DBURI, eb, cfg.Logger.WithField("system", "datastore"))
	if err != nil {
		return nil, err
	}

	rpc, err := rpc.NewServer(rpcConfig, ds, eb)
	if err != nil {
		return nil, err
	}

	ms, err := metrics.NewServer(cfg.MetricsAddr, cfg.Logger.WithField("system", "metrics-server"))
	if err != nil {
		return nil, err
	}

	mc := metrics.NewCollector(cfg.Name, ds)

	svc := &Service{
		cfg: cfg,
		rpc: rpc,
		ds:  ds,
		eb:  eb,
		mc:  mc,
		ms:  ms,
	}

	return svc, nil
}

func (s *Service) Start(errCh chan error) {
	go func() {
		errCh <- s.rpc.Start()
	}()

	go func() {
		errCh <- s.eb.Start()
	}()

	go func() {
		errCh <- s.ms.Start()
	}()

	s.mc.Start()
	s.ds.StartBackgroundTasks()
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
