package types

import (
	"math/rand"
	"sync"
	"time"
)

// GroupConfig for group configuration
type GroupConfig struct {
	// Variables to set for templates
	Variables map[string]string `json:"variables" mapstructure:"variables"`
	// ChaosEnabled to enable faults and delays
	ChaosEnabled bool `json:"chaos_enabled" mapstructure:"chaos_enabled"`
	// MeanTimeBetweenFailure for failure
	MeanTimeBetweenFailure float64 `json:"mean_time_between_failure" mapstructure:"mean_time_between_failure"`
	// MeanTimeBetweenAdditionalLatency for adding delay
	MeanTimeBetweenAdditionalLatency float64 `json:"mean_time_between_additional_latency" mapstructure:"mean_time_between_additional_latency"`
	// MaxAdditionalLatency for max delay
	MaxAdditionalLatencySecs float64 `json:"max_additional_latency_secs" mapstructure:"max_additional_latency_secs"`
	// HTTPErrors to return for failure
	HTTPErrors []int `json:"http_errors" mapstructure:"http_errors"`
	rnd        *rand.Rand
	lock       sync.RWMutex
}

// GetHTTPStatus accessor
func (gc *GroupConfig) GetHTTPStatus() int {
	if !gc.checkInit() {
		return 0
	}
	if !gc.checkProbability(gc.MeanTimeBetweenFailure) {
		return 0
	}
	return gc.HTTPErrors[gc.rnd.Intn(len(gc.HTTPErrors))]
}

func (gc *GroupConfig) checkProbability(mean float64) bool {
	var prob = 1.0 / mean

	// Sample uniformly over [0,1)
	sample := gc.rnd.Float64()

	return prob < sample
}

// GetDelayLatency calculates latency
func (gc *GroupConfig) GetDelayLatency() time.Duration {
	if !gc.checkInit() {
		return 0
	}
	if !gc.checkProbability(gc.MeanTimeBetweenFailure) {
		return 0
	}
	additional := float64(gc.rnd.Intn(int(gc.MaxAdditionalLatencySecs*100)) + 1)
	sample := gc.rnd.Float64() + 0.1
	d := time.Second * time.Duration(sample*additional)
	if d.Seconds() > gc.MaxAdditionalLatencySecs {
		d = time.Second * time.Duration(gc.MaxAdditionalLatencySecs)
	}
	return d
}

func (gc *GroupConfig) checkInit() bool {
	gc.lock.Lock()
	defer gc.lock.Unlock()
	if !gc.ChaosEnabled {
		return false
	}
	if gc.MeanTimeBetweenFailure <= 0 {
		gc.MeanTimeBetweenFailure = 2
	}
	if gc.MeanTimeBetweenAdditionalLatency <= 0 {
		gc.MeanTimeBetweenAdditionalLatency = 3
	}
	if gc.MaxAdditionalLatencySecs <= 0 {
		gc.MaxAdditionalLatencySecs = 2
	}
	if len(gc.HTTPErrors) == 0 {
		gc.HTTPErrors = []int{400, 401, 500}
	}
	if gc.rnd == nil {
		gc.rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	return true
}
