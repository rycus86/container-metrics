package docker

import (
	"strings"

	"github.com/docker/docker/api/types"

	"github.com/rycus86/container-metrics/model"
)

func convertStats(d *types.StatsJSON, osType string) *model.Stats {
	s := model.Stats{
		Id:   d.ID,
		Name: d.Name[1:],

		CpuStats: model.CpuStats{
			Total:   d.CPUStats.CPUUsage.TotalUsage,
			System:  d.CPUStats.CPUUsage.UsageInKernelmode,
			User:    d.CPUStats.CPUUsage.UsageInUsermode,
			Percent: calculateCPUPercent(d, osType),
		},

		MemoryStats: model.MemoryStats{
			Total:   d.MemoryStats.Limit,
			Usage:   calculateMemUsage(d, osType),
			Percent: calculateMemPercent(d, osType),
		},

		IOStats: model.IOStats{
			Read:    0,
			Written: 0,
		},

		NetworkStats: model.NetworkStats{
			RxBytes:   0,
			RxPackets: 0,
			RxDropped: 0,
			RxErrors:  0,

			TxBytes:   0,
			TxPackets: 0,
			TxDropped: 0,
			TxErrors:  0,
		},
	}

	for _, ioEntry := range d.BlkioStats.IoServiceBytesRecursive {
		switch strings.ToLower(ioEntry.Op) {
		case "read":
			s.IOStats.Read += ioEntry.Value
		case "write":
			s.IOStats.Written += ioEntry.Value
		}
	}

	for _, netEntry := range d.Networks {
		s.NetworkStats.RxBytes += netEntry.RxBytes
		s.NetworkStats.RxPackets += netEntry.RxPackets
		s.NetworkStats.RxDropped += netEntry.RxDropped
		s.NetworkStats.RxErrors += netEntry.RxErrors

		s.NetworkStats.TxBytes += netEntry.TxBytes
		s.NetworkStats.TxPackets += netEntry.TxPackets
		s.NetworkStats.TxDropped += netEntry.TxDropped
		s.NetworkStats.TxErrors += netEntry.TxErrors
	}

	return &s
}

func calculateCPUPercent(v *types.StatsJSON, osType string) float64 {
	if osType != "windows" {
		previousCPU := v.PreCPUStats.CPUUsage.TotalUsage
		previousSystem := v.PreCPUStats.SystemUsage
		return calculateCPUPercentUnix(previousCPU, previousSystem, v)
	} else {
		return calculateCPUPercentWindows(v)
	}
}

func calculateMemUsage(v *types.StatsJSON, osType string) float64 {
	if osType != "windows" {
		return calculateMemUsageUnixNoCache(v.MemoryStats)
	}

	return 0
}

func calculateMemPercent(v *types.StatsJSON, osType string) float64 {
	if osType != "windows" {
		mem := calculateMemUsageUnixNoCache(v.MemoryStats)
		memLimit := float64(v.MemoryStats.Limit)
		return calculateMemPercentUnixNoCache(memLimit, mem)
	}

	return 0
}

// taken from https://github.com/docker/docker-ce/blob/master/components/cli/cli/command/container/stats_helpers.go

func calculateCPUPercentUnix(previousCPU, previousSystem uint64, v *types.StatsJSON) float64 {
	var (
		cpuPercent = 0.0
		// calculate the change for the cpu usage of the container in between readings
		cpuDelta = float64(v.CPUStats.CPUUsage.TotalUsage) - float64(previousCPU)
		// calculate the change for the entire system between readings
		systemDelta = float64(v.CPUStats.SystemUsage) - float64(previousSystem)
		onlineCPUs  = float64(v.CPUStats.OnlineCPUs)
	)

	if onlineCPUs == 0.0 {
		onlineCPUs = float64(len(v.CPUStats.CPUUsage.PercpuUsage))
	}
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * onlineCPUs * 100.0
	}
	return cpuPercent
}

func calculateCPUPercentWindows(v *types.StatsJSON) float64 {
	// Max number of 100ns intervals between the previous time read and now
	possIntervals := uint64(v.Read.Sub(v.PreRead).Nanoseconds()) // Start with number of ns intervals
	possIntervals /= 100                                         // Convert to number of 100ns intervals
	possIntervals *= uint64(v.NumProcs)                          // Multiple by the number of processors

	// Intervals used
	intervalsUsed := v.CPUStats.CPUUsage.TotalUsage - v.PreCPUStats.CPUUsage.TotalUsage

	// Percentage avoiding divide-by-zero
	if possIntervals > 0 {
		return float64(intervalsUsed) / float64(possIntervals) * 100.0
	}
	return 0.00
}

// calculateMemUsageUnixNoCache calculate memory usage of the container.
// Page cache is intentionally excluded to avoid misinterpretation of the output.
func calculateMemUsageUnixNoCache(mem types.MemoryStats) float64 {
	return float64(mem.Usage - mem.Stats["cache"])
}

func calculateMemPercentUnixNoCache(limit float64, usedNoCache float64) float64 {
	// MemoryStats.Limit will never be 0 unless the container is not running and we haven't
	// got any data from cgroup
	if limit != 0 {
		return usedNoCache / limit * 100.0
	}
	return 0
}
