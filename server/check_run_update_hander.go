package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/go-github/v45/github"
	"github.com/mattermost/release-bot/client"
	"github.com/mattermost/release-bot/store"
	log "github.com/sirupsen/logrus"
)

type checkRunUpdateHandler struct {
	ClientCache       client.ClientCache
	EventContextStore store.EventContextStore
}

type CheckRunUpdateRequest struct {
	Token      string `json:"token"`
	CheckID    string `json:"checkId"`
	Name       string `json:"name"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	Summary    string `json:"summary"`
}

func (ucr *CheckRunUpdateRequest) Log() {
	log.WithFields(log.Fields{
		"type":       "finish_check_run",
		"checkId":    ucr.CheckID,
		"name":       ucr.Name,
		"title":      ucr.Title,
		"status":     ucr.Status,
		"conclusion": ucr.Conclusion,
		"summary":    ucr.Summary,
	}).Info("Update check request")
}

func (uh *checkRunUpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Info("Update Check Run Request is received!")
	if r.Method != "POST" {
		http.Error(w, "Only POST method is supported!", http.StatusMethodNotAllowed)
		return
	}
	var checkRunUpdateRequest CheckRunUpdateRequest

	err := json.NewDecoder(r.Body).Decode(&checkRunUpdateRequest)
	if err != nil {
		log.WithError(err).Error("Invalid finish check run request!")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	checkRunUpdateRequest.Log()
	eventContext, err := uh.EventContextStore.Get(checkRunUpdateRequest.Token)
	if err != nil {
		log.WithError(err).Error("Missing Event Context!")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	crOpts := github.UpdateCheckRunOptions{
		Name:       checkRunUpdateRequest.Name,
		Status:     &checkRunUpdateRequest.Status,
		Conclusion: &checkRunUpdateRequest.Conclusion,
		Output: &github.CheckRunOutput{
			Title:   &checkRunUpdateRequest.Title,
			Summary: &checkRunUpdateRequest.Summary},
	}

	client, err := uh.ClientCache.Get(eventContext.GetInstallationID())
	if err != nil {
		log.WithError(err).Error("Can not create installation Client.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	orgWithRepo := strings.Split(eventContext.GetRepository(), "/")
	checkID, err := strconv.Atoi(checkRunUpdateRequest.CheckID)
	checkRun, _, err := client.Checks.UpdateCheckRun(r.Context(), orgWithRepo[0], orgWithRepo[1], int64(checkID), crOpts)
	if err != nil {
		log.WithError(err).Error("Can not update check run.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json;charset=utf-8")
	createCheckRunResponse := &CreateCheckRunResponse{ID: checkRun.ID}
	response, _ := json.MarshalIndent(createCheckRunResponse, "", "  ")
	w.Header().Add("Content-Type", "application/json;charset=utf-8")
	w.Write(response)
}

func newCheckRunUpdateHandler(cc client.ClientCache, eventContextStore store.EventContextStore) http.Handler {
	return &checkRunUpdateHandler{
		ClientCache:       cc,
		EventContextStore: eventContextStore,
	}
}
