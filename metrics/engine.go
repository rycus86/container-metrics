package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/rycus86/container-metrics/model"
)

type EngineMapper func(*model.EngineStats) float64

type EngineGaugeMetric struct {
	Metric *prometheus.GaugeVec

	Mapper EngineMapper
}

func newEngineGauge(name, help string, mapper EngineMapper) *EngineGaugeMetric {
	return &EngineGaugeMetric{
		Metric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: defaultNamespace,
			Name:      name,
			Help:      help,
		}, []string{"engine_host"}),

		Mapper: mapper,
	}
}

func (m *EngineGaugeMetric) Describe(ch chan<- *prometheus.Desc) {
	m.Metric.Describe(ch)
}

func (m *EngineGaugeMetric) Collect(ch chan<- prometheus.Metric) {
	m.Metric.Collect(ch)
}

func (m *EngineGaugeMetric) Set(stats *model.EngineStats) {
	m.Metric.With(prometheus.Labels{
		"engine_host": stats.Host,
	}).Set(m.Mapper(stats))
}
