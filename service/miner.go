package service

import v1 "github.com/videocoin/cloud-api/miners/v1"

type Miner struct {
	ID       string `gorm:"primary_key"`
	UserID   string `gorm:"type:varchar(36)`
	CPUIdle  int64
	Status   v1.MinerStatus `gorm:"type:varchar(100)"`
	IsOnline bool
}

func (m *Miner) Id() string {
	return m.ID
}

func (m *Miner) UserId() string {
	return m.UserID
}

func (m *Miner) CpuIdle() int64 {
	return m.CPUIdle
}
