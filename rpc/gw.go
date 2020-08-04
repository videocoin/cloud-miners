package rpc

import (
	"context"
	"github.com/mailru/dbr"

	"github.com/AlekSi/pointer"
	protoempty "github.com/gogo/protobuf/types"
	"github.com/opentracing/opentracing-go"
	v1 "github.com/videocoin/cloud-api/miners/v1"
	"github.com/videocoin/cloud-api/rpc"
	"github.com/videocoin/cloud-miners/datastore"
)

func (s *Server) Create(ctx context.Context, req *v1.CreateMinerRequest) (*v1.MinerResponse, error) {
	userID, err := s.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	token, err := s.authToken(ctx)
	if err != nil {
		s.logger.Errorf("failed to extract auth header: %s", err)
		return nil, rpc.ErrRpcInternal
	}

	sa, err := s.iam.CreateServiceAccountJSON(token, userID)
	if err != nil {
		s.logger.WithError(err).Error("failed to create symphony service account")
		return nil, rpc.ErrRpcInternal
	}

	miner, err := s.ds.Miners.Create(ctx, userID, string(sa), req.K, req.S)
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
	span.SetTag("org_name", req.OrgName)
	span.SetTag("org_email", req.OrgEmail)
	span.SetTag("org_desc", req.OrgDesc)
	span.SetTag("allow_delegates", req.AllowThirdpartyDelegates)
	span.SetTag("delegate_policy", req.DelegatePolicy)

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

	updates := map[string]interface{}{
		"name": req.Name,
		"org_name": dbr.NewNullString(req.OrgName),
		"org_email": dbr.NewNullString(req.OrgEmail),
		"org_desc": dbr.NewNullString(req.OrgDesc),
		"allow_thirdparty_delegates": req.AllowThirdpartyDelegates,
		"delegate_policy": dbr.NewNullString(req.DelegatePolicy),
	}

	if err := s.ds.Miners.Update(ctx, miner, updates); err != nil {
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

func (s *Server) GetMiner(ctx context.Context, req *v1.MinerRequest) (*v1.MinerResponse, error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("id", req.Id)

	miner, err := s.ds.Miners.Get(ctx, req.Id, "")
	if err != nil {
		if err == datastore.ErrMinerNotFound {
			return nil, rpc.ErrRpcNotFound
		}
		return nil, err
	}

	return toMinerResponse(miner), nil
}