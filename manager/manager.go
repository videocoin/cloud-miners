package manager

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	streamsv1 "github.com/videocoin/cloud-api/streams/v1"
	"github.com/videocoin/cloud-miners/datastore"
	"github.com/videocoin/cloud-miners/eventbus"
)

type Manager struct {
	logger         *logrus.Entry
	ds             *datastore.Datastore
	eb             *eventbus.EventBus
	offlineTimeout time.Duration
	offlineTicker  *time.Ticker
}

func New(opts ...Option) (*Manager, error) {
	offlineTimeout := time.Second * 10
	ds := &Manager{
		offlineTimeout: offlineTimeout,
		offlineTicker:  time.NewTicker(offlineTimeout),
	}
	for _, o := range opts {
		if err := o(ds); err != nil {
			return nil, err
		}
	}

	return ds, nil
}

func (m *Manager) Start() {
	go m.checkOfflineTask()
}

func (m *Manager) Stop() {
	m.offlineTicker.Stop()
}

func (m *Manager) checkOfflineTask() {
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

				go func(miner *datastore.Miner) {
					if miner.CurrentTaskID.String != "" {
						err = m.eb.EmitUpdateStreamStatus(ctx, miner.CurrentTaskID.String, streamsv1.StreamStatusFailed)
						if err != nil {
							m.logger.Errorf("failed to update stream status: %s", err)
						}
					}
				}(miner)
			}
		}
	}
}
