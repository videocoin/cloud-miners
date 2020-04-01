package service

import (
	"context"
	"math"
	"math/big"

	"github.com/opentracing/opentracing-go"
	v1 "github.com/videocoin/cloud-api/miners/v1"
	"github.com/videocoin/cloud-api/rpc"
	usersv1 "github.com/videocoin/cloud-api/users/v1"
	"github.com/videocoin/cloud-pkg/auth"
	"github.com/videocoin/cloud-pkg/ethutils"
)

func (s *RPCServer) authenticate(ctx context.Context) (string, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "authenticate")
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

func (s *RPCServer) getTokenType(ctx context.Context) auth.TokenType {
	tokenType, ok := auth.TypeFromContext(ctx)
	if !ok {
		return auth.TokenType(usersv1.TokenTypeRegular)
	}

	return tokenType
}

func toMinerResponse(miner *Miner) *v1.MinerResponse {
	systemInfo := &v1.SystemInfo{}

	if cpuInfo, ok := miner.SystemInfo["cpu"]; ok {
		systemInfo.CpuCores = cpuInfo.(map[string]interface{})["cores"].(float64)
		systemInfo.CpuFreq = cpuInfo.(map[string]interface{})["freq"].(float64)
	}

	if cpuUsage, ok := miner.SystemInfo["cpu_usage"]; ok {
		systemInfo.CpuUsage = math.Round(cpuUsage.(float64)*100) / 100
	}

	if memInfo, ok := miner.SystemInfo["memory"]; ok {
		systemInfo.MemUsage = memInfo.(map[string]interface{})["used"].(float64)
		systemInfo.MemTotal = memInfo.(map[string]interface{})["total"].(float64)
	}

	if geoInfo, ok := miner.SystemInfo["geo"]; ok {
		systemInfo.Latitude = geoInfo.(map[string]interface{})["latitude"].(float64)
		systemInfo.Longitude = geoInfo.(map[string]interface{})["longitude"].(float64)
	}

	cryptoInfo := &v1.CryptoInfo{}
	if address, ok := miner.CryptoInfo["address"]; ok {
		cryptoInfo.Address = address.(string)
	}

	if balance, ok := miner.CryptoInfo["balance"]; ok {
		cryptoInfo.Balance = balance.(string)
	}

	if selfStake, ok := miner.CryptoInfo["stake"]; ok {
		cryptoInfo.Stake = selfStake.(string)
	}

	capacityInfo := &v1.CapacityInfo{}
	if value, ok := miner.CapacityInfo["encode"]; ok {
		capacityInfo.Encode = value.(int32)
	}
	if value, ok := miner.CapacityInfo["cpu"]; ok {
		capacityInfo.Cpu = value.(int32)
	}

	return &v1.MinerResponse{
		Id:           miner.ID,
		Name:         miner.Name,
		Status:       miner.Status,
		SystemInfo:   systemInfo,
		CryptoInfo:   cryptoInfo,
		CapacityInfo: capacityInfo,
	}
}

func toVid(amountWei string) float64 {
	amount := new(big.Int)
	amount, _ = amount.SetString(amountWei, 10)
	v, _ := ethutils.WeiToEth(amount)
	fv, _ := v.Float64()
	return fv
}
