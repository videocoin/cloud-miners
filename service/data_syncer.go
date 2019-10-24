package service

import (
	"context"
	"strings"
	"time"

	graphite "github.com/JensRantil/graphite-client"
	"github.com/sirupsen/logrus"
)

type DataSyncer struct {
	addr     string
	interval time.Duration
	cli      *graphite.Client
	ds       *Datastore
	ticker   *time.Ticker
	logger   *logrus.Entry
}

func NewDataSyncer(addr string, interval time.Duration, ds *Datastore, logger *logrus.Entry) (*DataSyncer, error) {
	cli, err := graphite.New(addr)
	if err != nil {
		return nil, err
	}

	ticker := time.NewTicker(interval)

	syncer := &DataSyncer{
		addr:     addr,
		interval: interval,
		cli:      cli,
		ticker:   ticker,
		logger:   logger,
		ds:       ds,
	}

	return syncer, nil
}

func (ds *DataSyncer) Start() error {
	for {
		select {
		case <-ds.ticker.C:
			data, err := ds.GetCPUIdleGroupByHost()
			if err != nil {
				ds.logger.Errorf("failed to get cpu idle group by host: %s", err)
				continue
			}

			ctx := context.Background()

			if len(data) > 0 {
				ids := []string{}
				for k, v := range data {
					miner, err := ds.ds.Miners.Get(ctx, k)
					if err != nil {
						ds.logger.Errorf("failed to get miner: %s", err)
					} else {
						ids = append(ids, miner.ID)

						miner.CPUIdle = v
						err = ds.ds.Miners.UpdateCPUIdle(ctx, miner)
						if err != nil {
							ds.logger.Error(err)
							continue
						}
					}
				}

				if len(ids) > 0 {
					ds.ds.Miners.MarkAllAsOffline(ctx)
					err := ds.ds.Miners.MarkAsOnline(ctx, ids)
					if err != nil {
						ds.logger.Errorf("failed to mark as online: %s", err)
						continue
					}
				}
			} else {
				err := ds.ds.Miners.MarkAllAsOffline(ctx)
				if err != nil {
					ds.logger.Errorf("failed to mark all as offline: %s", err)
					continue
				}
			}
		}
	}

	return nil
}

func (ds *DataSyncer) Stop() error {
	ds.ticker.Stop()
	return nil
}

func (ds *DataSyncer) GetCPUIdleGroupByHost() (map[string]int64, error) {
	tenSecs := 30 * time.Second
	datapoints, err := ds.cli.QueryMultiSince([]string{"cpu-total.*.miners.agents.cpu.usage_idle"}, tenSecs)
	if err != nil {
		return nil, err
	}

	data := map[string]int64{}

	for _, datapoint := range datapoints {
		parts := strings.Split(datapoint.Target, ".")
		host := parts[1]
		points, err := datapoint.AsInts()
		if err != nil {
			continue
		}

		for _, point := range points {
			ago := time.Now().Add(-tenSecs)
			if point.Time.After(ago) {
				if point.Value != nil && !point.Time.IsZero() {
					data[host] = *point.Value
				}
			}
		}
	}

	return data, nil
}
