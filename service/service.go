package service

import "time"

type Service struct {
	cfg    *Config
	rpc    *RPCServer
	syncer *DataSyncer
}

func NewService(cfg *Config) (*Service, error) {
	rpcConfig := &RPCServerOptions{
		Addr:            cfg.Addr,
		DBURI:           cfg.DBURI,
		AuthTokenSecret: cfg.AuthTokenSecret,
		Logger:          cfg.Logger,
	}

	ds, err := NewDatastore(cfg.DBURI)
	if err != nil {
		return nil, err
	}

	rpc, err := NewRPCServer(rpcConfig, ds)
	if err != nil {
		return nil, err
	}

	syncerLogger := cfg.Logger.WithField("system", "syncer")
	syncer, err := NewDataSyncer(cfg.GraphiteAddr, 10*time.Second, ds, syncerLogger)
	if err != nil {
		return nil, err
	}

	svc := &Service{
		cfg:    cfg,
		rpc:    rpc,
		syncer: syncer,
	}

	return svc, nil
}

func (s *Service) Start() error {
	go s.rpc.Start()
	go s.syncer.Start()
	return nil
}

func (s *Service) Stop() error {
	go s.syncer.Stop()
	return nil
}
