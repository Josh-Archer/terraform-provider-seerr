// Copyright (c) Josh Archer
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
		APIKey:     "test-api-key",
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
