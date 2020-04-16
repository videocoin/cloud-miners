package datastore

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" //nolint
	"github.com/videocoin/cloud-miners/eventbus"
)

type Datastore struct {
	Miners *MinerDatastore
	EB     *eventbus.EventBus
}

func NewDatastore(uri string, eb *eventbus.EventBus) (*Datastore, error) {
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

	return ds, nil
}
