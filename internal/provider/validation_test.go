package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestURLRegex(t *testing.T) {
	regex := urlRegex()
	tests := []struct {
		url      string
		expected bool
	}{
		{"http://localhost", true},
		{"https://example.com", true},
		{"http://localhost:5055", true},
		{"https://example.com/seerr", true},
		{"http://localhost/", false},
		{"https://example.com/", false},
		{"http://localhost:5055/", false},
		{"https://example.com/seerr/", false},
		{"localhost", false},
		{"ftp://example.com", false},
		{"http://", false},
	}

	for _, test := range tests {
		if got := regex.MatchString(test.url); got != test.expected {
			t.Errorf("urlRegex().MatchString(%q) = %v; want %v", test.url, got, test.expected)
		}
	}
}

func TestPortValidationOnServerAndSettingsResources(t *testing.T) {
	resources := map[string]func() resource.Resource{
		"radarr_server":     NewRadarrServerResource,
		"sonarr_server":     NewSonarrServerResource,
		"tautulli_settings": NewTautulliSettingsResource,
		"plex_settings":     NewPlexSettingsResource,
		"jellyfin_settings": NewJellyfinSettingsResource,
		"emby_settings":     NewEmbySettingsResource,
	}

	for name, factory := range resources {
		t.Run(name, func(t *testing.T) {
			var resp resource.SchemaResponse
			factory().Schema(context.Background(), resource.SchemaRequest{}, &resp)

			portAttr, ok := resp.Schema.Attributes["port"]
			if !ok {
				t.Fatalf("%s schema is missing 'port' attribute", name)
			}
			int64Attr, ok := portAttr.(schema.Int64Attribute)
			if !ok {
				t.Fatalf("%s 'port' attribute is not schema.Int64Attribute", name)
			}
			if len(int64Attr.Validators) == 0 {
				t.Fatalf("%s 'port' attribute has no validators", name)
			}
		})
	}
}

func TestDefaultsAndPlanModifiersOnHardenedResources(t *testing.T) {
	t.Run("plex_settings use_ssl default", func(t *testing.T) {
		var resp resource.SchemaResponse
		NewPlexSettingsResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
		attr := resp.Schema.Attributes["use_ssl"].(schema.BoolAttribute)
		if attr.Default == nil {
			t.Fatal("expected plex_settings.use_ssl to have a default value")
		}
	})

	t.Run("api_key plan modifiers", func(t *testing.T) {
		var resp resource.SchemaResponse
		NewAPIKeyResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
		attr := resp.Schema.Attributes["api_key"].(schema.StringAttribute)
		if len(attr.PlanModifiers) == 0 {
			t.Fatal("expected api_key.api_key to have plan modifiers")
		}
	})

	t.Run("discover_slider plan modifiers", func(t *testing.T) {
		var resp resource.SchemaResponse
		NewDiscoverSliderResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
		idAttr := resp.Schema.Attributes["id"].(schema.StringAttribute)
		if len(idAttr.PlanModifiers) == 0 {
			t.Fatal("expected discover_slider.id to have plan modifiers")
		}
	})

	t.Run("user permissions validator and defaults", func(t *testing.T) {
		var resp resource.SchemaResponse
		NewUserResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
		permAttr := resp.Schema.Attributes["permissions"].(schema.Int64Attribute)
		if permAttr.Default == nil {
			t.Fatal("expected user.permissions to have default")
		}
		if len(permAttr.Validators) == 0 {
			t.Fatal("expected user.permissions to have validator")
		}
		syncMovies := resp.Schema.Attributes["watchlist_sync_movies"].(schema.BoolAttribute)
		if syncMovies.Default == nil {
			t.Fatal("expected user.watchlist_sync_movies to have default")
		}
		syncTv := resp.Schema.Attributes["watchlist_sync_tv"].(schema.BoolAttribute)
		if syncTv.Default == nil {
			t.Fatal("expected user.watchlist_sync_tv to have default")
		}
	})
}
