package service

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	v1 "github.com/videocoin/cloud-api/miners/v1"
	"golang.org/x/net/context"
)

type Metrics struct {
	internalMinerStatus *prometheus.GaugeVec
}

func NewMetrics(namespace string) *Metrics {
	return &Metrics{
		internalMinerStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "internal_miner_status",
				Help:      "Status of internal miner",
			},
			[]string{"status", "hostname"},
		),
	}
}

func (m *Metrics) RegisterAll() {
	prometheus.MustRegister(m.internalMinerStatus)
}

type MetricsCollector struct {
	mutex   sync.RWMutex
	metrics *Metrics
	ds      *Datastore
	ticker  *time.Ticker
}

func NewMetricsCollector(namespace string, ds *Datastore) *MetricsCollector {
	metrics := NewMetrics(namespace)
	metrics.RegisterAll()
	return &MetricsCollector{
		metrics: metrics,
		ds:      ds,
		ticker:  time.NewTicker(time.Second * 5),
	}
}

func (mc *MetricsCollector) Collect() {
	for range mc.ticker.C {
		mc.mutex.Lock()
		mc.collectMetrics()
		mc.mutex.Unlock()
	}
}

func (mc *MetricsCollector) collectMetrics() {
	statuses := []string{
		v1.MinerStatusNew.String(),
		v1.MinerStatusIdle.String(),
		v1.MinerStatusBusy.String(),
		v1.MinerStatusOffline.String(),
	}

	ctx := context.Background()
	miners, err := mc.ds.Miners.List(ctx, nil)
	if err == nil {
		for _, miner := range miners {
			for _, status := range statuses {
				hostname := ""
				if status == miner.Status.String() {
					host, ok := miner.SystemInfo["host"].(map[string]interface{})
					if ok {
						hostname = host["hostname"].(string)
					}
					mc.metrics.internalMinerStatus.WithLabelValues(status, hostname).Set(1)
				} else {
					mc.metrics.internalMinerStatus.WithLabelValues(status, hostname).Set(0)
				}
			}
		}
	}
}

func (mc *MetricsCollector) Start() error {
	go mc.Collect()
	return nil
}

func (mc *MetricsCollector) Stop() error {
	mc.ticker.Stop()
	return nil
}
