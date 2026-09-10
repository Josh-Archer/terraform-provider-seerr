package workflows

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Protect the bot-merge release path independently of the ordinary push path.
func TestReleasePleaseSurvivesSuppressedBotPush(t *testing.T) {
	raw, err := os.ReadFile("../../.github/workflows/release-please.yml")
	require.NoError(t, err)
	var workflow struct {
		On struct {
			Dispatch *yaml.Node `yaml:"workflow_dispatch"`
			Run      struct {
				Workflows []string `yaml:"workflows"`
				Types     []string `yaml:"types"`
			} `yaml:"workflow_run"`
			Schedule []struct {
				Cron string `yaml:"cron"`
			} `yaml:"schedule"`
		} `yaml:"on"`
		Concurrency struct {
			Group  string `yaml:"group"`
			Cancel bool   `yaml:"cancel-in-progress"`
		} `yaml:"concurrency"`
		Jobs map[string]struct {
			If    string `yaml:"if"`
			Steps []struct {
				Uses string `yaml:"uses"`
				Run  string `yaml:"run"`
			} `yaml:"steps"`
		} `yaml:"jobs"`
	}
	require.NoError(t, yaml.Unmarshal(raw, &workflow))
	require.Contains(t, workflow.On.Run.Workflows, "CI")
	require.Equal(t, []string{"completed"}, workflow.On.Run.Types)
	require.NotEmpty(t, workflow.On.Schedule, "delayed merges need recovery")
	// A null YAML value still declares a manual trigger.
	var document map[string]any
	require.NoError(t, yaml.Unmarshal(raw, &document))
	triggers, ok := document["on"].(map[string]any)
	require.True(t, ok)
	require.Contains(t, triggers, "workflow_dispatch")
	require.NotEmpty(t, workflow.Concurrency.Group)
	require.False(t, workflow.Concurrency.Cancel, "do not cancel an in-progress release")
	job := workflow.Jobs["release-please"]
	require.Contains(t, job.If, "github.event.workflow_run.conclusion == 'success'")
	require.Contains(t, job.If, "github.event.workflow_run.head_repository.full_name == github.repository")
	var dispatchesRelease bool
	for _, step := range job.Steps {
		require.NotContains(t, step.Uses, "actions/checkout", "privileged completion handler must not execute PR code")
		if step.Run != "" {
			require.Contains(t, step.Run, "gh workflow run release.yml")
			dispatchesRelease = true
		}
	}
	require.True(t, dispatchesRelease, "GITHUB_TOKEN-created releases need explicit publication dispatch")
}

// GoReleaser owns publication, after signed assets have been uploaded.
func TestReleasePleaseHandsDraftToGoReleaser(t *testing.T) {
	raw, err := os.ReadFile("../../.github/release-please-config.json")
	require.NoError(t, err)
	var config struct {
		Packages map[string]struct {
			Draft bool `json:"draft"`
		} `json:"packages"`
	}
	require.NoError(t, json.Unmarshal(raw, &config))
	require.True(t, config.Packages["."].Draft, "publishing an empty release conflicts with GoReleaser's draft handoff")
	raw, err = os.ReadFile("../../.goreleaser.yml")
	require.NoError(t, err)
	var goreleaser struct {
		Release struct {
			UseDraft bool `yaml:"use_existing_draft"`
		} `yaml:"release"`
	}
	require.NoError(t, yaml.Unmarshal(raw, &goreleaser))
	require.True(t, goreleaser.Release.UseDraft)
}

func TestPipelineOnlyChangesDoNotCreateProviderReleases(t *testing.T) {
	raw, err := os.ReadFile("../../.github/release-please-config.json")
	require.NoError(t, err)
	var config struct {
		Sections []struct {
			Type   string `json:"type"`
			Hidden bool   `json:"hidden"`
		} `json:"changelog-sections"`
	}
	require.NoError(t, json.Unmarshal(raw, &config))
	sections := map[string]bool{}
	for _, section := range config.Sections {
		sections[section.Type] = section.Hidden
	}
	require.Contains(t, sections, "ci")
	require.True(t, sections["ci"], "visible ci entries generate patch releases even without provider changes")
	for _, kind := range []string{"feat", "fix", "perf"} {
		require.Contains(t, sections, kind)
		require.False(t, sections[kind], "provider changes must still create releases")
	}
}
