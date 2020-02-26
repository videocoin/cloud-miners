package service

import (
	"context"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/sirupsen/logrus"
	dispatcherv1 "github.com/videocoin/cloud-api/dispatcher/v1"
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

func (ds *Datastore) StartBackgroundTasks() error {
	go ds.startCheckOfflineTask()
	return nil
}

func (ds *Datastore) StopBackgroundTasks() error {
	ds.offlineTicker.Stop()
	return nil
}

func (ds *Datastore) startCheckOfflineTask() error {
	for {
		select {
		case <-ds.offlineTicker.C:
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

					go func() {
						if miner.CurrentTaskID.String != "" {
							err = ds.EB.EmitUpdateTaskStatus(ctx, miner.CurrentTaskID.String, dispatcherv1.TaskStatusPending)
							if err != nil {
								ds.logger.Errorf("failed to emit update task status: %s", err)
							}

							err = ds.Miners.UpdateCurrentTask(context.Background(), miner, "", true)
							if err != nil {
								ds.logger.Errorf("failed to update current task: %s", err)
							}
						}
					}()
				}
			}
		}
	}

	return nil
}
