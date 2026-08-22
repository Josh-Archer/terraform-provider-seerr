// Copyright (c) Josh Archer
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ImporterConfig holds configuration for the live Seerr importer.
type ImporterConfig struct {
	BaseURL    string
	APIKey     string
	Timeout    time.Duration
	HTTPClient *http.Client
}

// DiscoveredResource represents a live Seerr resource ready to be converted into HCL and import blocks.
type DiscoveredResource struct {
	ResourceType string
	ResourceName string
	ImportID     string
	Attributes   map[string]any
	Comments     []string
}

// Importer fetches configuration from a live Seerr instance and generates Terraform/OpenTofu code.
type Importer struct {
	cfg        ImporterConfig
	httpClient *http.Client
}

// NewImporter creates a new instance of Importer.
func NewImporter(cfg ImporterConfig) *Importer {
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{
			Timeout: cfg.Timeout,
		}
		if client.Timeout == 0 {
			client.Timeout = 30 * time.Second
		}
	}
	return &Importer{
		cfg:        cfg,
		httpClient: client,
	}
}

// sanitizeName transforms arbitrary names/titles into valid HCL identifiers.
func sanitizeName(name string) string {
	if name == "" {
		return "main"
	}
	var sb strings.Builder
	lastWasUnderscore := false
	for _, ch := range strings.ToLower(name) {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			sb.WriteRune(ch)
			lastWasUnderscore = false
		} else if ch == '_' || ch == ' ' || ch == '-' || ch == '.' || ch == '(' || ch == ')' {
			if !lastWasUnderscore && sb.Len() > 0 {
				sb.WriteRune('_')
				lastWasUnderscore = true
			}
		}
	}
	res := strings.Trim(sb.String(), "_")
	if res == "" {
		return "main"
	}
	if res[0] >= '0' && res[0] <= '9' {
		res = "item_" + res
	}
	return res
}

func (imp *Importer) get(ctx context.Context, apiPath string) (any, int, error) {
	baseURL := strings.TrimRight(imp.cfg.BaseURL, "/")
	reqURL := baseURL + apiPath
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Accept", "application/json")
	if imp.cfg.APIKey != "" {
		req.Header.Set("X-Api-Key", imp.cfg.APIKey)
	}

	resp, err := imp.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, resp.StatusCode, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.StatusCode, fmt.Errorf("API error %d from %s: %s", resp.StatusCode, apiPath, string(body))
	}

	var result any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, resp.StatusCode, err
	}

	return result, resp.StatusCode, nil
}

// DiscoverAll retrieves all manageable resources from the live instance.
func (imp *Importer) DiscoverAll(ctx context.Context) ([]DiscoveredResource, error) {
	var discovered []DiscoveredResource

	// 1. Main Settings
	mainRes, code, err := imp.get(ctx, "/api/v1/settings/main")
	if err != nil {
		if code != http.StatusNotFound {
			return nil, fmt.Errorf("failed to fetch main settings: %w", err)
		}
	} else if code == http.StatusOK {
		if m, ok := mainRes.(map[string]any); ok {
			attrs := map[string]any{}
			copyString(m, attrs, "applicationTitle", "app_title")
			copyString(m, attrs, "applicationUrl", "application_url")
			copyBool(m, attrs, "csrfProtection", "csrf_protection")
			copyInt(m, attrs, "defaultPermissions", "default_permissions")
			copyBool(m, attrs, "hideAvailable", "hide_available")
			copyBool(m, attrs, "localLogin", "local_login")
			copyString(m, attrs, "locale", "locale")
			copyBool(m, attrs, "newPlexLogin", "new_plex_login")
			copyBool(m, attrs, "partialRequests", "partial_requests")
			copyBool(m, attrs, "trustProxy", "trust_proxy")

			discovered = append(discovered, DiscoveredResource{
				ResourceType: "seerr_main_settings",
				ResourceName: "main",
				ImportID:     "main",
				Attributes:   attrs,
			})
		}
	}

	// 2. Plex Settings
	plexRes, code, err := imp.get(ctx, "/api/v1/settings/plex")
	if err == nil && code == 200 {
		if m, ok := plexRes.(map[string]any); ok {
			attrs := map[string]any{}
			copyString(m, attrs, "name", "name")
			copyString(m, attrs, "hostname", "hostname")
			copyInt(m, attrs, "port", "port")
			copyBool(m, attrs, "useSsl", "use_ssl")
			copyString(m, attrs, "machineId", "machine_id")
			copyString(m, attrs, "webAppUrl", "web_app_url")
			copyString(m, attrs, "ip", "ip")

			if len(attrs) > 0 {
				discovered = append(discovered, DiscoveredResource{
					ResourceType: "seerr_plex_settings",
					ResourceName: "plex",
					ImportID:     "plex",
					Attributes:   attrs,
				})
			}
		}
	}

	// 3. Jellyfin Settings
	jfRes, code, err := imp.get(ctx, "/api/v1/settings/jellyfin")
	if err == nil && code == 200 {
		if m, ok := jfRes.(map[string]any); ok {
			attrs := map[string]any{}
			copyString(m, attrs, "name", "name")
			copyString(m, attrs, "hostname", "hostname")
			copyInt(m, attrs, "port", "port")
			copyBool(m, attrs, "useSsl", "use_ssl")
			copyString(m, attrs, "url", "url")
			copyString(m, attrs, "userId", "user_id")
			copyString(m, attrs, "serverId", "server_id")

			if len(attrs) > 0 {
				discovered = append(discovered, DiscoveredResource{
					ResourceType: "seerr_jellyfin_settings",
					ResourceName: "jellyfin",
					ImportID:     "jellyfin",
					Attributes:   attrs,
				})
			}
		}
	}

	// 4. Emby Settings
	embyRes, code, err := imp.get(ctx, "/api/v1/settings/emby")
	if err == nil && code == 200 {
		if m, ok := embyRes.(map[string]any); ok {
			attrs := map[string]any{}
			copyString(m, attrs, "name", "name")
			copyString(m, attrs, "hostname", "hostname")
			copyInt(m, attrs, "port", "port")
			copyBool(m, attrs, "useSsl", "use_ssl")
			copyString(m, attrs, "url", "url")
			copyString(m, attrs, "userId", "user_id")
			copyString(m, attrs, "serverId", "server_id")

			if len(attrs) > 0 {
				discovered = append(discovered, DiscoveredResource{
					ResourceType: "seerr_emby_settings",
					ResourceName: "emby",
					ImportID:     "emby",
					Attributes:   attrs,
				})
			}
		}
	}

	// 5. Tautulli Settings
	tauRes, code, err := imp.get(ctx, "/api/v1/settings/tautulli")
	if err == nil && code == 200 {
		if m, ok := tauRes.(map[string]any); ok {
			attrs := map[string]any{}
			copyString(m, attrs, "hostname", "hostname")
			copyInt(m, attrs, "port", "port")
			copyBool(m, attrs, "useSsl", "use_ssl")
			copyString(m, attrs, "externalUrl", "external_url")

			if len(attrs) > 0 && (attrs["hostname"] != nil || attrs["external_url"] != nil) {
				discovered = append(discovered, DiscoveredResource{
					ResourceType: "seerr_tautulli_settings",
					ResourceName: "tautulli",
					ImportID:     "tautulli",
					Attributes:   attrs,
				})
			}
		}
	}

	// 6. Network Settings
	netRes, code, err := imp.get(ctx, "/api/v1/settings/network")
	if err == nil && code == 200 {
		if m, ok := netRes.(map[string]any); ok {
			attrs := map[string]any{}
			copyBool(m, attrs, "trustProxy", "trust_proxy")
			copyString(m, attrs, "proxyIpHeader", "proxy_ip_header")

			discovered = append(discovered, DiscoveredResource{
				ResourceType: "seerr_network_settings",
				ResourceName: "network",
				ImportID:     "network",
				Attributes:   attrs,
			})
		}
	}

	// 7. Radarr Servers
	radarrRes, code, err := imp.get(ctx, "/api/v1/settings/radarr")
	if err == nil && code == 200 {
		if list, ok := radarrRes.([]any); ok {
			for _, item := range list {
				if s, ok := item.(map[string]any); ok {
					idVal := getInt(s, "id")
					name := getString(s, "name")
					if name == "" {
						name = fmt.Sprintf("radarr_%d", idVal)
					}
					attrs := map[string]any{
						"name":                 name,
						"hostname":             getString(s, "hostname"),
						"port":                 getInt(s, "port"),
						"use_ssl":              getBool(s, "useSsl"),
						"is_4k":                getBool(s, "is4k"),
						"is_default":           getBool(s, "isDefault"),
						"active_profile_name":  getString(s, "activeProfileName"),
						"active_directory":     getString(s, "activeDirectory"),
						"minimum_availability": getString(s, "minimumAvailability"),
					}
					if v := getString(s, "baseUrl"); v != "" {
						attrs["base_url"] = v
					}
					discovered = append(discovered, DiscoveredResource{
						ResourceType: "seerr_radarr_server",
						ResourceName: sanitizeName(name),
						ImportID:     strconv.FormatInt(idVal, 10),
						Attributes:   attrs,
					})
				}
			}
		}
	}

	// 8. Sonarr Servers
	sonarrRes, code, err := imp.get(ctx, "/api/v1/settings/sonarr")
	if err == nil && code == 200 {
		if list, ok := sonarrRes.([]any); ok {
			for _, item := range list {
				if s, ok := item.(map[string]any); ok {
					idVal := getInt(s, "id")
					name := getString(s, "name")
					if name == "" {
						name = fmt.Sprintf("sonarr_%d", idVal)
					}
					attrs := map[string]any{
						"name":                name,
						"hostname":            getString(s, "hostname"),
						"port":                getInt(s, "port"),
						"use_ssl":             getBool(s, "useSsl"),
						"is_4k":               getBool(s, "is4k"),
						"is_default":          getBool(s, "isDefault"),
						"active_profile_name": getString(s, "activeProfileName"),
						"active_directory":    getString(s, "activeDirectory"),
						"season_folder":       getBool(s, "seasonFolder"),
					}
					if v := getString(s, "baseUrl"); v != "" {
						attrs["base_url"] = v
					}
					if v := getString(s, "animeProfileName"); v != "" {
						attrs["anime_profile_name"] = v
					}
					if v := getString(s, "animeDirectory"); v != "" {
						attrs["anime_directory"] = v
					}
					if v := getInt(s, "activeLanguageProfileId"); v != 0 {
						attrs["active_language_profile_id"] = v
					}
					discovered = append(discovered, DiscoveredResource{
						ResourceType: "seerr_sonarr_server",
						ResourceName: sanitizeName(name),
						ImportID:     strconv.FormatInt(idVal, 10),
						Attributes:   attrs,
					})
				}
			}
		}
	}

	// 9. Notification Agents
	notificationAgents := []struct {
		Endpoint     string
		ResourceType string
		TypeName     string
	}{
		{"discord", "seerr_notification_discord", "discord"},
		{"email", "seerr_notification_email", "email"},
		{"gotify", "seerr_notification_gotify", "gotify"},
		{"lunasea", "seerr_notification_lunasea", "lunasea"},
		{"ntfy", "seerr_notification_ntfy", "ntfy"},
		{"pushbullet", "seerr_notification_pushbullet", "pushbullet"},
		{"pushover", "seerr_notification_pushover", "pushover"},
		{"slack", "seerr_notification_slack", "slack"},
		{"telegram", "seerr_notification_telegram", "telegram"},
		{"webhook", "seerr_notification_webhook", "webhook"},
		{"webpush", "seerr_notification_webpush", "webpush"},
	}

	for _, na := range notificationAgents {
		agentRes, code, err := imp.get(ctx, "/api/v1/settings/notifications/"+na.Endpoint)
		if err == nil && code == 200 {
			if m, ok := agentRes.(map[string]any); ok {
				enabled := getBool(m, "enabled")
				typesVal := getInt(m, "types")
				if enabled || typesVal > 0 {
					attrs := map[string]any{
						"enabled": enabled,
						"types":   typesVal,
					}
					if options, ok := m["options"].(map[string]any); ok {
						for k, v := range options {
							if str, ok := v.(string); ok && str != "" {
								attrs[toSnakeCase(k)] = str
							} else if b, ok := v.(bool); ok {
								attrs[toSnakeCase(k)] = b
							} else if f, ok := v.(float64); ok {
								attrs[toSnakeCase(k)] = int64(f)
							}
						}
					}
					discovered = append(discovered, DiscoveredResource{
						ResourceType: na.ResourceType,
						ResourceName: na.TypeName,
						ImportID:     na.TypeName,
						Attributes:   attrs,
					})
				}
			}
		}
	}

	// 10. Override Rules
	rulesRes, code, err := imp.get(ctx, "/api/v1/settings/rules")
	if err == nil && code == 200 {
		if list, ok := rulesRes.([]any); ok {
			for _, item := range list {
				if r, ok := item.(map[string]any); ok {
					idVal := getInt(r, "id")
					name := getString(r, "name")
					if name == "" {
						name = fmt.Sprintf("rule_%d", idVal)
					}
					attrs := map[string]any{
						"name":       name,
						"enabled":    getBool(r, "enabled"),
						"media_type": getString(r, "type"),
					}
					if p := getInt(r, "priority"); p != 0 {
						attrs["priority"] = p
					}
					if s := getInt(r, "radarrServerId"); s >= 0 {
						attrs["radarr_server_id"] = s
					}
					if s := getInt(r, "sonarrServerId"); s >= 0 {
						attrs["sonarr_server_id"] = s
					}
					discovered = append(discovered, DiscoveredResource{
						ResourceType: "seerr_override_rule",
						ResourceName: sanitizeName(name),
						ImportID:     strconv.FormatInt(idVal, 10),
						Attributes:   attrs,
					})
				}
			}
		}
	}

	// 11. Users
	usersRes, code, err := imp.get(ctx, "/api/v1/user?take=100")
	if err == nil && code == 200 {
		if m, ok := usersRes.(map[string]any); ok {
			if results, ok := m["results"].([]any); ok {
				for _, item := range results {
					if u, ok := item.(map[string]any); ok {
						idVal := getInt(u, "id")
						username := getString(u, "username")
						email := getString(u, "email")
						resName := username
						if resName == "" {
							resName = fmt.Sprintf("user_%d", idVal)
						}

						attrs := map[string]any{
							"email":       email,
							"username":    username,
							"permissions": getInt(u, "permissions"),
							"user_type":   getInt(u, "userType"),
						}
						discovered = append(discovered, DiscoveredResource{
							ResourceType: "seerr_user",
							ResourceName: sanitizeName(resName),
							ImportID:     strconv.FormatInt(idVal, 10),
							Attributes:   attrs,
						})
					}
				}
			}
		}
	}

	// 12. Discover Sliders
	slidersRes, code, err := imp.get(ctx, "/api/v1/settings/discover")
	if err == nil && code == 200 {
		if list, ok := slidersRes.([]any); ok {
			for _, item := range list {
				if s, ok := item.(map[string]any); ok {
					idVal := getInt(s, "id")
					title := getString(s, "title")
					if title == "" {
						title = fmt.Sprintf("slider_%d", idVal)
					}
					attrs := map[string]any{
						"title":      title,
						"type":       getInt(s, "type"),
						"is_builtin": getBool(s, "isBuiltin"),
						"enabled":    getBool(s, "enabled"),
						"order":      getInt(s, "order"),
					}
					discovered = append(discovered, DiscoveredResource{
						ResourceType: "seerr_discover_slider",
						ResourceName: sanitizeName(title),
						ImportID:     strconv.FormatInt(idVal, 10),
						Attributes:   attrs,
					})
				}
			}
		}
	}

	return discovered, nil
}

// GenerateHCL produces clean Terraform/OpenTofu HCL code for discovered resources.
func GenerateHCL(resources []DiscoveredResource, providerHeader bool) string {
	var sb strings.Builder

	if providerHeader {
		sb.WriteString(`# --------------------------------------------------------------------------
# Terraform & OpenTofu Provider Configuration for Seerr
# Generated by seerr-import CLI
# --------------------------------------------------------------------------

terraform {
  required_version = ">= 1.5.0"
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "~> 0.38.0"
    }
  }
}

provider "seerr" {
  # Set via SEERR_URL and SEERR_API_KEY environment variables or specify below:
  # url     = "http://localhost:5055"
  # api_key = "YOUR_API_KEY"
}

`)
	}

	for _, res := range resources {
		sb.WriteString(fmt.Sprintf("resource \"%s\" \"%s\" {\n", res.ResourceType, res.ResourceName))

		keys := make([]string, 0, len(res.Attributes))
		for k := range res.Attributes {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			v := res.Attributes[k]
			switch val := v.(type) {
			case string:
				sb.WriteString(fmt.Sprintf("  %-25s = %s\n", k, strconv.Quote(val)))
			case bool:
				sb.WriteString(fmt.Sprintf("  %-25s = %t\n", k, val))
			case int64:
				sb.WriteString(fmt.Sprintf("  %-25s = %d\n", k, val))
			case int:
				sb.WriteString(fmt.Sprintf("  %-25s = %d\n", k, val))
			case []string:
				quoted := make([]string, len(val))
				for i, s := range val {
					quoted[i] = strconv.Quote(s)
				}
				sb.WriteString(fmt.Sprintf("  %-25s = [%s]\n", k, strings.Join(quoted, ", ")))
			case []int64:
				strs := make([]string, len(val))
				for i, n := range val {
					strs[i] = strconv.FormatInt(n, 10)
				}
				sb.WriteString(fmt.Sprintf("  %-25s = [%s]\n", k, strings.Join(strs, ", ")))
			}
		}
		sb.WriteString("}\n\n")
	}

	return sb.String()
}

// GenerateImportBlocks produces modern Terraform 1.5+ import blocks for discovered resources.
func GenerateImportBlocks(resources []DiscoveredResource) string {
	var sb strings.Builder
	sb.WriteString(`# --------------------------------------------------------------------------
# Terraform 1.5+ / OpenTofu Import Blocks
# Run: tofu plan -generate-config-out=generated.tf
#   or tofu plan to verify clean adoption
# --------------------------------------------------------------------------

`)
	for _, res := range resources {
		sb.WriteString("import {\n")
		sb.WriteString(fmt.Sprintf("  to = %s.%s\n", res.ResourceType, res.ResourceName))
		sb.WriteString(fmt.Sprintf("  id = \"%s\"\n", res.ImportID))
		sb.WriteString("}\n\n")
	}
	return sb.String()
}

// GenerateImportScript produces legacy shell import commands.
func GenerateImportScript(resources []DiscoveredResource) string {
	var sb strings.Builder
	sb.WriteString("#!/usr/bin/env bash\n")
	sb.WriteString("# Legacy Terraform Import Script\n")
	sb.WriteString("set -euo pipefail\n\n")

	for _, res := range resources {
		sb.WriteString(fmt.Sprintf("terraform import %s.%s %s\n", res.ResourceType, res.ResourceName, res.ImportID))
	}
	return sb.String()
}

// Helpers for data conversion.
func copyString(src, dst map[string]any, srcKey, dstKey string) {
	if v, ok := src[srcKey].(string); ok && v != "" {
		dst[dstKey] = v
	}
}

func copyBool(src, dst map[string]any, srcKey, dstKey string) {
	if v, ok := src[srcKey].(bool); ok {
		dst[dstKey] = v
	}
}

func copyInt(src, dst map[string]any, srcKey, dstKey string) {
	if v, ok := src[srcKey].(float64); ok {
		dst[dstKey] = int64(v)
	}
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getBool(m map[string]any, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func getInt(m map[string]any, key string) int64 {
	if v, ok := m[key].(float64); ok {
		return int64(v)
	}
	return 0
}

func toSnakeCase(s string) string {
	var sb strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				sb.WriteRune('_')
			}
			sb.WriteRune(r + ('a' - 'A'))
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
