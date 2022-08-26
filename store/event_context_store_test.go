package store

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/mattermost/release-bot/model"
	"github.com/stretchr/testify/assert"
)

func TestEventContextStore(t *testing.T) {
	t.Run("Context Store Test", func(t *testing.T) {
		token := "test"
		store := NewEventContextStore()
		event := createWorkflowRunEvent(t)
		store.Store(event, token)

		assert.NotEmpty(t, token)

		context, err := store.Get(token)
		assert.Nil(t, err)
		assert.Equal(t, event, context)
	})
	t.Run("Context Store Expiry Test", func(t *testing.T) {
		token := "test"
		cacheExpireInterval = 5 * time.Millisecond
		itemExpireDuration = time.Millisecond
		store := NewEventContextStore()
		event := createWorkflowRunEvent(t)
		store.Store(event, token)

		time.Sleep(10 * time.Millisecond)

		context, err := store.Get(token)
		assert.NotNil(t, err)
		assert.Nil(t, context)
	})
}
func createWorkflowRunEvent(t *testing.T) model.EventContext {
	source, err := os.ReadFile("testdata/workflow_run_event.json")
	if err != nil {
		t.Fatal("error reading source file:", err)
	}
	context, err := model.ConvertPayloadToEventContext("workflow_run", source)
	if err != nil {
		t.Fatal("error converting source file:", err)
	}
	if err = json.Unmarshal(source, &context); err != nil {
		t.Fatal("error reading source file:", err)
	}
	return context
}
