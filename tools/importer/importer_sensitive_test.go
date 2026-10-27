// Copyright (c) Josh Archer
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/Josh-Archer/terraform-provider-seerr/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// TestSensitiveAttributeMappingMatchesProviderSchema checks the importer's
// sensitive-attribute mapping against provider resource schemas in both
// directions: every field the importer classifies as sensitive is Sensitive
// in the schema, and every Sensitive schema attribute on those resources is
// classified as sensitive by the importer.
func TestSensitiveAttributeMappingMatchesProviderSchema(t *testing.T) {
	t.Run("notification_agents", func(t *testing.T) {
		compared := map[string]struct{}{}
		for _, agent := range notificationAgents {
			compared["seerr_notification_"+agent] = struct{}{}
			t.Run(agent, func(t *testing.T) {
				assertNotificationMappingMatchesSchema(t, agent)
			})
		}
		for _, typeName := range registeredResourceTypeNames(t) {
			if !strings.HasPrefix(typeName, "seerr_notification_") {
				continue
			}
			if _, ok := compared[typeName]; ok {
				continue
			}
			paths := schemaSensitivePaths(t, typeName)
			if len(paths) > 0 {
				t.Errorf("%s has sensitive attributes %v but the importer has no mapping", typeName, paths)
			}
		}
	})

	t.Run("api_keys", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/api/v1/settings/jellyfin":
				_ = json.NewEncoder(w).Encode(map[string]any{
					"ip": "jellyfin.local", "port": 8096, "apiKey": "fixture-jellyfin-api-key",
				})
			case "/api/v1/settings/emby":
				_ = json.NewEncoder(w).Encode(map[string]any{
					"ip": "emby.local", "port": 8096, "apiKey": "fixture-emby-api-key",
				})
			case "/api/v1/settings/tautulli":
				_ = json.NewEncoder(w).Encode(map[string]any{
					"hostname": "tautulli.local", "port": 8181, "apiKey": "fixture-tautulli-api-key",
				})
			case "/api/v1/settings/radarr":
				_ = json.NewEncoder(w).Encode([]any{map[string]any{
					"id": 1, "name": "Radarr", "hostname": "radarr.local", "port": 7878,
					"apiKey": "fixture-radarr-api-key", "activeProfileId": 1, "activeDirectory": "/movies",
				}})
			case "/api/v1/settings/sonarr":
				_ = json.NewEncoder(w).Encode([]any{map[string]any{
					"id": 1, "name": "Sonarr", "hostname": "sonarr.local", "port": 8989,
					"apiKey": "fixture-sonarr-api-key", "activeProfileId": 1, "activeDirectory": "/tv",
				}})
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		imp := NewImporter(ImporterConfig{BaseURL: server.URL, APIKey: "fixture-importer-key", HTTPClient: server.Client()})
		ctx := context.Background()
		cases := []struct {
			typeName string
			found    []DiscoveredResource
		}{
			{"seerr_jellyfin_settings", imp.discoverMediaServer(ctx, "/api/v1/settings/jellyfin", "seerr_jellyfin_settings", "jellyfin", "jellyfinForgotPasswordUrl", "jellyfin_forgot_password_url")},
			{"seerr_emby_settings", imp.discoverMediaServer(ctx, "/api/v1/settings/emby", "seerr_emby_settings", "emby", "embyForgotPasswordUrl", "emby_forgot_password_url")},
			{"seerr_tautulli_settings", imp.discoverTautulli(ctx)},
			{"seerr_radarr_server", imp.discoverArr(ctx, "/api/v1/settings/radarr", "seerr_radarr_server", "radarr", false)},
			{"seerr_sonarr_server", imp.discoverArr(ctx, "/api/v1/settings/sonarr", "seerr_sonarr_server", "sonarr", true)},
		}
		for _, tc := range cases {
			t.Run(tc.typeName, func(t *testing.T) {
				if len(tc.found) != 1 {
					t.Fatalf("discovered %d resources, want 1", len(tc.found))
				}
				got := sensitivePathsFromAttrs(tc.found[0].Attributes)
				want := schemaSensitivePaths(t, tc.typeName)
				if !slices.Equal(got, want) {
					t.Errorf("importer sensitive paths = %#v, schema Sensitive paths = %#v", got, want)
				}
			})
		}
	})
}

func assertNotificationMappingMatchesSchema(t *testing.T, agent string) {
	t.Helper()
	typeName := "seerr_notification_" + agent
	sch := providerResourceSchema(t, typeName)
	attr, ok := sch.Attributes[agent]
	if !ok {
		t.Fatalf("%s schema has no %q attribute", typeName, agent)
	}
	nested, ok := attr.(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("%s.%s is %T, want SingleNestedAttribute", typeName, agent, attr)
	}

	classified := map[string]bool{}
	var got []string
	for _, field := range notificationOptionFields(agent) {
		schemaAttr, ok := nested.Attributes[field.HCL]
		if !ok {
			t.Errorf("importer maps %s.%s but the provider schema has no such attribute", agent, field.HCL)
			continue
		}
		sensitive := field.Kind == "sensitive"
		if sensitive != schemaAttr.IsSensitive() {
			t.Errorf("%s.%s importer sensitive=%t, schema Sensitive=%t", agent, field.HCL, sensitive, schemaAttr.IsSensitive())
		}
		classified[field.HCL] = sensitive
		if sensitive {
			got = append(got, agent+"."+field.HCL)
		}
	}
	for name, schemaAttr := range nested.Attributes {
		if schemaAttr.IsSensitive() && !classified[name] {
			t.Errorf("schema %s.%s is Sensitive but the importer mapping does not classify it as sensitive", agent, name)
		}
	}
	sort.Strings(got)
	want := schemaSensitivePaths(t, typeName)
	if !slices.Equal(got, want) {
		t.Errorf("importer sensitive paths = %#v, schema Sensitive paths = %#v", got, want)
	}
}

func registeredResourceTypeNames(t *testing.T) []string {
	t.Helper()
	names := make([]string, 0, len(provider.RegisteredResources))
	for _, factory := range provider.RegisteredResources {
		names = append(names, resourceTypeName(t, factory()))
	}
	return names
}

func resourceTypeName(t *testing.T, res resource.Resource) string {
	t.Helper()
	var meta resource.MetadataResponse
	res.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "seerr"}, &meta)
	if meta.TypeName == "" {
		t.Fatalf("%T has an empty type name", res)
	}
	return meta.TypeName
}

func providerResourceSchema(t *testing.T, typeName string) schema.Schema {
	t.Helper()
	for _, factory := range provider.RegisteredResources {
		res := factory()
		if resourceTypeName(t, res) != typeName {
			continue
		}
		var resp resource.SchemaResponse
		res.Schema(context.Background(), resource.SchemaRequest{}, &resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("%s schema: %v", typeName, resp.Diagnostics)
		}
		return resp.Schema
	}
	t.Fatalf("resource %s is not registered", typeName)
	return schema.Schema{}
}

func schemaSensitivePaths(t *testing.T, typeName string) []string {
	t.Helper()
	sch := providerResourceSchema(t, typeName)
	var paths []string
	for name, attr := range sch.Attributes {
		collectSensitiveAttribute("", name, attr, &paths)
	}
	for name, block := range sch.Blocks {
		collectSensitiveBlock("", name, block, &paths)
	}
	sort.Strings(paths)
	return paths
}

func collectSensitiveAttribute(prefix, name string, attr schema.Attribute, out *[]string) {
	path := joinSchemaPath(prefix, name)
	switch nested := attr.(type) {
	case schema.SingleNestedAttribute:
		for childName, child := range nested.Attributes {
			collectSensitiveAttribute(path, childName, child, out)
		}
	case schema.ListNestedAttribute:
		for childName, child := range nested.NestedObject.Attributes {
			collectSensitiveAttribute(path, childName, child, out)
		}
	case schema.SetNestedAttribute:
		for childName, child := range nested.NestedObject.Attributes {
			collectSensitiveAttribute(path, childName, child, out)
		}
	case schema.MapNestedAttribute:
		for childName, child := range nested.NestedObject.Attributes {
			collectSensitiveAttribute(path, childName, child, out)
		}
	default:
		if attr.IsSensitive() {
			*out = append(*out, path)
		}
	}
}

func collectSensitiveBlock(prefix, name string, block schema.Block, out *[]string) {
	path := joinSchemaPath(prefix, name)
	switch nested := block.(type) {
	case schema.SingleNestedBlock:
		for childName, child := range nested.Attributes {
			collectSensitiveAttribute(path, childName, child, out)
		}
		for childName, child := range nested.Blocks {
			collectSensitiveBlock(path, childName, child, out)
		}
	case schema.ListNestedBlock:
		for childName, child := range nested.NestedObject.Attributes {
			collectSensitiveAttribute(path, childName, child, out)
		}
		for childName, child := range nested.NestedObject.Blocks {
			collectSensitiveBlock(path, childName, child, out)
		}
	case schema.SetNestedBlock:
		for childName, child := range nested.NestedObject.Attributes {
			collectSensitiveAttribute(path, childName, child, out)
		}
		for childName, child := range nested.NestedObject.Blocks {
			collectSensitiveBlock(path, childName, child, out)
		}
	}
}

func joinSchemaPath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

func sensitivePathsFromAttrs(attrs map[string]any) []string {
	var paths []string
	collectSensitiveAttrPaths("", attrs, &paths)
	sort.Strings(paths)
	return paths
}

func collectSensitiveAttrPaths(prefix string, attrs map[string]any, out *[]string) {
	for key, value := range attrs {
		path := joinSchemaPath(prefix, key)
		switch typed := value.(type) {
		case sensitiveValue:
			*out = append(*out, path)
		case nestedObject:
			collectSensitiveAttrPaths(path, typed.Attrs, out)
		}
	}
}

type secretCanary struct {
	value    string
	variable string
}

func secretCanaryFixture(t *testing.T) (*httptest.Server, []secretCanary) {
	t.Helper()
	specs := []struct {
		id       string
		variable string
	}{
		{"sonarr-api-key", sensitiveVarName("seerr_sonarr_server", "sonarr", "api_key")},
		{"jellyfin-api-key", sensitiveVarName("seerr_jellyfin_settings", "jellyfin", "api_key")},
		{"tautulli-api-key", sensitiveVarName("seerr_tautulli_settings", "tautulli", "api_key")},
		{"email-auth-pass", sensitiveVarName("seerr_notification_email", "email", "auth_pass")},
		{"pgp-private-key", sensitiveVarName("seerr_notification_email", "email", "pgp_private_key")},
		{"pgp-password", sensitiveVarName("seerr_notification_email", "email", "pgp_password")},
		{"telegram-bot-api", sensitiveVarName("seerr_notification_telegram", "telegram", "bot_api")},
		{"pushbullet-access-token", sensitiveVarName("seerr_notification_pushbullet", "pushbullet", "access_token")},
		{"pushover-access-token", sensitiveVarName("seerr_notification_pushover", "pushover", "access_token")},
		{"pushover-user-token", sensitiveVarName("seerr_notification_pushover", "pushover", "user_token")},
		{"ntfy-password", sensitiveVarName("seerr_notification_ntfy", "ntfy", "password")},
		{"ntfy-token", sensitiveVarName("seerr_notification_ntfy", "ntfy", "token")},
		{"webhook-auth-header", sensitiveVarName("seerr_notification_webhook", "webhook", "auth_header")},
		{"gotify-token", sensitiveVarName("seerr_notification_gotify", "gotify", "token")},
		{"discord-webhook-url", sensitiveVarName("seerr_notification_discord", "discord", "webhook_url")},
		{"slack-webhook-url", sensitiveVarName("seerr_notification_slack", "slack", "webhook_url")},
		{"lunasea-webhook-url", sensitiveVarName("seerr_notification_lunasea", "lunasea", "webhook_url")},
		{"webhook-url", sensitiveVarName("seerr_notification_webhook", "webhook", "webhook_url")},
	}
	values := make(map[string]string, len(specs))
	canaries := make([]secretCanary, len(specs))
	for i, spec := range specs {
		value := fakeCanary(spec.id)
		values[spec.id] = value
		canaries[i] = secretCanary{value: value, variable: spec.variable}
	}

	notify := func(options map[string]any) map[string]any {
		return map[string]any{"enabled": true, "types": 2, "options": options}
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var payload any
		switch r.URL.Path {
		case "/api/v1/settings/jellyfin":
			payload = map[string]any{"ip": "jellyfin.local", "port": 8096, "apiKey": values["jellyfin-api-key"]}
		case "/api/v1/settings/tautulli":
			payload = map[string]any{"hostname": "tautulli.local", "port": 8181, "apiKey": values["tautulli-api-key"]}
		case "/api/v1/settings/sonarr":
			payload = []any{map[string]any{
				"id": 1, "name": "Sonarr", "hostname": "sonarr.local", "port": 8989,
				"apiKey": values["sonarr-api-key"], "activeProfileId": 1, "activeDirectory": "/tv",
			}}
		case "/api/v1/settings/notifications/discord":
			payload = notify(map[string]any{"webhookUrl": values["discord-webhook-url"], "botUsername": "alerts"})
		case "/api/v1/settings/notifications/slack":
			payload = notify(map[string]any{"webhookUrl": values["slack-webhook-url"]})
		case "/api/v1/settings/notifications/email":
			payload = notify(map[string]any{
				"authPass": values["email-auth-pass"], "pgpPrivateKey": values["pgp-private-key"], "pgpPassword": values["pgp-password"],
			})
		case "/api/v1/settings/notifications/lunasea":
			payload = notify(map[string]any{"webhookUrl": values["lunasea-webhook-url"], "profileName": "default"})
		case "/api/v1/settings/notifications/telegram":
			payload = notify(map[string]any{"botAPI": values["telegram-bot-api"], "chatId": "1"})
		case "/api/v1/settings/notifications/pushbullet":
			payload = notify(map[string]any{"accessToken": values["pushbullet-access-token"]})
		case "/api/v1/settings/notifications/pushover":
			payload = notify(map[string]any{
				"accessToken": values["pushover-access-token"], "userToken": values["pushover-user-token"],
			})
		case "/api/v1/settings/notifications/ntfy":
			payload = notify(map[string]any{"password": values["ntfy-password"], "token": values["ntfy-token"]})
		case "/api/v1/settings/notifications/webhook":
			payload = notify(map[string]any{
				"webhookUrl": values["webhook-url"], "authHeader": values["webhook-auth-header"], "jsonPayload": "{}",
			})
		case "/api/v1/settings/notifications/gotify":
			payload = notify(map[string]any{"token": values["gotify-token"], "url": "https://gotify.example.test"})
		default:
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	return server, canaries
}

func assertCanariesAbsent(t *testing.T, label, body string, canaries []secretCanary) {
	t.Helper()
	for _, canary := range canaries {
		if strings.Contains(body, canary.value) {
			t.Errorf("%s contains %s", label, canary.variable)
		}
	}
}

func assertCanaryVariables(t *testing.T, hcl, variables string, canaries []secretCanary) {
	t.Helper()
	for _, canary := range canaries {
		if !strings.Contains(hcl, "var."+canary.variable) {
			t.Errorf("generated HCL missing var.%s", canary.variable)
		}
		if !strings.Contains(variables, `variable "`+canary.variable+`"`) {
			t.Errorf("variables.tf missing %s", canary.variable)
		}
	}
}

func TestCanarySecretNotWrittenToGeneratedFiles(t *testing.T) {
	server, canaries := secretCanaryFixture(t)
	defer server.Close()

	imp := NewImporter(ImporterConfig{BaseURL: server.URL, APIKey: fakeCanary("importer-key"), HTTPClient: server.Client()})
	resources, err := imp.DiscoverAll(context.Background())
	if err != nil {
		t.Fatalf("DiscoverAll: %v", err)
	}
	if len(resources) != 13 {
		types := make([]string, len(resources))
		for i, resource := range resources {
			types[i] = resource.ResourceType
		}
		t.Fatalf("discovered %d resources, want 13: %v", len(resources), types)
	}

	generated := map[string]string{
		"main.tf":      GenerateHCL(resources, true),
		"variables.tf": GenerateVariables(resources),
		"imports.tf":   GenerateImportBlocks(resources),
		"import.sh":    GenerateImportScript(resources),
	}
	for name, body := range generated {
		assertCanariesAbsent(t, name, body, canaries)
	}
	assertCanaryVariables(t, generated["main.tf"], generated["variables.tf"], canaries)
}

func TestCLISecretCanaryLeakProtection(t *testing.T) {
	server, canaries := secretCanaryFixture(t)
	defer server.Close()

	cliKey := fakeCanary("cli-api-key")
	outDir := t.TempDir()
	var stdout, stderr bytes.Buffer
	args := []string{"-url", server.URL, "-api-key", cliKey, "-out-dir", outDir, "-format", "all"}
	if code := run(args, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d\nstderr:\n%s", code, stderr.String())
	}

	combined := stdout.String() + stderr.String()
	assertCanariesAbsent(t, "CLI output", combined, canaries)
	if strings.Contains(combined, cliKey) {
		t.Errorf("CLI output contains the importer API key")
	}

	files := []string{"main.tf", "variables.tf", "imports.tf", "import.sh"}
	bodies := make(map[string]string, len(files))
	for _, name := range files {
		content, err := os.ReadFile(filepath.Join(outDir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		bodies[name] = string(content)
		assertCanariesAbsent(t, name, bodies[name], canaries)
		if strings.Contains(bodies[name], cliKey) {
			t.Errorf("%s contains the importer API key", name)
		}
	}
	assertCanaryVariables(t, bodies["main.tf"], bodies["variables.tf"], canaries)
}
