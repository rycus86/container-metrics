package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rycus86/container-metrics/container"
	"github.com/rycus86/container-metrics/docker"
	"github.com/rycus86/container-metrics/metrics"
	"github.com/rycus86/container-metrics/stats"
)

func recordMetrics(cli *docker.Client, containers []container.Container) {
	metrics.RecordAll(func (c *container.Container) *stats.Stats {
		stats, _ := cli.GetStats(c) // TODO error
		return stats
	})
	//for _, c := range containers {
	//	current := c
	//	go recordMetricsForOne(cli, &current)
	//}
}

func recordMetricsForOne(cli *docker.Client, c *container.Container) {
	stats, err := cli.GetStats(c)
	if err != nil {
		// TODO error handling?
		fmt.Println("Failed to get stats of", c.Name, "-", err)
		return
	}

	metrics.Record(c, stats)
}

func main() {
	cli, err := docker.NewClient()
	if err != nil {
		panic(err)
	}

	containers, err := cli.GetContainers()
	if err != nil {
		panic(err)
	}
	fmt.Println("container IDs:", containers)

	metrics.PrepareMetrics(containers)

	go recordMetrics(cli, containers)

	go metrics.Serve()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(3 * time.Second)

	for {
		select {

		case <-ticker.C:
			go recordMetrics(cli, containers)

		case s := <-signals:
			if s != syscall.SIGHUP {
				ticker.Stop()
				return
			} // TODO SIGHUP
		}
	}
}
