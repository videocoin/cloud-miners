package service

import (
	"github.com/videocoin/cloud-miners/datastore"
	"github.com/videocoin/cloud-miners/manager"
	"github.com/videocoin/cloud-miners/metrics"
	"github.com/videocoin/cloud-miners/rpc"
)

type Service struct {
	cfg *Config
	rpc *rpc.Server
	ds  *datastore.Datastore
	mc  *metrics.Collector
	ms  *metrics.Server
	dm  *manager.Manager
}

func NewService(cfg *Config) (*Service, error) {
	rpcConfig := &rpc.ServerOption{
		Logger:          cfg.Logger,
		Addr:            cfg.Addr,
		IamEndpoint:     cfg.IamEndpoint,
		DBURI:           cfg.DBURI,
		AuthTokenSecret: cfg.AuthTokenSecret,
	}

	ds, err := datastore.NewDatastore(cfg.DBURI)
	if err != nil {
		return nil, err
	}

	rpc, err := rpc.NewServer(rpcConfig, ds)
	if err != nil {
		return nil, err
	}

	ms, err := metrics.NewServer(cfg.MetricsAddr, cfg.Logger.WithField("system", "metrics-server"))
	if err != nil {
		return nil, err
	}

	mc := metrics.NewCollector(cfg.Name, ds)

	dm, err := manager.New(
		manager.WithLogger(cfg.Logger.WithField("system", "datamanager")),
		manager.WithDatastore(ds),
		manager.WithEmitterServiceClient(cfg.EmitterRPCAddr),
	)
	if err != nil {
		return nil, err
	}

	svc := &Service{
		cfg: cfg,
		rpc: rpc,
		ds:  ds,
		mc:  mc,
		ms:  ms,
		dm:  dm,
	}

	return svc, nil
}

func (s *Service) Start(errCh chan error) {
	go func() {
		s.cfg.Logger.Info("starting rpc server")
		errCh <- s.rpc.Start()
	}()

	go func() {
		s.cfg.Logger.Info("starting metrics server")
		errCh <- s.ms.Start()
	}()

	s.cfg.Logger.Info("starting metrics collector")
	s.mc.Start()

	s.cfg.Logger.Info("starting data manager")
	s.dm.Start()
}

func (s *Service) Stop() error {
	err := s.mc.Stop()
	if err != nil {
		return err
	}

	s.dm.Stop()

	return nil
}
