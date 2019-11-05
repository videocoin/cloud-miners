package service

import (
	"context"
	"encoding/json"
	"net"

	"github.com/AlekSi/pointer"
	protoempty "github.com/gogo/protobuf/types"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	v1 "github.com/videocoin/cloud-api/miners/v1"
	"github.com/videocoin/cloud-api/rpc"
	usersv1 "github.com/videocoin/cloud-api/users/v1"
	"github.com/videocoin/cloud-pkg/auth"
	"github.com/videocoin/cloud-pkg/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
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

func (s *RPCServer) Create(ctx context.Context, req *protoempty.Empty) (*v1.MinerResponse, error) {
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

	return &v1.MinerResponse{
		Id:     miner.ID,
		Status: miner.Status,
	}, nil
}

func (s *RPCServer) List(ctx context.Context, req *v1.MinerRequest) (*v1.MinerListResponse, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "List")
	defer span.Finish()

	userID, _, err := s.authenticate(ctx)
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	resp := &v1.MinerListResponse{Items: []*v1.MinerResponse{}}

	miners, err := s.ds.Miners.List(ctx, pointer.ToString(userID))
	if err != nil {
		return nil, err
	}

	for _, miner := range miners {
		resp.Items = append(resp.Items, &v1.MinerResponse{
			Id:     miner.ID,
			Status: miner.Status,
		})
	}

	return resp, nil
}

func (s *RPCServer) Get(ctx context.Context, req *v1.MinerRequest) (*v1.MinerResponse, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Get")
	defer span.Finish()

	span.SetTag("id", req.Id)

	userID, _, err := s.authenticate(ctx)
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	resp := &v1.MinerResponse{}

	miner, err := s.ds.Miners.Get(ctx, req.Id, userID)
	if err != nil {
		return nil, err
	}

	resp.Id = miner.ID
	resp.Status = miner.Status

	return resp, nil
}

func (s *RPCServer) GetByID(ctx context.Context, req *v1.MinerRequest) (*v1.MinerResponse, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GetByID")
	defer span.Finish()

	span.SetTag("id", req.Id)

	resp := &v1.MinerResponse{}

	miner, err := s.ds.Miners.Get(ctx, req.Id, "")
	if err != nil {
		return nil, err
	}

	resp.Id = miner.ID
	resp.Status = miner.Status
	resp.Tags = miner.Tags

	return resp, nil
}

func (s *RPCServer) Ping(ctx context.Context, req *v1.PingRequest) (*v1.PingResponse, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Ping")
	defer span.Finish()

	span.SetTag("client_id", req.ClientID)

	miner, err := s.ds.Miners.Get(ctx, req.ClientID, "")
	if err != nil {
		s.logger.Errorf("failed to get miner: %s", err)
		return nil, err
	}

	err = s.ds.Miners.UpdateLastPingAt(ctx, miner)
	if err != nil {
		s.logger.Errorf("failed to update last ping at: %s", err)
		return nil, err
	}

	sysInfo := map[string]interface{}{}
	err = json.Unmarshal(req.SystemInfo, &sysInfo)
	if err != nil {
		s.logger.Errorf("failed to unmarshal system info: %s", err)
	} else {
		err := s.ds.Miners.UpdateSystemInfo(ctx, miner, sysInfo)
		if err != nil {
			s.logger.Errorf("failed to update system info: %s", err)
			return nil, err
		}
	}

	return &v1.PingResponse{}, nil
}

func (s *RPCServer) AssignTask(ctx context.Context, req *v1.AssignTaskRequest) (*protoempty.Empty, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "AssignTask")
	defer span.Finish()

	span.SetTag("client_id", req.ClientID)

	miner, err := s.ds.Miners.Get(ctx, req.ClientID, "")
	if err != nil {
		s.logger.Errorf("failed to get miner: %s", err)
		return nil, err
	}

	err = s.ds.Miners.UpdateCurrentTask(ctx, miner, req.TaskID, false)
	if err != nil {
		s.logger.Errorf("failed to update current task: %s", err)
		return nil, err
	}

	return &protoempty.Empty{}, nil
}

func (s *RPCServer) UnassignTask(ctx context.Context, req *v1.AssignTaskRequest) (*protoempty.Empty, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "UnassignTask")
	defer span.Finish()

	span.SetTag("client_id", req.ClientID)

	miner, err := s.ds.Miners.Get(ctx, req.ClientID, "")
	if err != nil {
		s.logger.Errorf("failed to get miner: %s", err)
		return nil, err
	}

	err = s.ds.Miners.UpdateCurrentTask(ctx, miner, "", true)
	if err != nil {
		s.logger.Errorf("failed to update current task: %s", err)
		return nil, err
	}

	return &protoempty.Empty{}, nil
}

func (s *RPCServer) SetTags(ctx context.Context, req *v1.SetTagsRequest) (*v1.MinerResponse, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "SetTags")
	defer span.Finish()

	span.SetTag("id", req.Id)

	userID, _, err := s.authenticate(ctx)
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	resp := &v1.MinerResponse{}

	miner, err := s.ds.Miners.Get(ctx, req.Id, userID)
	if err != nil {
		return nil, err
	}

	resp.Id = miner.ID
	resp.Status = miner.Status

	if req.Tags != nil && len(req.Tags) > 0 {
		err = s.ds.Miners.SetTags(ctx, miner, req.Tags)
		if err != nil {
			return nil, err
		}
	}

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

func (s *RPCServer) GetForceTaskList(ctx context.Context, req *protoempty.Empty) (*v1.ForceTaskListResponse, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GetForceTaskList")
	defer span.Finish()

	resp := &v1.ForceTaskListResponse{}

	ids, err := s.ds.Miners.GetForceTaskIDs(ctx)
	if err != nil {
		return nil, err
	}

	resp.Ids = ids

	return resp, nil
}
