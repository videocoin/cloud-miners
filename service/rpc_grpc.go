package service

import (
	"context"
	"encoding/json"

	protoempty "github.com/gogo/protobuf/types"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	v1 "github.com/videocoin/cloud-api/miners/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *RPCServer) Register(ctx context.Context, req *v1.RegistrationRequest) (*v1.MinerResponse, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Register")
	defer span.Finish()

	span.SetTag("client_id", req.ClientID)
	span.SetTag("address", req.Address)

	logger := s.logger.WithFields(logrus.Fields{
		"client_id": req.ClientID,
		"address":   req.Address,
	})

	resp := &v1.MinerResponse{}

	miner, err := s.ds.Miners.Get(ctx, req.ClientID, "")
	if err != nil {
		logger.Errorf("failed to get miner: %s", err)
		return nil, err
	}

	err = s.ds.Miners.UpdateAddress(ctx, miner, req.Address)
	if err != nil {
		logger.Errorf("failed to update address: %s", err)
		return nil, err
	}

	err = s.eb.EmitAssignMinerAddress(ctx, miner.UserID, miner.Address.String)
	if err != nil {
		logger.Errorf("failed to miner created: %s", err)
		return nil, err
	}

	logger.Infof("miner status is %s", miner.Status.String())

	if miner.Status == v1.MinerStatusIdle || miner.Status == v1.MinerStatusBusy {
		logger.Warningf("miner is already running")
		return nil, status.Errorf(codes.AlreadyExists, "miner is already running")
	}

	minerList, err := s.ds.Miners.ListByAddress(ctx, req.Address)
	if err != nil {
		logger.Errorf("failed to list by address: %s", err)
		return nil, err
	}

	for _, m := range minerList {
		if m.Status == v1.MinerStatusIdle || m.Status == v1.MinerStatusBusy {
			logger.Warningf("miner is already running")
			return nil, status.Errorf(codes.AlreadyExists, "miner is already running")
		}
	}

	resp.Id = miner.ID
	resp.Name = miner.Name
	resp.Status = miner.Status
	resp.Tags = miner.Tags
	resp.UserID = miner.UserID

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

	if err := s.ds.Miners.UpdateLastPingAt(ctx, miner); err != nil {
		s.logger.Errorf("failed to update last ping at: %s", err)
		return nil, err
	}
	go func(logger *logrus.Entry, req *v1.PingRequest) {
		sysInfo := map[string]interface{}{}
		if err := json.Unmarshal(req.SystemInfo, &sysInfo); err != nil {
			logger.Errorf("failed to unmarshal system info: %s", err)
		} else {
			geo, hasGeo := miner.SystemInfo["geo"]
			currentIP := miner.SystemInfo["ip"]
			newIP, _ := sysInfo["ip"].(string)
			if currentIP != newIP || !hasGeo {
				latitude, longitude, err := GetGeoLocation(newIP)
				if err != nil {
					logger.WithField("ip", newIP).Errorf("Failed to get ip geolocation: %s", err)
				} else {
					geoInfo := map[string]interface{}{
						"latitude":  latitude,
						"longitude": longitude,
					}

					if err := s.ds.Miners.UpdateGeolocation(ctx, miner, geoInfo); err != nil {
						logger.Errorf("failed to update geolocation: %s", err)
					}
				}
			}
			if hasGeo {
				sysInfo["geo"] = geo
			}
			if err := s.ds.Miners.UpdateSystemInfo(ctx, miner, sysInfo); err != nil {
				logger.Errorf("failed to update system info: %s", err)
			}
		}

		if req.CryptoInfo != nil && len(req.CryptoInfo) > 0 {
			cryptoInfo := map[string]interface{}{}
			if err := json.Unmarshal(req.CryptoInfo, &cryptoInfo); err != nil {
				logger.Errorf("failed to unmarshal crypto info: %s", err)
			}
			if err := s.ds.Miners.UpdateCryptoInfo(ctx, miner, cryptoInfo); err != nil {
				logger.Errorf("failed to update crypto info: %s", err)
			}
		}

		if req.CapacityInfo != nil && len(req.CapacityInfo) > 0 {
			capacityInfo := map[string]interface{}{}
			if err := json.Unmarshal(req.CapacityInfo, &capacityInfo); err != nil {
				logger.Errorf("failed to unmarshal capacity info: %s", err)
			}

			if err := s.ds.Miners.UpdateCapacityInfo(ctx, miner, capacityInfo); err != nil {
				logger.Errorf("failed to update capacity info: %s", err)
			}
		}
	}(s.logger, req)

	return &v1.PingResponse{}, nil
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

func (s *RPCServer) GetByID(ctx context.Context, req *v1.MinerRequest) (*v1.MinerResponse, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GetByID")
	defer span.Finish()

	span.SetTag("id", req.Id)

	resp := &v1.MinerResponse{}

	miner, err := s.ds.Miners.Get(ctx, req.Id, "")
	if err != nil {
		s.logger.Errorf("failed to get miner: %s", err)
		return nil, err
	}

	resp.Id = miner.ID
	resp.Status = miner.Status
	resp.Tags = miner.Tags
	resp.Name = miner.Name
	resp.SystemInfo = &v1.SystemInfo{}
	resp.UserID = miner.UserID

	if miner.SystemInfo != nil {
		if hw, ok := miner.SystemInfo["hw"]; ok {
			resp.SystemInfo.Hw = hw.(string)
		}
	}

	return resp, nil
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

	if err := s.ds.Miners.UpdateCurrentTask(ctx, miner, req.TaskID, false); err != nil {
		s.logger.Errorf("failed to update current task: %s", err)
		return nil, err
	}

	return &protoempty.Empty{}, nil
}

func (s *RPCServer) UnassignTask(ctx context.Context, req *v1.AssignTaskRequest) (*protoempty.Empty, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "UnassignTask")
	defer span.Finish()

	span.SetTag("client_id", req.ClientID)

	if req.ClientID != "" {
		miner, err := s.ds.Miners.Get(ctx, req.ClientID, "")
		if err != nil {
			s.logger.Errorf("failed to get miner: %s", err)
			return nil, err
		}

		if err := s.ds.Miners.UpdateCurrentTask(ctx, miner, "", true); err != nil {
			s.logger.Errorf("failed to update current task: %s", err)
			return nil, err
		}
	} else {
		if req.TaskID != "" {
			miners, err := s.ds.Miners.ListByTag(ctx, "force_task_id", req.TaskID)
			if err != nil {
				s.logger.Errorf("failed to list miners by tag: %s", err)
				return nil, err
			}
			for _, miner := range miners {
				if err := s.ds.Miners.UpdateCurrentTask(ctx, miner, "", true); err != nil {
					s.logger.Errorf("failed to update current task: %s", err)
					return nil, err
				}
			}
		}
	}

	return &protoempty.Empty{}, nil
}

func (s *RPCServer) GetMinersWithForceTask(ctx context.Context, req *protoempty.Empty) (*v1.MinersWithForceTaskResponse, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GetMinersWithForceTask")
	defer span.Finish()

	resp := &v1.MinersWithForceTaskResponse{
		Items: []*v1.MinerWithForceTaskResponse{},
	}

	miners, err := s.ds.Miners.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	for _, miner := range miners {
		if ft, ok := miner.Tags["force_task_id"]; ok {
			if len(ft) > 0 {
				resp.Items = append(resp.Items, &v1.MinerWithForceTaskResponse{
					Id:     miner.ID,
					TaskId: ft,
				})
			}
		}
	}

	return resp, nil
}

func (s *RPCServer) GetMinersCandidates(ctx context.Context, req *v1.MinersCandidatesRequest) (*v1.MinersCandidatesResponse, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GetMinersCandidates")
	defer span.Finish()

	miners, err := s.ds.Miners.ListCandidates(ctx, req.EncodeCapacity, req.CpuCapacity)
	if err != nil {
		return nil, err
	}

	resp := &v1.MinersCandidatesResponse{
		Items: []*v1.MinerCandidateResponse{},
	}

	for _, miner := range miners {
		m := toMinerResponse(miner)
		resp.Items = append(resp.Items, &v1.MinerCandidateResponse{
			ID:    m.Id,
			Stake: int32(toVid(m.CryptoInfo.Stake)),
		})
	}

	return resp, nil
}
