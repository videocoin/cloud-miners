package service

import (
	"github.com/videocoin/cloud-miners/eventbus"
	"github.com/videocoin/cloud-pkg/grpcutil"
	iamv1 "github.com/videocoin/videocoinapis-admin/videocoin/admin/iam/admin/v1"

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
	conn, err := grpcutil.Connect(cfg.IAMRPCAddr, cfg.Logger.WithField("system", "iamcli"))
	if err != nil {
		return nil, err
	}
	iam := iamv1.NewIAMClient(conn)

	rpcConfig := &RPCServerOptions{
		Addr:            cfg.Addr,
		DBURI:           cfg.DBURI,
		Iam: iam,
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
	go s.rpc.Start()
	go s.eb.Start()
	go s.mc.Start()
	go s.ms.Start()
	s.ds.StartBackgroundTasks()
	return nil
}

func (s *Service) Stop() error {
	s.ds.StopBackgroundTasks()
	s.eb.Stop()
	s.mc.Stop()
	return nil
}
