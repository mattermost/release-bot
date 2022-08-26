package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/mattermost/release-bot/client"
	"github.com/mattermost/release-bot/config"
	"github.com/mattermost/release-bot/store"
	"github.com/stretchr/testify/assert"
)

func TestGithubHookHandlerFailureCases(t *testing.T) {
	eventContextStore := store.NewEventContextStore()
	config := &config.Config{
		Github: config.GithubConfig{
			WebhookSecret: "ABC",
		},
		Queue: config.QueueConfig{
			Limit:   10,
			Workers: 2,
		},
		Server: config.HTTPConfig{
			BaseURL: "http://abc.com",
		},
	}

	cc, _ := client.NewClientCache(int64(100), "Private Key File")
	handler, _ := newGithubHookHandler(cc, config, eventContextStore)
	t.Run("Missing Event Type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, githubHandlerDefaultRoute, nil)
		req.Header.Add("X-GitHub-Delivery", "100")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		res := w.Result()
		assert.Equal(t, "400 Bad Request", res.Status)
	})
	t.Run("Missing Delivery Id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, githubHandlerDefaultRoute, nil)
		req.Header.Add("X-GitHub-Event", "push")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		res := w.Result()
		assert.Equal(t, "400 Bad Request", res.Status)
	})
	t.Run("Invalid Payload", func(t *testing.T) {
		request, _ := os.Open("testdata/check_run_update_request.json")
		req := httptest.NewRequest(http.MethodPost, githubHandlerDefaultRoute, request)
		req.Header.Add("X-GitHub-Event", "push")
		req.Header.Add("X-GitHub-Delivery", "100")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		res := w.Result()
		assert.Equal(t, "400 Bad Request", res.Status)
	})
}

func TestGithubHookHandlerRouteWithNoPipelineTrigger(t *testing.T) {
	eventContextStore := store.NewEventContextStore()
	config := &config.Config{
		Github: config.GithubConfig{
			WebhookSecret: "",
		},
		Queue: config.QueueConfig{
			Limit:   10,
			Workers: 2,
		},
		Server: config.HTTPConfig{
			BaseURL: "http://abc.com",
		},
	}

	cc, _ := client.NewClientCache(int64(100), "Private Key File")
	handler, _ := newGithubHookHandler(cc, config, eventContextStore)

	request, _ := os.Open("testdata/workflow_run_event_pr.json")
	req := httptest.NewRequest(http.MethodPost, githubHandlerDefaultRoute, request)
	req.Header.Add("X-GitHub-Event", "workflow_run")
	req.Header.Add("X-GitHub-Delivery", "100")
	req.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	res := w.Result()
	assert.Equal(t, "200 OK", res.Status)
	time.Sleep(10 * time.Millisecond)
}

func TestGithubHookHandlerRouteWithPipelineTrigger(t *testing.T) {
	eventContextStore := store.NewEventContextStore()
	config := &config.Config{
		Github: config.GithubConfig{
			WebhookSecret: "",
		},
		Queue: config.QueueConfig{
			Limit:   10,
			Workers: 2,
		},
		Server: config.HTTPConfig{
			BaseURL: "http://abc.com",
		},
		Pipelines: []config.PipelineConfig{
			{
				Organization: "mattermost",
				Repository:   "test",
				Workflow:     "docker.yaml",
				Conditions: []config.PipelineCondition{
					{
						Webhook: []string{"workflow_run"},
						Type:    "pr",
					},
				},
			},
		},
	}

	handler, _ := newGithubHookHandler(&mockClientCache{}, config, eventContextStore)

	request, _ := os.Open("testdata/workflow_run_event_pr.json")
	req := httptest.NewRequest(http.MethodPost, githubHandlerDefaultRoute, request)
	req.Header.Add("X-GitHub-Event", "workflow_run")
	req.Header.Add("X-GitHub-Delivery", "100")
	req.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	res := w.Result()
	assert.Equal(t, "200 OK", res.Status)
	time.Sleep(10 * time.Millisecond)
}