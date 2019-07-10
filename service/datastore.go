package service

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Datastore struct {
	Miners *MinerDatastore
}

func NewDatastore(uri string) (*Datastore, error) {
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

	return ds, nil
}
