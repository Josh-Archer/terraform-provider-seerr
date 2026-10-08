package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestWatchlistInputsRejectedBeforeRequest(t *testing.T) {
	cliTools := []string{}
	for _, candidate := range []string{"tofu", "terraform"} {
		if _, err := exec.LookPath(candidate); err == nil {
			cliTools = append(cliTools, candidate)
		}
	}
	if len(cliTools) == 0 {
		t.Skip("requires OpenTofu or Terraform; CI installs OpenTofu")
	}

	var requestCount atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{}`)
	}))
	defer server.Close()

	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate integration test source")
	}
	repoRoot, err := filepath.Abs(filepath.Join(filepath.Dir(sourceFile), "../.."))
	if err != nil {
		t.Fatal(err)
	}
	providerDir := filepath.Join(t.TempDir(), "provider")
	if err := os.MkdirAll(providerDir, 0o755); err != nil {
		t.Fatal(err)
	}
	providerBinary := filepath.Join(providerDir, "terraform-provider-seerr")
	buildCtx, buildCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer buildCancel()
	buildCmd := exec.CommandContext(buildCtx, "go", "build", "-o", providerBinary, ".")
	buildCmd.Dir = repoRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("building local provider failed: %v\n%s", err, output)
	}

	cliConfigPath := filepath.Join(t.TempDir(), "terraform.rc")
	cliConfig := fmt.Sprintf(`provider_installation {
  dev_overrides {
    "josh-archer/seerr" = %q
  }
  direct {}
}
`, providerDir)
	if err := os.WriteFile(cliConfigPath, []byte(cliConfig), 0o600); err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name           string
		tmdbID         int
		mediaType      string
		wantDiagnostic string
	}{
		{name: "zero TMDB ID", tmdbID: 0, mediaType: "movie", wantDiagnostic: "value must be at least 1"},
		{name: "negative TMDB ID", tmdbID: -1, mediaType: "movie", wantDiagnostic: "value must be at least 1"},
		{name: "unknown media type", tmdbID: 123, mediaType: "episode", wantDiagnostic: `value must be one of: ["movie" "tv"]`},
		{name: "valid movie", tmdbID: 123, mediaType: "movie"},
		{name: "valid tv", tmdbID: 456, mediaType: "tv"},
	}
	for _, cliTool := range cliTools {
		t.Run(cliTool, func(t *testing.T) {
			for _, testCase := range testCases {
				t.Run(testCase.name, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
					defer cancel()
					workDir := t.TempDir()
					resourceName := "valid"
					if testCase.wantDiagnostic != "" {
						resourceName = "invalid"
					}
					config := fmt.Sprintf(`terraform {
  required_providers {
    seerr = {
      source = "josh-archer/seerr"
      version = "~> 2.0"
    }
  }
}

provider "seerr" {
  url = %q
  api_key = "fixture-key"
}

resource "seerr_watchlist" %q {
  tmdb_id = %d
  media_type = %q
}
`, server.URL, resourceName, testCase.tmdbID, testCase.mediaType)
					if err := os.WriteFile(filepath.Join(workDir, "main.tf"), []byte(config), 0o600); err != nil {
						t.Fatal(err)
					}

					env := append(os.Environ(), "TF_CLI_CONFIG_FILE="+cliConfigPath)
					planCmd := exec.CommandContext(ctx, cliTool, "plan", "-input=false", "-no-color")
					planCmd.Dir = workDir
					planCmd.Env = env
					output, err := planCmd.CombinedOutput()
					if testCase.wantDiagnostic == "" {
						if err != nil {
							t.Fatalf("%s rejected valid watchlist configuration: %v\n%s", cliTool, err, output)
						}
					} else {
						if err == nil {
							t.Fatalf("%s accepted invalid watchlist configuration:\n%s", cliTool, output)
						}
						if !strings.Contains(string(output), testCase.wantDiagnostic) {
							t.Fatalf("%s missing validator diagnostic %q:\n%s", cliTool, testCase.wantDiagnostic, output)
						}
					}
					if got := requestCount.Load(); got != 0 {
						t.Fatalf("configuration made %d API request(s), want none", got)
					}
				})
			}
		})
	}
}
