package rpc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"math"
	"net/http"

	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/opentracing/opentracing-go"
	v1 "github.com/videocoin/cloud-api/miners/v1"
	"github.com/videocoin/cloud-api/rpc"
	usersv1 "github.com/videocoin/cloud-api/users/v1"
	"github.com/videocoin/cloud-miners/datastore"
	"github.com/videocoin/cloud-pkg/auth"
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

func (s *Server) getSymphonyAccessKey(userID, authHeader string) ([]byte, error) {
	type Key struct {
		Type         string `json:"type"`
		ClientID     string `json:"client_id"`
		PrivateKeyID string `json:"private_key_id"`
		PrivateKey   string `json:"private_key"`
	}

	key := &Key{
		Type:     "service_account",
		ClientID: userID,
	}

	req, err := http.NewRequest("POST", s.iamEndpoint+"/v1/keys", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", authHeader)
	req.Header.Add("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	values := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&values); err != nil {
		return nil, err
	}

	privateKeyData := values["private_key_data"].(string)
	privateKeyID := values["id"].(string)
	pemBytes, err := base64.StdEncoding.DecodeString(privateKeyData)
	if err != nil {
		return nil, err
	}

	key.PrivateKey = string(pemBytes)
	key.PrivateKeyID = privateKeyID

	return json.Marshal(key)
}

func toMinerResponse(miner *datastore.Miner) *v1.MinerResponse {
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

	capacityInfo := &v1.CapacityInfo{}
	if value, ok := miner.CapacityInfo["encode"]; ok {
		capacityInfo.Encode = value.(float64)
	}
	if value, ok := miner.CapacityInfo["cpu"]; ok {
		capacityInfo.Cpu = value.(float64)
	}

	return &v1.MinerResponse{
		Id:           miner.ID,
		Name:         miner.Name,
		Status:       miner.Status,
		SystemInfo:   systemInfo,
		CapacityInfo: capacityInfo,
		UserID:       miner.UserID,
		Address:      miner.Address.String,
	}
}
