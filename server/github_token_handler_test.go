package server

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mattermost/release-bot/model"
	"github.com/mattermost/release-bot/store"
	"github.com/stretchr/testify/assert"
)

func TestGithubTokenHandlerInvalidRequests(t *testing.T) {
	eventContextStore := store.NewEventContextStore()
	eventContextStore.Store(createWorkflowRunEvent(t), "12345")

	handler := newGithubTokenHandler(&mockClientCache{}, eventContextStore)
	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, tokenGenerationHandlerDefaultRoute, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		res := w.Result()
		assert.Equal(t, "405 Method Not Allowed", res.Status)
	})
	t.Run("Missing Token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, tokenGenerationHandlerDefaultRoute, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		res := w.Result()
		assert.Equal(t, "400 Bad Request", res.Status)
	})
	t.Run("Invalid Token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, tokenGenerationHandlerDefaultRoute+"?token=1234", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		res := w.Result()
		assert.Equal(t, "400 Bad Request", res.Status)
	})
}

func TestGithubTokenHandler(t *testing.T) {
	eventContextStore := store.NewEventContextStore()
	eventContextStore.Store(createWorkflowRunEvent(t), "12345")

	handler := newGithubTokenHandler(&mockClientCache{}, eventContextStore)
	req := httptest.NewRequest(http.MethodGet, tokenGenerationHandlerDefaultRoute+"?token=12345", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	res := w.Result()
	assert.Equal(t, "200 OK", res.Status)
	defer res.Body.Close()
	assert.Equal(t, "text/plain;charset=utf-8", res.Header.Get("Content-Type"))
	data, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	assert.Equal(t, "gh-12345678", string(data))
}

func createWorkflowRunEvent(t *testing.T) model.EventContext {
	source, err := os.ReadFile("testdata/workflow_run_event_pr.json")
	if err != nil {
		t.Fatal("error reading source file:", err)
	}
	context, _ := model.ConvertPayloadToEventContext("workflow_run", source)
	return context
}
