package service

import (
	"context"
	"errors"
	"fmt"

	wrappers "github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"
	"github.com/opentracing/opentracing-go"
	"github.com/segmentio/ksuid"
	v1 "github.com/videocoin/cloud-api/miners/v1"
	"github.com/videocoin_/common/uuid4"
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

func (ds *MinerDatastore) Create(ctx context.Context, userId string) (*v1.Miner, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Create")
	defer span.Finish()

	span.LogKV("user_id", userId)

	tx := ds.db.Begin()

	id, err := uuid4.New()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	miner := &v1.Miner{
		Id:     id,
		UserId: userId,
		Key:    ksuid.New().String(),
	}

	err = tx.Create(miner).Error
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

	span.LogKV("id", id)

	miner := new(v1.Miner)

	if err := ds.db.Where("id = ?", id).First(&miner).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrMinerNotFound
		}

		return nil, fmt.Errorf("failed to get account by id: %s", err.Error())
	}

	return miner, nil
}

func (ds *MinerDatastore) List(ctx context.Context, isBusy *wrappers.BoolValue) ([]*v1.Miner, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "List")
	defer span.Finish()

	miners := []*v1.Miner{}

	qs := ds.db
	if isBusy != nil {
		qs = qs.Where("is_busy = ?", isBusy.Value)
	}

	qs = qs.Find(&miners)

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

func (ds *MinerDatastore) MarkAsBusy(ctx context.Context, miner *v1.Miner) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "MarkAsBusy")
	defer span.Finish()

	tx := ds.db.Begin()

	miner.IsBusy = true
	err := ds.db.Model(&miner).UpdateColumn("is_busy", true).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to mark miner as busy: %s", err)
	}

	tx.Commit()

	return nil
}

func (ds *MinerDatastore) MarkAsIdle(ctx context.Context, miner *v1.Miner) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "MarkAsIdle")
	defer span.Finish()

	tx := ds.db.Begin()

	miner.IsBusy = false
	err := ds.db.Model(&miner).UpdateColumn("is_busy", false).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to mark miner as idle: %s", err)
	}

	tx.Commit()

	return nil
}
