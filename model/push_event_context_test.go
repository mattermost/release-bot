package model

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-github/v45/github"
	"github.com/stretchr/testify/assert"
)

func TestPushEventContext(t *testing.T) {
	t.Run("Test Tag", func(t *testing.T) {
		context := newPushEventContext(createPushEvent(t, "push_event_tag.json"))
		assert.Equal(t, "f3c4bfb6bf87b9aa2a52a36ed213eec10ae0196c", context.GetCommitHash())
		assert.Equal(t, "", context.GetConclusion())
		assert.Equal(t, "push", context.GetEvent())
		assert.Equal(t, int64(28579677), context.GetInstallationID())
		assert.Equal(t, "test-tag", context.GetName())
		assert.Equal(t, "mattermost/release-bot", context.GetRepository())
		assert.Equal(t, "", context.GetStatus())
		assert.Equal(t, "tag", context.GetType())
		assert.Equal(t, "", context.GetWorkflow())
		assert.Equal(t, int64(-1), context.GetWorkflowRunID())
		assert.Equal(t, false, context.IsFork())
	})
	t.Run("Test Branch", func(t *testing.T) {
		context := newPushEventContext(createPushEvent(t, "push_event_branch.json"))
		assert.Equal(t, "e849bae468c92cc43a4bd18e8985a12ba06d7d64", context.GetCommitHash())
		assert.Equal(t, "", context.GetConclusion())
		assert.Equal(t, "push", context.GetEvent())
		assert.Equal(t, int64(28579677), context.GetInstallationID())
		assert.Equal(t, "feat/cld-3876-create-github-release-bot-for-unified-ci", context.GetName())
		assert.Equal(t, "mattermost/release-bot", context.GetRepository())
		assert.Equal(t, "", context.GetStatus())
		assert.Equal(t, "branch", context.GetType())
		assert.Equal(t, "", context.GetWorkflow())
		assert.Equal(t, int64(-1), context.GetWorkflowRunID())
		assert.Equal(t, false, context.IsFork())
	})
}

func createPushEvent(t *testing.T, filename string) *github.PushEvent {
	source, err := os.ReadFile(fmt.Sprintf("testdata/%s", filename))
	if err != nil {
		t.Fatal("error reading source file:", err)
	}
	var event github.PushEvent

	if err = json.Unmarshal(source, &event); err != nil {
		t.Fatal("error reading source file:", err)
	}
	return &event
}
