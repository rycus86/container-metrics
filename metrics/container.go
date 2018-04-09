package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rycus86/container-metrics/container"
	"github.com/rycus86/container-metrics/stats"
	"regexp"
	"time"
)

func newGauge(name, help string, baseLabels []string, mapper Mapper, extraLabels ...string) *GaugeMetric {
	return &GaugeMetric{
		Metric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "cntm",
			Name:      name,
			Help:      help,
		}, append(baseLabels, extraLabels...)),

		AdditionalLabels: extraLabels,
		Mapper:           mapper,
	}
}

func NewMetrics(containers []container.Container) *PrometheusMetrics {
	uniqueBaseLabels := map[string]string{"container_name": ""}

	nonLettersOrDigits := regexp.MustCompile("[^A-Za-z0-9]")

	for _, c := range containers {
		for labelName := range c.Labels {
			normalizedName := nonLettersOrDigits.ReplaceAllString(labelName, "_")
			uniqueBaseLabels[normalizedName] = labelName
		}
	}

	var baseLabels []string
	for name := range uniqueBaseLabels {
		baseLabels = append(baseLabels, name)
	}

	metrics := &PrometheusMetrics{
		Containers: containers,
		Labels:     baseLabels,
	}

	addAllMetrics(metrics)

	return metrics
}

func addAllMetrics(metrics *PrometheusMetrics) {
	// CPU
	metrics.Add(newGauge(
		"cpu_usage_total", "Total CPU usage", metrics.Labels,
		func(s *stats.Stats) float64 {
			return float64(s.CpuStats.Total) / float64(time.Second)
		}))
	metrics.Add(newGauge(
		"cpu_usage_system", "CPU usage in system mode", metrics.Labels,
		func(s *stats.Stats) float64 {
			return float64(s.CpuStats.System) / float64(time.Second)
		}))
	metrics.Add(newGauge(
		"cpu_usage_user", "CPU usage in user mode", metrics.Labels,
		func(s *stats.Stats) float64 {
			return float64(s.CpuStats.User) / float64(time.Second)
		}))

	//CpuStats: stats.CpuStats{
	//	Total:  d.CPUStats.CPUUsage.TotalUsage,
	//	System: d.CPUStats.CPUUsage.UsageInKernelmode,
	//	User:   d.CPUStats.CPUUsage.UsageInUsermode,
	//},
	//
	//	MemoryStats: stats.MemoryStats{
	//		Total: d.MemoryStats.Limit,
	//		Free:  d.MemoryStats.Limit - d.MemoryStats.Usage,
	//	},
	//
	//		IOStats: stats.IOStats{
	//		Read:    0,
	//		Written: 0,
	//	},
	//
	//		NetworkStats: stats.NetworkStats{
	//		RxBytes:   0,
	//		RxPackets: 0,
	//		RxDropped: 0,
	//		RxErrors:  0,
	//
	//		TxBytes:   0,
	//		TxPackets: 0,
	//		TxDropped: 0,
	//		TxErrors:  0,
	//	},
}
