package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rycus86/container-metrics/container"
	"github.com/rycus86/container-metrics/stats"
	"net/http"
	"regexp"
	"time"
	"fmt"
)

type currentCollectors struct{
	cpuTotal   *prometheus.GaugeVec
	containers []container.Container
}

var current currentCollectors

func PrepareMetrics(containers []container.Container) {
	uniqueBaseLabels := map[string]string{"container_name": ""}

	nonLettersOrDigits := regexp.MustCompile("[^A-Za-z0-9]")

	for _, container := range containers {
		for labelName, _ := range container.Labels {
			normalizedName := nonLettersOrDigits.ReplaceAllString(labelName, "_")
			uniqueBaseLabels[normalizedName] = labelName
		}
	}

	baseLabels := []string{}
	for name := range(uniqueBaseLabels) {
		baseLabels = append(baseLabels, name)
	}

	currentContainers := make([]container.Container, len(containers), len(containers))

	for idx, container := range containers {
		computed := map[string]string{
			"container_name": container.Name,
		}

		for key, labelName := range(uniqueBaseLabels) {
			if labelName == "" {
				continue
			}

			value, ok := container.Labels[labelName]
			fmt.Println("check:", container.Name, labelName, value, ok)
			if ok {
				computed[key] = value
			} else {
				computed[key] = ""
			}
		}

		fmt.Println("computed:", computed)
		container.ComputedLabels = computed  // TODO synchronization
		fmt.Println("computed:", container.ComputedLabels)

		currentContainers[idx] = container
	}

	// now prepare the metrics

	cpuTotal := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "cpu_total",
		Help:        "CPU total",
	}, baseLabels)

	current = currentCollectors{
		containers: currentContainers,

		cpuTotal: cpuTotal,
	}
	//current.cpuTotal = cpuTotal  // TODO synchronization
	registerOrReplace(current.cpuTotal)
}

func registerOrReplace(c prometheus.Collector) error {
	err := prometheus.Register(c)
	if err == nil {
		return nil
	}

	if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
		prometheus.Unregister(are.ExistingCollector)
		err = prometheus.Register(c)
		return err
	} else {
		return err
	}
}

func RecordAll(statsFunc func(*container.Container) *stats.Stats) {
	containers := current.containers

	for _, item := range containers {
		current := item

		// TODO go RecordOne
		s := statsFunc(&current)
		Record(&current, s)
	}
}

func Record(c *container.Container, s *stats.Stats) {
	fmt.Println("computed:", c.ComputedLabels)
	// TODO synchronization
	current.cpuTotal.With(prometheus.Labels(c.ComputedLabels)).Set(
		float64(s.CpuStats.Total) / float64(time.Second),
	)

	// TODO work out label values and set
	// setGauge(c, "cpu_total", "CPU total", float64(s.CpuStats.Total)/float64(time.Second))
}

var gauges = map[string]prometheus.Gauge{}

func setGauge(c *container.Container, name, help string, value float64) {
	key := name + "@" + c.Id

	gauge, ok := gauges[key]
	if !ok {
		labels := prometheus.Labels{}
		labels["name"] = c.Name[1:]

		for labelName, labelValue := range c.Labels {
			nonLettersOrDigits := regexp.MustCompile("[^A-Za-z0-9]")
			normalizedName := nonLettersOrDigits.ReplaceAllString(labelName, "_")
			labels[normalizedName] = labelValue
		}

		gauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        name,
			Help:        help,
			ConstLabels: labels,
		})

		prometheus.Register(gauge)

		gauges[key] = gauge
	}

	println("record", key)

	gauge.Set(value)
}

func Serve() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}
