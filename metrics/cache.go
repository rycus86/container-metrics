package metrics

import (
	"sync"

	"github.com/rycus86/container-metrics/model"
)

var (
	statsCache = map[string]*model.Stats{}
	statsLock  = sync.Mutex{}
)

func getCached(id string) *model.Stats {
	statsLock.Lock()
	defer statsLock.Unlock()

	return statsCache[id]
}

func cacheStats(id string, stats *model.Stats) {
	statsLock.Lock()
	defer statsLock.Unlock()

	statsCache[id] = stats
}
