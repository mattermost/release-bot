package model

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/google/go-github/v45/github"
	"github.com/mattermost/release-bot/config"
)

type EventContext interface {
	GetConclusion() string
	GetEvent() string
	GetInstallationID() int64
	GetName() string
	GetRepository() string
	GetStatus() string
	GetWorkflow() string
	GetWorkflowRunID() int64
	GetCommitHash() string
	GetType() string
	IsFork() bool
	Log()
}

type payloadToEventContextConverter func(payload []byte) (EventContext, error)

var eventContextConverters map[string]payloadToEventContextConverter = map[string]payloadToEventContextConverter{
	"workflow_run": workflowRunEventMapper,
}

func workflowRunEventMapper(payload []byte) (EventContext, error) {
	var event github.WorkflowRunEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, err
	}
	return newWorkflowRunEventContext(event), nil
}

func ConvertPayloadToEventContext(githubEventType string, payload []byte) (EventContext, error) {
	if converter, ok := eventContextConverters[githubEventType]; ok {
		return converter(payload)
	}
	return nil, fmt.Errorf("converter not found for %s event", githubEventType)
}

/*
Traverse all pipeline conditions from configuration for the github event.
If all conditions are matched then return pipeline

Rules:
1. Github event must be defined at condition allowed event list.
2. If event belongs to fork, fork option must be true at condition. For non-forks, condition is not important.
3. Condition type must be equal to event type (pr/branch or tag)
3. If event belongs to PR, pr option  must be true. Otherwise both must be false.
4. If event belongs to branch, branch option must be true. Otherwise both must be false.
5. If event belongs to tag, tag option must be true. Otherwise both must be false.
6. If event belongs to workflow, workflow must be equal to event workflow name. If condition field is empty, rule is skipped.
7. If event belongs to workflow, condition must be equal to event workflow conclusion. If conclusion field is empty, rule is skipped.
8. If event belongs to workflow, status must be equal to event workflow status. If status field is empty, rule is skipped.
9. [Regex]Repository must be matched to events repository name. If repository field is empty, rule is skipped.
10.[Regex]Name must be matched to events reference. If name field is empty, rule is skipped.
*/
func GetTargetPipeline(context EventContext, pipelines []config.PipelineConfig) *config.PipelineConfig {
	for _, pipeline := range pipelines {
		for _, condition := range pipeline.Conditions {
			if !contains(condition.Webhook, context.GetEvent()) {
				continue
			}
			if context.IsFork() && !condition.Fork {
				continue
			}
			if condition.Type != context.GetType() {
				continue
			}
			if condition.Workflow != "" && condition.Workflow != context.GetWorkflow() {
				continue
			}
			if condition.Conclusion != "" && condition.Conclusion != context.GetConclusion() {
				continue
			}
			if condition.Status != "" && condition.Status != context.GetStatus() {
				continue
			}
			if condition.Repository != "" && !isRegexpMatch(condition.Repository, context.GetRepository()) {
				continue
			}
			if condition.Name != "" && !isRegexpMatch(condition.Name, context.GetName()) {
				continue
			}
			return &pipeline
		}
	}
	return nil
}

func isRegexpMatch(rule string, check string) bool {
	match, _ := regexp.MatchString(rule, check)
	return match
}
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
