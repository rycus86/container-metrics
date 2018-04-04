package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rycus86/container-metrics/container"
	"github.com/rycus86/container-metrics/stats"
	"net/http"
	"regexp"
	"time"
)

func Record(c *container.Container, s *stats.Stats) {
	setGauge(c, "cpu_total", "CPU total", float64(s.CpuStats.Total)/float64(time.Second))
}

var gauges = map[string]prometheus.Gauge{}

func setGauge(c *container.Container, name, help string, value float64) {
	key := name + "@" + c.Id

	gauge, ok := gauges[key]
	if !ok {
		labels := prometheus.Labels{}
		labels["name"] = c.Name[1:]

		for labelName, labelValue := range c.Labels {
			nonLettersOrDigits := regexp.MustCompile("[^A-Za-z0-9]")
			normalizedName := nonLettersOrDigits.ReplaceAllString(labelName, "_")
			labels[normalizedName] = labelValue
		}

		gauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        name,
			Help:        help,
			ConstLabels: labels,
		})

		prometheus.Register(gauge)

		gauges[key] = gauge
	}

	gauge.Set(value)
}

func Serve() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}
