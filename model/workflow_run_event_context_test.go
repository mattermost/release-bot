package model

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-github/v45/github"
	"github.com/stretchr/testify/assert"
)

func TestWorkflowRunEventContext(t *testing.T) {
	t.Run("Test PR", func(t *testing.T) {
		context := newWorkflowRunEventContext(createWorkflowRunEvent(t, "workflow_run_event_pr.json"))
		assert.Equal(t, "ab7a32c308ac42df77385bbb5e97f0e3aac5c42f", context.GetCommitHash())
		assert.Equal(t, "queued", context.GetConclusion())
		assert.Equal(t, "workflow_run", context.GetEvent())
		assert.Equal(t, int64(1854), context.GetInstallationID())
		assert.Equal(t, "feat/cld-3876-create-github-release-bot-for-unified-ci", context.GetName())
		assert.Equal(t, "mattermost/release-bot", context.GetRepository())
		assert.Equal(t, "queued", context.GetStatus())
		assert.Equal(t, "pull_request", context.GetType())
		assert.Equal(t, "Build", context.GetWorkflow())
		assert.Equal(t, int64(2926155304), context.GetWorkflowRunID())
		assert.Equal(t, false, context.IsFork())
	})
	t.Run("Test Branch", func(t *testing.T) {
		context := newWorkflowRunEventContext(createWorkflowRunEvent(t, "workflow_run_event_branch.json"))
		assert.Equal(t, "ab7a32c308ac42df77385bbb5e97f0e3aac5c42f", context.GetCommitHash())
		assert.Equal(t, "completed", context.GetConclusion())
		assert.Equal(t, "workflow_run", context.GetEvent())
		assert.Equal(t, int64(1854), context.GetInstallationID())
		assert.Equal(t, "feat/cld-3876-create-github-release-bot-for-unified-ci", context.GetName())
		assert.Equal(t, "mattermost/release-bot", context.GetRepository())
		assert.Equal(t, "finished", context.GetStatus())
		assert.Equal(t, "branch", context.GetType())
		assert.Equal(t, "Build", context.GetWorkflow())
		assert.Equal(t, int64(2926155304), context.GetWorkflowRunID())
		assert.Equal(t, false, context.IsFork())
	})
}

func createWorkflowRunEvent(t *testing.T, filename string) github.WorkflowRunEvent {
	source, err := os.ReadFile(fmt.Sprintf("testdata/%s", filename))
	if err != nil {
		t.Fatal("error reading source file:", err)
	}
	var event github.WorkflowRunEvent

	if err = json.Unmarshal(source, &event); err != nil {
		t.Fatal("error reading source file:", err)
	}
	return event
}
