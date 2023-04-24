package model

import (
	"encoding/json"
	"testing"

	"github.com/mattermost/release-bot/config"
	"github.com/stretchr/testify/assert"
)

type eventContextFixture struct {
	action     string
	conclusion string
	event      string
	name       string
	repository string
	status     string
	workflow   string
	commitHash string
	_type      string
	fork       bool
}

func (f *eventContextFixture) GetAction() string {
	return f.action
}
func (f *eventContextFixture) GetEvent() string {
	return f.event
}
func (f *eventContextFixture) IsFork() bool {
	return f.fork
}
func (f *eventContextFixture) GetType() string {
	return f._type
}
func (f *eventContextFixture) GetWorkflow() string {
	return f.workflow
}
func (f *eventContextFixture) GetWorkflowRunID() int64 {
	return -1
}
func (f *eventContextFixture) GetConclusion() string {
	return f.conclusion
}
func (f *eventContextFixture) GetStatus() string {
	return f.status
}
func (f *eventContextFixture) GetRepository() string {
	return f.repository
}
func (f *eventContextFixture) GetName() string {
	return f.name
}
func (f *eventContextFixture) GetInstallationID() int64 {
	return -1
}
func (f *eventContextFixture) GetCommitHash() string {
	return f.commitHash
}
func (f *eventContextFixture) Log() {
}
func (f *eventContextFixture) JSON() string {
	b, _ := json.Marshal(f)
	return string(b)
}

func createPipelineConfiguration() []config.PipelineConfig {
	return []config.PipelineConfig{
		{
			Organization: "a",
			Repository:   "b",
			Workflow:     "dedicated-repo-pr",
			Conditions: []config.PipelineCondition{
				{
					Webhook:    []string{"workflow_run"},
					Repository: "^mattermost/dedicated$",
					Type:       "pull_request",
					// Fork:       false,
				},
			},
		},
		{
			Organization: "a",
			Repository:   "b",
			Workflow:     "dedicated-repo-push",
			Conditions: []config.PipelineCondition{
				{
					Webhook:    []string{"workflow_run"},
					Repository: "^mattermost/dedicated$",
					Type:       "push",
					// Fork:       true,
				},
			},
		},
		{
			Organization: "a",
			Repository:   "b",
			Workflow:     "dedicated-repo-tag",
			Conditions: []config.PipelineCondition{
				{
					Webhook:    []string{"workflow_run"},
					Repository: "^mattermost/dedicated$",
					Type:       "tag",
				},
			},
		},
		{
			Organization: "a",
			Repository:   "b",
			Workflow:     "repo-workflow",
			Conditions: []config.PipelineCondition{
				{
					Webhook:    []string{"workflow_run"},
					Repository: "^mattermost/workflow$",
					Type:       "pull_request",
					Workflow:   "Build",
				},
			},
		},
		{
			Organization: "a",
			Repository:   "b",
			Workflow:     "repo-workflow-conclusion",
			Conditions: []config.PipelineCondition{
				{
					Webhook:    []string{"workflow_run"},
					Repository: "^mattermost/conclusion$",
					Type:       "pull_request",
					Workflow:   "Build",
					Conclusion: "success",
				},
			},
		},
		{
			Organization: "a",
			Repository:   "b",
			Workflow:     "repo-workflow-status",
			Conditions: []config.PipelineCondition{
				{
					Webhook:    []string{"workflow_run"},
					Repository: "^mattermost/status$",
					Type:       "pull_request",
					Workflow:   "Build",
					Status:     "queued",
				},
			},
		},
		{
			Organization: "a",
			Repository:   "b",
			Workflow:     "repo-workflow-name",
			Conditions: []config.PipelineCondition{
				{
					Webhook:    []string{"workflow_run"},
					Repository: "^mattermost/a.*$",
					// Name:       "feat/a.*",
					Type:     "pull_request",
					Workflow: "Build",
					Status:   "queued",
				},
			},
		},
	}
}

func TestEventContext(t *testing.T) {
	pipelineConfiguration := createPipelineConfiguration()
	t.Run("Non Supported Event", func(t *testing.T) {
		eventContext := &eventContextFixture{
			event: "non-supported",
		}
		assert.Nil(t, GetTargetPipeline(eventContext, pipelineConfiguration))
	})
	t.Run("All PR events in the repository", func(t *testing.T) {
		eventContext := &eventContextFixture{
			event:      "workflow_run",
			repository: "mattermost/dedicated",
			fork:       false,
			_type:      "pull_request",
			conclusion: "",
			name:       "",
			status:     "",
			workflow:   "",
			commitHash: "",
		}
		pipeline := GetTargetPipeline(eventContext, pipelineConfiguration)
		assert.NotNil(t, pipeline)
		assert.Equal(t, "dedicated-repo-pr", pipeline.Workflow)
	})
	t.Run("All Push events in the repository", func(t *testing.T) {
		eventContext := &eventContextFixture{
			event:      "workflow_run",
			repository: "mattermost/dedicated",
			fork:       false,
			_type:      "push",
			conclusion: "",
			name:       "",
			status:     "",
			workflow:   "",
			commitHash: "",
		}
		pipeline := GetTargetPipeline(eventContext, pipelineConfiguration)
		assert.NotNil(t, pipeline)
		assert.Equal(t, "dedicated-repo-push", pipeline.Workflow)
	})
	t.Run("All Tag events in the repository", func(t *testing.T) {
		eventContext := &eventContextFixture{
			event:      "workflow_run",
			repository: "mattermost/dedicated",
			fork:       false,
			_type:      "tag",
			conclusion: "",
			name:       "",
			status:     "",
			workflow:   "",
			commitHash: "",
		}
		pipeline := GetTargetPipeline(eventContext, pipelineConfiguration)
		assert.NotNil(t, pipeline)
		assert.Equal(t, "dedicated-repo-tag", pipeline.Workflow)
	})
	// t.Run("No pipeline for forked pr", func(t *testing.T) {
	// 	eventContext := &eventContextFixture{
	// 		event:      "workflow_run",
	// 		repository: "mattermost/dedicated",
	// 		fork:       true,
	// 		_type:      "pull_request",
	// 		conclusion: "",
	// 		name:       "",
	// 		status:     "",
	// 		workflow:   "",
	// 		commitHash: "",
	// 	}
	// 	pipeline := GetTargetPipeline(eventContext, pipelineConfiguration)
	// 	assert.Nil(t, pipeline)
	// })
	// t.Run("Forked Branch event pipeline check", func(t *testing.T) {
	// 	eventContext := &eventContextFixture{
	// 		event:      "workflow_run",
	// 		repository: "mattermost/dedicated",
	// 		fork:       true,
	// 		_type:      "branch",
	// 		conclusion: "",
	// 		name:       "",
	// 		status:     "",
	// 		workflow:   "",
	// 		commitHash: "",
	// 	}
	// 	pipeline := GetTargetPipeline(eventContext, pipelineConfiguration)
	// 	assert.NotNil(t, pipeline)
	// 	assert.Equal(t, "dedicated-repo-push", pipeline.Workflow)
	// })
	t.Run("Workflow event pipeline check", func(t *testing.T) {
		eventContext := &eventContextFixture{
			event:      "workflow_run",
			repository: "mattermost/workflow",
			fork:       false,
			_type:      "pull_request",
			conclusion: "",
			name:       "",
			status:     "",
			workflow:   "Build",
			commitHash: "",
		}
		pipeline := GetTargetPipeline(eventContext, pipelineConfiguration)
		assert.NotNil(t, pipeline)
		assert.Equal(t, "repo-workflow", pipeline.Workflow)
	})
	t.Run("Workflow event pipeline negative check", func(t *testing.T) {
		eventContext := &eventContextFixture{
			event:      "workflow_run",
			repository: "mattermost/workflow",
			fork:       false,
			_type:      "pull_request",
			conclusion: "",
			name:       "",
			status:     "",
			workflow:   "Builds",
			commitHash: "",
		}
		pipeline := GetTargetPipeline(eventContext, pipelineConfiguration)
		assert.Nil(t, pipeline)
	})
	t.Run("Workflow conclusion pipeline check", func(t *testing.T) {
		eventContext := &eventContextFixture{
			event:      "workflow_run",
			repository: "mattermost/conclusion",
			fork:       false,
			_type:      "pull_request",
			conclusion: "success",
			name:       "",
			status:     "",
			workflow:   "Build",
			commitHash: "",
		}
		pipeline := GetTargetPipeline(eventContext, pipelineConfiguration)
		assert.NotNil(t, pipeline)
		assert.Equal(t, "repo-workflow-conclusion", pipeline.Workflow)
	})
	t.Run("Workflow conclusion pipeline negative check", func(t *testing.T) {
		eventContext := &eventContextFixture{
			event:      "workflow_run",
			repository: "mattermost/conclusion",
			fork:       false,
			_type:      "pull_request",
			conclusion: "failure",
			name:       "",
			status:     "",
			workflow:   "Build",
			commitHash: "",
		}
		pipeline := GetTargetPipeline(eventContext, pipelineConfiguration)
		assert.Nil(t, pipeline)
	})
	t.Run("Workflow status pipeline check", func(t *testing.T) {
		eventContext := &eventContextFixture{
			event:      "workflow_run",
			repository: "mattermost/status",
			fork:       false,
			_type:      "pull_request",
			conclusion: "",
			name:       "",
			status:     "queued",
			workflow:   "Build",
			commitHash: "",
		}
		pipeline := GetTargetPipeline(eventContext, pipelineConfiguration)
		assert.NotNil(t, pipeline)
		assert.Equal(t, "repo-workflow-status", pipeline.Workflow)
	})
	t.Run("Workflow status pipeline negative check", func(t *testing.T) {
		eventContext := &eventContextFixture{
			event:      "workflow_run",
			repository: "mattermost/status",
			fork:       false,
			_type:      "pull_request",
			conclusion: "",
			name:       "",
			status:     "finished",
			workflow:   "Build",
			commitHash: "",
		}
		pipeline := GetTargetPipeline(eventContext, pipelineConfiguration)
		assert.Nil(t, pipeline)
	})
	t.Run("Name pipeline check", func(t *testing.T) {
		eventContext := &eventContextFixture{
			event:      "workflow_run",
			repository: "mattermost/abc",
			fork:       false,
			_type:      "pull_request",
			conclusion: "",
			// name:       "feat/abc",
			status:     "queued",
			workflow:   "Build",
			commitHash: "",
		}
		pipeline := GetTargetPipeline(eventContext, pipelineConfiguration)
		assert.NotNil(t, pipeline)
		assert.Equal(t, "repo-workflow-name", pipeline.Workflow)
	})
	t.Run("Name pipeline with conlcusion does not exist", func(t *testing.T) {
		eventContext := &eventContextFixture{
			event:      "workflow_run",
			repository: "mattermost/abc",
			fork:       false,
			_type:      "pull_request",
			conclusion: "success",
			// name:       "feat/bac",
			status:     "completed",
			workflow:   "Build",
			commitHash: "",
		}
		pipeline := GetTargetPipeline(eventContext, pipelineConfiguration)
		assert.Nil(t, pipeline)
	})
}
