// Copyright (c) Josh Archer
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"encoding/json"
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

func TestImporterCLIOutputValidatesWithLocalProvider(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	cliTools := []string{}
	for _, candidate := range []string{"tofu", "terraform"} {
		if _, err := exec.LookPath(candidate); err == nil {
			cliTools = append(cliTools, candidate)
		}
	}
	if len(cliTools) == 0 {
		t.Skip("requires OpenTofu or Terraform; CI installs OpenTofu")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/settings/main" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"applicationTitle": "Fixture Seerr",
				"applicationUrl":   "https://seerr.example.test",
				"locale":           "en",
			})
			return
		}
		http.NotFound(w, r)
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
	outputDir := t.TempDir()

	importerCmd := exec.CommandContext(ctx, "go", "run", ".", "-url", server.URL, "-out-dir", outputDir, "-format", "hcl")
	importerCmd.Dir = filepath.Join(repoRoot, "tools", "importer")
	if output, err := importerCmd.CombinedOutput(); err != nil {
		t.Fatalf("importer CLI failed: %v\n%s", err, output)
	}

	generatedPath := filepath.Join(outputDir, "main.tf")
	generated, err := os.ReadFile(generatedPath)
	if err != nil {
		t.Fatalf("CLI did not generate main.tf: %v", err)
	}
	if !strings.Contains(string(generated), `version = "~> 2.0"`) {
		t.Fatalf("generated provider configuration does not target v2:\n%s", generated)
	}

	providerDir := filepath.Join(t.TempDir(), "provider")
	if err := os.MkdirAll(providerDir, 0o755); err != nil {
		t.Fatal(err)
	}
	providerBinary := filepath.Join(providerDir, "terraform-provider-seerr")
	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", providerBinary, ".")
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
			toolCtx, toolCancel := context.WithTimeout(ctx, 2*time.Minute)
			defer toolCancel()
			env := append(os.Environ(), "TF_CLI_CONFIG_FILE="+cliConfigPath)
			cmd := exec.CommandContext(toolCtx, cliTool, "validate")
			cmd.Dir = outputDir
			cmd.Env = env
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("%s validate failed: %v\n%s", cliTool, err, output)
			}
		})
	}
}

func TestImporterCLIOutputWithSensitiveVariablesValidatesWithLocalProvider(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	cliTools := []string{}
	for _, candidate := range []string{"tofu", "terraform"} {
		if _, err := exec.LookPath(candidate); err == nil {
			cliTools = append(cliTools, candidate)
		}
	}
	if len(cliTools) == 0 {
		t.Skip("requires OpenTofu or Terraform; CI installs OpenTofu")
	}

	canaryTautulli := strings.Join([]string{"canary", "integ-tautulli", "not-real"}, "-")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/settings/main":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"applicationTitle": "Integration Seerr",
				"applicationUrl":   "https://seerr.example.test",
				"locale":           "en",
			})
		case "/api/v1/settings/tautulli":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"hostname": "tautulli.local",
				"port":     8181,
				"apiKey":   canaryTautulli,
			})
		default:
			http.NotFound(w, r)
		}
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
	outputDir := t.TempDir()

	importerCmd := exec.CommandContext(ctx, "go", "run", ".", "-url", server.URL, "-out-dir", outputDir, "-format", "all")
	importerCmd.Dir = filepath.Join(repoRoot, "tools", "importer")
	importerCmd.Env = append(os.Environ(), "SEERR_API_KEY=test-api-key-not-real")
	if output, err := importerCmd.CombinedOutput(); err != nil {
		t.Fatalf("importer CLI failed: %v\n%s", err, output)
	}

	mainPath := filepath.Join(outputDir, "main.tf")
	varsPath := filepath.Join(outputDir, "variables.tf")
	importsPath := filepath.Join(outputDir, "imports.tf")
	scriptPath := filepath.Join(outputDir, "import.sh")

	mainContent, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatalf("CLI did not generate main.tf: %v", err)
	}
	varsContent, err := os.ReadFile(varsPath)
	if err != nil {
		t.Fatalf("CLI did not generate variables.tf: %v", err)
	}
	importsContent, err := os.ReadFile(importsPath)
	if err != nil {
		t.Fatalf("CLI did not generate imports.tf: %v", err)
	}
	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("CLI did not generate import.sh: %v", err)
	}

	// Verify no canary leaked
	for _, content := range [][]byte{mainContent, varsContent, importsContent, scriptContent} {
		if strings.Contains(string(content), canaryTautulli) {
			t.Fatalf("Canary %q leaked into output:\n%s", canaryTautulli, content)
		}
	}

	// Verify variable references and declarations
	if !strings.Contains(string(mainContent), "api_key                   = var.tautulli_api_key") {
		t.Errorf("main.tf missing var.tautulli_api_key:\n%s", mainContent)
	}
	if !strings.Contains(string(varsContent), `variable "tautulli_api_key"`) {
		t.Errorf("variables.tf missing variable tautulli_api_key:\n%s", varsContent)
	}
	if !strings.Contains(string(varsContent), "sensitive   = true") {
		t.Errorf("variables.tf missing sensitive = true:\n%s", varsContent)
	}

	// Verify restrictive permissions
	mainStat, _ := os.Stat(mainPath)
	if perm := mainStat.Mode().Perm(); perm != 0600 {
		t.Errorf("main.tf permissions = %04o; want 0600", perm)
	}
	varsStat, _ := os.Stat(varsPath)
	if perm := varsStat.Mode().Perm(); perm != 0600 {
		t.Errorf("variables.tf permissions = %04o; want 0600", perm)
	}
	scriptStat, _ := os.Stat(scriptPath)
	if perm := scriptStat.Mode().Perm(); perm != 0700 {
		t.Errorf("import.sh permissions = %04o; want 0700", perm)
	}

	// Build local provider for validation
	providerDir := filepath.Join(t.TempDir(), "provider")
	if err := os.MkdirAll(providerDir, 0o755); err != nil {
		t.Fatal(err)
	}
	providerBinary := filepath.Join(providerDir, "terraform-provider-seerr")
	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", providerBinary, ".")
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

	// Validate with both engines
	for _, cliTool := range cliTools {
		t.Run(cliTool, func(t *testing.T) {
			toolCtx, toolCancel := context.WithTimeout(ctx, 2*time.Minute)
			defer toolCancel()
			env := append(os.Environ(), "TF_CLI_CONFIG_FILE="+cliConfigPath)
			cmd := exec.CommandContext(toolCtx, cliTool, "validate")
			cmd.Dir = outputDir
			cmd.Env = env
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("%s validate failed on generated configuration with sensitive variables: %v\n%s", cliTool, err, output)
			}
		})
	}
}
