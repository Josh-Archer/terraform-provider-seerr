// Copyright (c) Josh Archer
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Josh-Archer/terraform-provider-seerr/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Main Server", "main_server"},
		{"Sonarr - 4K (Anime)", "sonarr_4k_anime"},
		{"123_test", "item_123_test"},
		{"Radarr.HD", "radarr_hd"},
		{"", "main"},
		{"---", "main"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeName(tt.input)
			if got != tt.expected {
				t.Errorf("sanitizeName(%q) = %q, expected %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestImporterDiscoveryAndGeneration(t *testing.T) {
	mux := http.NewServeMux()

	// Mock Main Settings.
	mux.HandleFunc("/api/v1/settings/main", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"applicationTitle":   "Homelab Seerr",
			"applicationUrl":     "https://seerr.homelab.local",
			"csrfProtection":     true,
			"defaultPermissions": 32,
			"hideAvailable":      false,
			"localLogin":         true,
			"locale":             "en",
		})
	})

	// Mock Radarr Servers.
	mux.HandleFunc("/api/v1/settings/radarr", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{
				"id":                0,
				"name":              "Radarr HD",
				"hostname":          "radarr.local",
				"port":              7878,
				"useSsl":            false,
				"is4k":              false,
				"isDefault":         true,
				"activeProfileName": "HD-1080p",
				"activeDirectory":   "/movies",
			},
		})
	})

	// Mock Discord Notification.
	mux.HandleFunc("/api/v1/settings/notifications/discord", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"enabled": true,
			"types":   15,
			"options": map[string]any{
				"webhookUrl":  "https://discord.com/api/webhooks/123/abc",
				"botUsername": "Seerr Alerts",
			},
		})
	})

	// Mock Users.
	mux.HandleFunc("/api/v1/user", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": []map[string]any{
				{
					"id":          1,
					"username":    "admin",
					"email":       "admin@example.com",
					"permissions": 2097150,
					"userType":    1,
				},
			},
		})
	})

	// Fallback 404 for un-mocked endpoints.
	server := httptest.NewServer(mux)
	defer server.Close()

	imp := NewImporter(ImporterConfig{
		BaseURL:    server.URL,
		APIKey:     "test-api-key-not-real",
		HTTPClient: server.Client(),
	})

	ctx := context.Background()
	resources, err := imp.DiscoverAll(ctx)
	if err != nil {
		t.Fatalf("DiscoverAll failed: %v", err)
	}

	if len(resources) < 4 {
		t.Fatalf("Expected at least 4 discovered resources, got %d", len(resources))
	}

	// Verify HCL Generation.
	hcl := GenerateHCL(resources, true)
	if !strings.Contains(hcl, `version = "~> 2.0"`) {
		t.Errorf("Expected generated provider configuration to target the current v2 release line, got:\n%s", hcl)
	}
	if !strings.Contains(hcl, `resource "seerr_main_settings" "main"`) {
		t.Errorf("Expected seerr_main_settings in HCL, got:\n%s", hcl)
	}
	if !strings.Contains(hcl, `resource "seerr_radarr_server" "radarr_hd"`) {
		t.Errorf("Expected seerr_radarr_server in HCL, got:\n%s", hcl)
	}
	if !strings.Contains(hcl, `resource "seerr_notification_discord" "discord"`) {
		t.Errorf("Expected seerr_notification_discord in HCL, got:\n%s", hcl)
	}
	if !strings.Contains(hcl, `resource "seerr_user" "admin"`) {
		t.Errorf("Expected seerr_user in HCL, got:\n%s", hcl)
	}

	// Verify Import Blocks.
	imports := GenerateImportBlocks(resources)
	if !strings.Contains(imports, "to = seerr_main_settings.main") {
		t.Errorf("Expected import block for seerr_main_settings, got:\n%s", imports)
	}
	if !strings.Contains(imports, `id = "0"`) {
		t.Errorf("Expected import id '0' for Radarr server, got:\n%s", imports)
	}

	// Verify Import Shell Script.
	script := GenerateImportScript(resources)
	if !strings.Contains(script, "terraform import seerr_main_settings.main main") {
		t.Errorf("Expected terraform import command in script, got:\n%s", script)
	}
}

// TestSensitiveAttributeMappingMatchesProviderSchema checks that every sensitive attribute
// declared in the provider's resource schemas is classified as sensitive by the importer.
func TestSensitiveAttributeMappingMatchesProviderSchema(t *testing.T) {
	ctx := context.Background()

	for _, factory := range provider.RegisteredResources {
		res := factory()
		var mResp resource.MetadataResponse
		res.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "seerr"}, &mResp)
		var sResp resource.SchemaResponse
		res.Schema(ctx, resource.SchemaRequest{}, &sResp)

		resourceType := mResp.TypeName

		for attrName, attr := range sResp.Schema.Attributes {
			if attr.IsSensitive() {
				if !IsSensitiveAttribute(resourceType, attrName) {
					t.Errorf("Resource %s attribute %s is marked Sensitive in provider schema but not classified as sensitive in importer", resourceType, attrName)
				}
			}

			// If attribute is a single nested attribute (e.g. notification agent options)
			if nested, ok := attr.(schema.SingleNestedAttribute); ok {
				for nestedName, nestedAttr := range nested.Attributes {
					if nestedAttr.IsSensitive() {
						if !IsSensitiveAttribute(resourceType, nestedName) {
							t.Errorf("Resource %s nested attribute %s is marked Sensitive in provider schema but not classified as sensitive in importer", resourceType, nestedName)
						}
					}
				}
			}
		}
	}
}

// fakeCanary builds obviously-fake, unique test values at runtime so secret
// scanners do not treat fixture literals as credentials.
func fakeCanary(name string) string {
	return strings.Join([]string{"canary", name, "not-real"}, "-")
}

var (
	canaryRadarrKey        = fakeCanary("radarr-key-99999")
	canarySonarrKey        = fakeCanary("sonarr-key-88888")
	canaryEmailPass        = fakeCanary("email-pass-77777")
	canaryEmailPgpPass     = fakeCanary("email-pgp-pass-66666")
	canaryEmailPgpKey      = fakeCanary("email-pgp-key-55555")
	canaryTelegramBot      = fakeCanary("telegram-bot-44444")
	canaryGotifyToken      = fakeCanary("gotify-token-33333")
	canaryWebhookAuth      = fakeCanary("webhook-auth-22222")
	canaryNtfyPass         = fakeCanary("ntfy-pass-11111")
	canaryNtfyToken        = fakeCanary("ntfy-token-10101")
	canaryPushbulletToken  = fakeCanary("pushbullet-token-20202")
	canaryPushoverAppToken = fakeCanary("pushover-app-30303")
	canaryPushoverUserKey  = fakeCanary("pushover-user-40404")
	canaryTautulliKey      = fakeCanary("tautulli-key-50505")
	canaryJellyfinKey      = fakeCanary("jellyfin-key-60606")
	canaryEmbyKey          = fakeCanary("emby-key-70707")
	canaryUserPlexToken    = fakeCanary("user-plex-token-80808")
	canaryUserJfPass       = fakeCanary("user-jf-pass-90909")
	canaryImporterAPIKey   = fakeCanary("importer-api-key-00000")
)

var allCanaries = []string{
	canaryRadarrKey,
	canarySonarrKey,
	canaryEmailPass,
	canaryEmailPgpPass,
	canaryEmailPgpKey,
	canaryTelegramBot,
	canaryGotifyToken,
	canaryWebhookAuth,
	canaryNtfyPass,
	canaryNtfyToken,
	canaryPushbulletToken,
	canaryPushoverAppToken,
	canaryPushoverUserKey,
	canaryTautulliKey,
	canaryJellyfinKey,
	canaryEmbyKey,
	canaryUserPlexToken,
	canaryUserJfPass,
	canaryImporterAPIKey,
}

func setupCanaryServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/settings/main", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"applicationTitle": "Canary Seerr",
			"applicationUrl":   "https://seerr.canary.test",
		})
	})

	mux.HandleFunc("/api/v1/settings/radarr", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{
				"id":       10,
				"name":     "Radarr Primary",
				"hostname": "radarr.internal",
				"port":     7878,
				"apiKey":   canaryRadarrKey,
			},
		})
	})

	mux.HandleFunc("/api/v1/settings/sonarr", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{
				"id":       20,
				"name":     "Sonarr Anime",
				"hostname": "sonarr.internal",
				"port":     8989,
				"apiKey":   canarySonarrKey,
			},
		})
	})

	mux.HandleFunc("/api/v1/settings/notifications/email", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"enabled": true,
			"types":   1,
			"options": map[string]any{
				"emailFrom":     "alerts@example.com",
				"smtpHost":      "smtp.example.com",
				"smtpPort":      587,
				"authPass":      canaryEmailPass,
				"pgpPassword":   canaryEmailPgpPass,
				"pgpPrivateKey": canaryEmailPgpKey,
			},
		})
	})

	mux.HandleFunc("/api/v1/settings/notifications/telegram", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"enabled": true,
			"types":   2,
			"options": map[string]any{
				"botApi": canaryTelegramBot,
				"chatId": "123456",
			},
		})
	})

	mux.HandleFunc("/api/v1/settings/notifications/gotify", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"enabled": true,
			"types":   4,
			"options": map[string]any{
				"url":   "https://gotify.example.com",
				"token": canaryGotifyToken,
			},
		})
	})

	mux.HandleFunc("/api/v1/settings/notifications/webhook", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"enabled": true,
			"types":   8,
			"options": map[string]any{
				"webhookUrl": "https://webhook.example.com/alerts",
				"authHeader": canaryWebhookAuth,
			},
		})
	})

	mux.HandleFunc("/api/v1/settings/notifications/ntfy", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"enabled": true,
			"types":   16,
			"options": map[string]any{
				"url":      "https://ntfy.sh",
				"topic":    "alerts",
				"password": canaryNtfyPass,
				"token":    canaryNtfyToken,
			},
		})
	})

	mux.HandleFunc("/api/v1/settings/notifications/pushbullet", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"enabled": true,
			"types":   32,
			"options": map[string]any{
				"accessToken": canaryPushbulletToken,
			},
		})
	})

	mux.HandleFunc("/api/v1/settings/notifications/pushover", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"enabled": true,
			"types":   64,
			"options": map[string]any{
				"accessToken": canaryPushoverAppToken,
				"userToken":   canaryPushoverUserKey,
			},
		})
	})

	mux.HandleFunc("/api/v1/settings/tautulli", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"hostname": "tautulli.internal",
			"port":     8181,
			"apiKey":   canaryTautulliKey,
		})
	})

	mux.HandleFunc("/api/v1/settings/jellyfin", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"name":     "Jellyfin Main",
			"hostname": "jellyfin.internal",
			"port":     8096,
			"apiKey":   canaryJellyfinKey,
		})
	})

	mux.HandleFunc("/api/v1/settings/emby", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"name":     "Emby Main",
			"hostname": "emby.internal",
			"port":     8096,
			"apiKey":   canaryEmbyKey,
		})
	})

	mux.HandleFunc("/api/v1/user", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": []map[string]any{
				{
					"id":               1,
					"username":         "alice",
					"email":            "alice@example.com",
					"plexAuthToken":    canaryUserPlexToken,
					"jellyfinPassword": canaryUserJfPass,
				},
			},
		})
	})

	return httptest.NewServer(mux)
}

// TestCanarySecretNotWrittenToGeneratedFiles verifies that fetched secret canary values
// are NEVER written into generated HCL, variables, import blocks, or scripts.
func TestCanarySecretNotWrittenToGeneratedFiles(t *testing.T) {
	server := setupCanaryServer()
	defer server.Close()

	imp := NewImporter(ImporterConfig{
		BaseURL:    server.URL,
		APIKey:     canaryImporterAPIKey,
		HTTPClient: server.Client(),
	})

	ctx := context.Background()
	resources, err := imp.DiscoverAll(ctx)
	if err != nil {
		t.Fatalf("DiscoverAll failed: %v", err)
	}

	hcl := GenerateHCL(resources, true)
	vars := GenerateVariables(resources)
	imports := GenerateImportBlocks(resources)
	script := GenerateImportScript(resources)

	// Verify no canaries appear in any generated content
	for _, canary := range allCanaries {
		if strings.Contains(hcl, canary) {
			t.Errorf("Canary %q leaked into generated HCL:\n%s", canary, hcl)
		}
		if strings.Contains(vars, canary) {
			t.Errorf("Canary %q leaked into generated variables:\n%s", canary, vars)
		}
		if strings.Contains(imports, canary) {
			t.Errorf("Canary %q leaked into generated imports:\n%s", canary, imports)
		}
		if strings.Contains(script, canary) {
			t.Errorf("Canary %q leaked into generated script:\n%s", canary, script)
		}
	}

	// Verify variable references in HCL
	expectedVarRefs := []string{
		"auth_pass                 = var.email_auth_pass",
		"pgp_password              = var.email_pgp_password",
		"pgp_private_key           = var.email_pgp_private_key",
		"bot_api                   = var.telegram_bot_api",
		"token                     = var.gotify_token",
		"auth_header               = var.webhook_auth_header",
		"password                  = var.ntfy_password",
		"token                     = var.ntfy_token",
		"access_token              = var.pushbullet_access_token",
		"access_token              = var.pushover_access_token",
		"user_token                = var.pushover_user_token",
		"api_key                   = var.radarr_primary_api_key",
		"api_key                   = var.sonarr_anime_api_key",
		"api_key                   = var.tautulli_api_key",
		"api_key                   = var.jellyfin_api_key",
		"api_key                   = var.emby_api_key",
		"plex_auth_token           = var.alice_plex_auth_token",
		"jellyfin_password         = var.alice_jellyfin_password",
	}

	for _, ref := range expectedVarRefs {
		if !strings.Contains(hcl, ref) {
			t.Errorf("Expected HCL to contain variable reference %q, got:\n%s", ref, hcl)
		}
	}

	// Verify sensitive variable declarations
	expectedVars := []string{
		`variable "email_auth_pass"`,
		`variable "telegram_bot_api"`,
		`variable "radarr_primary_api_key"`,
		`variable "sonarr_anime_api_key"`,
		`variable "tautulli_api_key"`,
		`variable "jellyfin_api_key"`,
		`variable "emby_api_key"`,
		`variable "alice_plex_auth_token"`,
		`variable "alice_jellyfin_password"`,
	}

	for _, v := range expectedVars {
		if !strings.Contains(vars, v) {
			t.Errorf("Expected variables.tf to contain declaration %q, got:\n%s", v, vars)
		}
	}

	if !strings.Contains(vars, "sensitive   = true") {
		t.Errorf("Expected variables.tf to declare variables as sensitive = true, got:\n%s", vars)
	}

	// Verify security state storage notice
	if !strings.Contains(hcl, "SECURITY NOTICE - STATE STORAGE") {
		t.Errorf("Expected HCL header to contain state storage security notice")
	}
	if !strings.Contains(vars, "SECURITY NOTICE - STATE STORAGE") {
		t.Errorf("Expected variables header to contain state storage security notice")
	}
}

// TestCLISecretCanaryLeakProtection verifies stdout, stderr, and all generated files
// on disk contain NO secret canaries.
func TestCLISecretCanaryLeakProtection(t *testing.T) {
	server := setupCanaryServer()
	defer server.Close()

	outDir := t.TempDir()

	var stdout, stderr bytes.Buffer
	args := []string{
		"-url", server.URL,
		"-api-key", canaryImporterAPIKey,
		"-out-dir", outDir,
		"-format", "all",
	}

	exitCode := run(args, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("run failed with exit code %d; stderr:\n%s", exitCode, stderr.String())
	}

	stdoutStr := stdout.String()
	stderrStr := stderr.String()

	for _, canary := range allCanaries {
		if strings.Contains(stdoutStr, canary) {
			t.Errorf("Canary %q leaked to stdout:\n%s", canary, stdoutStr)
		}
		if strings.Contains(stderrStr, canary) {
			t.Errorf("Canary %q leaked to stderr:\n%s", canary, stderrStr)
		}
	}

	// Read every file in outDir
	files, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("failed to read outDir: %v", err)
	}

	if len(files) == 0 {
		t.Fatalf("expected generated files in outDir, found none")
	}

	for _, f := range files {
		content, err := os.ReadFile(filepath.Join(outDir, f.Name()))
		if err != nil {
			t.Fatalf("failed reading %s: %v", f.Name(), err)
		}
		contentStr := string(content)
		for _, canary := range allCanaries {
			if strings.Contains(contentStr, canary) {
				t.Errorf("Canary %q leaked into file %s:\n%s", canary, f.Name(), contentStr)
			}
		}
	}
}

// TestCLIAPIErrorPathCanaryProtection verifies that API error paths do not leak
// credentials or error response canaries into diagnostics or stderr.
func TestCLIAPIErrorPathCanaryProtection(t *testing.T) {
	errorCanary := fakeCanary("error-internal-token")
	apiKeyCanary := fakeCanary("api-key-auth-fail")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": fmt.Sprintf("Authentication failed for key %s with secret token %s", apiKeyCanary, errorCanary),
			"token":   errorCanary,
			"apiKey":  apiKeyCanary,
		})
	}))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	args := []string{
		"-url", server.URL,
		"-api-key", apiKeyCanary,
		"-out-dir", t.TempDir(),
	}

	exitCode := run(args, &stdout, &stderr)
	if exitCode == 0 {
		t.Fatalf("expected nonzero exit code on API failure, got 0")
	}

	errOutput := stderr.String()

	if strings.Contains(errOutput, errorCanary) {
		t.Errorf("Error canary %q leaked into stderr diagnostics:\n%s", errorCanary, errOutput)
	}
	if strings.Contains(errOutput, apiKeyCanary) {
		t.Errorf("API key canary %q leaked into stderr diagnostics:\n%s", apiKeyCanary, errOutput)
	}

	// Verify [REDACTED] placeholder was used
	if !strings.Contains(errOutput, "[REDACTED]") {
		t.Errorf("Expected [REDACTED] in sanitized error output, got:\n%s", errOutput)
	}
}

// TestOutputPermissionsAndOverwriteControl tests restrictive permissions, refusal to overwrite by default,
// and allowing overwrite when -overwrite is set.
func TestOutputPermissionsAndOverwriteControl(t *testing.T) {
	server := setupCanaryServer()
	defer server.Close()

	outDir := filepath.Join(t.TempDir(), "secure-out")

	var stdout, stderr bytes.Buffer
	args := []string{
		"-url", server.URL,
		"-api-key", canaryImporterAPIKey,
		"-out-dir", outDir,
		"-format", "all",
	}

	// Run 1: Create files
	exitCode := run(args, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("First run failed with exit code %d; stderr:\n%s", exitCode, stderr.String())
	}

	// Check directory permissions (0700)
	dirInfo, err := os.Stat(outDir)
	if err != nil {
		t.Fatalf("failed to stat outDir: %v", err)
	}
	if perm := dirInfo.Mode().Perm(); perm != 0700 {
		t.Errorf("outDir permissions = %04o; want 0700", perm)
	}

	// Check file permissions
	tfFiles := []string{"main.tf", "variables.tf", "imports.tf"}
	for _, name := range tfFiles {
		info, err := os.Stat(filepath.Join(outDir, name))
		if err != nil {
			t.Fatalf("failed to stat %s: %v", name, err)
		}
		if perm := info.Mode().Perm(); perm != 0600 {
			t.Errorf("File %s permissions = %04o; want 0600", name, perm)
		}
	}

	scriptInfo, err := os.Stat(filepath.Join(outDir, "import.sh"))
	if err != nil {
		t.Fatalf("failed to stat import.sh: %v", err)
	}
	if perm := scriptInfo.Mode().Perm(); perm != 0700 {
		t.Errorf("import.sh permissions = %04o; want 0700", perm)
	}

	// Run 2: Re-run without -overwrite; must fail closed with exit code 1
	var stdout2, stderr2 bytes.Buffer
	exitCode2 := run(args, &stdout2, &stderr2)
	if exitCode2 != 1 {
		t.Errorf("Expected exit code 1 when refusing to overwrite, got %d", exitCode2)
	}
	if !strings.Contains(stderr2.String(), "refusing to overwrite existing file") {
		t.Errorf("Expected refusal message in stderr, got:\n%s", stderr2.String())
	}

	// Run 3: Re-run with -overwrite; must succeed
	var stdout3, stderr3 bytes.Buffer
	argsOverwrite := append(args, "-overwrite")
	exitCode3 := run(argsOverwrite, &stdout3, &stderr3)
	if exitCode3 != 0 {
		t.Errorf("Expected exit code 0 when -overwrite is provided, got %d; stderr:\n%s", exitCode3, stderr3.String())
	}
}

// TestWriteFailureReturnsNonzero verifies that write failures (e.g. invalid target path) return nonzero.
func TestWriteFailureReturnsNonzero(t *testing.T) {
	server := setupCanaryServer()
	defer server.Close()

	// Create a regular file and use it as outDir path so MkdirAll fails
	blockingFile := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blockingFile, []byte("blocker"), 0600); err != nil {
		t.Fatal(err)
	}
	invalidOutDir := filepath.Join(blockingFile, "sub")

	var stdout, stderr bytes.Buffer
	args := []string{
		"-url", server.URL,
		"-out-dir", invalidOutDir,
	}

	exitCode := run(args, &stdout, &stderr)
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 on directory creation/write failure, got %d", exitCode)
	}
}

// TestAuthDiagnosticsAdvisesEnvVar verifies that passing -api-key on the command line
// triggers an advisory notice pointing to SEERR_API_KEY.
func TestAuthDiagnosticsAdvisesEnvVar(t *testing.T) {
	server := setupCanaryServer()
	defer server.Close()

	var stdout, stderr bytes.Buffer
	args := []string{
		"-url", server.URL,
		"-api-key", "sample-key",
		"-out-dir", t.TempDir(),
	}

	exitCode := run(args, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("unexpected exit code %d; stderr:\n%s", exitCode, stderr.String())
	}

	if !strings.Contains(stderr.String(), "SEERR_API_KEY environment variable is preferred") {
		t.Errorf("Expected advisory notice in stderr about SEERR_API_KEY, got:\n%s", stderr.String())
	}
}
