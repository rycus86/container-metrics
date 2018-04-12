package metrics

import "sync"

var (
	currentMetrics *PrometheusMetrics
	currentLock    sync.Mutex
)

func getCurrent() *PrometheusMetrics {
	currentLock.Lock()
	defer currentLock.Unlock()

	return currentMetrics
}

func setCurrent(pm *PrometheusMetrics) {
	currentLock.Lock()
	defer currentLock.Unlock()

	currentMetrics = pm
}
