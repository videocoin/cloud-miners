package metrics

import (
	"sync"
	"time"

	v1 "github.com/videocoin/cloud-api/miners/v1"
	"github.com/videocoin/cloud-miners/datastore"
	"golang.org/x/net/context"
)

type Collector struct {
	mutex   sync.RWMutex
	metrics *Metrics
	ds      *datastore.Datastore
	ticker  *time.Ticker
}

func NewCollector(namespace string, ds *datastore.Datastore) *Collector {
	metrics := NewMetrics(namespace)
	metrics.RegisterAll()
	return &Collector{
		metrics: metrics,
		ds:      ds,
		ticker:  time.NewTicker(time.Second * 5),
	}
}

func (mc *Collector) Collect() {
	for range mc.ticker.C {
		mc.mutex.Lock()
		mc.collectMetrics()
		mc.mutex.Unlock()
	}
}

func (mc *Collector) collectMetrics() {
	statuses := []string{
		v1.MinerStatusNew.String(),
		v1.MinerStatusIdle.String(),
		v1.MinerStatusBusy.String(),
		v1.MinerStatusOffline.String(),
	}

	mc.metrics.internalMinerStatus.Reset()

	ctx := context.Background()
	miners, err := mc.ds.Miners.ListByInternal(ctx)
	if err == nil {
		for _, miner := range miners {
			for _, status := range statuses {
				hostname := ""
				host, ok := miner.SystemInfo["host"].(map[string]interface{})
				if ok {
					hostname = host["hostname"].(string)
				}
				if status == miner.Status.String() {
					mc.metrics.internalMinerStatus.WithLabelValues(status, hostname).Set(1)
				} else {
					mc.metrics.internalMinerStatus.WithLabelValues(status, hostname).Set(0)
				}
			}
		}
	}
}

func (mc *Collector) Start() {
	go mc.Collect()
}

func (mc *Collector) Stop() error {
	mc.ticker.Stop()
	return nil
}
