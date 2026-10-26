// Copyright (c) Josh Archer
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
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

// TestImporterImportPreservesLiveSettings seeds a disposable Seerr, generates
// HCL and import blocks, supplies secret variables, imports, and requires a
// second plan to be empty. Metadata settings are outside the generated
// configuration: the importer must not request them, and import must not
// change them or send any write.
func TestImporterImportPreservesLiveSettings(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
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

	const (
		radarrKey   = "tf-test-radarr-api-key"
		sonarrKey   = "tf-test-sonarr-api-key"
		telegramKey = "tf-test-telegram-bot-api"
	)
	metadata := []byte(`{"preservedMarker":"keep-me-357","unmanaged":true}`)
	proof := newPreserveProof(t, metadata)

	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate integration test source")
	}
	repoRoot, err := filepath.Abs(filepath.Join(filepath.Dir(sourceFile), "../.."))
	if err != nil {
		t.Fatal(err)
	}
	outputDir := t.TempDir()

	importerCmd := exec.CommandContext(ctx, "go", "run", ".", "-url", proof.URL, "-api-key", "fixture-key", "-out-dir", outputDir)
	importerCmd.Dir = filepath.Join(repoRoot, "tools", "importer")
	importerOut, err := importerCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("importer CLI failed: %v\n%s\nrequests:\n%s", err, importerOut, proof.log())
	}
	if hits, mutations := proof.snapshot(); hits != 0 || len(mutations) != 0 {
		t.Fatalf("importer metadata hits=%d writes=%v\nrequests:\n%s", hits, mutations, proof.log())
	}

	generated := map[string]string{}
	for _, name := range []string{"main.tf", "variables.tf", "imports.tf", "import.sh"} {
		body, err := os.ReadFile(filepath.Join(outputDir, name))
		if err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
		generated[name] = string(body)
	}
	for _, secret := range []string{radarrKey, sonarrKey, telegramKey, "keep-me-357"} {
		if strings.Contains(string(importerOut), secret) {
			t.Fatalf("importer output included %s", secret)
		}
		for name, body := range generated {
			if strings.Contains(body, secret) {
				t.Fatalf("%s included %s", name, secret)
			}
		}
	}
	mainTF := generated["main.tf"]
	for _, address := range []string{
		`resource "seerr_main_settings" "main"`,
		`resource "seerr_radarr_server" "radarr_hd"`,
		`resource "seerr_sonarr_server" "sonarr_hd"`,
		`resource "seerr_notification_telegram" "telegram"`,
		`resource "seerr_user" "admin"`,
		`resource "seerr_user" "admin_2"`,
	} {
		if !strings.Contains(mainTF, address) {
			t.Fatalf("generated HCL missing %s\n%s", address, mainTF)
		}
	}
	for _, unmanaged := range []string{"seerr_network_settings", "seerr_discover_slider", "seerr_override_rule", "seerr_metadata_settings"} {
		if strings.Contains(mainTF, unmanaged) {
			t.Fatalf("generated HCL emitted unmanaged %s\n%s", unmanaged, mainTF)
		}
	}
	if !strings.Contains(generated["imports.tf"], "to = seerr_radarr_server.radarr_hd") || !strings.Contains(generated["imports.tf"], `id = "0"`) {
		t.Fatalf("import blocks:\n%s", generated["imports.tf"])
	}
	if !strings.Contains(generated["variables.tf"], "sensitive   = true") {
		t.Fatalf("variables.tf:\n%s", generated["variables.tf"])
	}
	varNames := regexp.MustCompile(`(?m)^variable "([^"]+)"`).FindAllStringSubmatch(generated["variables.tf"], -1)
	secrets := map[string]string{
		"seerr_radarr_server_radarr_hd_api_key":        radarrKey,
		"seerr_sonarr_server_sonarr_hd_api_key":        sonarrKey,
		"seerr_notification_telegram_telegram_bot_api": telegramKey,
	}
	if len(varNames) != len(secrets) {
		t.Fatalf("variables = %#v", varNames)
	}
	var tfvars strings.Builder
	for _, match := range varNames {
		value, ok := secrets[match[1]]
		if !ok {
			t.Fatalf("unexpected variable %s\n%s", match[1], generated["variables.tf"])
		}
		fmt.Fprintf(&tfvars, "%s = %q\n", match[1], value)
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
			toolCtx, toolCancel := context.WithTimeout(ctx, 4*time.Minute)
			defer toolCancel()
			workDir := t.TempDir()
			for name, body := range generated {
				if err := os.WriteFile(filepath.Join(workDir, name), []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(filepath.Join(workDir, "terraform.tfvars"), []byte(tfvars.String()), 0o600); err != nil {
				t.Fatal(err)
			}
			env := append(os.Environ(),
				"TF_CLI_CONFIG_FILE="+cliConfigPath,
				"TF_IN_AUTOMATION=1",
				"SEERR_URL="+proof.URL,
				"SEERR_API_KEY=fixture-key",
				"SEERR_MAX_RETRIES=0",
				"CHECKPOINT_DISABLE=1",
			)
			run := func(args ...string) []byte {
				t.Helper()
				cmd := exec.CommandContext(toolCtx, cliTool, args...)
				cmd.Dir = workDir
				cmd.Env = env
				output, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("%s %s failed: %v\n%s\nrequests:\n%s", cliTool, strings.Join(args, " "), err, output, proof.log())
				}
				return output
			}
			run("init", "-input=false", "-no-color")
			_, beforeWrites := proof.snapshot()
			run("apply", "-auto-approve", "-input=false", "-no-color")
			_, mutations := proof.snapshot()
			if mismatch := mismatchedMutations(mutations[len(beforeWrites):]); mismatch != "" {
				t.Fatalf("import apply changed live settings: %s\nrequests:\n%s", mismatch, proof.log())
			}
			planCmd := exec.CommandContext(toolCtx, cliTool, "plan", "-detailed-exitcode", "-input=false", "-no-color")
			planCmd.Dir = workDir
			planCmd.Env = env
			planOut, planErr := planCmd.CombinedOutput()
			if planErr != nil {
				var exitErr *exec.ExitError
				if errors.As(planErr, &exitErr) && exitErr.ExitCode() == 2 {
					t.Fatalf("re-plan reported changes:\n%s\nrequests:\n%s", planOut, proof.log())
				}
				t.Fatalf("re-plan failed: %v\n%s\nrequests:\n%s", planErr, planOut, proof.log())
			}
			if !strings.Contains(string(planOut), "No changes") {
				t.Fatalf("re-plan was not clean:\n%s", planOut)
			}
			hits, mutations := proof.snapshot()
			if hits != 0 || !bytes.Equal(proof.metadata, metadata) {
				t.Fatalf("metadata hits=%d body=%s writes=%v", hits, proof.metadata, mutations)
			}
			if mismatch := mismatchedMutations(mutations); mismatch != "" {
				t.Fatalf("import changed live settings: %s", mismatch)
			}
		})
	}
}

type preserveProof struct {
	URL          string
	metadata     []byte
	metadataHits int
	mutations    []string
	requests     []string
	mu           sync.Mutex
}

func (p *preserveProof) log() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return strings.Join(p.requests, "\n")
}

func (p *preserveProof) snapshot() (int, []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.metadataHits, append([]string(nil), p.mutations...)
}

func newPreserveProof(t *testing.T, metadata []byte) *preserveProof {
	t.Helper()
	proof := &preserveProof{metadata: append([]byte(nil), metadata...)}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proof.mu.Lock()
		proof.requests = append(proof.requests, r.Method+" "+r.URL.RequestURI())
		if r.URL.Path == "/api/v1/settings/metadata" {
			proof.metadataHits++
		}
		var body []byte
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			body, _ = io.ReadAll(io.LimitReader(r.Body, 1<<20))
			mismatch := preserveArrWrite(r.Method, r.URL.Path, body)
			if mismatch != "" {
				proof.mutations = append(proof.mutations, r.Method+" "+r.URL.Path+" MISMATCH "+mismatch+" body="+string(body))
			} else {
				proof.mutations = append(proof.mutations, r.Method+" "+r.URL.Path)
			}
			proof.mu.Unlock()
			if mismatch != "" {
				http.Error(w, mismatch, http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/api/v1/settings/radarr/0":
				_ = json.NewEncoder(w).Encode(proofRadarr())
			case "/api/v1/settings/sonarr/1":
				_ = json.NewEncoder(w).Encode(proofSonarr())
			default:
				http.Error(w, "unexpected write", http.StatusInternalServerError)
			}
			return
		}
		proof.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/settings/main":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"applicationTitle":       "Homelab Seerr",
				"applicationUrl":         "https://seerr.example.test",
				"locale":                 "en",
				"hideAvailable":          false,
				"localLogin":             true,
				"partialRequestsEnabled": true,
				"originalLanguage":       "en",
				"mediaServerType":        "plex",
				"movieRequestLimit":      4,
				"seriesRequestLimit":     6,
			})
		case "/api/v1/settings/radarr":
			_ = json.NewEncoder(w).Encode([]map[string]any{proofRadarr()})
		case "/api/v1/settings/sonarr":
			_ = json.NewEncoder(w).Encode([]map[string]any{proofSonarr()})
		case "/api/v1/settings/notifications/telegram":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"enabled":     true,
				"embedPoster": false,
				"types":       15,
				"options": map[string]any{
					"botUsername":  "homelab_bot",
					"botAPI":       "tf-test-telegram-bot-api",
					"chatId":       "12345",
					"sendSilently": false,
				},
			})
		case "/api/v1/user":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{proofUser(1), proofUser(2)},
				"pageInfo": map[string]any{
					"page": 1, "pages": 1, "total": 2, "pageSize": 100,
				},
			})
		case "/api/v1/user/1":
			_ = json.NewEncoder(w).Encode(proofUser(1))
		case "/api/v1/user/2":
			_ = json.NewEncoder(w).Encode(proofUser(2))
		case "/api/v1/user/1/settings/main", "/api/v1/user/2/settings/main":
			_ = json.NewEncoder(w).Encode(proofUserSettings())
		case "/api/v1/settings/metadata":
			_, _ = w.Write(metadata)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	proof.URL = server.URL
	return proof
}

func mismatchedMutations(mutations []string) string {
	var bad []string
	for _, mutation := range mutations {
		if strings.Contains(mutation, "MISMATCH") || strings.Contains(mutation, "DELETE") {
			bad = append(bad, mutation)
		}
	}
	return strings.Join(bad, "\n")
}

// preserveArrWrite allows the provider's import apply to PUT an Arr server
// when state has no API key yet. The body must still carry the seeded
// settings and the same API key. Every other write is rejected.
func preserveArrWrite(method, path string, body []byte) string {
	if method == http.MethodDelete {
		return "delete is not an import"
	}
	var want map[string]any
	switch path {
	case "/api/v1/settings/radarr/0":
		if method != http.MethodPut {
			return "unexpected method"
		}
		want = proofRadarr()
	case "/api/v1/settings/sonarr/1":
		if method != http.MethodPut {
			return "unexpected method"
		}
		want = proofSonarr()
	default:
		return "unexpected write"
	}
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		return err.Error()
	}
	var diffs []string
	for key, wantValue := range want {
		if key == "id" {
			continue
		}
		if !jsonEquivalent(got[key], wantValue) {
			diffs = append(diffs, fmt.Sprintf("%s=%v want %v", key, got[key], wantValue))
		}
	}
	return strings.Join(diffs, ", ")
}

func jsonEquivalent(got, want any) bool {
	gotJSON, err := json.Marshal(got)
	if err != nil {
		return false
	}
	wantJSON, err := json.Marshal(want)
	if err != nil {
		return false
	}
	return bytes.Equal(gotJSON, wantJSON)
}

func proofUser(id int) map[string]any {
	if id == 1 {
		return map[string]any{
			"id": 1, "username": "Admin", "email": "admin@example.com",
			"permissions": 32, "userType": 2,
		}
	}
	return map[string]any{
		"id": 2, "username": "admin", "email": "other@example.com",
		"permissions": 0, "userType": 2,
	}
}

func proofUserSettings() map[string]any {
	return map[string]any{
		"locale": "en", "discoverRegion": "US", "streamingRegion": "US", "originalLanguage": "en",
		"movieQuotaLimit": 10, "movieQuotaDays": 7, "tvQuotaLimit": 8, "tvQuotaDays": 14,
		"watchlistSyncMovies": false, "watchlistSyncTv": false,
	}
}

func proofRadarr() map[string]any {
	return map[string]any{
		"id": 0, "name": "Radarr HD", "hostname": "radarr.local", "port": 7878,
		"useSsl": false, "is4k": false, "isDefault": true,
		"activeProfileId": 3, "activeProfileName": "HD-1080p", "activeDirectory": "/movies",
		"minimumAvailability": "released", "enableScan": true, "syncEnabled": true,
		"preventSearch": false, "tagRequests": true, "tags": []any{},
		"apiKey": "tf-test-radarr-api-key",
	}
}

func proofSonarr() map[string]any {
	return map[string]any{
		"id": 1, "name": "Sonarr HD", "hostname": "sonarr.local", "port": 8989,
		"useSsl": false, "is4k": false, "isDefault": true,
		"activeProfileId": 4, "activeProfileName": "HD-1080p", "activeDirectory": "/tv",
		"activeAnimeDirectory": "/anime", "enableSeasonFolders": true,
		"enableScan": true, "syncEnabled": true, "preventSearch": false, "tagRequests": true,
		"tags": []any{}, "animeTags": []any{},
		"apiKey": "tf-test-sonarr-api-key",
	}
}
