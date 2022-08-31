package server

import (
	"encoding/json"
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
	type test struct {
		testFile   string
		httpMethod string
		want       string
	}

	eventContextStore := store.NewEventContextStore()
	eventContextStore.Store(createWorkflowRunEvent(t), "12345")
	handler := newGithubTokenHandler(&mockClientCache{}, eventContextStore)
	tests := []test{
		{testFile: "", httpMethod: http.MethodGet, want: "405 Method Not Allowed"},
		{testFile: "github_token_request_missing_token.json", httpMethod: http.MethodPost, want: "400 Bad Request"},
		{testFile: "github_token_request_missing_repo.json", httpMethod: http.MethodPost, want: "400 Bad Request"},
		{testFile: "github_token_request_missing_run_id.json", httpMethod: http.MethodPost, want: "400 Bad Request"},
		{testFile: "github_token_request_invalid_token.json", httpMethod: http.MethodPost, want: "400 Bad Request"},
	}

	for _, tc := range tests {
		var request *os.File
		if tc.testFile != "" {
			request, _ = os.Open("testdata/github_token_request_missing_token.json")
		}
		req := httptest.NewRequest(tc.httpMethod, tokenGenerationHandlerDefaultRoute, request)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		res := w.Result()
		assert.Equal(t, tc.want, res.Status)
	}
}

func TestGithubTokenHandler(t *testing.T) {
	eventContextStore := store.NewEventContextStore()
	eventContextStore.Store(createWorkflowRunEvent(t), "bot_token")

	handler := newGithubTokenHandler(&mockClientCache{}, eventContextStore)
	request, _ := os.Open("testdata/github_token_request.json")
	req := httptest.NewRequest(http.MethodPost, tokenGenerationHandlerDefaultRoute, request)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	res := w.Result()
	assert.Equal(t, "200 OK", res.Status)
	defer res.Body.Close()
	assert.Equal(t, "application/json;charset=utf-8", res.Header.Get("Content-Type"))
	data, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	gtr := &githubTokenResponse{}
	err = json.Unmarshal(data, gtr)
	assert.Nil(t, err)
	assert.Equal(t, "gh-12345678", gtr.Token)
}

func createWorkflowRunEvent(t *testing.T) model.EventContext {
	source, err := os.ReadFile("testdata/workflow_run_event_pr.json")
	if err != nil {
		t.Fatal("error reading source file:", err)
	}
	context, _ := model.ConvertPayloadToEventContext("workflow_run", source)
	return context
}
