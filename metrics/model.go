package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/rycus86/container-metrics/model"
)

type PrometheusMetrics struct {
	Containers []model.Container
	Labels     map[string]string // {container.label} -> {prometheus_label}
	Metrics    []SingleMetric

	EngineStats   *model.EngineStats
	EngineMetrics []EngineMetric
}

type SingleMetric interface {
	prometheus.Collector

	WithParent(*PrometheusMetrics) SingleMetric
	Set(*model.Container, *model.Stats)
}

type EngineMetric interface {
	prometheus.Collector

	Set(*model.EngineStats)
}

func (pm *PrometheusMetrics) Add(metric SingleMetric) {
	pm.Metrics = append(pm.Metrics, metric.WithParent(pm))
}

func (pm *PrometheusMetrics) AddEngine(metric EngineMetric) {
	pm.EngineMetrics = append(pm.EngineMetrics, metric)
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
