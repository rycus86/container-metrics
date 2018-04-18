package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/rycus86/container-metrics/model"
)

type Mapper func(*model.Stats) float64

type GaugeMetric struct {
	Metric *prometheus.GaugeVec
	Mapper Mapper

	Parent *PrometheusMetrics
}

func newGauge(name, help string, baseLabels []string, mapper Mapper) *GaugeMetric {
	return &GaugeMetric{
		Metric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: defaultNamespace,
			Name:      name,
			Help:      help,
		}, baseLabels),

		Mapper: mapper,
	}
}

func (m *GaugeMetric) Describe(ch chan<- *prometheus.Desc) {
	m.Metric.Describe(ch)
}

func (m *GaugeMetric) Collect(ch chan<- prometheus.Metric) {
	m.Metric.Collect(ch)
}

func (m *GaugeMetric) WithParent(pm *PrometheusMetrics) SingleMetric {
	m.Parent = pm
	return m
}

func (m *GaugeMetric) Set(c *model.Container, s *model.Stats) {
	m.Metric.With(m.extractLabels(c)).Set(m.Mapper(s))
}

func (m *GaugeMetric) extractLabels(c *model.Container) map[string]string {
	values := map[string]string{
		"container_name":  c.Name,
		"container_image": c.Image,
	}

	if m.Parent != nil && m.Parent.EngineStats != nil {
		values["engine_host"] = m.Parent.EngineStats.Host
	}

	for name, key := range m.Parent.Labels {
		_, exists := values[key]
		if exists {
			continue
		}

		label, _ := c.Labels[name]
		values[key] = label
	}

	return values
}
