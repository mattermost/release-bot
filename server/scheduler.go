package server

import (
	"github.com/mattermost/release-bot/metric"
	"github.com/pkg/errors"
)

type GithubEventProcessor func(eventType string, deliveryID string, payload []byte)

type dispatch struct {
	Processor GithubEventProcessor

	EventType  string
	DeliveryID string
	Payload    []byte
}

type Scheduler interface {
	Schedule(d dispatch)
}

type scheduler struct {
	queue chan dispatch
}

func NewGithubEventScheduler(queueSize int, workers int) (Scheduler, error) {
	if queueSize < 0 {
		return nil, errors.New("Queue size must be non-negative")
	}
	if workers < 1 {
		return nil, errors.New("Worker count must be positive")
	}

	s := &scheduler{queue: make(chan dispatch, queueSize)}
	for i := 0; i < workers; i++ {
		go func() {
			for d := range s.queue {
				metric.DecreaseGauge(metric.QueuedRequests)
				metric.IncreaseGauge(metric.ActiveWorkers)
				d.Processor(d.EventType, d.DeliveryID, d.Payload)
				metric.DecreaseGauge(metric.ActiveWorkers)
			}
		}()
	}

	return s, nil
}

func (s *scheduler) Schedule(d dispatch) {
	metric.IncreaseGauge(metric.QueuedRequests)
	s.queue <- d
}
