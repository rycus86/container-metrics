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

var cli *docker.Client

func recordMetrics() {
	metrics.RecordAll(statsFunc)
}

func statsFunc(c *container.Container) (*stats.Stats, error) {
	return cli.GetStats(c)
}

func main() {
	dockerClient, err := docker.NewClient()
	if err != nil {
		panic(err)
	}

	cli = dockerClient

	containers, err := cli.GetContainers()
	if err != nil {
		panic(err)
	}
	fmt.Println("container IDs:", containers)

	metrics.PrepareMetrics(containers)

	updates := make(chan []container.Container, 10)

	go cli.ListenForEvents(updates)

	go recordMetrics()

	go metrics.Serve()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(3 * time.Second)

	for {
		select {

		case updatedContainers := <-updates:
			fmt.Println("Updating with", updatedContainers)
			metrics.PrepareMetrics(updatedContainers)

		case <-ticker.C:
			go recordMetrics()

		case s := <-signals:
			if s != syscall.SIGHUP {
				ticker.Stop()
				return
			} // TODO SIGHUP
		}
	}
}
