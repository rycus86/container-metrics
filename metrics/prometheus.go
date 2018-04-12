package metrics

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rycus86/container-metrics/model"
	"net/http"
	"runtime/debug"
)

var (
	current       *PrometheusMetrics
	noCachedStats = errors.New("Previous stats not available")
)

type currentMetricsCollector struct{}

func (c *currentMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	// This only needs to output *something*
	prometheus.NewGauge(prometheus.GaugeOpts{Name: "Dummy", Help: "Dummy"}).Describe(ch)
}

func (c *currentMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	if current == nil {
		return
	}

	for _, metric := range current.Metrics {
		metric.Collect(ch)
	}

	for _, metric := range current.EngineMetrics {
		metric.Collect(ch)
	}
}

func init() {
	prometheus.Register(&currentMetricsCollector{})
}

func RecordEngineStats(stats *model.EngineStats) {
	if current == nil {
		return
	}

	recordEngineStatsOn(current, stats)
}

func recordEngineStatsOn(pm *PrometheusMetrics, stats *model.EngineStats) {
	if stats == nil {
		return
	}

	pm.EngineStats = stats

	for _, metric := range pm.EngineMetrics {
		metric.Set(stats)
	}
}

func PrepareMetrics(containers []model.Container) {
	current = NewMetrics(containers)

	RecordAll(recordCached)
}

func recordCached(c *model.Container) (*model.Stats, error) {
	cached := getCached(c.Id)
	if cached != nil {
		return cached, nil
	} else {
		return nil, noCachedStats
	}
}

func RecordAll(statsFunc func(*model.Container) (*model.Stats, error)) {
	containers := current.Containers

	for _, item := range containers {
		current := item
		go record(&current, statsFunc)
	}
}

func record(c *model.Container, statsFunc func(*model.Container) (*model.Stats, error)) {
	s, err := statsFunc(c)
	if err != nil {
		if err != noCachedStats {
			fmt.Println("Failed to collect stats:", err)
		}

		return
	}

	defer func() {
		err := recover()
		if err != nil {
			fmt.Println("Recovered:", err)
			fmt.Println(string(debug.Stack()))
		}
	}()

	for _, metric := range current.Metrics {
		metric.Set(c, s)
	}

	cacheStats(c.Id, s)
}

func Serve() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}
