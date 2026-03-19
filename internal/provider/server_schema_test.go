package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestRadarrServerResourceSchemaOmitsRawJSON(t *testing.T) {
	var resp resource.SchemaResponse
	(&RadarrServerResource{}).Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if _, ok := resp.Schema.Attributes["response_json"]; ok {
		t.Fatal("radarr server resource should not expose response_json")
	}
}

func TestSonarrServerResourceSchemaOmitsRawJSON(t *testing.T) {
	var resp resource.SchemaResponse
	(&SonarrServerResource{}).Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if _, ok := resp.Schema.Attributes["response_json"]; ok {
		t.Fatal("sonarr server resource should not expose response_json")
	}
}

func TestRadarrServerDataSourceSchemaIsTyped(t *testing.T) {
	var resp datasource.SchemaResponse
	(&RadarrServerDataSource{}).Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	for _, name := range []string{
		"id",
		"server_id",
		"name",
		"hostname",
		"port",
		"api_key",
		"use_ssl",
		"base_url",
		"quality_profile_id",
		"quality_profile_name",
		"active_directory",
		"is_4k",
		"minimum_availability",
		"tags",
		"is_default",
		"enable_scan",
		"sync_enabled",
		"prevent_search",
		"tag_requests_with_user",
	} {
		if _, ok := resp.Schema.Attributes[name]; !ok {
			t.Fatalf("radarr server data source missing %q attribute", name)
		}
	}

	for _, name := range []string{"response_json", "status_code"} {
		if _, ok := resp.Schema.Attributes[name]; ok {
			t.Fatalf("radarr server data source should not expose %q", name)
		}
	}
}

func TestSonarrServerDataSourceSchemaIsTyped(t *testing.T) {
	var resp datasource.SchemaResponse
	(&SonarrServerDataSource{}).Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	for _, name := range []string{
		"id",
		"server_id",
		"name",
		"hostname",
		"port",
		"api_key",
		"use_ssl",
		"base_url",
		"quality_profile_id",
		"quality_profile_name",
		"active_directory",
		"active_anime_directory",
		"tags",
		"anime_tags",
		"is_4k",
		"is_default",
		"enable_scan",
		"enable_season_folders",
		"sync_enabled",
		"prevent_search",
		"tag_requests_with_user",
	} {
		if _, ok := resp.Schema.Attributes[name]; !ok {
			t.Fatalf("sonarr server data source missing %q attribute", name)
		}
	}

	for _, name := range []string{"response_json", "status_code"} {
		if _, ok := resp.Schema.Attributes[name]; ok {
			t.Fatalf("sonarr server data source should not expose %q", name)
		}
	}
}
