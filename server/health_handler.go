package server

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/release-bot/version"
	log "github.com/sirupsen/logrus"
)

type healthHandler struct {
}

func (vh *healthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Info("Version Request is received!")
	response, _ := json.MarshalIndent(version.Full(), "", "  ")
	w.Header().Add("Content-Type", "application/json;charset=utf-8")
	w.Write(response)
}

func newHealthHandler() http.Handler {
	return &healthHandler{}
}
