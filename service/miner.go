package service

import (
	"time"

	"github.com/mailru/dbr"
	v1 "github.com/videocoin/cloud-api/miners/v1"
)

type Miner struct {
	ID            string         `gorm:"primary_key"`
	UserID        string         `gorm:"type:varchar(36)`
	Status        v1.MinerStatus `gorm:"type:varchar(100)"`
	LastPingAt    *time.Time
	CurrentTaskID dbr.NullString
}

func (m *Miner) IsOnline() bool {
	return m.LastPingAt != nil && m.LastPingAt.After(time.Now().Add(-5*time.Second))
}
