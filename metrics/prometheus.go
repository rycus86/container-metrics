package metrics

import (
	"log"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/rycus86/container-metrics/model"
)

const defaultNamespace = "cntm"

var noCachedStats = errors.New("Previous stats not available")

type currentMetricsCollector struct{}

func (c *currentMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	// This only needs to output *something*
	prometheus.NewGauge(prometheus.GaugeOpts{Name: "Dummy", Help: "Dummy"}).Describe(ch)
}

func (c *currentMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	current := getCurrent()

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
	if current := getCurrent(); current != nil {
		recordEngineStatsOn(current, stats)
	}
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
	setCurrent(NewMetrics(containers))
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
	containers := getCurrent().Containers

	for _, item := range containers {
		current := item
		go record(&current, statsFunc)
	}
}

func record(c *model.Container, statsFunc func(*model.Container) (*model.Stats, error)) {
	s, err := statsFunc(c)
	if err != nil {
		if err != noCachedStats {
			log.Println("Failed to collect stats for", c.Name, err)
		}

		return
	}

	for _, metric := range getCurrent().Metrics {
		metric.Set(c, s)
	}

	cacheStats(c.Id, s)
}

func Serve(port int) {
	log.Println("Serving metrics on port", port)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
