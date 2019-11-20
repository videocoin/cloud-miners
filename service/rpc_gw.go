package service

import (
	"context"

	"github.com/AlekSi/pointer"
	protoempty "github.com/gogo/protobuf/types"
	"github.com/opentracing/opentracing-go"
	v1 "github.com/videocoin/cloud-api/miners/v1"
	"github.com/videocoin/cloud-api/rpc"
)

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

	return toMinerResponse(miner), nil
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
		resp.Items = append(resp.Items, toMinerResponse(miner))
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

	miner, err := s.ds.Miners.Get(ctx, req.Id, userID)
	if err != nil {
		if err == ErrMinerNotFound {
			return nil, rpc.ErrRpcNotFound
		}
		return nil, err
	}

	return toMinerResponse(miner), nil
}

func (s *RPCServer) Update(ctx context.Context, req *v1.UpdateMinerRequest) (*v1.MinerResponse, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Update")
	defer span.Finish()

	span.SetTag("id", req.Id)

	userID, _, err := s.authenticate(ctx)
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	if req.Name == "" {
		return nil, rpc.ErrRpcBadRequest
	}

	miner, err := s.ds.Miners.Get(ctx, req.Id, userID)
	if err != nil {
		if err == ErrMinerNotFound {
			return nil, rpc.ErrRpcNotFound
		}
		return nil, err
	}

	err = s.ds.Miners.UpdateName(ctx, miner, req.Name)
	if err != nil {
		return nil, err
	}

	return toMinerResponse(miner), nil
}

func (s *RPCServer) Delete(ctx context.Context, req *v1.MinerRequest) (*v1.MinerResponse, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Delete")
	defer span.Finish()

	span.SetTag("id", req.Id)

	userID, _, err := s.authenticate(ctx)
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	miner, err := s.ds.Miners.Get(ctx, req.Id, userID)
	if err != nil {
		if err == ErrMinerNotFound {
			return nil, rpc.ErrRpcNotFound
		}
		return nil, err
	}

	if miner.Status != v1.MinerStatusOffline && miner.Status != v1.MinerStatusNew {
		return nil, rpc.NewRpcPermissionError("Worker must be offline to delete")
	}

	err = s.ds.Miners.Delete(ctx, miner.ID)
	if err != nil {
		return nil, err
	}

	return toMinerResponse(miner), nil
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

	miner, err := s.ds.Miners.Get(ctx, req.Id, userID)
	if err != nil {
		return nil, err
	}

	if req.Tags != nil && len(req.Tags) > 0 {
		err = s.ds.Miners.SetTags(ctx, miner, req.Tags)
		if err != nil {
			return nil, err
		}
	}

	return toMinerResponse(miner), nil
}
