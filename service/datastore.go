package service

import (
	"context"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/sirupsen/logrus"
)

type Datastore struct {
	Miners         *MinerDatastore
	logger         *logrus.Entry
	offlineTicker  *time.Ticker
	offlineTimeout time.Duration
}

func NewDatastore(uri string, logger *logrus.Entry) (*Datastore, error) {
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

	ds.Miners = minersDs
	ds.offlineTimeout = 5 * time.Second
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
			err := ds.Miners.MarkAsOffline(context.Background(), ds.offlineTimeout)
			if err != nil {
				ds.logger.Errorf("failed to mark miners as offline: %s", err)
				continue
			}
		}
	}

	return nil
}
