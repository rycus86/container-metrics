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
		Read: d.Read,

		CpuStats: model.CpuStats{
			Total:  d.CPUStats.CPUUsage.TotalUsage,
			System: d.CPUStats.CPUUsage.UsageInKernelmode,
			User:   d.CPUStats.CPUUsage.UsageInUsermode,
		},

		MemoryStats: model.MemoryStats{
			Total: d.MemoryStats.Limit,
			Usage: d.MemoryStats.Usage,
			Free:  d.MemoryStats.Limit - d.MemoryStats.Usage,
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

func calculateCPUPercentUnix(previousCPU, previousSystem uint64, v *types.StatsJSON) float64 {
	var (
		cpuPercent = 0.0
		// calculate the change for the cpu usage of the container in between readings
		cpuDelta = float64(v.CPUStats.CPUUsage.TotalUsage) - float64(previousCPU)
		// calculate the change for the entire system between readings
		systemDelta = float64(v.CPUStats.SystemUsage) - float64(previousSystem)
		onlineCPUs = 0.0
		// TODO docker/docker-ce ?
		// onlineCPUs  = float64(v.CPUStats.OnlineCPUs)
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

func calculateBlockIO(blkio types.BlkioStats) (uint64, uint64) {
	var blkRead, blkWrite uint64
	for _, bioEntry := range blkio.IoServiceBytesRecursive {
		switch strings.ToLower(bioEntry.Op) {
		case "read":
			blkRead = blkRead + bioEntry.Value
		case "write":
			blkWrite = blkWrite + bioEntry.Value
		}
	}
	return blkRead, blkWrite
}

func calculateNetwork(network map[string]types.NetworkStats) (float64, float64) {
	var rx, tx float64

	for _, v := range network {
		rx += float64(v.RxBytes)
		tx += float64(v.TxBytes)
	}
	return rx, tx
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
