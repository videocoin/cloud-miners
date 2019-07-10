package service

import (
	"context"
	"net"

	protoempty "github.com/gogo/protobuf/types"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	v1 "github.com/videocoin/cloud-api/miners/v1"
	"github.com/videocoin/cloud-api/rpc"
	usersv1 "github.com/videocoin/cloud-api/users/v1"
	"github.com/videocoin/cloud-pkg/auth"
	"github.com/videocoin/cloud-pkg/grpcutil"
	"google.golang.org/grpc"
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

	authTokenSecret string
}

func NewRPCServer(opts *RPCServerOptions, ds *Datastore) (*RPCServer, error) {
	grpcOpts := grpcutil.DefaultServerOpts(opts.Logger)
	grpcServer := grpc.NewServer(grpcOpts...)

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

func (s *RPCServer) Health(ctx context.Context, req *protoempty.Empty) (*rpc.HealthStatus, error) {
	return &rpc.HealthStatus{Status: "OK"}, nil
}

func (s *RPCServer) Create(ctx context.Context, req *protoempty.Empty) (*v1.CreateResponse, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Create")
	defer span.Finish()

	userId, _, err := s.authenticate(ctx)
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	miner, err := s.ds.Miners.Create(ctx, userId)
	if err != nil {
		s.logger.Errorf("failed to create miner: %s", err)
		return nil, rpc.ErrRpcInternal
	}

	return &v1.CreateResponse{
		Id:      miner.Id,
		KeyHash: GetMD5Hash(miner.Key),
	}, nil
}

func (s *RPCServer) List(ctx context.Context, req *v1.ListRequest) (*v1.ListResponse, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "List")
	defer span.Finish()

	resp := &v1.ListResponse{Items: []*v1.Miner{}}
	miners, err := s.ds.Miners.List(ctx, req.IsBusy)
	if err != nil {
		return nil, err
	}

	for _, miner := range miners {
		resp.Items = append(resp.Items, &v1.Miner{
			Id:       miner.Id,
			Hostname: miner.Hostname,
			IsBusy:   miner.IsBusy,
			CpuIdle:  miner.CpuIdle,
		})
	}

	return resp, nil
}

func (s *RPCServer) Get(ctx context.Context, req *v1.Request) (*v1.Response, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Get")
	defer span.Finish()

	span.SetTag("id", req.Id)

	resp := &v1.Response{}

	miner, err := s.ds.Miners.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	resp.Id = miner.Id
	resp.IsBusy = miner.IsBusy
	resp.Hostname = miner.Hostname
	resp.CpuIdle = miner.CpuIdle

	return resp, nil
}

func (s *RPCServer) MarkAsBusy(ctx context.Context, req *v1.Request) (*v1.Response, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "MarkAsBusy")
	defer span.Finish()

	span.SetTag("id", req.Id)

	resp := &v1.Response{}

	miner, err := s.ds.Miners.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	err = s.ds.Miners.MarkAsBusy(ctx, miner)
	if err != nil {
		return nil, err
	}

	resp.Id = miner.Id
	resp.IsBusy = miner.IsBusy
	resp.Hostname = miner.Hostname
	resp.CpuIdle = miner.CpuIdle

	return resp, nil
}

func (s *RPCServer) MarkAsIdle(ctx context.Context, req *v1.Request) (*v1.Response, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "MarkAsIdle")
	defer span.Finish()

	span.SetTag("id", req.Id)

	resp := &v1.Response{}

	miner, err := s.ds.Miners.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	err = s.ds.Miners.MarkAsIdle(ctx, miner)
	if err != nil {
		return nil, err
	}

	resp.Id = miner.Id
	resp.IsBusy = miner.IsBusy
	resp.Hostname = miner.Hostname
	resp.CpuIdle = miner.CpuIdle

	return resp, nil
}

func (s *RPCServer) authenticate(ctx context.Context) (string, context.Context, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "authenticate")
	defer span.Finish()

	ctx = auth.NewContextWithSecretKey(ctx, s.authTokenSecret)
	ctx, err := auth.AuthFromContext(ctx)
	if err != nil {
		return "", ctx, rpc.ErrRpcUnauthenticated
	}

	if s.getTokenType(ctx) == auth.TokenType(usersv1.TokenTypeAPI) {
		return "", nil, rpc.ErrRpcPermissionDenied
	}

	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return "", ctx, rpc.ErrRpcUnauthenticated
	}

	return userID, ctx, nil
}

func (s *RPCServer) getTokenType(ctx context.Context) auth.TokenType {
	tokenType, ok := auth.TypeFromContext(ctx)
	if !ok {
		return auth.TokenType(usersv1.TokenTypeRegular)
	}

	return tokenType
}
