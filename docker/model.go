package docker

import (
	"github.com/docker/docker/api/types"
	"github.com/rycus86/container-metrics/model"
)

func convertStats(d *types.StatsJSON) *model.Stats {
	s := model.Stats{
		Id:   d.ID,
		Name: d.Name,

		CpuStats: model.CpuStats{
			Total:  d.CPUStats.CPUUsage.TotalUsage,
			System: d.CPUStats.CPUUsage.UsageInKernelmode,
			User:   d.CPUStats.CPUUsage.UsageInUsermode,
		},

		MemoryStats: model.MemoryStats{
			Total: d.MemoryStats.Limit,
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
