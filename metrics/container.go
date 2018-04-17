package metrics

import (
	"regexp"
	"time"

	"github.com/rycus86/container-metrics/model"
)

var (
	nonLettersOrDigits = regexp.MustCompile("[^A-Za-z0-9_]")
)

func NewMetrics(containers []model.Container) *PrometheusMetrics {
	baseLabels := map[string]string{
		"container.name":  "container_name",
		"container.image": "container_image",
		"engine.host":     "engine_host",
	}

	for _, c := range containers {
		for labelName := range c.Labels {
			normalizedName := nonLettersOrDigits.ReplaceAllString(labelName, "_")
			baseLabels[labelName] = normalizedName
		}
	}

	metrics := &PrometheusMetrics{
		Containers: containers,
		Labels:     baseLabels,
	}

	addAllMetrics(metrics)

	if current := getCurrent(); current != nil {
		recordEngineStatsOn(metrics, current.EngineStats)
	}

	return metrics
}

func addAllMetrics(metrics *PrometheusMetrics) {
	baseLabels := metrics.GetLabelNames()

	// Engine metrics
	metrics.AddEngine(newEngineGauge(
		"engine_num_images", "Number of images",
		func(stats *model.EngineStats) float64 {
			return float64(stats.Images)
		},
	))
	metrics.AddEngine(newEngineGauge(
		"engine_num_containers", "Number of containers",
		func(stats *model.EngineStats) float64 {
			return float64(stats.Containers)
		},
	))
	metrics.AddEngine(newEngineGauge(
		"engine_num_containers_running", "Number of running containers",
		func(stats *model.EngineStats) float64 {
			return float64(stats.ContainersRunning)
		},
	))
	metrics.AddEngine(newEngineGauge(
		"engine_num_containers_stopped", "Number of stopped containers",
		func(stats *model.EngineStats) float64 {
			return float64(stats.ContainersStopped)
		},
	))
	metrics.AddEngine(newEngineGauge(
		"engine_num_containers_paused", "Number of paused containers",
		func(stats *model.EngineStats) float64 {
			return float64(stats.ContainersPaused)
		},
	))

	// CPU metrics
	metrics.Add(newGauge(
		"cpu_usage_total_seconds", "Total CPU usage", baseLabels,
		func(s *model.Stats) float64 {
			return float64(s.CpuStats.Total) / float64(time.Second)
		}))
	metrics.Add(newGauge(
		"cpu_usage_system_seconds", "CPU usage in system mode", baseLabels,
		func(s *model.Stats) float64 {
			return float64(s.CpuStats.System) / float64(time.Second)
		}))
	metrics.Add(newGauge(
		"cpu_usage_user_seconds", "CPU usage in user mode", baseLabels,
		func(s *model.Stats) float64 {
			return float64(s.CpuStats.User) / float64(time.Second)
		}))
	metrics.Add(newGauge(
		"cpu_usage_percent", "Total CPU usage in percent", baseLabels,
		func(s *model.Stats) float64 {
			previous := getCached(s.Id)
			if previous == nil {
				return 0
			}

			delta := s.Read.Sub(previous.Read)
			diff := float64(s.CpuStats.Total - previous.CpuStats.Total) / float64(time.Second)

			// TODO how does the docker CLI compute this?
			// TODO https://github.com/docker/docker-ce/blob/master/components/cli/cli/command/container/stats_helpers.go
			return diff / delta.Seconds()
		}))

	// Memory metrics
	metrics.Add(newGauge(
		"memory_total_bytes", "Total memory available", baseLabels,
		func(s *model.Stats) float64 {
			return float64(s.MemoryStats.Total)
		}))
	metrics.Add(newGauge(
		"memory_usage_bytes", "Memory usage", baseLabels,
		func(s *model.Stats) float64 {
			return float64(s.MemoryStats.Usage)
		}))
	metrics.Add(newGauge(
		"memory_usage_percent", "Memory usage in percent", baseLabels,
		func(s *model.Stats) float64 {
			return float64(s.MemoryStats.Usage * 100) / float64(s.MemoryStats.Total)
		}))
	metrics.Add(newGauge(
		"memory_free_bytes", "Free memory", baseLabels,
		func(s *model.Stats) float64 {
			return float64(s.MemoryStats.Free)
		}))
	metrics.Add(newGauge(
		"memory_free_percent", "Free memory in percent", baseLabels,
		func(s *model.Stats) float64 {
			return float64(s.MemoryStats.Free * 100) / float64(s.MemoryStats.Total)
		}))

	// I/O metrics
	metrics.Add(newGauge(
		"io_read_bytes", "I/O bytes read", baseLabels,
		func(s *model.Stats) float64 {
			return float64(s.IOStats.Read)
		}))
	metrics.Add(newGauge(
		"io_write_bytes", "I/O bytes written", baseLabels,
		func(s *model.Stats) float64 {
			return float64(s.IOStats.Written)
		}))

	// Network metrics
	metrics.Add(newGauge(
		"net_rx_bytes", "Network receive bytes", baseLabels,
		func(s *model.Stats) float64 {
			return float64(s.NetworkStats.RxBytes)
		}))
	metrics.Add(newGauge(
		"net_rx_packets", "Network receive packets", baseLabels,
		func(s *model.Stats) float64 {
			return float64(s.NetworkStats.RxPackets)
		}))
	metrics.Add(newGauge(
		"net_rx_dropped", "Network receive packets dropped", baseLabels,
		func(s *model.Stats) float64 {
			return float64(s.NetworkStats.RxDropped)
		}))
	metrics.Add(newGauge(
		"net_rx_errors", "Network receive errors", baseLabels,
		func(s *model.Stats) float64 {
			return float64(s.NetworkStats.RxErrors)
		}))

	metrics.Add(newGauge(
		"net_tx_bytes", "Network transmit bytes", baseLabels,
		func(s *model.Stats) float64 {
			return float64(s.NetworkStats.TxBytes)
		}))
	metrics.Add(newGauge(
		"net_tx_packets", "Network transmit packets", baseLabels,
		func(s *model.Stats) float64 {
			return float64(s.NetworkStats.TxPackets)
		}))
	metrics.Add(newGauge(
		"net_tx_dropped", "Network transmit packets dropped", baseLabels,
		func(s *model.Stats) float64 {
			return float64(s.NetworkStats.TxDropped)
		}))
	metrics.Add(newGauge(
		"net_tx_errors", "Network transmit errors", baseLabels,
		func(s *model.Stats) float64 {
			return float64(s.NetworkStats.TxErrors)
		}))
}
