package service

import (
	"github.com/videocoin/cloud-miners/eventbus"
)

type Service struct {
	cfg *Config
	rpc *RPCServer
	ds  *Datastore
	eb  *eventbus.EventBus
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

	svc := &Service{
		cfg: cfg,
		rpc: rpc,
		ds:  ds,
		eb:  eb,
	}

	return svc, nil
}

func (s *Service) Start() error {
	go s.rpc.Start()
	go s.eb.Start()
	s.ds.StartBackgroundTasks()
	return nil
}

func (s *Service) Stop() error {
	s.ds.StopBackgroundTasks()
	s.eb.Stop()
	return nil
}
