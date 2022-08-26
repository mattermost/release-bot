package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestData struct {
	Called bool
}

func TestScheduler(t *testing.T) {
	t.Run("Negative Queue Length", func(t *testing.T) {
		s, err := NewGithubEventScheduler(-1, 1)
		assert.Nil(t, s)
		assert.Error(t, err)
	})
	t.Run("Zero Workers", func(t *testing.T) {
		s, err := NewGithubEventScheduler(100, 0)
		assert.Nil(t, s)
		assert.Error(t, err)
	})
	t.Run("Valid Scheduler Test", func(t *testing.T) {
		done := make(chan bool)
		test := &TestData{Called: false}
		s, _ := NewGithubEventScheduler(1, 1)
		d := dispatch{
			Processor: func(eventType string, deliveryID string, payload []byte) {
				test.Called = true
				done <- true
			},
		}
		s.Schedule(d)
		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
			panic("timeout")
		}
		assert.True(t, test.Called)
	})
}
