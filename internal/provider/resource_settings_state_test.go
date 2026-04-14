package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPlexSettingsApplyDecodedSettingsClearsMissingOptionalValues(t *testing.T) {
	r := &PlexSettingsResource{}
	data := PlexSettingsModel{
		Name:   types.StringValue("Plex"),
		UseSSL: types.BoolValue(true),
	}

	r.applyDecodedPlexSettings(&data, map[string]any{
		"ip":   "127.0.0.1",
		"port": float64(32400),
	})

	if !data.Name.IsNull() {
		t.Fatalf("expected name to become null when omitted from response")
	}
	if !data.UseSSL.IsNull() {
		t.Fatalf("expected use_ssl to become null when omitted from response")
	}
	if got := data.IP.ValueString(); got != "127.0.0.1" {
		t.Fatalf("expected ip 127.0.0.1, got %q", got)
	}
	if got := data.Port.ValueInt64(); got != 32400 {
		t.Fatalf("expected port 32400, got %d", got)
	}
}

func TestJellyfinSettingsApplyDecodedSettingsClearsMissingOptionalValuesAndPreservesAPIKey(t *testing.T) {
	r := &JellyfinSettingsResource{}
	data := JellyfinSettingsModel{
		Name:                      types.StringValue("Jellyfin"),
		UseSSL:                    types.BoolValue(true),
		URLBase:                   types.StringValue("/jf"),
		ExternalHostname:          types.StringValue("media.example"),
		JellyfinForgotPasswordURL: types.StringValue("https://media.example/forgot"),
		ServerID:                  types.StringValue("server-1"),
		APIKey:                    types.StringValue("secret"),
	}

	r.applyDecodedJellyfinSettings(&data, map[string]any{
		"ip":   "127.0.0.1",
		"port": float64(8096),
	})

	if !data.Name.IsNull() || !data.UseSSL.IsNull() || !data.URLBase.IsNull() || !data.ExternalHostname.IsNull() || !data.JellyfinForgotPasswordURL.IsNull() || !data.ServerID.IsNull() {
		t.Fatalf("expected omitted optional/computed jellyfin fields to become null")
	}
	if got := data.APIKey.ValueString(); got != "secret" {
		t.Fatalf("expected api key to be preserved, got %q", got)
	}
}

func TestEmbySettingsApplyDecodedSettingsClearsMissingOptionalValuesAndPreservesAPIKey(t *testing.T) {
	r := &EmbySettingsResource{}
	data := EmbySettingsModel{
		Name:                  types.StringValue("Emby"),
		UseSSL:                types.BoolValue(true),
		URLBase:               types.StringValue("/emby"),
		ExternalHostname:      types.StringValue("media.example"),
		EmbyForgotPasswordURL: types.StringValue("https://media.example/forgot"),
		ServerID:              types.StringValue("server-1"),
		APIKey:                types.StringValue("secret"),
	}

	r.applyDecodedEmbySettings(&data, map[string]any{
		"ip":   "127.0.0.1",
		"port": float64(8097),
	})

	if !data.Name.IsNull() || !data.UseSSL.IsNull() || !data.URLBase.IsNull() || !data.ExternalHostname.IsNull() || !data.EmbyForgotPasswordURL.IsNull() || !data.ServerID.IsNull() {
		t.Fatalf("expected omitted optional/computed emby fields to become null")
	}
	if got := data.APIKey.ValueString(); got != "secret" {
		t.Fatalf("expected api key to be preserved, got %q", got)
	}
}

func TestTautulliSettingsApplyDecodedSettingsClearsMissingOptionalValuesAndPreservesAPIKey(t *testing.T) {
	r := &TautulliSettingsResource{}
	data := TautulliSettingsModel{
		Hostname:    types.StringValue("tautulli"),
		Port:        types.Int64Value(8181),
		UseSSL:      types.BoolValue(true),
		URLBase:     types.StringValue("/tautulli"),
		ExternalURL: types.StringValue("https://media.example/tautulli"),
		APIKey:      types.StringValue("secret"),
	}

	r.applyDecodedTautulliSettings(&data, map[string]any{})

	if !data.Hostname.IsNull() || !data.Port.IsNull() || !data.UseSSL.IsNull() || !data.URLBase.IsNull() || !data.ExternalURL.IsNull() {
		t.Fatalf("expected omitted optional/computed tautulli fields to become null")
	}
	if got := data.APIKey.ValueString(); got != "secret" {
		t.Fatalf("expected api key to be preserved, got %q", got)
	}
}

func TestUserSettingsPermissionsReadPermissionsPreservesMissingValues(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/user/7/settings/permissions" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"autoApproveMovies":true}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	r := &UserSettingsPermissionsResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}
	data := UserSettingsPermissionsModel{
		UserID:              types.Int64Value(7),
		AutoApproveMovies:   types.BoolValue(false),
		AutoApproveTV:       types.BoolValue(true),
		AutoApprove4KMovies: types.BoolValue(true),
		AutoApprove4KTV:     types.BoolValue(true),
	}

	if err := r.readPermissions(context.Background(), &data); err != nil {
		t.Fatal(err)
	}

	if got := data.AutoApproveMovies.ValueBool(); !got {
		t.Fatalf("expected auto_approve_movies true, got %v", got)
	}
	if got := data.AutoApproveTV.ValueBool(); !got {
		t.Fatalf("expected auto_approve_tv to retain prior value, got %v", got)
	}
	if got := data.AutoApprove4KMovies.ValueBool(); !got {
		t.Fatalf("expected auto_approve_4k_movies to retain prior value, got %v", got)
	}
	if got := data.AutoApprove4KTV.ValueBool(); !got {
		t.Fatalf("expected auto_approve_4k_tv to retain prior value, got %v", got)
	}
}
