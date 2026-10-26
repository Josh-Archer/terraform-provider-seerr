package provider

import (
	"context"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	pschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// credentialAttributeNames must stay Sensitive wherever they appear. The test
// fails if a listed attribute loses Sensitive, and if a new attribute with one
// of these names is added without Sensitive.
var credentialAttributeNames = map[string]struct{}{
	"access_token":               {},
	"api_key":                    {},
	"auth_header":                {},
	"auth_pass":                  {},
	"bot_api":                    {},
	"jellyfin_password":          {},
	"password":                   {},
	"pgp_password":               {},
	"pgp_private_key":            {},
	"plex_auth_token":            {},
	"plex_token":                 {},
	"pushbullet_access_token":    {},
	"pushover_application_token": {},
	"pushover_user_key":          {},
	"token":                      {},
	"user_token":                 {},
	"webhook_url":                {},
}

// reviewedPublicURLNames are URL attributes that do not embed credentials.
// A new *url* leaf must be added here or to credentialAttributeNames.
var reviewedPublicURLNames = map[string]struct{}{
	"application_url":              {},
	"base_url":                     {},
	"bot_avatar_url":               {},
	"emby_forgot_password_url":     {},
	"external_url":                 {},
	"jellyfin_forgot_password_url": {},
	"url":                          {},
	"url_base":                     {},
}

type schemaLeaf struct {
	target    string
	path      string
	name      string
	sensitive bool
}

func TestSensitiveSchemaInventory(t *testing.T) {
	leaves := collectAllSchemaLeaves(t)

	t.Run("credential_attributes_are_sensitive", func(t *testing.T) {
		var missing []string
		for _, leaf := range leaves {
			if _, ok := credentialAttributeNames[leaf.name]; !ok {
				continue
			}
			if !leaf.sensitive {
				missing = append(missing, leaf.target+"."+leaf.path)
			}
		}
		if len(missing) > 0 {
			t.Errorf("credential-bearing attributes missing Sensitive: %v", missing)
		}
	})

	t.Run("url_attributes_are_classified", func(t *testing.T) {
		var unclassified []string
		for _, leaf := range leaves {
			if !isURLAttributeName(leaf.name) {
				continue
			}
			_, credential := credentialAttributeNames[leaf.name]
			_, reviewed := reviewedPublicURLNames[leaf.name]
			if !credential && !reviewed {
				unclassified = append(unclassified, leaf.target+"."+leaf.path)
			}
			if credential && !leaf.sensitive {
				t.Errorf("%s.%s is a credential URL but is not Sensitive", leaf.target, leaf.path)
			}
		}
		if len(unclassified) > 0 {
			t.Errorf("URL attributes not in credentialAttributeNames or reviewedPublicURLNames: %v", unclassified)
		}
	})

	t.Run("sensitive_paths_match_golden_list", func(t *testing.T) {
		got := map[string][]string{}
		for _, leaf := range leaves {
			if !leaf.sensitive {
				continue
			}
			got[leaf.target] = append(got[leaf.target], leaf.path)
		}
		for target, paths := range got {
			sort.Strings(paths)
			got[target] = paths
		}

		want := expectedSensitivePaths()
		for target, paths := range want {
			sort.Strings(paths)
			want[target] = paths
		}

		var targets []string
		seen := map[string]struct{}{}
		for target := range got {
			targets = append(targets, target)
			seen[target] = struct{}{}
		}
		for target := range want {
			if _, ok := seen[target]; !ok {
				targets = append(targets, target)
			}
		}
		sort.Strings(targets)

		for _, target := range targets {
			if !slices.Equal(got[target], want[target]) {
				t.Errorf("%s sensitive paths = %#v, want %#v", target, got[target], want[target])
			}
		}
	})
}

func TestNotificationAgentWebhookURLSensitive(t *testing.T) {
	found := 0
	for agent, attr := range notificationAgentResourceOptionAttributes() {
		nested, ok := attr.(rschema.SingleNestedAttribute)
		if !ok {
			continue
		}
		if w, ok := nested.Attributes["webhook_url"]; ok {
			found++
			if !w.IsSensitive() {
				t.Errorf("resource %s.webhook_url is not Sensitive", agent)
			}
		}
	}
	if found != 4 {
		t.Errorf("expected 4 resource webhook_url attributes, found %d", found)
	}

	found = 0
	for agent, attr := range notificationAgentDataSourceOptionAttributes() {
		nested, ok := attr.(dschema.SingleNestedAttribute)
		if !ok {
			continue
		}
		if w, ok := nested.Attributes["webhook_url"]; ok {
			found++
			if !w.IsSensitive() {
				t.Errorf("data source %s.webhook_url is not Sensitive", agent)
			}
		}
	}
	if found != 4 {
		t.Errorf("expected 4 data source webhook_url attributes, found %d", found)
	}
}

func isURLAttributeName(name string) bool {
	return name == "url" || strings.HasPrefix(name, "url_") || strings.HasSuffix(name, "_url")
}

func collectAllSchemaLeaves(t *testing.T) []schemaLeaf {
	t.Helper()
	var leaves []schemaLeaf

	p := New("test")()
	var providerSchema provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &providerSchema)
	if providerSchema.Diagnostics.HasError() {
		t.Fatalf("provider schema diagnostics: %v", providerSchema.Diagnostics)
	}
	for name, attr := range providerSchema.Schema.Attributes {
		walkProviderAttribute("provider.seerr", "", name, attr, &leaves)
	}

	for _, newRes := range RegisteredResources {
		res := newRes()
		target := lifecycleTypeName(t, res)
		sch := lifecycleSchema(t, res)
		for name, attr := range sch.Attributes {
			walkResourceAttribute("resource."+target, "", name, attr, &leaves)
		}
		for name, block := range sch.Blocks {
			walkResourceBlock("resource."+target, "", name, block, &leaves)
		}
	}

	for _, newDS := range RegisteredDataSources {
		ds := newDS()
		target := dataSourceTypeName(t, ds)
		sch := dataSourceSchema(t, ds)
		for name, attr := range sch.Attributes {
			walkDataSourceAttribute("data."+target, "", name, attr, &leaves)
		}
		for name, block := range sch.Blocks {
			walkDataSourceBlock("data."+target, "", name, block, &leaves)
		}
	}

	return leaves
}

func dataSourceTypeName(t *testing.T, ds datasource.DataSource) string {
	t.Helper()
	var resp datasource.MetadataResponse
	ds.Metadata(t.Context(), datasource.MetadataRequest{ProviderTypeName: "seerr"}, &resp)
	if resp.TypeName == "" {
		t.Fatalf("%T has an empty type name", ds)
	}
	return resp.TypeName
}

func dataSourceSchema(t *testing.T, ds datasource.DataSource) dschema.Schema {
	t.Helper()
	var resp datasource.SchemaResponse
	ds.Schema(t.Context(), datasource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("%T schema diagnostics: %v", ds, resp.Diagnostics)
	}
	return resp.Schema
}

func walkProviderAttribute(target, prefix, name string, attr pschema.Attribute, out *[]schemaLeaf) {
	path := joinSchemaPath(prefix, name)
	switch nested := attr.(type) {
	case pschema.SingleNestedAttribute:
		for childName, child := range nested.Attributes {
			walkProviderAttribute(target, path, childName, child, out)
		}
	case pschema.ListNestedAttribute:
		for childName, child := range nested.NestedObject.Attributes {
			walkProviderAttribute(target, path, childName, child, out)
		}
	case pschema.SetNestedAttribute:
		for childName, child := range nested.NestedObject.Attributes {
			walkProviderAttribute(target, path, childName, child, out)
		}
	case pschema.MapNestedAttribute:
		for childName, child := range nested.NestedObject.Attributes {
			walkProviderAttribute(target, path, childName, child, out)
		}
	default:
		*out = append(*out, schemaLeaf{target: target, path: path, name: name, sensitive: attr.IsSensitive()})
	}
}

func walkResourceAttribute(target, prefix, name string, attr rschema.Attribute, out *[]schemaLeaf) {
	path := joinSchemaPath(prefix, name)
	switch nested := attr.(type) {
	case rschema.SingleNestedAttribute:
		for childName, child := range nested.Attributes {
			walkResourceAttribute(target, path, childName, child, out)
		}
	case rschema.ListNestedAttribute:
		for childName, child := range nested.NestedObject.Attributes {
			walkResourceAttribute(target, path, childName, child, out)
		}
	case rschema.SetNestedAttribute:
		for childName, child := range nested.NestedObject.Attributes {
			walkResourceAttribute(target, path, childName, child, out)
		}
	case rschema.MapNestedAttribute:
		for childName, child := range nested.NestedObject.Attributes {
			walkResourceAttribute(target, path, childName, child, out)
		}
	default:
		*out = append(*out, schemaLeaf{target: target, path: path, name: name, sensitive: attr.IsSensitive()})
	}
}

func walkResourceBlock(target, prefix, name string, block rschema.Block, out *[]schemaLeaf) {
	path := joinSchemaPath(prefix, name)
	switch nested := block.(type) {
	case rschema.SingleNestedBlock:
		for childName, child := range nested.Attributes {
			walkResourceAttribute(target, path, childName, child, out)
		}
		for childName, child := range nested.Blocks {
			walkResourceBlock(target, path, childName, child, out)
		}
	case rschema.ListNestedBlock:
		for childName, child := range nested.NestedObject.Attributes {
			walkResourceAttribute(target, path, childName, child, out)
		}
		for childName, child := range nested.NestedObject.Blocks {
			walkResourceBlock(target, path, childName, child, out)
		}
	case rschema.SetNestedBlock:
		for childName, child := range nested.NestedObject.Attributes {
			walkResourceAttribute(target, path, childName, child, out)
		}
		for childName, child := range nested.NestedObject.Blocks {
			walkResourceBlock(target, path, childName, child, out)
		}
	}
}

func walkDataSourceAttribute(target, prefix, name string, attr dschema.Attribute, out *[]schemaLeaf) {
	path := joinSchemaPath(prefix, name)
	switch nested := attr.(type) {
	case dschema.SingleNestedAttribute:
		for childName, child := range nested.Attributes {
			walkDataSourceAttribute(target, path, childName, child, out)
		}
	case dschema.ListNestedAttribute:
		for childName, child := range nested.NestedObject.Attributes {
			walkDataSourceAttribute(target, path, childName, child, out)
		}
	case dschema.SetNestedAttribute:
		for childName, child := range nested.NestedObject.Attributes {
			walkDataSourceAttribute(target, path, childName, child, out)
		}
	case dschema.MapNestedAttribute:
		for childName, child := range nested.NestedObject.Attributes {
			walkDataSourceAttribute(target, path, childName, child, out)
		}
	default:
		*out = append(*out, schemaLeaf{target: target, path: path, name: name, sensitive: attr.IsSensitive()})
	}
}

func walkDataSourceBlock(target, prefix, name string, block dschema.Block, out *[]schemaLeaf) {
	path := joinSchemaPath(prefix, name)
	switch nested := block.(type) {
	case dschema.SingleNestedBlock:
		for childName, child := range nested.Attributes {
			walkDataSourceAttribute(target, path, childName, child, out)
		}
		for childName, child := range nested.Blocks {
			walkDataSourceBlock(target, path, childName, child, out)
		}
	case dschema.ListNestedBlock:
		for childName, child := range nested.NestedObject.Attributes {
			walkDataSourceAttribute(target, path, childName, child, out)
		}
		for childName, child := range nested.NestedObject.Blocks {
			walkDataSourceBlock(target, path, childName, child, out)
		}
	case dschema.SetNestedBlock:
		for childName, child := range nested.NestedObject.Attributes {
			walkDataSourceAttribute(target, path, childName, child, out)
		}
		for childName, child := range nested.NestedObject.Blocks {
			walkDataSourceBlock(target, path, childName, child, out)
		}
	}
}

func joinSchemaPath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

func expectedSensitivePaths() map[string][]string {
	return map[string][]string{
		"provider.seerr": {"api_key", "plex_token"},

		"resource.seerr_api_key":           {"api_key"},
		"resource.seerr_emby_settings":     {"api_key"},
		"resource.seerr_jellyfin_settings": {"api_key"},
		"resource.seerr_network_settings":  {"proxy.password"},
		"resource.seerr_notification_discord": {
			"discord.webhook_url",
		},
		"resource.seerr_notification_email": {
			"email.auth_pass",
			"email.pgp_password",
			"email.pgp_private_key",
		},
		"resource.seerr_notification_gotify":     {"gotify.token"},
		"resource.seerr_notification_lunasea":    {"lunasea.webhook_url"},
		"resource.seerr_notification_ntfy":       {"ntfy.password", "ntfy.token"},
		"resource.seerr_notification_pushbullet": {"pushbullet.access_token"},
		"resource.seerr_notification_pushover":   {"pushover.access_token", "pushover.user_token"},
		"resource.seerr_notification_slack":      {"slack.webhook_url"},
		"resource.seerr_notification_telegram":   {"telegram.bot_api"},
		"resource.seerr_notification_webhook":    {"webhook.auth_header", "webhook.webhook_url"},
		"resource.seerr_radarr_server":           {"api_key"},
		"resource.seerr_radarr_server_test":      {"api_key"},
		"resource.seerr_sonarr_server":           {"api_key"},
		"resource.seerr_sonarr_server_test":      {"api_key"},
		"resource.seerr_tautulli_settings":       {"api_key"},
		"resource.seerr_user": {
			"jellyfin_password",
			"notification_settings.pushbullet_access_token",
			"notification_settings.pushover_application_token",
			"notification_settings.pushover_user_key",
			"plex_auth_token",
		},
		"resource.seerr_user_notification_settings": {
			"pushbullet_access_token",
			"pushover_application_token",
			"pushover_user_key",
		},

		"data.seerr_api_key":           {"api_key"},
		"data.seerr_api_request":       {"response_body_json"},
		"data.seerr_email_settings":    {"email.auth_pass", "email.pgp_password", "email.pgp_private_key"},
		"data.seerr_jellyfin_settings": {"api_key"},
		"data.seerr_network_settings":  {"proxy.password"},
		"data.seerr_notification_discord": {
			"discord.webhook_url",
		},
		"data.seerr_notification_email": {
			"email.auth_pass",
			"email.pgp_password",
			"email.pgp_private_key",
		},
		"data.seerr_notification_gotify":      {"gotify.token"},
		"data.seerr_notification_lunasea":     {"lunasea.webhook_url"},
		"data.seerr_notification_ntfy":        {"ntfy.password", "ntfy.token"},
		"data.seerr_notification_pushbullet":  {"pushbullet.access_token"},
		"data.seerr_notification_pushover":    {"pushover.access_token", "pushover.user_token"},
		"data.seerr_notification_slack":       {"slack.webhook_url"},
		"data.seerr_notification_telegram":    {"telegram.bot_api"},
		"data.seerr_notification_webhook":     {"webhook.auth_header", "webhook.webhook_url"},
		"data.seerr_plex_devices":             {"devices.access_token"},
		"data.seerr_pushbullet_settings":      {"pushbullet.access_token"},
		"data.seerr_pushover_sounds":          {"token"},
		"data.seerr_radarr_quality_profile":   {"api_key"},
		"data.seerr_radarr_root_folders":      {"api_key"},
		"data.seerr_radarr_server":            {"api_key"},
		"data.seerr_radarr_tags":              {"api_key"},
		"data.seerr_sonarr_language_profiles": {"api_key"},
		"data.seerr_sonarr_quality_profile":   {"api_key"},
		"data.seerr_sonarr_root_folders":      {"api_key"},
		"data.seerr_sonarr_server":            {"api_key"},
		"data.seerr_sonarr_tags":              {"api_key"},
		"data.seerr_tautulli_settings":        {"api_key"},
		"data.seerr_user_notification_settings": {
			"pushbullet_access_token",
			"pushover_application_token",
			"pushover_user_key",
		},
	}
}
