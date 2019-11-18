package service

import (
	"context"

	"github.com/opentracing/opentracing-go"
	v1 "github.com/videocoin/cloud-api/miners/v1"
	"github.com/videocoin/cloud-api/rpc"
	usersv1 "github.com/videocoin/cloud-api/users/v1"
	"github.com/videocoin/cloud-pkg/auth"
)

func (s *RPCServer) authenticate(ctx context.Context) (string, context.Context, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "authenticate")
	defer span.Finish()

	ctx = auth.NewContextWithSecretKey(ctx, s.authTokenSecret)
	ctx, _, err := auth.AuthFromContext(ctx)
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

func toMinerResponse(miner *Miner) *v1.MinerResponse {
	info := &v1.SystemInfo{}

	if cpuInfo, ok := miner.SystemInfo["cpu"]; ok {
		info.CpuCores = int32(cpuInfo.(map[string]interface{})["cores"].(float64))
		info.CpuFreq = int32(cpuInfo.(map[string]interface{})["freq"].(float64))
	}

	if memInfo, ok := miner.SystemInfo["memory"]; ok {
		info.MemUsage = float32(memInfo.(map[string]interface{})["used"].(float64))
		info.MemTotal = float32(memInfo.(map[string]interface{})["total"].(float64))
	}

	return &v1.MinerResponse{
		Id:         miner.ID,
		Name:       miner.Name,
		Status:     miner.Status,
		SystemInfo: info,
	}
}
