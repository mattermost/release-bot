package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Counter int
type Gauge int

const (
	TotalRequestCount Counter = iota
	TotalSuccessCount
	TotalFailureCount
	HealthRequestCount
	TokenRequestCount
	GithubHookCount
	TagRequestCount
	BranchRequestCount
)

const (
	QueuedRequests Gauge = iota
	ActiveWorkers
)

type metricCollector struct {
	counters map[Counter]prometheus.Counter
	gauges   map[Gauge]prometheus.Gauge
}

var collector *metricCollector

func RegisterMetrics() {
	collector = &metricCollector{
		counters: make(map[Counter]prometheus.Counter),
		gauges:   make(map[Gauge]prometheus.Gauge),
	}
	collector.counters[TotalRequestCount] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "release_bot",
		Subsystem: "request",
		Name:      "total",
		Help:      "The total number of requests",
	})
	collector.counters[TotalSuccessCount] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "release_bot",
		Subsystem: "request",
		Name:      "success",
		Help:      "The total number of successful requests",
	})
	collector.counters[TotalFailureCount] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "release_bot",
		Subsystem: "request",
		Name:      "failure",
		Help:      "The total number of failed requests",
	})
	collector.counters[HealthRequestCount] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "release_bot",
		Subsystem: "request",
		Name:      "health",
		Help:      "The total number of health requests",
	})
	collector.counters[TokenRequestCount] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "release_bot",
		Subsystem: "request",
		Name:      "token",
		Help:      "The total number of token requests",
	})
	collector.counters[GithubHookCount] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "release_bot",
		Subsystem: "request",
		Name:      "hook",
		Help:      "The total number of github hook requests",
	})
	collector.counters[TagRequestCount] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "release_bot",
		Subsystem: "request",
		Name:      "tag",
		Help:      "The total number of tag requests",
	})
	collector.counters[BranchRequestCount] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "release_bot",
		Subsystem: "request",
		Name:      "branch",
		Help:      "The total number of branch requests",
	})
	collector.gauges[QueuedRequests] = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "release_bot",
		Subsystem: "queue",
		Name:      "waiting",
		Help:      "The total number of reuqests queued",
	})
	collector.gauges[ActiveWorkers] = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "release_bot",
		Subsystem: "queue",
		Name:      "workers",
		Help:      "The total number of active workers",
	})

	for _, counter := range collector.counters {
		prometheus.MustRegister(counter)
	}
	for _, gauge := range collector.gauges {
		prometheus.MustRegister(gauge)
	}
}

func IncreaseCounter(counters ...Counter) {
	go func() {
		for _, counter := range counters {
			if c, ok := collector.counters[counter]; ok {
				c.Inc()
			}
		}
	}()
}

func IncreaseGauge(gauges ...Gauge) {
	go func() {
		for _, gauge := range gauges {
			if c, ok := collector.gauges[gauge]; ok {
				c.Inc()
			}
		}
	}()
}
func DecreaseGauge(gauges ...Gauge) {
	go func() {
		for _, gauge := range gauges {
			if c, ok := collector.gauges[gauge]; ok {
				c.Dec()
			}
		}
	}()
}
