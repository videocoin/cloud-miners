package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/mailru/dbr"
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
	db.AutoMigrate(&Miner{})
	return &MinerDatastore{db: db}, nil
}

func (ds *MinerDatastore) Create(ctx context.Context, userID string) (*Miner, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Create")
	defer span.Finish()

	span.SetTag("user_id", userID)

	tx := ds.db.Begin()

	id := uuid.New()

	miner := &Miner{
		ID:     id.String(),
		UserID: userID,
	}

	err := tx.Create(miner).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return miner, nil
}

func (ds *MinerDatastore) Get(ctx context.Context, id string, userID string) (*Miner, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Get")
	defer span.Finish()

	span.SetTag("id", id)

	miner := new(Miner)

	qs := ds.db.Where("id = ?", id)
	if userID != "" {
		qs = qs.Where("user_id = ?", userID)
	}

	if err := qs.First(&miner).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrMinerNotFound
		}

		return nil, fmt.Errorf("failed to get miner by id: %s", err.Error())
	}

	return miner, nil
}

func (ds *MinerDatastore) List(ctx context.Context, userID *string) ([]*Miner, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "List")
	defer span.Finish()

	miners := []*Miner{}

	qs := ds.db
	if userID != nil {
		qs = qs.Where("user_id = ?", &userID)
	}
	qs = qs.Find(&miners)

	if err := qs.Error; err != nil {
		return nil, fmt.Errorf("failed to get miners list: %s", err)
	}

	return miners, nil
}

func (ds *MinerDatastore) UpdateLastPingAt(ctx context.Context, miner *Miner) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "UpdateLastPingAt")
	defer span.Finish()

	tx := ds.db.Begin()

	miner.LastPingAt = pointer.ToTime(time.Now())
	err := ds.db.Model(&miner).UpdateColumn("last_ping_at", miner.LastPingAt).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update last_ping_at: %s", err)
	}

	miner.Status = v1.MinerStatusIdle

	if miner.CurrentTaskID.String != "" {
		miner.Status = v1.MinerStatusBusy
	}

	err = ds.db.Model(&miner).UpdateColumn("status", miner.Status).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update last_ping_at: %s", err)
	}

	tx.Commit()

	return nil
}

func (ds *MinerDatastore) UpdateSystemInfo(ctx context.Context, miner *Miner, systemInfo SystemInfo) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "UpdateSystemInfo")
	defer span.Finish()

	tx := ds.db.Begin()

	miner.SystemInfo = systemInfo
	err := ds.db.Model(&miner).UpdateColumn("system_info", miner.SystemInfo).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update system_info: %s", err)
	}

	tx.Commit()

	return nil
}

func (ds *MinerDatastore) UpdateCurrentTask(ctx context.Context, miner *Miner, taskID string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "UpdateCurrentTask")
	defer span.Finish()

	tx := ds.db.Begin()

	if taskID == "" {
		miner.CurrentTaskID = dbr.NewNullString(nil)
	} else {
		miner.CurrentTaskID = dbr.NewNullString(taskID)
	}

	err := ds.db.Model(&miner).UpdateColumn("current_task_id", miner.CurrentTaskID).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update current_task_id: %s", err)
	}

	tx.Commit()

	return nil
}

func (ds *MinerDatastore) UpdateStatus(ctx context.Context, minerID string, status v1.MinerStatus) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "UpdateStatus")
	defer span.Finish()

	span.SetTag("id", minerID)
	span.SetTag("status", status)

	tx := ds.db.Begin()

	miner := Miner{
		ID:     minerID,
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

	err := ds.db.
		Table("miners").
		Updates(map[string]interface{}{
			"is_online": false,
			"status":    v1.MinerStatusOffline,
		}).
		Error
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

	err := ds.db.
		Table("miners").
		Where("id IN (?)", ids).
		Updates(map[string]interface{}{
			"is_online": true,
			"status":    v1.MinerStatusIdle,
		}).
		Error
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

func (ds *MinerDatastore) MarkAsOffline(ctx context.Context, d time.Duration) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "MarkAsOffline")
	defer span.Finish()

	tx := ds.db.Begin()

	t := time.Now().Add(-d)

	err := ds.db.
		Table("miners").
		Where("last_ping_at < ?", t).
		Updates(map[string]interface{}{
			"status": v1.MinerStatusOffline,
		}).
		Error
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

func (ds *MinerDatastore) SetTags(ctx context.Context, miner *Miner, tags []*v1.Tag) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "SetTags")
	defer span.Finish()

	span.SetTag("id", miner.ID)

	tx := ds.db.Begin()

	if miner.Tags == nil {
		miner.Tags = Tags{}
	}

	for _, tag := range tags {
		if tag.Value == "" {
			delete(miner.Tags, tag.Key)
		} else {
			miner.Tags[tag.Key] = tag.Value
		}
	}

	keysCount := 0
	for range miner.Tags {
		keysCount++
	}

	if keysCount == 0 {
		miner.Tags = nil
	}

	err := ds.db.Model(&miner).UpdateColumn("tags", miner.Tags).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to set tags: %s", err)
	}

	tx.Commit()

	return nil
}
