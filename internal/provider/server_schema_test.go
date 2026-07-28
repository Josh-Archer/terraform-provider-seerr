package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

func TestServerResourceAttributesNoStaticDefaults(t *testing.T) {
	radarrOptComputed := map[string]string{
		"name":                   "string",
		"hostname":               "string",
		"port":                   "int64",
		"use_ssl":                "bool",
		"base_url":               "string",
		"quality_profile_name":   "string",
		"is_4k":                  "bool",
		"minimum_availability":   "string",
		"is_default":             "bool",
		"enable_scan":            "bool",
		"sync_enabled":           "bool",
		"prevent_search":         "bool",
		"tag_requests_with_user": "bool",
	}

	sonarrOptComputed := map[string]string{
		"name":                   "string",
		"hostname":               "string",
		"port":                   "int64",
		"use_ssl":                "bool",
		"base_url":               "string",
		"quality_profile_name":   "string",
		"active_anime_directory": "string",
		"is_4k":                  "bool",
		"is_default":             "bool",
		"enable_scan":            "bool",
		"enable_season_folders":  "bool",
		"sync_enabled":           "bool",
		"prevent_search":         "bool",
		"tag_requests_with_user": "bool",
	}

	tests := []struct {
		name        string
		buildSchema func(context.Context, resource.SchemaRequest, *resource.SchemaResponse)
		attributes  map[string]string
	}{
		{"radarr", (&RadarrServerResource{}).Schema, radarrOptComputed},
		{"sonarr", (&SonarrServerResource{}).Schema, sonarrOptComputed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp resource.SchemaResponse
			tt.buildSchema(context.Background(), resource.SchemaRequest{}, &resp)

			for attrName, attrType := range tt.attributes {
				attr, ok := resp.Schema.Attributes[attrName]
				if !ok {
					t.Fatalf("attribute %q missing from schema", attrName)
				}

				switch attrType {
				case "string":
					sa, ok := attr.(rschema.StringAttribute)
					if !ok {
						t.Fatalf("%s: attribute %q is not StringAttribute", tt.name, attrName)
					}
					if sa.Default != nil {
						t.Fatalf("%s: attribute %q must not have static default to avoid import clobber risk", tt.name, attrName)
					}
					if len(sa.PlanModifiers) == 0 {
						t.Fatalf("%s: attribute %q must have plan modifiers to preserve state", tt.name, attrName)
					}
				case "bool":
					ba, ok := attr.(rschema.BoolAttribute)
					if !ok {
						t.Fatalf("%s: attribute %q is not BoolAttribute", tt.name, attrName)
					}
					if ba.Default != nil {
						t.Fatalf("%s: attribute %q must not have static default to avoid import clobber risk", tt.name, attrName)
					}
					if len(ba.PlanModifiers) == 0 {
						t.Fatalf("%s: attribute %q must have plan modifiers to preserve state", tt.name, attrName)
					}
				case "int64":
					ia, ok := attr.(rschema.Int64Attribute)
					if !ok {
						t.Fatalf("%s: attribute %q is not Int64Attribute", tt.name, attrName)
					}
					if ia.Default != nil {
						t.Fatalf("%s: attribute %q must not have static default to avoid import clobber risk", tt.name, attrName)
					}
					if len(ia.PlanModifiers) == 0 {
						t.Fatalf("%s: attribute %q must have plan modifiers to preserve state", tt.name, attrName)
					}
				}
			}
		})
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
