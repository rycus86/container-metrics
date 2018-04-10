package metrics

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rycus86/container-metrics/container"
	"github.com/rycus86/container-metrics/stats"
	"net/http"
	"runtime/debug"
)

var (
	current *PrometheusMetrics
)

type currentMetricsCollector struct{}

func (c *currentMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	// This only needs to ouput *something*
	prometheus.NewGauge(prometheus.GaugeOpts{Name: "Dummy", Help: "Dummy"}).Describe(ch)
}

func (c *currentMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	if current == nil {
		return
	}

	for _, metric := range current.Metrics {
		metric.Collect(ch)
	}
}

func init() {
	prometheus.Register(&currentMetricsCollector{})
}

func PrepareMetrics(containers []container.Container) {
	current = NewMetrics(containers)
}

func RecordAll(statsFunc func(*container.Container) (*stats.Stats, error)) {
	containers := current.Containers

	for _, item := range containers {
		current := item
		go record(&current, statsFunc)
	}
}

func record(c *container.Container, statsFunc func(*container.Container) (*stats.Stats, error)) {
	s, err := statsFunc(c)
	if err != nil {
		fmt.Println("Failed to collect stats", err)
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
}

func Serve() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}
