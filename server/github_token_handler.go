package server

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/release-bot/client"
	"github.com/mattermost/release-bot/store"
	log "github.com/sirupsen/logrus"
)

type githubTokenHandler struct {
	ClientManager     client.GithubClientManager
	EventContextStore store.EventContextStore
}

type githubTokenRequest struct {
	BotToken string `json:"bot_token"`
}

func (gtr *githubTokenRequest) Log() {
	log.WithFields(log.Fields{
		"type":      "github_token",
		"bot_token": gtr.BotToken,
	}).Info("GitHub token request")
}

type githubTokenResponse struct {
	Token string `json:"token"`
}

func (gh *githubTokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Info("New Token Request is Received!")
	if r.Method != "POST" {
		http.Error(w, "Only POST method is supported!", http.StatusMethodNotAllowed)
		return
	}
	var request githubTokenRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.WithError(err).Error("Invalid GitHub Token request!")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	request.Log()
	if request.BotToken == "" {
		http.Error(w, "Provide Bot Token", http.StatusBadRequest)
		return
	}

	context, err := gh.EventContextStore.Get(request.BotToken)
	if err != nil {
		http.Error(w, "Invalid Bot Token", http.StatusBadRequest)
		return
	}
	accessToken, err := gh.ClientManager.CreateToken(context.GetInstallationID())
	if err != nil {
		log.WithError(err).Error("Unable to create GitHub Access Token!")
		http.Error(w, "Invalid Bot Token", http.StatusInternalServerError)
		return
	}
	response, _ := json.MarshalIndent(githubTokenResponse{Token: accessToken.GetToken()}, "", "  ")
	w.Header().Add("Content-Type", "application/json;charset=utf-8")
	w.Write(response)
}

func newGithubTokenHandler(clientManager client.GithubClientManager, eventContextStore store.EventContextStore) http.Handler {
	return &githubTokenHandler{
		ClientManager:     clientManager,
		EventContextStore: eventContextStore,
	}
}
