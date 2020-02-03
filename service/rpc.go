package service

import (
	"net"

	"github.com/sirupsen/logrus"
	v1 "github.com/videocoin/cloud-api/miners/v1"
	"github.com/videocoin/cloud-pkg/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	iamv1 "github.com/videocoin/videocoinapis-admin/videocoin/admin/iam/admin/v1"
)

type RPCServerOptions struct {
	Addr            string
	DBURI           string

	Iam        iamv1.IAMClient

	AuthTokenSecret string

	Logger *logrus.Entry
}

type RPCServer struct {
	addr   string
	grpc   *grpc.Server
	listen net.Listener
	logger *logrus.Entry
	ds     *Datastore

	iam        iamv1.IAMClient

	authTokenSecret string
}

func NewRPCServer(opts *RPCServerOptions, ds *Datastore) (*RPCServer, error) {
	grpcOpts := grpcutil.DefaultServerOpts(opts.Logger)
	grpcServer := grpc.NewServer(grpcOpts...)
	healthService := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthService)
	listen, err := net.Listen("tcp", opts.Addr)
	if err != nil {
		return nil, err
	}

	rpcServer := &RPCServer{
		addr:            opts.Addr,
		grpc:            grpcServer,
		listen:          listen,
		logger:          opts.Logger,
		ds:              ds,
		iam:             opts.Iam,
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
