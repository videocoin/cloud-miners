package service

import (
	"context"
	"fmt"
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
			for k, v := range data {
				id := GetMD5Hash(k)
				ctx := context.Background()
				miner, err := ds.ds.Miners.Get(ctx, id)

				if err != nil {
					ds.logger.Errorf("failed to create miner: %s", err)
				} else {
					miner.CpuIdle = v
					err = ds.ds.Miners.UpdateCPUIdle(ctx, miner)
					if err != nil {
						ds.logger.Error(err)
					}
				}

				continue
			}
		}
	}

	return nil
}

func (ds *DataSyncer) Stop() error {
	fmt.Println("ticker stop")
	ds.ticker.Stop()
	return nil
}

func (ds *DataSyncer) GetCPUIdleGroupByHost() (map[string]int64, error) {
	datapoints, err := ds.cli.QueryMultiSince([]string{"cpu-total.*.miners.agents.cpu.usage_idle"}, 10*time.Second)
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
			if point.Value != nil {
				data[host] = *point.Value
			}
		}
	}

	return data, nil
}
