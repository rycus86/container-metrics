package model

type Stats struct {
	Id   string
	Name string

	CpuStats     CpuStats
	MemoryStats  MemoryStats
	IOStats      IOStats
	NetworkStats NetworkStats
}

type CpuStats struct {
	Total  uint64
	User   uint64
	System uint64
}

type MemoryStats struct {
	Total uint64
	Free  uint64
}

type IOStats struct {
	Read    uint64
	Written uint64
}

type NetworkStats struct {
	RxBytes   uint64
	RxPackets uint64
	RxDropped uint64
	RxErrors  uint64

	TxBytes   uint64
	TxPackets uint64
	TxDropped uint64
	TxErrors  uint64
}

type EngineStats struct {
	Host string

	Images            int
	Containers        int
	ContainersRunning int
	ContainersPaused  int
	ContainersStopped int
}
