package manager

import (
	"context"
	"time"

	prototypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
	emitterv1 "github.com/videocoin/cloud-api/emitter/v1"
	streamsv1 "github.com/videocoin/cloud-api/streams/v1"
	"github.com/videocoin/cloud-miners/datastore"
	"github.com/videocoin/cloud-miners/eventbus"
)

type Manager struct {
	logger         *logrus.Entry
	offlineTimeout time.Duration
	offlineTicker  *time.Ticker
	lwTicker       *time.Ticker
	ds             *datastore.Datastore
	eb             *eventbus.EventBus
	emitter        emitterv1.EmitterServiceClient
}

func New(opts ...Option) (*Manager, error) {
	offlineTimeout := time.Second * 10
	ds := &Manager{
		offlineTimeout: offlineTimeout,
		offlineTicker:  time.NewTicker(offlineTimeout),
		lwTicker:       time.NewTicker(time.Second * 10),
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
	go m.listWorkers()
	go m.checkStuckMiners()
}

func (m *Manager) Stop() {
	m.offlineTicker.Stop()
	m.lwTicker.Stop()
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

				if miner.CurrentTaskID.String != "" {
					err = m.eb.EmitUpdateStreamStatus(ctx, miner.CurrentTaskID.String, streamsv1.StreamStatusFailed)
					if err != nil {
						m.logger.Errorf("failed to update stream status: %s", err)
					}
				}
			}
		}
	}
}

func (m *Manager) listWorkers() {
	for range m.lwTicker.C {
		workers, err := m.emitter.ListWorkers(context.Background(), &prototypes.Empty{})
		if err != nil {
			m.logger.Infof("failed to list workers: %s", err)
			continue
		}

		for _, worker := range workers.Items {
			ctx := context.Background()

			_, err := m.ds.Miners.GetByAddress(ctx, worker.Address)
			if err != nil {
				m.logger.
					WithField("address", worker.Address).
					Warningf("failed to get worker by address: %s", err)
				continue
			}

			err = m.ds.Miners.UpdateWorkerInfoByAddress(ctx, worker.Address, worker)
			if err != nil {
				m.logger.
					WithField("address", worker.Address).
					Errorf("failed to update worker info: %s", err)
				continue
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
