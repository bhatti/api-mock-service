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
	MaxAdditionalLatency time.Duration `json:"max_additional_latency" mapstructure:"max_additional_latency"`
	// HTTPErrors to return for failure
	HTTPErrors []int        `json:"http_errors" mapstructure:"http_errors"`
	rnd        *rand.Rand   `json:"-" mapstructure:"-"`
	lock       sync.RWMutex `json:"-" mapstructure:"-"`
}

func (gc *GroupConfig) GetHTTPStatus() int {
	if !gc.checkInit() {
		return 0
	}
	var prob = 1.0 / float64(gc.MeanTimeBetweenFailure)

	// Sample uniformly over [0,1)
	sample := gc.rnd.Float64()

	if prob < sample {
		return 0
	}
	return gc.HTTPErrors[gc.rnd.Intn(len(gc.HTTPErrors))]
}

func (gc *GroupConfig) GetDelayLatency() time.Duration {
	if !gc.checkInit() {
		return 0
	}

	var prob = 1.0 / float64(gc.MeanTimeBetweenFailure)

	// Sample uniformly over [0,1)
	sample := gc.rnd.Float64()

	if prob >= sample {
		additional := float64(gc.rnd.Intn(int(gc.MaxAdditionalLatency.Seconds()*10)) + 1)
		d := time.Second * time.Duration(sample*additional)
		if d.Seconds() > gc.MaxAdditionalLatency.Seconds() {
			d = gc.MaxAdditionalLatency
		}
		return d
	}
	return 0
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
	if gc.MaxAdditionalLatency <= 0 {
		gc.MaxAdditionalLatency = time.Second * 2
	}
	if len(gc.HTTPErrors) == 0 {
		gc.HTTPErrors = []int{400, 401, 500}
	}
	if gc.rnd == nil {
		gc.rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	return true
}
