package server

import (
	"net/http"

	"github.com/mattermost/release-bot/client"
	"github.com/mattermost/release-bot/store"
	log "github.com/sirupsen/logrus"
)

type githubTokenHandler struct {
	ClientManager     client.GithubClientManager
	EventContextStore store.EventContextStore
}

func (gh *githubTokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Info("New Token Request is Received!")
	if r.Method != "GET" {
		http.Error(w, "Only GET method is supported!", http.StatusMethodNotAllowed)
		return
	}
	botToken := r.URL.Query().Get("token")
	if botToken == "" {
		http.Error(w, "Provide Bot Token", http.StatusBadRequest)
		return
	}
	context, err := gh.EventContextStore.Get(botToken)
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
	w.Header().Add("Content-Type", "text/plain;charset=utf-8")
	w.Write([]byte(accessToken.GetToken()))
}

func newGithubTokenHandler(clientManager client.GithubClientManager, eventContextStore store.EventContextStore) http.Handler {
	return &githubTokenHandler{
		ClientManager:     clientManager,
		EventContextStore: eventContextStore,
	}
}
