package model

import (
	"encoding/json"
	"strings"

	"github.com/google/go-github/v45/github"
	log "github.com/sirupsen/logrus"
)

type PushEventContext struct {
	Event     string `json:"event"`
	action    string
	PushEvent *github.PushEvent `json:"eventPayload"`
}

func newPushEventContext(event *github.PushEvent) EventContext {
	return &PushEventContext{
		Event:     "push",
		action:    "push",
		PushEvent: event,
	}
}

func (pec *PushEventContext) Log() {
	log.WithFields(log.Fields{
		"event":           pec.GetEvent(),
		"action":          pec.GetAction(),
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
	return pec.Event
}
func (pec *PushEventContext) GetAction() string {
	return pec.action
}
func (pec *PushEventContext) IsFork() bool {
	return pec.PushEvent.GetRepo().GetFork()
}
func (pec *PushEventContext) GetType() string {
	if strings.HasPrefix(pec.PushEvent.GetRef(), "refs/tags") {
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
	return pec.PushEvent.GetRepo().GetFullName()
}
func (pec *PushEventContext) GetName() string {
	if strings.HasPrefix(pec.PushEvent.GetRef(), "refs/tags") {
		return strings.TrimPrefix(pec.PushEvent.GetRef(), "refs/tags/")
	}
	return strings.TrimPrefix(pec.PushEvent.GetRef(), "refs/heads/")
}
func (pec *PushEventContext) GetInstallationID() int64 {
	return pec.PushEvent.GetInstallation().GetID()
}
func (pec *PushEventContext) GetCommitHash() string {
	return pec.PushEvent.GetAfter()
}

func (pec *PushEventContext) JSON() string {
	// We never care for the error
	b, _ := json.Marshal(pec)
	return string(b)
}
