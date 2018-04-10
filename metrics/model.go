package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rycus86/container-metrics/container"
	"github.com/rycus86/container-metrics/stats"
)

type PrometheusMetrics struct {
	Containers []container.Container
	Labels     map[string]string
	Metrics    []SingleMetric
}

type SingleMetric interface {
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
	m.Metric.With(m.extractLabels(c)).Set(m.Mapper(s))
}

func (m *GaugeMetric) extractLabels(c *container.Container) map[string]string {
	values := map[string]string{
		"container_name": c.Name[1:],
	}

	for name, key := range m.Parent.Labels {
		_, exists := values[key]
		if exists {
			continue
		}

		label, _ := c.Labels[name]
		values[key] = label
	}

	// TODO additional labels - is it needed?
	return values
}

func (pm *PrometheusMetrics) Add(metric SingleMetric) {
	metric.WithParent(pm)
	pm.Metrics = append(pm.Metrics, metric)
}

func (pm *PrometheusMetrics) GetLabelNames() []string {
	labelNames := make([]string, len(pm.Labels), len(pm.Labels))

	idx := 0
	for _, name := range pm.Labels {
		labelNames[idx] = name
		idx++
	}

	return labelNames
}
