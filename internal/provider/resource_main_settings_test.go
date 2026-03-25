package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMainSettingsApplyDecodedSettingsPreservesExistingValues(t *testing.T) {
	r := &MainSettingsResource{}
	data := MainSettingsModel{
		AppTitle:              types.StringValue("Jellyseerr"),
		PartialRequests:       types.BoolValue(true),
		MovieRequestsEnabled:  types.BoolValue(true),
		SeriesRequestsEnabled: types.BoolValue(true),
		TrustProxy:            types.BoolValue(false),
	}

	r.applyDecodedSettings(&data, map[string]any{
		"applicationUrl": "https://jellyseerr.example",
	})

	if got := data.AppTitle.ValueString(); got != "Jellyseerr" {
		t.Fatalf("expected app title to remain Jellyseerr, got %q", got)
	}
	if got := data.PartialRequests.ValueBool(); !got {
		t.Fatalf("expected partial requests to remain true")
	}
	if got := data.MovieRequestsEnabled.ValueBool(); !got {
		t.Fatalf("expected movie requests to remain true")
	}
	if got := data.SeriesRequestsEnabled.ValueBool(); !got {
		t.Fatalf("expected series requests to remain true")
	}
	if got := data.ApplicationURL.ValueString(); got != "https://jellyseerr.example" {
		t.Fatalf("expected application url to be updated, got %q", got)
	}
}

func TestMainSettingsRefreshStateReadsCanonicalValues(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/settings/main" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"appTitle":"Jellyseerr",
			"applicationUrl":"https://jellyseerr.example",
			"trustProxy":false,
			"csrfProtection":false,
			"imageProxy":true,
			"locale":"en",
			"region":"US",
			"originalLanguage":"en",
			"hideAvailable":true,
			"partialRequests":true,
			"localLogin":true,
			"newPlexLogin":true,
			"plexLogin":true,
			"movieRequestsEnabled":true,
			"seriesRequestsEnabled":true,
			"enableReportAnIssue":false,
			"movieRequestLimit":7,
			"seriesRequestLimit":11
		}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	r := &MainSettingsResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}
	data := MainSettingsModel{
		AppTitle:        types.StringValue("stale"),
		PartialRequests: types.BoolValue(false),
	}

	if err := r.refreshState(context.Background(), &data); err != nil {
		t.Fatal(err)
	}

	if got := data.AppTitle.ValueString(); got != "Jellyseerr" {
		t.Fatalf("expected app title Jellyseerr, got %q", got)
	}
	if got := data.PartialRequests.ValueBool(); !got {
		t.Fatalf("expected partial requests true, got %v", got)
	}
	if got := data.MovieRequestsEnabled.ValueBool(); !got {
		t.Fatalf("expected movie requests enabled true, got %v", got)
	}
	if got := data.SeriesRequestsEnabled.ValueBool(); !got {
		t.Fatalf("expected series requests enabled true, got %v", got)
	}
	if got := data.MovieRequestLimit.ValueInt64(); got != 7 {
		t.Fatalf("expected movie request limit 7, got %d", got)
	}
	if got := data.SeriesRequestLimit.ValueInt64(); got != 11 {
		t.Fatalf("expected series request limit 11, got %d", got)
	}
}
