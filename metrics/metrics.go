package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
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
