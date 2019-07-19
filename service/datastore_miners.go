package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/opentracing/opentracing-go"
	v1 "github.com/videocoin/cloud-api/miners/v1"
)

var (
	ErrMinerNotFound = errors.New("miner is not found")
)

type MinerDatastore struct {
	db *gorm.DB
}

func NewMinerDatastore(db *gorm.DB) (*MinerDatastore, error) {
	db.AutoMigrate(&v1.Miner{})
	return &MinerDatastore{db: db}, nil
}

func (ds *MinerDatastore) Create(ctx context.Context, userId, hash string) (*v1.Miner, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Create")
	defer span.Finish()

	span.SetTag("user_id", userId)
	span.SetTag("hash", hash)

	tx := ds.db.Begin()

	id := uuid.New()

	miner := &v1.Miner{
		Id:     id.String(),
		UserId: userId,
		Hash:   hash,
	}

	err := tx.Create(miner).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return miner, nil
}

func (ds *MinerDatastore) Get(ctx context.Context, id string) (*v1.Miner, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Get")
	defer span.Finish()

	span.SetTag("id", id)

	miner := new(v1.Miner)

	if err := ds.db.Where("id = ?", id).First(&miner).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrMinerNotFound
		}

		return nil, fmt.Errorf("failed to get account by id: %s", err.Error())
	}

	return miner, nil
}

func (ds *MinerDatastore) GetByHash(ctx context.Context, hash string) (*v1.Miner, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GetByHash")
	defer span.Finish()

	span.SetTag("hash", hash)

	miner := new(v1.Miner)

	if err := ds.db.Where("hash = ?", hash).First(&miner).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrMinerNotFound
		}

		return nil, fmt.Errorf("failed to get account by hash: %s", err.Error())
	}

	return miner, nil
}

func (ds *MinerDatastore) List(ctx context.Context, status v1.MinerStatus) ([]*v1.Miner, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "List")
	defer span.Finish()

	miners := []*v1.Miner{}

	qs := ds.db.Where("status = ?", status).Find(&miners)

	if err := qs.Error; err != nil {
		return nil, fmt.Errorf("failed to get miners list: %s", err)
	}

	return miners, nil
}

func (ds *MinerDatastore) UpdateCPUIdle(ctx context.Context, miner *v1.Miner) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "Update")
	defer span.Finish()

	tx := ds.db.Begin()

	err := ds.db.Model(&miner).UpdateColumn("cpu_idle", miner.CpuIdle).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update miner: %s", err)
	}

	tx.Commit()

	return nil
}

func (ds *MinerDatastore) UpdateStatus(ctx context.Context, minerId string, status v1.MinerStatus) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "UpdateStatus")
	defer span.Finish()

	span.SetTag("id", minerId)
	span.SetTag("status", status)

	tx := ds.db.Begin()

	miner := v1.Miner{
		Id:     minerId,
		Status: status,
	}

	err := ds.db.Model(&miner).UpdateColumn("status", status).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to mark miner as busy: %s", err)
	}

	tx.Commit()

	return nil
}

func (ds *MinerDatastore) MarkAllAsOffline(ctx context.Context) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "MarkAllAsOffline")
	defer span.Finish()

	tx := ds.db.Begin()

	err := ds.db.Table("miners").Updates(map[string]interface{}{"is_online": false}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

func (ds *MinerDatastore) MarkAsOnline(ctx context.Context, ids []string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "MarAsOnline")
	defer span.Finish()

	tx := ds.db.Begin()

	err := ds.db.Table("miners").Where("id IN (?)", ids).Updates(map[string]interface{}{"is_online": true}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}
