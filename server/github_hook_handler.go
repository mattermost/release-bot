package server

import (
	"context"
	"net/http"
	"strconv"

	"github.com/google/go-github/v45/github"
	"github.com/google/uuid"
	"github.com/mattermost/release-bot/client"
	"github.com/mattermost/release-bot/config"
	"github.com/mattermost/release-bot/metric"
	"github.com/mattermost/release-bot/model"
	"github.com/mattermost/release-bot/store"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type githubHookHandler struct {
	WebhookSecret     []byte
	Pipelines         []config.PipelineConfig
	BaseURL           string
	ClientManager     client.GithubClientManager
	EventContextStore store.EventContextStore
	Scheduler         Scheduler
}

func (gh *githubHookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metric.IncreaseCounter(metric.GithubHookCount, metric.TotalRequestCount)
	eventType := r.Header.Get("X-GitHub-Event")
	deliveryID := r.Header.Get("X-GitHub-Delivery")

	if eventType == "" {
		http.Error(w, "Missing Event Type", http.StatusBadRequest)
		metric.IncreaseCounter(metric.TotalFailureCount)
		return
	}
	if deliveryID == "" {
		http.Error(w, "Missing Delivery Id", http.StatusBadRequest)
		metric.IncreaseCounter(metric.TotalFailureCount)
		return
	}
	payload, err := github.ValidatePayload(r, gh.WebhookSecret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		metric.IncreaseCounter(metric.TotalFailureCount)
		return
	}

	log.WithField("type", eventType).Infof("%s is received!", eventType)
	gh.Scheduler.Schedule(dispatch{
		Processor:  gh.processEvent,
		EventType:  eventType,
		DeliveryID: deliveryID,
		Payload:    payload,
	})

	w.WriteHeader(http.StatusOK)
	metric.IncreaseCounter(metric.TotalSuccessCount)
}

func newGithubHookHandler(cc client.GithubClientManager, config *config.Config, eventContextStore store.EventContextStore) (http.Handler, error) {
	scheduler, err := NewGithubEventScheduler(config.Queue.Limit, config.Queue.Workers)
	if err != nil {
		return nil, errors.Wrap(err, "Scheduler error!")
	}
	return &githubHookHandler{
		WebhookSecret:     []byte(config.Github.WebhookSecret),
		Pipelines:         config.Pipelines,
		BaseURL:           config.Server.BaseURL,
		ClientManager:     cc,
		EventContextStore: eventContextStore,
		Scheduler:         scheduler,
	}, nil
}

func (gh *githubHookHandler) processEvent(eventType string, deliveryID string, payload []byte) {
	eventContext, err := model.ConvertPayloadToEventContext(eventType, payload)
	if err != nil {
		log.WithError(err).Error("Error occurred while deserializing request")
		return
	}
	if eventContext.GetType() == "tag" {
		metric.IncreaseCounter(metric.TagRequestCount)
	} else {
		metric.IncreaseCounter(metric.BranchRequestCount)
	}
	eventContext.Log()

	if "workflow_run" == eventContext.GetEvent() && "completed" == eventContext.GetAction() {
		if err := gh.ClientManager.RevokeToken(eventContext.GetRepository(), eventContext.GetWorkflowRunID()); err != nil {
			log.WithError(err).Error("Error occurred while revoking pipeline token")
		}
	}

	pipeline := model.GetTargetPipeline(eventContext, gh.Pipelines)

	if pipeline == nil {
		log.WithFields(log.Fields{
			"type":       eventContext.GetType(),
			"workflow":   eventContext.GetWorkflow(),
			"runId":      eventContext.GetWorkflowRunID(),
			"repository": eventContext.GetRepository(),
			"sha":        eventContext.GetCommitHash(),
		}).Info("No pipeline configured")
		return
	}

	if err := gh.triggerPipeline(context.Background(), eventContext, *pipeline); err != nil {
		log.WithError(err).Error("Error occurred while triggering pipeline request")
	}
}

func (h *githubHookHandler) triggerPipeline(ctx context.Context, eventContext model.EventContext, pipeline config.PipelineConfig) error {
	log.WithFields(log.Fields{
		"type":     "trigger",
		"org":      pipeline.Organization,
		"repo":     pipeline.Repository,
		"workflow": pipeline.Workflow,
	}).Info("Will trigger pipeline!")

	client, err := h.ClientManager.Get(eventContext.GetInstallationID())

	if err != nil {
		log.
			WithError(err).
			WithField("installation_id", eventContext.GetInstallationID()).
			Error("Can not find installation id at cache!")
		return err
	}
	token := uuid.New().String()
	h.EventContextStore.Store(eventContext, token)
	inputs := map[string]interface{}{
		"repository":    eventContext.GetRepository(),
		"name":          eventContext.GetName(),
		"workflowRunId": strconv.FormatInt(eventContext.GetWorkflowRunID(), 10),
		"commmitHash":   eventContext.GetCommitHash(),
		"fork":          strconv.FormatBool(eventContext.IsFork()),
		"type":          eventContext.GetType(),
		"botToken":      token,
		"botBaseUrl":    h.BaseURL,
	}
	deRequest := github.CreateWorkflowDispatchEventRequest{
		Ref:    "main",
		Inputs: inputs,
	}
	_, err = client.Actions.CreateWorkflowDispatchEventByFileName(ctx, pipeline.Organization, pipeline.Repository, pipeline.Workflow, deRequest)
	if err != nil {
		log.
			WithError(err).
			WithFields(log.Fields{
				"installation_id": eventContext.GetInstallationID(),
				"org":             pipeline.Organization,
				"repo":            pipeline.Repository,
				"workflow":        pipeline.Workflow,
			}).
			Error("Error occurred while triggering pipeline!")
		return errors.Wrap(err, "")
	}

	return nil
}
