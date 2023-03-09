package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	dto "github.com/prometheus/client_model/go"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strings"
	"sync"
)

// Metrics structure
type Metrics struct {
	registry   *prometheus.Registry
	histograms map[string]*prometheus.Histogram
	lock       sync.RWMutex
}

// NewMetrics constructor
func NewMetrics() *Metrics {
	metrics := &Metrics{
		registry:   prometheus.NewRegistry(),
		histograms: make(map[string]*prometheus.Histogram),
	}
	if err := metrics.registry.Register(collectors.NewGoCollector()); err != nil {
		log.WithFields(log.Fields{
			"Component": "Metrics",
			"Error":     err,
		}).Warn("failed to register GO collector")
	}
	if err := metrics.registry.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
		log.WithFields(log.Fields{
			"Component": "Metrics",
			"Error":     err,
		}).Warn("failed to register process collector")
	}
	return metrics
}

// RegisterHistogram registers a new metric
func (m *Metrics) RegisterHistogram(name string) {
	name = sanitizeName(name)
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.histograms[name] != nil {
		return
	}
	requestDurations := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    name + "_duration_seconds",
		Help:    name,
		Buckets: prometheus.ExponentialBuckets(0.1, 1.5, 5),
		//Buckets:                     prometheus.LinearBuckets(normMean-5*normDomain, .5*normDomain, 20),
		NativeHistogramBucketFactor: 1.1,
	})
	//requestDurations := prometheus.NewSummaryVec(
	//	prometheus.SummaryOpts{
	//		Name:       name + "_request_duration_seconds",
	//		Help:       name,
	//		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	//	},
	//	[]string{"service"},
	//)
	if err := m.registry.Register(requestDurations); err != nil {
		log.WithFields(log.Fields{
			"Component": "Metrics",
			"Error":     err,
			"Name":      name,
		}).Warn("failed to register collector")
	}
	m.histograms[name] = &requestDurations
}

// AddHistogram adds latency
func (m *Metrics) AddHistogram(name string, value float64, labels map[string]string) {
	name = sanitizeName(name)
	m.lock.RLock()
	defer m.lock.RUnlock()
	requestDurations := m.histograms[name]
	if requestDurations == nil {
		return
	}
	(*requestDurations).(prometheus.ExemplarObserver).ObserveWithExemplar(value, labels)
}

// Summary returns histogram summary
func (m *Metrics) Summary() map[string]float64 {
	m.lock.RLock()
	defer m.lock.RUnlock()
	res := make(map[string]float64)
	metrics, _ := m.registry.Gather()
	for _, metric := range metrics {
		for _, m := range metric.Metric {
			if metric.Name != nil && metric.Type != nil &&
				*metric.Type == dto.MetricType_HISTOGRAM {
				if m.Histogram.SampleSum != nil {
					res[*metric.Name] = *m.Histogram.SampleSum
					res[strings.ReplaceAll(*metric.Name, "duration_seconds", "counts")] = float64(*m.Histogram.SampleCount)
				}
			}
		}
	}
	return res
}

func sanitizeName(name string) string {
	if re, err := regexp.Compile(`[^a-zA-Z0-9_]`); err == nil {
		name = re.ReplaceAllString(name, "")
	}
	return name
}
