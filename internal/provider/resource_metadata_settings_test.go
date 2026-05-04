package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMetadataSettingsApplyUsesMetadataEndpoint(t *testing.T) {
	seenPut := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/settings/metadata" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPut:
			seenPut = true
			var payload map[string]string
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			if payload["tv"] != "tvdb" || payload["anime"] != "tmdb" {
				t.Fatalf("unexpected payload %#v", payload)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "tv": "tvdb", "anime": "tmdb"})
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{"tv": "tvdb", "anime": "tmdb"})
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	resource := &MetadataSettingsResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}
	data := MetadataSettingsModel{
		TV:    types.StringValue("tvdb"),
		Anime: types.StringValue("tmdb"),
	}
	if err := resource.applyMetadataSettings(context.Background(), &data); err != nil {
		t.Fatal(err)
	}
	if !seenPut {
		t.Fatal("expected PUT request")
	}
	if got := data.ID.ValueString(); got != "metadata" {
		t.Fatalf("expected metadata id, got %q", got)
	}
}
