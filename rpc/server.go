package rpc

import (
	"net"

	"github.com/sirupsen/logrus"
	v1 "github.com/videocoin/cloud-api/miners/v1"
	"github.com/videocoin/cloud-miners/datastore"
	"github.com/videocoin/cloud-pkg/grpcutil"
	"github.com/videocoin/cloud-pkg/iam"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type ServerOption struct {
	Logger          *logrus.Entry
	Addr            string
	DBURI           string
	AuthTokenSecret string
	IAM             *iam.Client
}

type Server struct {
	logger          *logrus.Entry
	addr            string
	authTokenSecret string
	grpc            *grpc.Server
	listen          net.Listener
	ds              *datastore.Datastore
	iam             *iam.Client
}

func NewServer(opts *ServerOption, ds *datastore.Datastore) (*Server, error) {
	grpcOpts := grpcutil.DefaultServerOpts(opts.Logger)
	grpcServer := grpc.NewServer(grpcOpts...)

	healthService := health.NewServer()
	healthv1.RegisterHealthServer(grpcServer, healthService)

	listen, err := net.Listen("tcp", opts.Addr)
	if err != nil {
		return nil, err
	}

	rpcServer := &Server{
		logger:          opts.Logger,
		addr:            opts.Addr,
		authTokenSecret: opts.AuthTokenSecret,
		iam:             opts.IAM,
		grpc:            grpcServer,
		listen:          listen,
		ds:              ds,
	}

	v1.RegisterMinersServiceServer(grpcServer, rpcServer)
	reflection.Register(grpcServer)

	return rpcServer, nil
}

func (s *Server) Start() error {
	s.logger.Infof("starting rpc server on %s", s.addr)
	return s.grpc.Serve(s.listen)
}
