package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/go-github/v45/github"
	"github.com/mattermost/release-bot/client"
	"github.com/mattermost/release-bot/model"
	"github.com/mattermost/release-bot/store"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
)

type (
	mockClientCache struct {
	}
)

func (cc *mockClientCache) Get(installationID int64) (*github.Client, error) {
	checkRunId := int64(100)
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.PostReposCheckRunsByOwnerByRepo,
			github.CheckRun{
				ID: &checkRunId,
			},
		),
		mock.WithRequestMatch(
			mock.PatchReposCheckRunsByOwnerByRepoByCheckRunId,
			github.CheckRun{
				ID: &checkRunId,
			},
		),
	)
	return github.NewClient(mockedHTTPClient), nil
}

func TestCheckRunCreateRouteInvalidMethod(t *testing.T) {
	eventContextStore := store.NewEventContextStore()

	cc, _ := client.NewClientCache(int64(100), "Private Key File")
	handler := newCheckRunCreateHandler(cc, eventContextStore)
	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, createCheckRunDefaultRoute, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		res := w.Result()
		assert.Equal(t, "405 Method Not Allowed", res.Status)
	})
	t.Run("Bad Request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, createCheckRunDefaultRoute, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		res := w.Result()
		assert.Equal(t, "400 Bad Request", res.Status)
	})
}

func TestCheckRunCreateRoute(t *testing.T) {
	token := "token"
	eventContextStore := store.NewEventContextStore()
	eventContext := createWorkflowRunEvent(t)
	eventContextStore.Store(eventContext, token)

	handler := newCheckRunCreateHandler(&mockClientCache{}, eventContextStore)
	request, _ := os.Open("testdata/check_run_request.json")

	req := httptest.NewRequest(http.MethodPost, createCheckRunDefaultRoute, request)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	res := w.Result()
	assert.Equal(t, "200 OK", res.Status)
	defer res.Body.Close()
	assert.Equal(t, "application/json;charset=utf-8", res.Header.Get("Content-Type"))
	data, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	ccr := &CreateCheckRunResponse{}
	err = json.Unmarshal(data, ccr)
	assert.Nil(t, err)
	assert.Equal(t, int64(100), *ccr.ID)
}

func createWorkflowRunEvent(t *testing.T) model.EventContext {
	source, err := os.ReadFile("testdata/workflow_run_event_pr.json")
	if err != nil {
		t.Fatal("error reading source file:", err)
	}
	context, _ := model.ConvertPayloadToEventContext("workflow_run", source)
	return context
}
