package service

import (
	"net"

	"github.com/sirupsen/logrus"
	v1 "github.com/videocoin/cloud-api/miners/v1"
	"github.com/videocoin/cloud-miners/eventbus"
	"github.com/videocoin/cloud-pkg/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type RPCServerOptions struct {
	Addr            string
	DBURI           string
	AuthTokenSecret string

	Logger *logrus.Entry
}

type RPCServer struct {
	addr   string
	grpc   *grpc.Server
	listen net.Listener
	logger *logrus.Entry
	ds     *Datastore
	eb     *eventbus.EventBus

	authTokenSecret string
}

func NewRPCServer(opts *RPCServerOptions, ds *Datastore, eb *eventbus.EventBus) (*RPCServer, error) {
	grpcOpts := grpcutil.DefaultServerOpts(opts.Logger)
	grpcServer := grpc.NewServer(grpcOpts...)

	healthService := health.NewServer()
	healthv1.RegisterHealthServer(grpcServer, healthService)

	listen, err := net.Listen("tcp", opts.Addr)
	if err != nil {
		return nil, err
	}

	rpcServer := &RPCServer{
		logger:          opts.Logger,
		addr:            opts.Addr,
		grpc:            grpcServer,
		listen:          listen,
		ds:              ds,
		eb:              eb,
		authTokenSecret: opts.AuthTokenSecret,
	}

	v1.RegisterMinersServiceServer(grpcServer, rpcServer)
	reflection.Register(grpcServer)

	return rpcServer, nil
}

func (s *RPCServer) Start() error {
	s.logger.Infof("starting rpc server on %s", s.addr)
	return s.grpc.Serve(s.listen)
}
