package docker

type Container struct{
	Id string
	Names []string
	Image string
}

type ioStat struct{
	Operation string `json:"op"`
	Value int64
}

type Stats struct {
	Id string
	Name string
	Read string

	CpuStats struct{
		CpuUsage struct{
			TotalUsage int64 `json:"total_usage"`
			UserMode int64 `json:"usage_in_usermode"`
			KernelMode int64 `json:"usage_in_kernelmode"`
		} `json:"cpu_usage"`

		SystemCpuUsage int64 `json:"system_cpu_usage"`
	} `json:"cpu_stats"`

	MemoryStats struct{
		Usage int64
		MaxUsage int64 `json:"max_usage"`
		Limit int64
	} `json:"memory_stats"`

	IOStats struct{
		ServiceBytes []ioStat `json:"io_service_bytes_recursive"`
		Serviced []ioStat `json:"io_serviced_recursive"`
	} `json:"blkio_stats"`

	Networks map[string]struct{
		RxBytes int64 `json:"rx_bytes"`
		RxPackets int64 `json:"rx_packets"`
		RxDropped int64 `json:"rx_dropped"`
		RxErrors int64 `json:"rx_errors"`

		TxBytes int64 `json:"tx_bytes"`
		TxPackets int64 `json:"tx_packets"`
		TxDropped int64 `json:"tx_dropped"`
		TxErrors int64 `json:"tx_errors"`
	}

	PidStats struct{
		Current int64
	} `json:"pids_stats"`
}
