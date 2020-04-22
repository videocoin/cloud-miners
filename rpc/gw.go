package rpc

import (
	"context"
	"fmt"

	"github.com/AlekSi/pointer"
	protoempty "github.com/gogo/protobuf/types"
	"github.com/opentracing/opentracing-go"
	v1 "github.com/videocoin/cloud-api/miners/v1"
	"github.com/videocoin/cloud-api/rpc"
	"github.com/videocoin/cloud-miners/datastore"
)

func (s *Server) Create(ctx context.Context, req *protoempty.Empty) (*v1.MinerResponse, error) {
	userID, err := s.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	token, err := s.authToken(ctx)
	if err != nil {
		s.logger.Errorf("failed to extract auth header: %s", err)
		return nil, rpc.ErrRpcInternal
	}

	accessKey, err := s.getSymphonyAccessKey(userID, fmt.Sprintf("Bearer %s", token))
	if err != nil {
		s.logger.Errorf("failed to get symphony access key: %s", err)
		return nil, rpc.ErrRpcInternal
	}

	miner, err := s.ds.Miners.Create(ctx, userID, string(accessKey))
	if err != nil {
		s.logger.Errorf("failed to create miner: %s", err)
		return nil, rpc.ErrRpcInternal
	}

	return toMinerResponse(miner), nil
}

func (s *Server) List(ctx context.Context, req *v1.MinerRequest) (*v1.MinerListResponse, error) {
	userID, err := s.authenticate(ctx)
	if err != nil {
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

func (s *Server) Get(ctx context.Context, req *v1.MinerRequest) (*v1.MinerResponse, error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("id", req.Id)

	userID, err := s.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	miner, err := s.ds.Miners.Get(ctx, req.Id, userID)
	if err != nil {
		if err == datastore.ErrMinerNotFound {
			return nil, rpc.ErrRpcNotFound
		}
		return nil, err
	}

	return toMinerResponse(miner), nil
}

func (s *Server) Update(ctx context.Context, req *v1.UpdateMinerRequest) (*v1.MinerResponse, error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("id", req.Id)
	span.SetTag("name", req.Name)

	userID, err := s.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	if req.Name == "" {
		return nil, rpc.ErrRpcBadRequest
	}

	miner, err := s.ds.Miners.Get(ctx, req.Id, userID)
	if err != nil {
		if err == datastore.ErrMinerNotFound {
			return nil, rpc.ErrRpcNotFound
		}
		return nil, err
	}

	if err := s.ds.Miners.UpdateName(ctx, miner, req.Name); err != nil {
		return nil, err
	}

	return toMinerResponse(miner), nil
}

func (s *Server) Delete(ctx context.Context, req *v1.MinerRequest) (*v1.MinerResponse, error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("id", req.Id)

	userID, err := s.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	miner, err := s.ds.Miners.Get(ctx, req.Id, userID)
	if err != nil {
		if err == datastore.ErrMinerNotFound {
			return nil, rpc.ErrRpcNotFound
		}
		return nil, err
	}

	if miner.Status != v1.MinerStatusOffline && miner.Status != v1.MinerStatusNew {
		return nil, rpc.NewRpcPermissionError("Worker must be offline to delete")
	}

	if err := s.ds.Miners.Delete(ctx, miner.ID); err != nil {
		return nil, err
	}

	return toMinerResponse(miner), nil
}

func (s *Server) SetTags(ctx context.Context, req *v1.SetTagsRequest) (*v1.MinerResponse, error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("id", req.Id)
	span.SetTag("tags", req.Tags)

	_, err := s.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	miner, err := s.ds.Miners.Get(ctx, req.Id, "")
	if err != nil {
		if err == datastore.ErrMinerNotFound {
			return nil, rpc.ErrRpcNotFound
		}

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

func (s *Server) All(ctx context.Context, req *protoempty.Empty) (*v1.MinerListResponse, error) {
	resp := &v1.MinerListResponse{Items: []*v1.MinerResponse{}}

	miners, err := s.ds.Miners.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	for _, miner := range miners {
		resp.Items = append(resp.Items, toMinerResponse(miner))
	}

	return resp, nil
}
