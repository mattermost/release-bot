package server

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/release-bot/client"
	"github.com/mattermost/release-bot/metric"
	"github.com/mattermost/release-bot/store"
	log "github.com/sirupsen/logrus"
)

type githubTokenHandler struct {
	ClientManager     client.GithubClientManager
	EventContextStore store.EventContextStore
}

type githubTokenRequest struct {
	BotToken   string `json:"bot_token"`
	Repository string `json:"repository"`
	RunID      int64  `json:"run_id"`
}

func (gtr *githubTokenRequest) Log() {
	log.WithFields(log.Fields{
		"type":       "github_token",
		"bot_token":  gtr.BotToken,
		"repository": gtr.Repository,
		"run_id":     gtr.RunID,
	}).Info("GitHub token request")
}

type githubTokenResponse struct {
	Token string `json:"token"`
}

func (gh *githubTokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metric.IncreaseCounter(metric.TokenRequestCount, metric.TotalRequestCount)
	log.Info("New Token Request is Received!")
	if r.Method != "POST" {
		http.Error(w, "Only POST method is supported!", http.StatusMethodNotAllowed)
		metric.IncreaseCounter(metric.TotalFailureCount)
		return
	}
	var request githubTokenRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.WithError(err).Error("Invalid GitHub Token request!")
		http.Error(w, err.Error(), http.StatusBadRequest)
		metric.IncreaseCounter(metric.TotalFailureCount)
		return
	}

	request.Log()

	if request.BotToken == "" {
		http.Error(w, "Provide Bot Token", http.StatusBadRequest)
		metric.IncreaseCounter(metric.TotalFailureCount)
		return
	}
	if request.Repository == "" {
		http.Error(w, "Provide Repository", http.StatusBadRequest)
		metric.IncreaseCounter(metric.TotalFailureCount)
		return
	}
	if request.RunID == 0 {
		http.Error(w, "Provide Workflow Run ID", http.StatusBadRequest)
		metric.IncreaseCounter(metric.TotalFailureCount)
		return
	}
	context, err := gh.EventContextStore.Get(request.BotToken)
	if err != nil {
		http.Error(w, "Invalid Bot Token", http.StatusBadRequest)
		metric.IncreaseCounter(metric.TotalFailureCount)
		return
	}
	accessToken, err := gh.ClientManager.CreateToken(
		request.Repository,
		request.RunID,
		context.GetInstallationID(),
	)
	if err != nil {
		log.WithError(err).Error("Unable to create GitHub Access Token!")
		http.Error(w, "Invalid Bot Token", http.StatusInternalServerError)
		metric.IncreaseCounter(metric.TotalFailureCount)
		return
	}
	response, _ := json.MarshalIndent(githubTokenResponse{Token: accessToken.GetToken()}, "", "  ")
	w.Header().Add("Content-Type", "application/json;charset=utf-8")
	w.Write(response)
	metric.IncreaseCounter(metric.TotalSuccessCount)
}

func newGithubTokenHandler(clientManager client.GithubClientManager, eventContextStore store.EventContextStore) http.Handler {
	return &githubTokenHandler{
		ClientManager:     clientManager,
		EventContextStore: eventContextStore,
	}
}
