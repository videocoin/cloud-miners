package rpc

import (
	"context"
	"math"

	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/opentracing/opentracing-go"
	emitterv1 "github.com/videocoin/cloud-api/emitter/v1"
	v1 "github.com/videocoin/cloud-api/miners/v1"
	"github.com/videocoin/cloud-api/rpc"
	usersv1 "github.com/videocoin/cloud-api/users/v1"
	"github.com/videocoin/cloud-miners/datastore"
	"github.com/videocoin/cloud-pkg/auth"
	"github.com/videocoin/cloud-pkg/ethutils"
)

func (s *Server) authToken(ctx context.Context) (string, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "rpc.authToken")
	defer span.Finish()

	return grpcauth.AuthFromMD(ctx, "bearer")
}

func (s *Server) authenticate(ctx context.Context) (string, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "rpc.authenticate")
	defer span.Finish()

	ctx = auth.NewContextWithSecretKey(ctx, s.authTokenSecret)
	ctx, _, err := auth.AuthFromContext(ctx)
	if err != nil {
		return "", rpc.ErrRpcUnauthenticated
	}

	if s.getTokenType(ctx) == auth.TokenType(usersv1.TokenTypeAPI) {
		return "", rpc.ErrRpcPermissionDenied
	}

	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return "", rpc.ErrRpcUnauthenticated
	}

	return userID, nil
}

func (s *Server) getTokenType(ctx context.Context) auth.TokenType {
	tokenType, ok := auth.TypeFromContext(ctx)
	if !ok {
		return auth.TokenType(usersv1.TokenTypeRegular)
	}

	return tokenType
}

func toMinerResponse(miner *datastore.Miner) *v1.MinerResponse {
	systemInfo := &v1.SystemInfo{}

	if cpuInfo, ok := miner.SystemInfo["cpu"]; ok {
		if cpuInfo != nil {
			systemInfo.CpuCores = cpuInfo.(map[string]interface{})["cores"].(float64)
			systemInfo.CpuFreq = cpuInfo.(map[string]interface{})["freq"].(float64)
		}
	}

	if cpuUsage, ok := miner.SystemInfo["cpu_usage"]; ok {
		if cpuUsage != nil {
			systemInfo.CpuUsage = math.Round(cpuUsage.(float64)*100) / 100
		}
	}

	if memInfo, ok := miner.SystemInfo["memory"]; ok {
		if memInfo != nil {
			systemInfo.MemUsage = memInfo.(map[string]interface{})["used"].(float64)
			systemInfo.MemTotal = memInfo.(map[string]interface{})["total"].(float64)
		}
	}

	if geoInfo, ok := miner.SystemInfo["geo"]; ok {
		if geoInfo != nil {
			systemInfo.Latitude = geoInfo.(map[string]interface{})["latitude"].(float64)
			systemInfo.Longitude = geoInfo.(map[string]interface{})["longitude"].(float64)
		}
	}

	capacityInfo := &v1.CapacityInfo{}
	if value, ok := miner.CapacityInfo["encode"]; ok {
		capacityInfo.Encode = value.(float64)
	}
	if value, ok := miner.CapacityInfo["cpu"]; ok {
		capacityInfo.Cpu = value.(float64)
	}

	var totalStake, delegatedStake, selfStake float64
	if miner.WorkerInfo != nil {
		if miner.WorkerInfo.TotalStake != "" {
			wei, err := ethutils.ParseBigInt(miner.WorkerInfo.TotalStake)
			if err == nil {
				vid, e := ethutils.WeiToEth(&wei)
				if e == nil {
					totalStake, _ = vid.Float64()
				}
			}
		}

		if miner.WorkerInfo.SelfStake != "" {
			wei, err := ethutils.ParseBigInt(miner.WorkerInfo.SelfStake)
			if err == nil {
				vid, e := ethutils.WeiToEth(&wei)
				if e == nil {
					selfStake, _ = vid.Float64()
				}
			}
		}

		if miner.WorkerInfo.DelegatedStake != "" {
			wei, err := ethutils.ParseBigInt(miner.WorkerInfo.DelegatedStake)
			if err == nil {
				vid, e := ethutils.WeiToEth(&wei)
				if e == nil {
					delegatedStake, _ = vid.Float64()
				}
			}
		}
	}

	workerState := emitterv1.WorkerStateBonding
	if miner.WorkerInfo != nil {
		workerState = miner.WorkerInfo.State
	}

	return &v1.MinerResponse{
		Id:                       miner.ID,
		Name:                     miner.Name,
		Status:                   miner.Status,
		SystemInfo:               systemInfo,
		CapacityInfo:             capacityInfo,
		UserID:                   miner.UserID,
		Address:                  miner.Address.String,
		TotalStake:               totalStake,
		DelegatedStake:           delegatedStake,
		SelfStake:                selfStake,
		Reward:                   miner.Reward,
		IsBlock:                  miner.IsBlock,
		IsInternal:               miner.IsInternal,
		WorkerState:              workerState,
		OrgName:                  miner.OrgName.String,
		OrgEmail:                 miner.OrgEmail.String,
		OrgDesc:                  miner.OrgDesc.String,
		AllowThirdpartyDelegates: miner.AllowThirdpartyDelegates,
		DelegatePolicy:           miner.DelegatePolicy.String,
	}
}
