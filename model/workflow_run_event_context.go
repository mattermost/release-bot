package model

import (
	"github.com/google/go-github/v45/github"
	log "github.com/sirupsen/logrus"
)

type WorkflowRunEventContext struct {
	event          string
	action         string
	repository     string
	installationID int64
	workflowRun    *github.WorkflowRun
}

func newWorkflowRunEventContext(event github.WorkflowRunEvent) EventContext {
	workflowRun := event.GetWorkflowRun()
	return &WorkflowRunEventContext{
		event:          "workflow_run",
		action:         event.GetAction(),
		repository:     event.GetRepo().GetFullName(),
		installationID: event.GetInstallation().GetID(),
		workflowRun:    workflowRun,
	}
}

func (wrec *WorkflowRunEventContext) Log() {
	log.WithFields(log.Fields{
		"event":           wrec.GetEvent(),
		"action":          wrec.GetAction(),
		"fork":            wrec.IsFork(),
		"type":            wrec.GetType(),
		"workflow":        wrec.GetWorkflow(),
		"run_id":          wrec.GetWorkflowRunID(),
		"conclusion":      wrec.GetConclusion(),
		"status":          wrec.GetStatus(),
		"repo":            wrec.GetRepository(),
		"name":            wrec.GetName(),
		"installation_id": wrec.GetInstallationID(),
		"sha":             wrec.GetCommitHash(),
	}).Info("Workflow Run Event!")
}

func (wrec *WorkflowRunEventContext) GetEvent() string {
	return wrec.event
}

func (wrec *WorkflowRunEventContext) GetAction() string {
	// Can be one of completed in_progress requested
	return wrec.action
}

func (wrec *WorkflowRunEventContext) IsFork() bool {
	return wrec.workflowRun.GetRepository().GetFullName() != wrec.workflowRun.GetHeadRepository().GetFullName()
}

func (wrec *WorkflowRunEventContext) GetType() string {
	switch wrec.workflowRun.GetEvent() {
	case "push":
		if wrec.workflowRun.GetHeadBranch() != "" {
			return "branch"
		}
		return "tag"
	default:
		return wrec.workflowRun.GetEvent()
	}
}

func (wrec *WorkflowRunEventContext) GetWorkflow() string {
	return wrec.workflowRun.GetName()
}

func (wrec *WorkflowRunEventContext) GetWorkflowRunID() int64 {
	return wrec.workflowRun.GetID()
}

func (wrec *WorkflowRunEventContext) GetConclusion() string {
	// Can be one of: success, failure, neutral, cancelled, timed_out, action_required, stale, null, skipped, startup_failure
	return wrec.workflowRun.GetConclusion()
}

func (wrec *WorkflowRunEventContext) GetStatus() string {
	// Can be one of: requested, in_progress, completed, queued, pending, waiting
	return wrec.workflowRun.GetStatus()
}

func (wrec *WorkflowRunEventContext) GetRepository() string {
	return wrec.repository
}

func (wrec *WorkflowRunEventContext) GetName() string {
	return wrec.workflowRun.GetHeadBranch()
}

func (wrec *WorkflowRunEventContext) GetInstallationID() int64 {
	return wrec.installationID
}

func (wrec *WorkflowRunEventContext) GetCommitHash() string {
	return wrec.workflowRun.GetHeadSHA()
}
