package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

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
		t.Errorf("APIKey should not be overwritten by helper, got %q", data.APIKey.ValueString())
	}
	if data.URL.ValueString() != "http://radarr:7878" {
		t.Errorf("URL should not be overwritten by helper, got %q", data.URL.ValueString())
	}
	if data.ExtraPayloadJSON.ValueString() != `{"custom":"value"}` {
		t.Errorf("ExtraPayloadJSON should not be overwritten by helper, got %q", data.ExtraPayloadJSON.ValueString())
	}
}

// TestReadRadarrStateFromJSON_PopulatesAPIKeyWhenUnset verifies that the helper
// can populate api_key for the data source path when the field is not already
// present in state.
func TestReadRadarrStateFromJSON_PopulatesAPIKeyWhenUnset(t *testing.T) {
	raw := `{"id": 1, "name": "Radarr", "apiKey": "server-key"}`
	data := &RadarrServerModel{}
	if err := readRadarrStateFromJSON(context.Background(), []byte(raw), data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.APIKey.ValueString() != "server-key" {
		t.Errorf("APIKey: got %q, want %q", data.APIKey.ValueString(), "server-key")
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

// TestReadRadarrStateFromJSON_NullValuesPreserved verifies explicit API nulls
// remain null in model state.
func TestReadRadarrStateFromJSON_NullValuesPreserved(t *testing.T) {
	raw := `{
		"id":                  1,
		"name":                null,
		"hostname":            null,
		"port":                null,
		"useSsl":              null,
		"baseUrl":             null,
		"is4k":                null,
		"minimumAvailability": null,
		"isDefault":           null,
		"enableScan":          null,
		"syncEnabled":         null,
		"preventSearch":       null,
		"tagRequests":         null,
		"tags":                null
	}`

	tagsVal, _ := types.ListValueFrom(context.Background(), types.Int64Type, []int64{1})
	data := &RadarrServerModel{
		Name:                types.StringValue("Prior"),
		Hostname:            types.StringValue("prior.local"),
		Port:                types.Int64Value(7878),
		UseSSL:              types.BoolValue(true),
		BaseURL:             types.StringValue("/prior"),
		Is4K:                types.BoolValue(true),
		MinimumAvailability: types.StringValue("released"),
		IsDefault:           types.BoolValue(true),
		EnableScan:          types.BoolValue(true),
		SyncEnabled:         types.BoolValue(true),
		PreventSearch:       types.BoolValue(true),
		TagRequestsWithUser: types.BoolValue(true),
		Tags:                tagsVal,
	}

	if err := readRadarrStateFromJSON(context.Background(), []byte(raw), data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !data.Name.IsNull() {
		t.Errorf("Name should be null, got %v", data.Name)
	}
	if !data.Hostname.IsNull() {
		t.Errorf("Hostname should be null, got %v", data.Hostname)
	}
	if !data.Port.IsNull() {
		t.Errorf("Port should be null, got %v", data.Port)
	}
	if !data.UseSSL.IsNull() {
		t.Errorf("UseSSL should be null, got %v", data.UseSSL)
	}
	if !data.BaseURL.IsNull() {
		t.Errorf("BaseURL should be null, got %v", data.BaseURL)
	}
	if !data.Is4K.IsNull() {
		t.Errorf("Is4K should be null, got %v", data.Is4K)
	}
	if !data.MinimumAvailability.IsNull() {
		t.Errorf("MinimumAvailability should be null, got %v", data.MinimumAvailability)
	}
	if !data.IsDefault.IsNull() {
		t.Errorf("IsDefault should be null, got %v", data.IsDefault)
	}
	if !data.EnableScan.IsNull() {
		t.Errorf("EnableScan should be null, got %v", data.EnableScan)
	}
	if !data.SyncEnabled.IsNull() {
		t.Errorf("SyncEnabled should be null, got %v", data.SyncEnabled)
	}
	if !data.PreventSearch.IsNull() {
		t.Errorf("PreventSearch should be null, got %v", data.PreventSearch)
	}
	if !data.TagRequestsWithUser.IsNull() {
		t.Errorf("TagRequestsWithUser should be null, got %v", data.TagRequestsWithUser)
	}
	if !data.Tags.IsNull() {
		t.Errorf("Tags should be null, got %v", data.Tags)
	}
}

// TestRadarrServerPayload_SeerrProxyTest verifies that RadarrServerResource.payload
// validates connectivity and resolves quality profile name via Seerr's proxy test endpoint
// (/api/v1/settings/radarr/test), supporting internal container hostnames without
// client-side network calls (fixes issue #231).
func TestRadarrServerPayload_SeerrProxyTest(t *testing.T) {
	var testedHostname string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/settings/radarr/test" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		var reqBody map[string]any
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("decode test request: %v", err)
		}
		if h, ok := reqBody["hostname"].(string); ok {
			testedHostname = h
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"profiles": [
				{"id": 1, "name": "Any"},
				{"id": 3, "name": "Ultra-HD"}
			],
			"rootFolders": [
				{"id": 1, "path": "/data/media/movies"}
			]
		}`))
	}))
	defer srv.Close()

	serverURL, _ := url.Parse(srv.URL)
	client := NewClient(serverURL, "apiKey", "test-agent", false, 5*time.Second, 0, 0)

	res := &RadarrServerResource{
		client: client,
	}

	tagsVal, _ := types.ListValueFrom(context.Background(), types.Int64Type, []int64{})

	// Hostname "radarr" represents an internal Docker bridge DNS name
	model := RadarrServerModel{
		Name:               types.StringValue("Radarr"),
		Hostname:           types.StringValue("radarr"),
		Port:               types.Int64Value(7878),
		APIKey:             types.StringValue("secret-radarr-key"),
		QualityProfileID:   types.Int64Value(3),
		QualityProfileName: types.StringNull(),
		ActiveDirectory:    types.StringValue("/data/media/movies"),
		Tags:               tagsVal,
	}

	updatedModel, payloadStr, err := res.payload(context.Background(), model)
	if err != nil {
		t.Fatalf("payload() failed: %v", err)
	}

	if testedHostname != "radarr" {
		t.Errorf("expected tested hostname 'radarr', got %q", testedHostname)
	}
	if updatedModel.QualityProfileName.ValueString() != "Ultra-HD" {
		t.Errorf("expected QualityProfileName 'Ultra-HD', got %q", updatedModel.QualityProfileName.ValueString())
	}
	if payloadStr == "" {
		t.Error("expected non-empty payload string")
	}
}

// TestRadarrServerPayload_ExplicitQualityProfileNameNoNetworkCall verifies that
// when quality_profile_name is provided, payload() succeeds without any network calls.
func TestRadarrServerPayload_ExplicitQualityProfileNameNoNetworkCall(t *testing.T) {
	res := &RadarrServerResource{}

	tagsVal, _ := types.ListValueFrom(context.Background(), types.Int64Type, []int64{})

	model := RadarrServerModel{
		Name:               types.StringValue("Radarr"),
		Hostname:           types.StringValue("radarr"),
		Port:               types.Int64Value(7878),
		APIKey:             types.StringValue("secret-radarr-key"),
		QualityProfileID:   types.Int64Value(1),
		QualityProfileName: types.StringValue("Ultra-HD"),
		ActiveDirectory:    types.StringValue("/data/media/movies"),
		Tags:               tagsVal,
	}

	updatedModel, payloadStr, err := res.payload(context.Background(), model)
	if err != nil {
		t.Fatalf("payload() failed: %v", err)
	}
	if updatedModel.QualityProfileName.ValueString() != "Ultra-HD" {
		t.Errorf("expected 'Ultra-HD', got %q", updatedModel.QualityProfileName.ValueString())
	}
	if payloadStr == "" {
		t.Error("expected non-empty payload")
	}
}

// TestRadarrServerPayload_UnresolvableQualityProfileError verifies that when quality_profile_name
// cannot be resolved, an error is returned prompting for explicit configuration.
func TestRadarrServerPayload_UnresolvableQualityProfileError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message": "Host unreachable"}`))
	}))
	defer srv.Close()

	serverURL, _ := url.Parse(srv.URL)
	client := NewClient(serverURL, "apiKey", "test-agent", false, 5*time.Second, 0, 0)

	res := &RadarrServerResource{
		client: client,
	}

	tagsVal, _ := types.ListValueFrom(context.Background(), types.Int64Type, []int64{})

	model := RadarrServerModel{
		Name:             types.StringValue("Radarr"),
		Hostname:         types.StringValue("radarr-unreachable"),
		Port:             types.Int64Value(7878),
		APIKey:           types.StringValue("bad-key"),
		QualityProfileID: types.Int64Value(99),
		ActiveDirectory:  types.StringValue("/data/media/movies"),
		Tags:             tagsVal,
	}

	_, _, err := res.payload(context.Background(), model)
	if err == nil {
		t.Fatal("expected error from payload(), got nil")
	}
	if expected := "could not resolve quality_profile_name for profile id 99"; !strings.Contains(err.Error(), expected) {
		t.Errorf("expected error containing %q, got %q", expected, err.Error())
	}
}
