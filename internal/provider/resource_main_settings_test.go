package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMainSettingsApplyDecodedSettingsClearsMissingValues(t *testing.T) {
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

	if !data.AppTitle.IsNull() {
		t.Fatalf("expected app title to become null when omitted from response")
	}
	if !data.PartialRequests.IsNull() {
		t.Fatalf("expected partial requests to become null when omitted from response")
	}
	if !data.MovieRequestsEnabled.IsNull() {
		t.Fatalf("expected movie requests to become null when omitted from response")
	}
	if !data.SeriesRequestsEnabled.IsNull() {
		t.Fatalf("expected series requests to become null when omitted from response")
	}
	if got := data.ApplicationURL.ValueString(); got != "https://jellyseerr.example" {
		t.Fatalf("expected application url to be updated, got %q", got)
	}
}

func TestMainSettingsApplyDecodedSettingsReadsSeerrV3Keys(t *testing.T) {
	r := &MainSettingsResource{}
	var data MainSettingsModel

	r.applyDecodedSettings(&data, map[string]any{
		"applicationTitle":       "Seerr",
		"applicationUrl":         "https://seerr.example",
		"cacheImages":            true,
		"locale":                 "fr",
		"discoverRegion":         "",
		"streamingRegion":        "FR",
		"originalLanguage":       "",
		"hideAvailable":          false,
		"partialRequestsEnabled": true,
		"localLogin":             true,
		"mediaServerLogin":       true,
		"newPlexLogin":           true,
	})

	if got := data.AppTitle.ValueString(); got != "Seerr" {
		t.Fatalf("expected app title Seerr, got %q", got)
	}
	if got := data.CacheImages.ValueBool(); !got {
		t.Fatalf("expected cache_images true, got %v", got)
	}
	if got := data.ImageProxy.ValueBool(); !got {
		t.Fatalf("expected deprecated image_proxy alias true, got %v", got)
	}
	if got := data.StreamingRegion.ValueString(); got != "FR" {
		t.Fatalf("expected streaming region FR, got %q", got)
	}
	if got := data.Region.ValueString(); got != "FR" {
		t.Fatalf("expected deprecated region alias FR, got %q", got)
	}
	if got := data.PartialRequests.ValueBool(); !got {
		t.Fatalf("expected partial requests true, got %v", got)
	}
	if got := data.MediaServerLogin.ValueBool(); !got {
		t.Fatalf("expected media_server_login true, got %v", got)
	}
	if got := data.PlexLogin.ValueBool(); !got {
		t.Fatalf("expected deprecated plex_login alias true, got %v", got)
	}
	if !data.TrustProxy.IsNull() {
		t.Fatalf("expected moved trust_proxy to remain null from v3 main settings")
	}
}

func TestMainSettingsBuildPayloadUsesSeerrV3Keys(t *testing.T) {
	r := &MainSettingsResource{}
	payload := r.buildPayload(&MainSettingsModel{
		AppTitle:              types.StringValue("Seerr"),
		ApplicationURL:        types.StringValue("https://seerr.example"),
		ImageProxy:            types.BoolValue(true),
		Locale:                types.StringValue("fr"),
		DiscoverRegion:        types.StringValue(""),
		Region:                types.StringValue("FR"),
		OriginalLanguage:      types.StringValue(""),
		HideAvailable:         types.BoolValue(false),
		PartialRequests:       types.BoolValue(true),
		LocalLogin:            types.BoolValue(true),
		PlexLogin:             types.BoolValue(true),
		NewPlexLogin:          types.BoolValue(true),
		TrustProxy:            types.BoolValue(true),
		CSRFProtection:        types.BoolValue(false),
		MovieRequestsEnabled:  types.BoolValue(true),
		SeriesRequestsEnabled: types.BoolValue(true),
	})

	if got := payload["applicationTitle"]; got != "Seerr" {
		t.Fatalf("expected applicationTitle, got %#v", got)
	}
	if _, ok := payload["appTitle"]; ok {
		t.Fatalf("did not expect legacy appTitle key in payload")
	}
	if got := payload["cacheImages"]; got != true {
		t.Fatalf("expected cacheImages true from image_proxy alias, got %#v", got)
	}
	if got := payload["streamingRegion"]; got != "FR" {
		t.Fatalf("expected streamingRegion FR from region alias, got %#v", got)
	}
	if got := payload["partialRequestsEnabled"]; got != true {
		t.Fatalf("expected partialRequestsEnabled true, got %#v", got)
	}
	if got := payload["mediaServerLogin"]; got != true {
		t.Fatalf("expected mediaServerLogin true from plex_login alias, got %#v", got)
	}
	for _, key := range []string{"trustProxy", "csrfProtection", "movieRequestsEnabled", "seriesRequestsEnabled"} {
		if _, ok := payload[key]; ok {
			t.Fatalf("did not expect removed or moved key %q in payload", key)
		}
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
