package service

import (
	"context"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" //nolint
	"github.com/sirupsen/logrus"
	streamsv1 "github.com/videocoin/cloud-api/streams/v1"
	"github.com/videocoin/cloud-miners/eventbus"
)

type Datastore struct {
	logger         *logrus.Entry
	offlineTicker  *time.Ticker
	offlineTimeout time.Duration

	Miners *MinerDatastore
	EB     *eventbus.EventBus
}

func NewDatastore(uri string, eb *eventbus.EventBus, logger *logrus.Entry) (*Datastore, error) {
	ds := new(Datastore)

	db, err := gorm.Open("mysql", uri)
	if err != nil {
		return nil, err
	}

	db.LogMode(true)

	minersDs, err := NewMinerDatastore(db)
	if err != nil {
		return nil, err
	}

	ds.EB = eb
	ds.Miners = minersDs
	ds.offlineTimeout = 10 * time.Second
	ds.offlineTicker = time.NewTicker(ds.offlineTimeout)

	return ds, nil
}

func (ds *Datastore) StartBackgroundTasks() {
	go ds.startCheckOfflineTask()
}

func (ds *Datastore) StopBackgroundTasks() error {
	ds.offlineTicker.Stop()
	return nil
}

func (ds *Datastore) startCheckOfflineTask() {
	for range ds.offlineTicker.C {
		ctx := context.Background()
		err := ds.Miners.MarkAsOffline(ctx, ds.offlineTimeout)
		if err != nil {
			ds.logger.Errorf("failed to mark miners as offline: %s", err)
			continue
		}

		miners, err := ds.Miners.GetStuckMinerList(ctx, ds.offlineTimeout)
		if err != nil {
			ds.logger.Errorf("failed to get stuck miners: %s", err)
			continue
		}

		if len(miners) > 0 {
			for _, miner := range miners {
				err := ds.Miners.MarkMinerAsOffline(ctx, miner)
				if err != nil {
					ds.logger.Errorf("failed to mark miner as offline: %s", err)
					continue
				}

				go func(m *Miner) {
					if m.CurrentTaskID.String != "" {
						err = ds.EB.EmitUpdateStreamStatus(ctx, m.CurrentTaskID.String, streamsv1.StreamStatusFailed)
						if err != nil {
							ds.logger.Errorf("failed to update stream status: %s", err)
						}
					}
				}(miner)
			}
		}
	}
}
