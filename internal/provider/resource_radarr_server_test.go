package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestReadRadarrStateFromJSON_AllFields verifies that every API-mapped field is
// populated correctly from a full JSON payload.
func TestReadRadarrStateFromJSON_AllFields(t *testing.T) {
	raw := `{
		"id":                  1,
		"name":                "Radarr",
		"hostname":            "radarr.local",
		"port":                7878,
		"useSsl":              true,
		"baseUrl":             "/radarr",
		"activeProfileId":     3,
		"activeProfileName":   "HD-1080p",
		"activeDirectory":     "/movies",
		"is4k":                false,
		"minimumAvailability": "released",
		"tags":                [10, 20],
		"isDefault":           true,
		"enableScan":          false,
		"syncEnabled":         true,
		"preventSearch":       true,
		"tagRequests":         false
	}`

	data := &RadarrServerModel{}
	if err := readRadarrStateFromJSON(context.Background(), []byte(raw), data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	check := func(field, got, want string) {
		t.Helper()
		if got != want {
			t.Errorf("%s: got %q, want %q", field, got, want)
		}
	}
	checkBool := func(field string, got types.Bool, want bool) {
		t.Helper()
		if got.ValueBool() != want {
			t.Errorf("%s: got %v, want %v", field, got.ValueBool(), want)
		}
	}
	checkInt := func(field string, got types.Int64, want int64) {
		t.Helper()
		if got.ValueInt64() != want {
			t.Errorf("%s: got %d, want %d", field, got.ValueInt64(), want)
		}
	}

	check("Name", data.Name.ValueString(), "Radarr")
	check("Hostname", data.Hostname.ValueString(), "radarr.local")
	checkInt("Port", data.Port, 7878)
	checkBool("UseSSL", data.UseSSL, true)
	check("BaseURL", data.BaseURL.ValueString(), "/radarr")
	checkInt("QualityProfileID", data.QualityProfileID, 3)
	check("QualityProfileName", data.QualityProfileName.ValueString(), "HD-1080p")
	check("ActiveDirectory", data.ActiveDirectory.ValueString(), "/movies")
	checkBool("Is4K", data.Is4K, false)
	check("MinimumAvailability", data.MinimumAvailability.ValueString(), "released")
	checkBool("IsDefault", data.IsDefault, true)
	checkBool("EnableScan", data.EnableScan, false)
	checkBool("SyncEnabled", data.SyncEnabled, true)
	checkBool("PreventSearch", data.PreventSearch, true)
	checkBool("TagRequestsWithUser", data.TagRequestsWithUser, false)

	// tags
	if data.Tags.IsNull() || data.Tags.IsUnknown() {
		t.Fatal("Tags should not be null/unknown")
	}
	tagIDs, err := listInt64(context.Background(), data.Tags)
	if err != nil {
		t.Fatalf("listInt64: %v", err)
	}
	if len(tagIDs) != 2 || tagIDs[0] != 10 || tagIDs[1] != 20 {
		t.Errorf("Tags: got %v, want [10 20]", tagIDs)
	}
}

// TestReadRadarrStateFromJSON_EmptyTags ensures an empty tags array results in
// an empty (non-null) list in state.
func TestReadRadarrStateFromJSON_EmptyTags(t *testing.T) {
	raw := `{
		"id":   1,
		"name": "Radarr",
		"tags": []
	}`
	data := &RadarrServerModel{}
	if err := readRadarrStateFromJSON(context.Background(), []byte(raw), data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.Tags.IsNull() {
		t.Error("Tags should not be null for an explicit empty array")
	}
	tagIDs, err := listInt64(context.Background(), data.Tags)
	if err != nil {
		t.Fatalf("listInt64: %v", err)
	}
	if len(tagIDs) != 0 {
		t.Errorf("expected empty tags slice, got %v", tagIDs)
	}
}

// TestReadRadarrStateFromJSON_InvalidJSON verifies an error is returned for
// malformed JSON.
func TestReadRadarrStateFromJSON_InvalidJSON(t *testing.T) {
	data := &RadarrServerModel{}
	if err := readRadarrStateFromJSON(context.Background(), []byte(`{not valid json`), data); err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

// TestReadRadarrStateFromJSON_UserSuppliedFieldsUntouched verifies that
// api_key, url, and extra_payload_json are not overwritten.
func TestReadRadarrStateFromJSON_UserSuppliedFieldsUntouched(t *testing.T) {
	raw := `{"id": 1, "name": "Radarr"}`
	data := &RadarrServerModel{
		APIKey:           types.StringValue("my-secret-key"),
		URL:              types.StringValue("http://radarr:7878"),
		ExtraPayloadJSON: types.StringValue(`{"custom":"value"}`),
	}
	if err := readRadarrStateFromJSON(context.Background(), []byte(raw), data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.APIKey.ValueString() != "my-secret-key" {
		t.Errorf("APIKey should not be overwritten, got %q", data.APIKey.ValueString())
	}
	if data.URL.ValueString() != "http://radarr:7878" {
		t.Errorf("URL should not be overwritten, got %q", data.URL.ValueString())
	}
	if data.ExtraPayloadJSON.ValueString() != `{"custom":"value"}` {
		t.Errorf("ExtraPayloadJSON should not be overwritten, got %q", data.ExtraPayloadJSON.ValueString())
	}
}

// TestParseURLIntoModel_HTTPS verifies that an https URL is decomposed into
// hostname, port, use_ssl, and base_url correctly.
func TestParseURLIntoModel_HTTPS(t *testing.T) {
	data := &RadarrServerModel{
		URL: types.StringValue("https://radarr.example.com:9090/radarr"),
	}
	parseURLIntoModel(data)

	if data.Hostname.ValueString() != "radarr.example.com" {
		t.Errorf("Hostname: got %q", data.Hostname.ValueString())
	}
	if data.Port.ValueInt64() != 9090 {
		t.Errorf("Port: got %d", data.Port.ValueInt64())
	}
	if !data.UseSSL.ValueBool() {
		t.Error("UseSSL should be true for https URL")
	}
	if data.BaseURL.ValueString() != "/radarr" {
		t.Errorf("BaseURL: got %q", data.BaseURL.ValueString())
	}
}

// TestParseURLIntoModel_HTTP verifies http URL decomposition.
func TestParseURLIntoModel_HTTP(t *testing.T) {
	data := &RadarrServerModel{
		URL: types.StringValue("http://radarr.local:7878"),
	}
	parseURLIntoModel(data)

	if data.Hostname.ValueString() != "radarr.local" {
		t.Errorf("Hostname: got %q", data.Hostname.ValueString())
	}
	if data.Port.ValueInt64() != 7878 {
		t.Errorf("Port: got %d", data.Port.ValueInt64())
	}
	if data.UseSSL.ValueBool() {
		t.Error("UseSSL should be false for http URL")
	}
}

// TestParseURLIntoModel_NoURL verifies the default port when URL is omitted.
func TestParseURLIntoModel_NoURL(t *testing.T) {
	data := &RadarrServerModel{
		URL:  types.StringNull(),
		Port: types.Int64Null(),
	}
	parseURLIntoModel(data)
	if data.Port.ValueInt64() != 7878 {
		t.Errorf("Port default: got %d, want 7878", data.Port.ValueInt64())
	}
}
