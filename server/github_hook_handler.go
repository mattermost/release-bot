package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v45/github"
	"github.com/mattermost/release-bot/client"
	"github.com/mattermost/release-bot/config"
	"github.com/mattermost/release-bot/metric"
	"github.com/mattermost/release-bot/model"
	"github.com/mattermost/release-bot/store"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	// Commit statuses
	statePending = "pending"
	stateSuccess = "success"
	stateError   = "error"
	// stateFailure = "failure"

	// Available Workflow Run statuses
	// https://docs.github.com/en/webhooks-and-events/webhooks/webhook-events-and-payloads?actionType=requested#workflow_run
	// requested, in_progress, completed, queued, pending, waiting
	statusCompleted = "completed"
	// statusRequested  = "requested"
	// statusInProgress = "in_progress"
	// statusQueued     = "queued"
	// statusPending    = "pending"
	// statusWaiting    = "waiting"
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
	// Adding event for processing
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

// Process event
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

	// if "workflow_run" == eventContext.GetEvent() && "completed" == eventContext.GetAction() {
	// 	if err := gh.ClientManager.RevokeToken(eventContext.GetRepository(), eventContext.GetWorkflowRunID()); err != nil {
	// 		log.WithError(err).Error("Error occurred while revoking pipeline token")
	// 	}
	// }

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

func (gh *githubHookHandler) triggerPipeline(ctx context.Context, eventContext model.EventContext, pipeline config.PipelineConfig) error {
	log.WithFields(log.Fields{
		"type":     "trigger",
		"org":      pipeline.Organization,
		"repo":     pipeline.Repository,
		"workflow": pipeline.Workflow,
		"branch":   pipeline.TargetBranch,
	}).Info("Will trigger pipeline!")

	client, err := gh.ClientManager.Get(eventContext.GetInstallationID())
	if err != nil {
		log.WithError(err).
			WithField("installation_id", eventContext.GetInstallationID()).Error("Can not find installation id at cache!")
		return err
	}

	// This needs to be parameterized
	inputs := map[string]interface{}{
		"logLevel":    "info",
		"tags":        true,
		"environment": eventContext.GetCommitHash(),
	}

	deRequest := github.CreateWorkflowDispatchEventRequest{
		Ref:    pipeline.TargetBranch,
		Inputs: inputs,
	}

	_, err = client.Actions.CreateWorkflowDispatchEventByFileName(ctx, pipeline.Organization, pipeline.Repository, pipeline.Workflow, deRequest)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"org":      pipeline.Organization,
				"repo":     pipeline.Repository,
				"workflow": pipeline.Workflow,
				"branch":   pipeline.TargetBranch,
			}).Error("Error occurred while triggering pipeline!")
		return errors.Wrap(err, "")
	}

	if err := gh.waitForWorkflow(ctx, eventContext, pipeline); err != nil {
		return err
	}

	return nil
}

func (gh *githubHookHandler) waitForWorkflow(ctx context.Context, eventContext model.EventContext, pipeline config.PipelineConfig) error {
	wf, _ := gh.getCreatedPipeline(ctx, eventContext, pipeline)
	gh.sendStatus(ctx, eventContext, pipeline, wf, statePending)
	state, _ := gh.waitForWorkflowStatus(ctx, eventContext, pipeline, wf, 15*time.Second)
	gh.sendStatus(ctx, eventContext, pipeline, wf, state)

	return nil
}

func (gh *githubHookHandler) getCreatedPipeline(ctx context.Context, eventContext model.EventContext, pipeline config.PipelineConfig) (*github.WorkflowRun, error) {
	log.WithFields(log.Fields{
		"type":     "get",
		"org":      pipeline.Organization,
		"repo":     pipeline.Repository,
		"workflow": pipeline.Workflow,
	}).Info("Getting current running Workflow")

	client, err := gh.ClientManager.Get(eventContext.GetInstallationID())
	if err != nil {
		log.WithError(err).
			WithField("installation_id", eventContext.GetInstallationID()).Error("Cannot find installation id at cache!")
		return nil, err
	}

	log.Info("Waiting for 5 seconds for the Workflow to settle")
	time.Sleep(5 * time.Second)

	listRequest := github.ListWorkflowRunsOptions{
		Branch:  pipeline.TargetBranch,
		Event:   "workflow_dispatch",
		Created: fmt.Sprintf(">=%s", "2006-01-02"),
		ListOptions: github.ListOptions{
			PerPage: 1,
		},
	}

	// It's always the last one since it's the one triggered . We will have to create a state within the webhook to keep track
	workflowRuns, _, err := client.Actions.ListWorkflowRunsByFileName(ctx, pipeline.Organization, pipeline.Repository, pipeline.Workflow, &listRequest)
	if err != nil {
		return nil, err
	}
	log.Info(fmt.Sprintf("Found %d workflows", *workflowRuns.TotalCount))

	return workflowRuns.WorkflowRuns[0], nil
}

func (gh *githubHookHandler) sendStatus(ctx context.Context, eventContext model.EventContext, pipeline config.PipelineConfig, wf *github.WorkflowRun, state string) {
	client, err := gh.ClientManager.Get(eventContext.GetInstallationID())
	if err != nil {
		log.WithError(err).
			WithField("installation_id", eventContext.GetInstallationID()).Error("Cannot find installation id at cache!")
	}

	var status *github.RepoStatus
	var finalState string
	var finalDescription string

	// The conclusion of workflow run
	// Can be one of: success, failure, neutral, cancelled, timed_out, action_required, stale, null, skipped, startup_failure
	switch state {
	case stateSuccess:
		finalState = stateSuccess
		finalDescription = fmt.Sprintf("Private workflow %s on %s/%s finished with %s", pipeline.Workflow, pipeline.Organization, pipeline.Repository, stateSuccess)

	case statePending:
		finalState = statePending
		finalDescription = fmt.Sprintf("Private workflow %s on %s/%s started", pipeline.Workflow, pipeline.Organization, pipeline.Repository)

	default:
		finalState = stateError
		finalDescription = fmt.Sprintf("Private workflow %s on %s/%s finished with %s", pipeline.Workflow, pipeline.Organization, pipeline.Repository, state)
	}

	status = &github.RepoStatus{
		State:       github.String(finalState),
		Context:     github.String(pipeline.Context),
		Description: github.String(finalDescription),
		TargetURL:   github.String(*wf.HTMLURL),
	}

	_, _, err = client.Repositories.CreateStatus(ctx, strings.Split(eventContext.GetRepository(), "/")[0], strings.Split(eventContext.GetRepository(), "/")[1], eventContext.GetCommitHash(), status)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"org":    strings.Split(eventContext.GetRepository(), "/")[0],
				"repo":   strings.Split(eventContext.GetRepository(), "/")[1],
				"sha":    eventContext.GetCommitHash(),
				"status": status,
			}).Error("Failed to create status update for the referenced commit hash")
	}
}

func (gh *githubHookHandler) waitForWorkflowStatus(ctx context.Context, eventContext model.EventContext, pipeline config.PipelineConfig, wf *github.WorkflowRun, t time.Duration) (string, error) {
	ticker := time.NewTicker(t)

	duration, err := time.ParseDuration(pipeline.Timeout)
	if err != nil {
		log.Error(err)
		duration = 45 * time.Minute
	}

	ctxWithDeadline, cancel := context.WithDeadline(ctx, time.Now().Add(duration))
	defer cancel()

	client, err := gh.ClientManager.Get(eventContext.GetInstallationID())
	if err != nil {
		log.WithError(err).
			WithField("installation_id", eventContext.GetInstallationID()).Error("Cannot find installation id at cache!")
		return stateError, err
	}

	for {
		select {
		case <-ctxWithDeadline.Done():
			ticker.Stop()
			return stateError, errors.New("Timed out waiting for status")

		case <-ticker.C:
			wfRun, _, err := client.Actions.GetWorkflowRunByID(ctx, pipeline.Organization, pipeline.Repository, *wf.ID)
			if err != nil {
				log.WithField("error", err).Warn("Could not fetch status")
			}

			log.WithFields(log.Fields{
				"conslusion": wfRun.GetConclusion(),
				"status":     wfRun.GetStatus(),
			}).Info("Fetched status of workflow")
			switch wfRun.GetStatus() {
			case statusCompleted:
				ticker.Stop()
				return wfRun.GetConclusion(), nil

			default:
				continue
			}
		}
	}
}
