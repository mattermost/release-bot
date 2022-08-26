package model

import (
	"strings"

	"github.com/google/go-github/v45/github"
	log "github.com/sirupsen/logrus"
)

type PushEventContext struct {
	event     string
	pushEvent *github.PushEvent
}

func newPushEventContext(event *github.PushEvent) EventContext {
	return &PushEventContext{
		event:     "push",
		pushEvent: event,
	}
}
func (pec *PushEventContext) Log() {
	log.WithFields(log.Fields{
		"event":           pec.GetEvent(),
		"fork":            pec.IsFork(),
		"type":            pec.GetType(),
		"workflow":        pec.GetWorkflow(),
		"run_id":          pec.GetWorkflowRunID(),
		"conclusion":      pec.GetConclusion(),
		"status":          pec.GetStatus(),
		"repo":            pec.GetRepository(),
		"name":            pec.GetName(),
		"installation_id": pec.GetInstallationID(),
		"sha":             pec.GetCommitHash(),
	}).Info("Push Event!")
}

func (pec *PushEventContext) GetEvent() string {
	return pec.event
}
func (pec *PushEventContext) IsFork() bool {
	return pec.pushEvent.GetRepo().GetFork()
}
func (pec *PushEventContext) GetType() string {
	if strings.HasPrefix(pec.pushEvent.GetRef(), "refs/tags") {
		return "tag"
	}
	return "branch"
}

func (pec *PushEventContext) GetWorkflow() string {
	return ""
}
func (pec *PushEventContext) GetWorkflowRunID() int64 {
	return int64(-1)
}
func (pec *PushEventContext) GetConclusion() string {
	return ""
}
func (pec *PushEventContext) GetStatus() string {
	return ""
}
func (pec *PushEventContext) GetRepository() string {
	return pec.pushEvent.GetRepo().GetFullName()
}
func (pec *PushEventContext) GetName() string {
	if strings.HasPrefix(pec.pushEvent.GetRef(), "refs/tags") {
		return strings.TrimPrefix(pec.pushEvent.GetRef(), "refs/tags/")
	}
	return strings.TrimPrefix(pec.pushEvent.GetRef(), "refs/heads/")
}
func (pec *PushEventContext) GetInstallationID() int64 {
	return pec.pushEvent.GetInstallation().GetID()
}
func (pec *PushEventContext) GetCommitHash() string {
	return pec.pushEvent.GetAfter()
}
