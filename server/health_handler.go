package server

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/release-bot/metric"
	"github.com/mattermost/release-bot/version"
	log "github.com/sirupsen/logrus"
)

type healthHandler struct {
}

func (vh *healthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metric.IncreaseCounter(metric.HealthRequestCount, metric.TotalRequestCount)
	log.Debug("Version Request is received!")
	response, _ := json.MarshalIndent(version.Full(), "", "  ")
	w.Header().Add("Content-Type", "application/json;charset=utf-8")
	w.Write(response)
	metric.IncreaseCounter(metric.TotalSuccessCount)
}

func newHealthHandler() http.Handler {
	return &healthHandler{}
}
