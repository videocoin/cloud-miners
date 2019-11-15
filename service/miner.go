package service

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/mailru/dbr"
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
		return errors.New("type assertion .([]byte) failed.")
	}

	return json.Unmarshal(source, t)
}

type SystemInfo map[string]interface{}

func (info SystemInfo) Value() (driver.Value, error) {
	b, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (info *SystemInfo) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("type assertion .([]byte) failed.")
	}

	return json.Unmarshal(source, info)
}

type Miner struct {
	ID            string         `gorm:"primary_key"`
	UserID        string         `gorm:"type:varchar(36)`
	Name          string         `gorm:"type:varchar(255)"`
	Status        v1.MinerStatus `gorm:"type:varchar(100)"`
	LastPingAt    *time.Time
	CurrentTaskID dbr.NullString
	Address       dbr.NullString
	Tags          Tags       `sql:"type:json"`
	SystemInfo    SystemInfo `sql:"type:json"`
}

func (m *Miner) IsOnline() bool {
	return m.LastPingAt != nil && m.LastPingAt.After(time.Now().Add(-5*time.Second))
}
