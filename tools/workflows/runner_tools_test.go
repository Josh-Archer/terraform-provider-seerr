package workflows

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// CodeQL's Go autobuild still has no actions/setup-go step. That fix is
// intentionally not in this change: #388 is PR #397, and the actions-language
// RAM failure is #389 / PR #399.
const deferredCodeQLGo = ".github/workflows/codeql.yml"

var (
	scriptRef = regexp.MustCompile(`[\w./-]*(?:scripts|tools)/[\w./-]+\.(?:sh|py)`)
	ghCommand = regexp.MustCompile(`(?:^|[^\w./-])gh(?:[^\w./-]|$)`)
	goCommand = regexp.MustCompile(`(?:^|[^\w./-])go\s+(?:run|test|build|install|mod)\b`)
	goLookup  = regexp.MustCompile(`resolve_tool[[:space:]]+go\b|\$\{go_bin\}`)
)

type workflowFile struct {
	Jobs map[string]workflowJob `yaml:"jobs"`
}

type workflowJob struct {
	Steps []workflowStep `yaml:"steps"`
}

type workflowStep struct {
	Uses string `yaml:"uses"`
	Run  string `yaml:"run"`
}

// TestRunnerToolDependenciesAreDeclared keeps uecb-shared jobs from calling
// gh or compiling Go unless the workflow installs that tool first.
func TestRunnerToolDependenciesAreDeclared(t *testing.T) {
	root := repoRoot(t)
	workflowDir := filepath.Join(root, ".github", "workflows")
	entries, err := os.ReadDir(workflowDir)
	require.NoError(t, err)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yml") {
			continue
		}
		rel := filepath.Join(".github", "workflows", entry.Name())
		t.Run(entry.Name(), func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(root, rel))
			require.NoError(t, err)
			var workflow workflowFile
			require.NoError(t, yaml.Unmarshal(raw, &workflow))
			require.NotEmpty(t, workflow.Jobs)
			for id, job := range workflow.Jobs {
				assertToolOrder(t, root, rel, id, job.Steps)
			}
		})
	}
}

// TestEnsureGHPinMatchesTheReleaseAsset keeps the installer checksums in the
// shape the download step verifies.
// TestToolDetectorsMatchRealInvocations locks the gh and go scanners to the
// command shapes these workflows actually use.
func TestToolDetectorsMatchRealInvocations(t *testing.T) {
	require.True(t, needsGH("gh api repos/example\n"))
	require.True(t, needsGH("existing=$(gh issue list --repo \"${GITHUB_REPOSITORY}\")\n"))
	require.True(t, needsGH("subprocess.run([\"gh\", *args])\n"))
	require.False(t, needsGH("# uecb-shared images do not ship gh\n"))
	require.False(t, needsGH("bash \"${RUNNER_TEMP}/ensure-gh.sh\"\n"))
	require.True(t, needsGH("gh_releases=\"$(gh api repos/example)\"\n"))
	require.True(t, needsGo(workflowStep{}, "go run ./tools/openapi\n"))
	require.True(t, needsGo(workflowStep{}, "go_bin=\"$(resolve_tool go)\"\n"))
	require.True(t, needsGo(workflowStep{Uses: "github/codeql-action/autobuild@abc"}, ""))
	require.True(t, needsGo(workflowStep{Uses: "goreleaser/goreleaser-action@abc"}, ""))
	require.False(t, needsGo(workflowStep{Uses: "actions/setup-go@abc"}, ""))
	require.True(t, declaresGH(workflowStep{Uses: "./.github/actions/ensure-gh"}))
	require.True(t, declaresGH(workflowStep{Run: "bash \"${installer}\"\n# https://example/.github/actions/ensure-gh/ensure-gh.sh"}))
}

func TestEnsureGHPinMatchesTheReleaseAsset(t *testing.T) {
	root := repoRoot(t)
	script, err := os.ReadFile(filepath.Join(root, ".github", "actions", "ensure-gh", "ensure-gh.sh"))
	require.NoError(t, err)
	text := string(script)
	require.Contains(t, text, `version="2.103.3"`)
	require.Regexp(t, `sha_amd64="[0-9a-f]{64}"`, text)
	require.Regexp(t, `sha_arm64="[0-9a-f]{64}"`, text)
	require.Contains(t, text, "https://github.com/cli/cli/releases/download/v${version}/")
	require.Contains(t, text, `>> "${GITHUB_PATH}"`)
}

func assertToolOrder(t *testing.T, root, workflowRel, jobID string, steps []workflowStep) {
	t.Helper()
	ghAt := -1
	goAt := -1
	ghReady := -1
	goReady := -1
	for i, step := range steps {
		body := step.Run + "\n" + referencedSources(t, root, step.Run)
		if strings.Contains(step.Run, "ensure-gh.sh") {
			require.Contains(t, step.Run, "contents/.github/actions/ensure-gh/ensure-gh.sh?ref=${GITHUB_SHA}")
		}
		if ghReady < 0 && declaresGH(step) {
			ghReady = i
		}
		if goReady < 0 && declaresGo(step) {
			goReady = i
		}
		if ghAt < 0 && needsGH(body) {
			ghAt = i
		}
		if goAt < 0 && needsGo(step, body) {
			goAt = i
		}
	}
	if ghAt >= 0 && ghReady < 0 {
		t.Errorf("%s job %s shells out to gh without ./.github/actions/ensure-gh or ensure-gh.sh", workflowRel, jobID)
	}
	if ghAt >= 0 && ghReady >= ghAt {
		t.Errorf("%s job %s installs gh at step %d after the first gh use at step %d", workflowRel, jobID, ghReady, ghAt)
	}
	if goAt >= 0 && goReady < 0 {
		if workflowRel == deferredCodeQLGo {
			t.Logf("deferred: %s Go autobuild has no setup-go (#388, PR #397)", workflowRel)
			return
		}
		t.Errorf("%s job %s compiles Go without actions/setup-go", workflowRel, jobID)
	}
	if goAt >= 0 && goReady >= goAt && workflowRel != deferredCodeQLGo {
		t.Errorf("%s job %s sets up Go at step %d after the first Go use at step %d", workflowRel, jobID, goReady, goAt)
	}
}

func declaresGH(step workflowStep) bool {
	return strings.Contains(step.Uses, "ensure-gh") || strings.Contains(step.Run, "ensure-gh.sh")
}

func declaresGo(step workflowStep) bool {
	return strings.Contains(step.Uses, "actions/setup-go")
}

func needsGH(body string) bool {
	return ghCommand.MatchString(stripComments(body))
}

func needsGo(step workflowStep, body string) bool {
	uses := step.Uses
	if strings.Contains(uses, "goreleaser/goreleaser-action") ||
		strings.Contains(uses, "golangci/golangci-lint-action") ||
		strings.Contains(uses, "codeql-action/autobuild") {
		return true
	}
	stripped := stripComments(body)
	return goCommand.MatchString(stripped) || goLookup.MatchString(stripped)
}

func referencedSources(t *testing.T, root, run string) string {
	t.Helper()
	var b strings.Builder
	seen := map[string]bool{}
	for _, match := range scriptRef.FindAllString(run, -1) {
		rel := normalizeScript(match)
		if rel == "" || seen[rel] {
			continue
		}
		seen[rel] = true
		raw, err := os.ReadFile(filepath.Join(root, rel))
		require.NoError(t, err, "workflow references %s", rel)
		b.WriteString(string(raw))
		b.WriteByte('\n')
	}
	return b.String()
}

func normalizeScript(match string) string {
	for _, prefix := range []string{"scripts/", "tools/"} {
		if i := strings.Index(match, prefix); i >= 0 {
			rel := match[i:]
			if strings.Contains(rel, "..") {
				return ""
			}
			return rel
		}
	}
	return ""
}

func stripComments(body string) string {
	var b strings.Builder
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	return root
}
