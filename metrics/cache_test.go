package metrics

import (
	"testing"

	"github.com/rycus86/container-metrics/model"
)

func TestCache(t *testing.T) {
	cacheStats("sample", &model.Stats{
		Name: "for-testing",
	})

	missing := getCached("missing")
	if missing != nil {
		t.Error("Unexpected item in the cache:", missing)
	}

	valid := getCached("sample")
	if valid == nil {
		t.Error("Expected item not found")
	}
	if valid.Name != "for-testing" {
		t.Error("Unexpected name:", valid.Name)
	}
}
