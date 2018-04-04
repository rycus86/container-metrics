package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rycus86/docker-metrics/docker"
)

func main() {
	cli, err := docker.NewClient()
	if err != nil {
		panic(err)
	}

	containers, err := cli.GetContainerIDs()
	if err != nil {
		panic(err)
	}
	fmt.Println("container IDs:", containers)

	for _, container := range containers {
		stats, err := cli.GetStats(container)
		if err != nil {
			fmt.Println("Failed to get stats for", container, err)
		} else {
			fmt.Println("Stats for", container, stats)
		}
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {

		case <-ticker.C:
			for _, c := range containers {
				go func() {
					stats, _ := cli.GetStats(c)
					fmt.Println("Stats:", stats)
				}()
			}

		case s := <-signals:
			if s != syscall.SIGHUP {
				ticker.Stop()
				return
			} // TODO SIGHUP
		}
	}
}
