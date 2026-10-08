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
	"testing"
	"time"
)

func TestAPIRequestResponseIsRedactedInTerraformPlan(t *testing.T) {
	cliTools := []string{}
	for _, candidate := range []string{"tofu", "terraform"} {
		if _, err := exec.LookPath(candidate); err == nil {
			cliTools = append(cliTools, candidate)
		}
	}
	if len(cliTools) == 0 {
		t.Skip("requires OpenTofu or Terraform; CI installs OpenTofu")
	}

	const fixtureSecret = "integration-only-seerr-token-6a97"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/private" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"token":%q}`, fixtureSecret)
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

data "seerr_api_request" "private" {
  path = "/api/v1/private"
}
`, server.URL)
	providerDir := filepath.Join(t.TempDir(), "provider")
	if err := os.MkdirAll(providerDir, 0o755); err != nil {
		t.Fatal(err)
	}
	providerBinary := filepath.Join(providerDir, "terraform-provider-seerr")
	buildCmd := exec.Command("go", "build", "-o", providerBinary, ".")
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

	for _, cliTool := range cliTools {
		t.Run(cliTool, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			workDir := t.TempDir()
			env := append(os.Environ(), "TF_CLI_CONFIG_FILE="+cliConfigPath)
			configPath := filepath.Join(workDir, "main.tf")
			if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
				t.Fatal(err)
			}
			initCmd := exec.CommandContext(ctx, cliTool, "init", "-backend=false", "-input=false")
			initCmd.Dir = workDir
			initCmd.Env = env
			if output, err := initCmd.CombinedOutput(); err != nil {
				t.Fatalf("%s init failed: %v\n%s", cliTool, err, output)
			}

			unsafeOutputConfig := config + `
output "raw_response" {
  value = data.seerr_api_request.private.response_body_json
}
`
			if err := os.WriteFile(configPath, []byte(unsafeOutputConfig), 0o600); err != nil {
				t.Fatal(err)
			}
			validateCmd := exec.CommandContext(ctx, cliTool, "plan", "-input=false", "-no-color")
			validateCmd.Dir = workDir
			validateCmd.Env = env
			validateOutput, err := validateCmd.CombinedOutput()
			if err == nil || !strings.Contains(strings.ToLower(string(validateOutput)), "sensitive") {
				t.Fatalf("%s did not reject exposing the API response through a non-sensitive output (err=%v):\n%s", cliTool, err, validateOutput)
			}

			safeOutputConfig := config + `
output "raw_response" {
  value     = data.seerr_api_request.private.response_body_json
  sensitive = true
}
`
			if err := os.WriteFile(configPath, []byte(safeOutputConfig), 0o600); err != nil {
				t.Fatal(err)
			}
			applyCmd := exec.CommandContext(ctx, cliTool, "apply", "-auto-approve", "-input=false", "-no-color")
			applyCmd.Dir = workDir
			applyCmd.Env = env
			applyOutput, err := applyCmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s apply failed: %v\n%s", cliTool, err, applyOutput)
			}
			if strings.Contains(string(applyOutput), fixtureSecret) {
				t.Fatalf("%s apply output exposed the arbitrary API response secret:\n%s", cliTool, applyOutput)
			}
			if !strings.Contains(string(applyOutput), "sensitive") {
				t.Fatalf("%s apply output did not mark the response as sensitive:\n%s", cliTool, applyOutput)
			}
		})
	}
}
