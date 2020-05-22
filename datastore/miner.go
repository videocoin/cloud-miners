package datastore

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/mailru/dbr"
	emitterv1 "github.com/videocoin/cloud-api/emitter/v1"
	v1 "github.com/videocoin/cloud-api/miners/v1"
)

type Tags map[string]string

func (t Tags) Value() (driver.Value, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (t *Tags) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("type assertion .([]byte) failed")
	}

	return json.Unmarshal(source, t)
}

type Info map[string]interface{}

func (info Info) Value() (driver.Value, error) {
	b, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (info *Info) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("type assertion .([]byte) failed")
	}

	return json.Unmarshal(source, info)
}

type Miner struct {
	ID            string
	UserID        string
	Name          string
	AccessKey     string
	Status        v1.MinerStatus
	LastPingAt    *time.Time
	CurrentTaskID dbr.NullString
	Address       dbr.NullString
	DeletedAt     *time.Time
	Tags          Tags                      `sql:"type:json"`
	SystemInfo    Info                      `sql:"type:json"`
	CapacityInfo  Info                      `sql:"type:json"`
	WorkerInfo    *emitterv1.WorkerResponse `sql:"type:json"`
	Key           dbr.NullString
	Secret        dbr.NullString
	IsInternal    bool
	IsLock        bool
	Reward        float64
	IsBlock       bool
}

func (m *Miner) IsOnline() bool {
	return m.LastPingAt != nil && m.LastPingAt.After(time.Now().Add(-5*time.Second))
}
