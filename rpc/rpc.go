package rpc

import (
	"context"
	"encoding/json"

	protoempty "github.com/gogo/protobuf/types"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	v1 "github.com/videocoin/cloud-api/miners/v1"
	"github.com/videocoin/cloud-api/rpc"
	"github.com/videocoin/cloud-miners/datastore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) Register(ctx context.Context, req *v1.RegistrationRequest) (*v1.MinerResponse, error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("client_id", req.ClientID)
	span.SetTag("address", req.Address)
	span.SetTag("is_rpi", req.IsRaspberry)
	span.SetTag("is_jetson", req.IsJetson)

	logger := s.logger.WithFields(logrus.Fields{
		"client_id": req.ClientID,
		"address":   req.Address,
	})

	resp := &v1.MinerResponse{}

	ctx = opentracing.ContextWithSpan(context.Background(), span)

	miner, err := s.ds.Miners.Get(ctx, req.ClientID, "")
	if err != nil {
		logger.Errorf("failed to get miner: %s", err)
		return nil, err
	}

	defer func() {
		err = s.ds.Miners.Unlock(ctx, miner)
		if err != nil {
			logger.Errorf("failed to unlock miner: %s", err)
		}
	}()

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

	err = s.ds.Miners.UpdateAddress(ctx, miner, req.Address)
	if err != nil {
		logger.Errorf("failed to update address: %s", err)
		return nil, err
	}

	var tags []*v1.Tag
	if req.IsRaspberry {
		tags = []*v1.Tag{{Key: "hw", Value: "raspberrypi"}}
	} else if req.IsJetson {
		tags = []*v1.Tag{{Key: "hw", Value: "jetson"}}
	} else {
		tags = []*v1.Tag{{Key: "hw", Value: ""}}
	}

	if len(tags) > 0 {
		err = s.ds.Miners.SetTags(ctx, miner, tags)
		if err != nil {
			logger.Errorf("failed to update tags: %s", err)
			return nil, err
		}
	}

	err = s.ds.Miners.MarkMinerAsIdle(ctx, miner)
	if err != nil {
		logger.Errorf("failed to mark miner as idle: %s", err)
		return nil, err
	}

	resp.Id = miner.ID
	resp.Name = miner.Name
	resp.Status = miner.Status
	resp.Tags = miner.Tags
	resp.UserID = miner.UserID

	return resp, nil
}

func (s *Server) Ping(ctx context.Context, req *v1.PingRequest) (*v1.PingResponse, error) {
	span := opentracing.SpanFromContext(ctx)
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
				latitude, longitude, err := GetLatLon(newIP)
				if err != nil {
					logger.WithField("ip", newIP).Errorf("failed to get location by ip: %s", err)
				} else {
					geoInfo := map[string]interface{}{
						"latitude":  latitude,
						"longitude": longitude,
					}

					if err := s.ds.Miners.UpdateGeolocation(ctx, miner, geoInfo); err != nil {
						logger.Errorf("failed to update geolocation: %s", err)
					}

					sysInfo["geo"] = geoInfo
				}
			}
			if hasGeo {
				sysInfo["geo"] = geo
			}
			if err := s.ds.Miners.UpdateSystemInfo(ctx, miner, sysInfo); err != nil {
				logger.Errorf("failed to update system info: %s", err)
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

func (s *Server) GetForceTaskList(ctx context.Context, req *protoempty.Empty) (*v1.ForceTaskListResponse, error) {
	ids, err := s.ds.Miners.GetForceTaskIDs(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.ForceTaskListResponse{Ids: ids}, nil
}

func (s *Server) GetByID(ctx context.Context, req *v1.MinerRequest) (*v1.MinerResponse, error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("id", req.Id)

	miner, err := s.ds.Miners.Get(ctx, req.Id, "")
	if err != nil {
		if err == datastore.ErrMinerNotFound {
			return nil, rpc.ErrRpcNotFound
		}
		return nil, rpc.NewRpcInternalError(err)
	}

	resp := &v1.MinerResponse{
		Id:         miner.ID,
		Status:     miner.Status,
		Tags:       miner.Tags,
		Name:       miner.Name,
		SystemInfo: &v1.SystemInfo{},
		UserID:     miner.UserID,
		IsBlock:    miner.IsBlock,
		Address:    miner.Address.String,
		IsInternal: miner.IsInternal,
	}

	if miner.SystemInfo != nil {
		if hw, ok := miner.SystemInfo["hw"]; ok {
			resp.SystemInfo.Hw = hw.(string)
		}
	}

	return resp, nil
}

func (s *Server) AssignTask(ctx context.Context, req *v1.AssignTaskRequest) (*protoempty.Empty, error) {
	span := opentracing.SpanFromContext(ctx)
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

func (s *Server) UnassignTask(ctx context.Context, req *v1.AssignTaskRequest) (*protoempty.Empty, error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("client_id", req.ClientID)

	logger := s.logger.WithFields(logrus.Fields{
		"client_id": req.ClientID,
		"task_id":   req.TaskID,
	})
	logger.Info("unassigning task")

	if req.ClientID != "" {
		miner, err := s.ds.Miners.Get(ctx, req.ClientID, "")
		if err != nil {
			logger.Errorf("failed to get miner: %s", err)
			return nil, err
		}

		if err := s.ds.Miners.UpdateCurrentTask(ctx, miner, "", true); err != nil {
			logger.Errorf("failed to update current task: %s", err)
			return nil, err
		}
	} else {
		if req.TaskID != "" {
			miners, err := s.ds.Miners.ListByTag(ctx, "force_task_id", req.TaskID)
			if err != nil {
				logger.Errorf("failed to list miners by tag: %s", err)
				return nil, err
			}
			for _, miner := range miners {
				if err := s.ds.Miners.UpdateCurrentTask(ctx, miner, "", true); err != nil {
					logger.Errorf("failed to update current task: %s", err)
					return nil, err
				}
			}
		}
	}

	return &protoempty.Empty{}, nil
}

func (s *Server) GetMinersWithForceTask(ctx context.Context, req *protoempty.Empty) (*v1.MinersWithForceTaskResponse, error) {
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

func (s *Server) GetMinersCandidates(ctx context.Context, req *v1.MinersCandidatesRequest) (*v1.MinersCandidatesResponse, error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("encode_capacity", req.EncodeCapacity)
	span.SetTag("cpu_capacity", req.CpuCapacity)

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
			ID:         miner.ID,
			Stake:      m.TotalStake,
			IsInternal: miner.IsInternal,
		})
	}

	return resp, nil
}

func (s *Server) GetKey(ctx context.Context, req *v1.KeyRequest) (*v1.KeyResponse, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GetKey")
	defer span.Finish()

	span.SetTag("id", req.ClientID)

	miner, err := s.ds.Miners.Get(ctx, req.ClientID, "")
	if err != nil {
		s.logger.Errorf("failed to get miner: %s", err)
		return nil, err
	}

	return &v1.KeyResponse{
		Key: miner.AccessKey,
	}, nil
}

func (s *Server) GetInternalMiner(ctx context.Context, req *v1.InternalMinerRequest) (*v1.InternalMinerResponse, error) {
	miner, err := s.ds.Miners.GetInternal(ctx)
	if err != nil {
		s.logger.WithError(err).Error("failed to get internal miner")
		return nil, err
	}

	resp := &v1.InternalMinerResponse{
		ID:     miner.ID,
		Key:    miner.Key.String,
		Secret: miner.Secret.String,
	}
	if ft, ok := miner.Tags["force_task_id"]; ok {
		if len(ft) > 0 {
			resp.TaskID = ft
		}
	}

	return resp, nil
}
