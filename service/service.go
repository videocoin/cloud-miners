package service

type Service struct {
	cfg *Config
	rpc *RPCServer
	ds  *Datastore
}

func NewService(cfg *Config) (*Service, error) {
	rpcConfig := &RPCServerOptions{
		Addr:            cfg.Addr,
		DBURI:           cfg.DBURI,
		AuthTokenSecret: cfg.AuthTokenSecret,
		Logger:          cfg.Logger,
	}

	ds, err := NewDatastore(cfg.DBURI, cfg.Logger.WithField("system", "datastore"))
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
	}

	return svc, nil
}

func (s *Service) Start() error {
	go s.rpc.Start()
	s.ds.StartBackgroundTasks()
	return nil
}

func (s *Service) Stop() error {
	s.ds.StopBackgroundTasks()
	return nil
}
