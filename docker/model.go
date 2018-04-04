package docker

import (
	"github.com/docker/docker/api/types"
	"github.com/rycus86/container-metrics/stats"
)

func convertStats(d *types.StatsJSON) *stats.Stats {
	s := stats.Stats{
		Id:   d.ID,
		Name: d.Name,

		CpuStats: stats.CpuStats{
			Total:  d.CPUStats.CPUUsage.TotalUsage,
			System: d.CPUStats.CPUUsage.UsageInKernelmode,
			User:   d.CPUStats.CPUUsage.UsageInUsermode,
		},

		MemoryStats: stats.MemoryStats{
			Total: d.MemoryStats.Limit,
			Free:  d.MemoryStats.Limit - d.MemoryStats.Usage,
		},

		IOStats: stats.IOStats{
			Read:    0,
			Written: 0,
		},

		NetworkStats: stats.NetworkStats{
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
		if ioEntry.Op == "Read" {
			s.IOStats.Read += ioEntry.Value
		} else if ioEntry.Op == "Write" {
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
