package manager

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	emitterv1 "github.com/videocoin/cloud-api/emitter/v1"
	"github.com/videocoin/cloud-miners/datastore"
)

type Manager struct {
	logger         *logrus.Entry
	offlineTimeout time.Duration
	offlineTicker  *time.Ticker
	wiTicker       *time.Ticker
	wrTicker       *time.Ticker
	ds             *datastore.Datastore
	emitter        emitterv1.EmitterServiceClient
}

func New(opts ...Option) (*Manager, error) {
	offlineTimeout := time.Second * 10
	ds := &Manager{
		offlineTimeout: offlineTimeout,
		offlineTicker:  time.NewTicker(offlineTimeout),
		wiTicker:       time.NewTicker(time.Second * 10),
		wrTicker:       time.NewTicker(time.Second * 30),
	}
	for _, o := range opts {
		if err := o(ds); err != nil {
			return nil, err
		}
	}

	return ds, nil
}

func (m *Manager) Start() {
	go m.checkOffline()
	go m.checkStuckMiners()
	go m.updateWorkerInfo()
	go m.updateWorkerReward()
}

func (m *Manager) Stop() {
	m.offlineTicker.Stop()
	m.wiTicker.Stop()
	m.wrTicker.Stop()
}

func (m *Manager) checkOffline() {
	for range m.offlineTicker.C {
		ctx := context.Background()
		err := m.ds.Miners.MarkAsOffline(ctx, m.offlineTimeout)
		if err != nil {
			m.logger.Errorf("failed to mark miners as offline: %s", err)
			continue
		}

		miners, err := m.ds.Miners.GetStuckMinerList(ctx, m.offlineTimeout)
		if err != nil {
			m.logger.Errorf("failed to get stuck miners: %s", err)
			continue
		}

		if len(miners) > 0 {
			for _, miner := range miners {
				err := m.ds.Miners.MarkMinerAsOffline(ctx, miner)
				if err != nil {
					m.logger.Errorf("failed to mark miner as offline: %s", err)
					continue
				}
			}
		}
	}
}

func (m *Manager) updateWorkerInfo() {
	for range m.wiTicker.C {
		emptyCtx := context.Background()
		// miners, err := m.ds.Miners.ListByOnline(emptyCtx)
		miners, err := m.ds.Miners.List(emptyCtx, nil)
		if err != nil {
			m.logger.Infof("failed to list workers by online: %s", err)
			continue
		}

		for _, miner := range miners {
			if miner.IsInternal {
				continue
			}

			logger := m.logger.WithField("miner_id", miner.ID)
			if miner.Address.String != "" {
				workerReq := &emitterv1.WorkerRequest{Address: miner.Address.String}
				worker, err := m.emitter.GetWorker(emptyCtx, workerReq)
				if err != nil {
					logger.Infof("failed to get worker: %s", err)
					continue
				}
				err = m.ds.Miners.UpdateWorkerInfoByAddress(emptyCtx, worker.Address, worker)
				if err != nil {
					logger.
						WithField("address", worker.Address).
						Errorf("failed to update worker info: %s", err)
					continue
				}
			}
		}
	}
}

func (m *Manager) checkStuckMiners() {
	for range m.offlineTicker.C {
		ctx := context.Background()
		miners, err := m.ds.Miners.GetStuckOfflineMinerList(ctx, m.offlineTimeout)
		if err != nil {
			m.logger.Errorf("failed to get stuck offline miners: %s", err)
			continue
		}

		if len(miners) > 0 {
			for _, miner := range miners {
				logger := m.logger.WithField("miner_id", miner.ID)
				err := m.ds.Miners.UpdateCurrentTask(ctx, miner, "", false)
				if err != nil {
					logger.WithError(err).Error("failed to clear current task")
					continue
				}
			}
		}
	}
}

func (m *Manager) updateWorkerReward() {
	for range m.wrTicker.C {
		emptyCtx := context.Background()
		miners, err := m.ds.Miners.List(emptyCtx, nil)
		if err != nil {
			m.logger.Infof("failed to list workers: %s", err)
			continue
		}

		for _, miner := range miners {
			logger := m.logger.WithField("miner_id", miner.ID)
			if miner.Address.String != "" {
				rewardReq := &emitterv1.RewardRequest{Address: miner.Address.String}
				reward, err := m.emitter.GetReward(emptyCtx, rewardReq)
				if err != nil {
					logger.Infof("failed to get worker reward: %s", err)
					continue
				}

				if reward.Reward > 0 {
					err = m.ds.Miners.UpdateMinerReward(emptyCtx, miner, reward.Reward)
					if err != nil {
						logger.WithError(err).Error("failed to update worker reward")
						continue
					}
				}
			}
		}
	}
}
