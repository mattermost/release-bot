package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/go-github/v45/github"
	"github.com/mattermost/release-bot/client"
	"github.com/mattermost/release-bot/config"
	"github.com/mattermost/release-bot/store"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
)

type (
	mockClientCache struct {
	}
	mockAccessToken struct {
	}
)

func (t *mockAccessToken) IsExpired() bool {
	return false
}
func (t *mockAccessToken) GetToken() string {
	return "gh-12345678"
}
func (cc *mockClientCache) Get(installationID int64) (*github.Client, error) {
	token := "gh-12345678"
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.PostAppInstallationsAccessTokensByInstallationId,
			&github.InstallationToken{
				Token: &token,
			},
		),
	)
	return github.NewClient(mockedHTTPClient), nil
}

func (cc *mockClientCache) CreateToken(installationID int64) (client.AccessToken, error) {
	return &mockAccessToken{}, nil
}

func TestGithubHookHandlerFailureCases(t *testing.T) {
	eventContextStore := store.NewEventContextStore()
	config := &config.Config{
		Github: config.GithubConfig{
			IntegrationID: int64(100),
			PrivateKey:    "Private Key File",
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

	cc, _ := client.BuildFromConfig(config)
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
			IntegrationID: int64(100),
			PrivateKey:    "Private Key File",
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

	cc, _ := client.BuildFromConfig(config)
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
