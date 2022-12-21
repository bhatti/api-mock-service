package metrics

import (
	"github.com/stretchr/testify/require"
	"math/rand"
	"sync"
	"testing"
)

func Test_ShouldRegisterAndAddMetrics(t *testing.T) {
	names := []string{"key1", "key2", "key3"}
	metrics := NewMetrics()
	var wg sync.WaitGroup
	for _, name := range names {
		name := name
		metrics.RegisterHistogram(name)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				metrics.AddHistogram(name, rand.Float64(), nil)
			}
		}()
	}
	wg.Wait()
	summary := metrics.Summary()
	for k, v := range summary {
		t.Log(k, v)
	}
	require.Equal(t, 6, len(summary))
}
