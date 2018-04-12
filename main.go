package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rycus86/container-metrics/docker"
	"github.com/rycus86/container-metrics/logging"
	"github.com/rycus86/container-metrics/metrics"
	"github.com/rycus86/container-metrics/model"
)

type MetricsCollector struct {
	client   *docker.Client
	httpPort int
	ticker   *time.Ticker
	updates  chan []model.Container
}

func (mc *MetricsCollector) Setup() {
	containers, err := mc.client.GetContainers()
	if err != nil {
		log.Panicln("Failed to load the containers", err)
	}

	log.Println("Starting ...")

	metrics.PrepareMetrics(containers)

	if logging.IsVerboseEnabled() {
		log.Println("Metrics ready")
	}

	go metrics.Serve(mc.httpPort)

	go mc.client.ListenForEvents(mc.updates)

	if logging.IsVerboseEnabled() {
		log.Println("Now listening for Docker events")
	}
}

func (mc *MetricsCollector) Run() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Running ...")

	go mc.recordMetrics()

	for {
		select {

		case containers := <-mc.updates:
			if logging.IsDebugEnabled() {
				log.Println("Reloading with", len(containers), "containers")
			}

			metrics.PrepareMetrics(containers)

		case <-mc.ticker.C:
			if logging.IsVerboseEnabled() {
				log.Println("Recording metrics")
			}

			go mc.recordMetrics()

		case s := <-signals:
			if s != syscall.SIGHUP {
				mc.ticker.Stop()
				log.Println("Exiting ...")
				return
			} // TODO SIGHUP
		}
	}
}

func (mc *MetricsCollector) recordMetrics() {
	go mc.recordEngineStats()

	metrics.RecordAll(mc.statsFunc)
}

func (mc *MetricsCollector) recordEngineStats() {
	engineStats, err := mc.client.GetEngineStats()
	if err != nil {
		log.Println("Failed to collect engine stats")
		return
	}

	if logging.IsVerboseEnabled() {
		log.Printf("Engine stats: %+v\n", engineStats)
	}

	go metrics.RecordEngineStats(engineStats)
}

func (mc *MetricsCollector) statsFunc(c *model.Container) (*model.Stats, error) {
	stats, err := mc.client.GetStats(c)

	if err == nil && logging.IsVerboseEnabled() {
		log.Printf("Container stats for %s: %+v\n", c.Name, stats)
	}

	return stats, err
}

func main() {
	var (
		port     int
		interval time.Duration
		timeout  time.Duration
		debug    bool
		verbose  bool
	)

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// -p or -port
	flag.IntVar(&port, "port", 8080,
		"HTTP port to listen on")
	flag.IntVar(&port, "p", 8080,
		"HTTP port to listen on (shorthand)")
	// -i or -interval
	flag.DurationVar(&interval, "interval", 5*time.Second,
		"Interval for reading metrics from the engine")
	flag.DurationVar(&interval, "i", 5*time.Second,
		"Interval for reading metrics from the engine (shorthand)")
	// -t or -timeout
	flag.DurationVar(&timeout, "timeout", 30*time.Second,
		"Timeout for calling endpoints on the engine")
	flag.DurationVar(&timeout, "t", 30*time.Second,
		"Timeout for calling endpoints on the engine (shorthand)")
	// -d or -debug
	flag.BoolVar(&debug, "debug", false,
		"Enable debug messages")
	flag.BoolVar(&debug, "d", false,
		"Enable debug messages (shorthand)")
	// -v or -verbose
	flag.BoolVar(&verbose, "verbose", false,
		"Enable verbose messages - assumes debug")
	flag.BoolVar(&verbose, "v", false,
		"Enable verbose messages - assumes debug (shorthand)")

	flag.Parse()

	logging.Setup(debug, verbose)

	dockerClient, err := docker.NewClient(timeout)
	if err != nil {
		log.Panicln("Failed to connect to the Docker daemon", err)
	}

	collector := &MetricsCollector{
		client:   dockerClient,
		httpPort: port,
		ticker:   time.NewTicker(interval),
		updates:  make(chan []model.Container),
	}

	collector.Setup()
	collector.Run()
}
