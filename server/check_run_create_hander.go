package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/go-github/v45/github"
	"github.com/mattermost/release-bot/client"
	"github.com/mattermost/release-bot/store"
	log "github.com/sirupsen/logrus"
)

type checkRunCreateHandler struct {
	ClientCache       client.ClientCache
	EventContextStore store.EventContextStore
}

type CheckRunCreateRequest struct {
	Token   string `json:"token"`
	Name    string `json:"name"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

func (ccr *CheckRunCreateRequest) Log() {
	log.WithFields(log.Fields{
		"type":    "create_check_run",
		"name":    ccr.Name,
		"title":   ccr.Title,
		"summary": ccr.Summary,
	}).Info("Create check request")
}

type CreateCheckRunResponse struct {
	ID *int64 `json:"id"`
}

func (vh *checkRunCreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Info("Create Check Run Request is received!")
	if r.Method != "POST" {
		http.Error(w, "Only POST method is supported!", http.StatusMethodNotAllowed)
		return
	}
	var createCheckRunRequest CheckRunCreateRequest

	err := json.NewDecoder(r.Body).Decode(&createCheckRunRequest)
	if err != nil {
		log.WithError(err).Error("Invalid check run request!")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	createCheckRunRequest.Log()
	eventContext, err := vh.EventContextStore.Get(createCheckRunRequest.Token)
	if err != nil {
		log.WithError(err).Error("Missing Event Context!")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	status := "in_progress"
	opts := github.CreateCheckRunOptions{
		Name:    createCheckRunRequest.Name,
		HeadSHA: eventContext.GetCommitHash(),
		Status:  &status,
		Output: &github.CheckRunOutput{
			Title:   &createCheckRunRequest.Title,
			Summary: &createCheckRunRequest.Summary,
		},
	}
	client, err := vh.ClientCache.Get(eventContext.GetInstallationID())
	if err != nil {
		log.WithError(err).Error("Can not create installation Client.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	orgWithRepo := strings.Split(eventContext.GetRepository(), "/")
	checkRun, _, err := client.Checks.CreateCheckRun(r.Context(), orgWithRepo[0], orgWithRepo[1], opts)
	if err != nil {
		log.WithError(err).Error("Can not create check run.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json;charset=utf-8")
	createCheckRunResponse := &CreateCheckRunResponse{ID: checkRun.ID}
	response, _ := json.MarshalIndent(createCheckRunResponse, "", "  ")
	w.Header().Add("Content-Type", "application/json;charset=utf-8")
	w.Write(response)
}

func newCheckRunCreateHandler(cc client.ClientCache, eventContextStore store.EventContextStore) http.Handler {
	return &checkRunCreateHandler{
		ClientCache:       cc,
		EventContextStore: eventContextStore,
	}
}
