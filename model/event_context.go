package model

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/google/go-github/v45/github"
	"github.com/mattermost/release-bot/config"
)

type EventContext interface {
	GetAction() string
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
	"push":         pushEventMapper,
	"workflow_run": workflowRunEventMapper,
}

func pushEventMapper(payload []byte) (EventContext, error) {
	var event github.PushEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, err
	}
	return newPushEventContext(&event), nil
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
			// Check if webhook conditions allowed match the event from Github
			// Can be pull_request push workflow_run
			if !contains(condition.Webhook, context.GetEvent()) {
				continue
			}

			// Commenting out fork check since we are going to use Github built in functionality to approve workflows for forks
			// We are going to use only workflow_run events for now in order to be secure
			// if context.IsFork() && !condition.Fork {
			// 	fmt.Println(context.IsFork() && !condition.Fork)
			// 	continue
			// }

			// Check if type that triggered the webhook matches with the one configured
			// Can be one of pull_request push tag
			if condition.Type != context.GetType() {
				continue
			}

			// Check if workflow name is the one that want to trigger the workflow
			if condition.Workflow != "" && condition.Workflow != context.GetWorkflow() {
				continue
			}

			// Check if workflow conclusion is the one that is condigured
			// Can be one of: success, failure, neutral, cancelled, timed_out, action_required, stale, null, skipped, startup_failure
			if condition.Conclusion != "" && condition.Conclusion != context.GetConclusion() {
				continue
			}

			// Check if workflow status is the one that is condigured
			// Can be one of: requested, in_progress, completed, queued, pending, waiting
			if condition.Status != "" && condition.Status != context.GetStatus() {
				continue
			}

			// Check if the repository that triggered the webhook is allowed
			if condition.Repository != "" && !isRegexpMatch(condition.Repository, context.GetRepository()) {
				continue
			}

			// We will suspend branch matching for now since we will only allow workflow_run events
			// if condition.Name != "" && !isRegexpMatch(condition.Name, context.GetName()) {
			// 	continue
			// }

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
