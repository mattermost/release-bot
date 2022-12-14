package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/mattermost/release-bot/client"
	"github.com/mattermost/release-bot/config"
	"github.com/mattermost/release-bot/metric"
	"github.com/mattermost/release-bot/store"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const (
	githubHandlerDefaultRoute          string = "/hook"
	healthHandlerDefaultRoute          string = "/healthz"
	tokenGenerationHandlerDefaultRoute string = "/token"
	metricsHandlerDetaultRoute         string = "/metrics"
)

type Server interface {
	Start(ctx context.Context) error
	Stop() error
}

type server struct {
	server *http.Server
}

func New() Server {
	return &server{}
}

func (s *server) Start(ctx context.Context) error {
	log.Info("Starting release bot server...")
	config, err := config.ReadConfig("config.yaml", "config")
	if err != nil {
		return errors.Wrap(err, "Invalid Configuration File.")
	}

	if err = s.registerHandlers(config); err != nil {
		return errors.Wrap(err, "Handler registration error.")
	}

	log.WithFields(log.Fields{
		"baseURL": config.Server.BaseURL,
		"address": config.Server.Address,
		"port":    config.Server.Port,
	}).Info("Starting release bot server...")
	metric.RegisterMetrics()
	s.server = &http.Server{Addr: fmt.Sprintf("%s:%d", config.Server.Address, config.Server.Port)}
	log.Info("Server Started")

	return s.server.ListenAndServe()
}

func (s *server) Stop() error {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(ctx)
	}
	return nil
}

func (s *server) registerHandlers(config *config.Config) error {
	eventContextStore := store.NewEventContextStore()

	cc, err := client.BuildFromConfig(config)
	if err != nil {
		log.WithError(err).Error("Can not create github client creator! Check configuration settings.")
		return err
	}
	githubHookHandler, err := newGithubHookHandler(cc, config, eventContextStore)
	if err != nil {
		log.WithError(err).Error("Can not create github request scheduler! Check configuration settings.")
		return err
	}
	http.Handle(healthHandlerDefaultRoute, newHealthHandler())
	http.Handle(githubHandlerDefaultRoute, githubHookHandler)
	http.Handle(tokenGenerationHandlerDefaultRoute, newGithubTokenHandler(cc, eventContextStore))
	http.Handle(metricsHandlerDetaultRoute, promhttp.Handler())
	return nil
}
