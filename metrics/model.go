package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rycus86/container-metrics/container"
	"github.com/rycus86/container-metrics/stats"
)

type PrometheusMetrics struct {
	Containers []container.Container
	Labels     []string
	Metrics    []SettableMetric
}

type SettableMetric interface {
	prometheus.Collector

	WithParent(*PrometheusMetrics)
	Set(*container.Container, *stats.Stats)
}

type Mapper func(*stats.Stats) float64

type GaugeMetric struct {
	Metric           *prometheus.GaugeVec
	Mapper           Mapper
	AdditionalLabels []string

	Parent *PrometheusMetrics
}

func (m *GaugeMetric) Describe(ch chan<- *prometheus.Desc) {
	m.Metric.Describe(ch)
}

func (m *GaugeMetric) Collect(ch chan<- prometheus.Metric) {
	m.Metric.Collect(ch)
}

func (m *GaugeMetric) WithParent(pm *PrometheusMetrics) {
	m.Parent = pm
}

func (m *GaugeMetric) Set(c *container.Container, s *stats.Stats) {
	labelNames := append(m.Parent.Labels, m.AdditionalLabels...)
	m.Metric.With(extractLabels(c, labelNames...)).Set(m.Mapper(s))
}

func extractLabels(c *container.Container, names ...string) map[string]string {
	values := map[string]string{
		"container_name": c.Name[1:],
	}

	for _, name := range names {
		_, exists := values[name]
		if exists {
			continue
		}

		label, _ := c.Labels[name]
		values[name] = label
	}

	return values
}

func (pm *PrometheusMetrics) Add(metric SettableMetric) {
	metric.WithParent(pm)
	pm.Metrics = append(pm.Metrics, metric)
}
