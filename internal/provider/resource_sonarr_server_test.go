package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestReadSonarrStateFromJSON_AllFields verifies that every API-mapped field
// (including Sonarr-specific ones) is populated correctly.
func TestReadSonarrStateFromJSON_AllFields(t *testing.T) {
	raw := `{
		"id":                   1,
		"name":                 "Sonarr",
		"hostname":             "sonarr.local",
		"port":                 8989,
		"useSsl":               false,
		"baseUrl":              "",
		"activeProfileId":      2,
		"activeProfileName":    "HD-1080p",
		"activeDirectory":      "/tv",
		"activeAnimeDirectory": "/anime",
		"is4k":                 false,
		"isDefault":            true,
		"enableScan":           true,
		"enableSeasonFolders":  true,
		"syncEnabled":          true,
		"preventSearch":        false,
		"tagRequests":          true,
		"tags":                 [1, 2, 3],
		"animeTags":            [4, 5]
	}`

	data := &SonarrServerModel{}
	if err := readSonarrStateFromJSON(context.Background(), []byte(raw), data); err != nil {
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

	check("Name", data.Name.ValueString(), "Sonarr")
	check("Hostname", data.Hostname.ValueString(), "sonarr.local")
	checkInt("Port", data.Port, 8989)
	checkBool("UseSSL", data.UseSSL, false)
	check("BaseURL", data.BaseURL.ValueString(), "")
	checkInt("QualityProfileID", data.QualityProfileID, 2)
	check("QualityProfileName", data.QualityProfileName.ValueString(), "HD-1080p")
	check("ActiveDirectory", data.ActiveDirectory.ValueString(), "/tv")
	check("ActiveAnimeDirectory", data.ActiveAnimeDirectory.ValueString(), "/anime")
	checkBool("Is4K", data.Is4K, false)
	checkBool("IsDefault", data.IsDefault, true)
	checkBool("EnableScan", data.EnableScan, true)
	checkBool("EnableSeasonFolders", data.EnableSeasonFolders, true)
	checkBool("SyncEnabled", data.SyncEnabled, true)
	checkBool("PreventSearch", data.PreventSearch, false)
	checkBool("TagRequestsWithUser", data.TagRequestsWithUser, true)

	// tags
	tagIDs, err := listInt64(context.Background(), data.Tags)
	if err != nil {
		t.Fatalf("listInt64 tags: %v", err)
	}
	if len(tagIDs) != 3 || tagIDs[0] != 1 || tagIDs[1] != 2 || tagIDs[2] != 3 {
		t.Errorf("Tags: got %v, want [1 2 3]", tagIDs)
	}

	// animeTags
	animeTagIDs, err := listInt64(context.Background(), data.AnimeTags)
	if err != nil {
		t.Fatalf("listInt64 animeTags: %v", err)
	}
	if len(animeTagIDs) != 2 || animeTagIDs[0] != 4 || animeTagIDs[1] != 5 {
		t.Errorf("AnimeTags: got %v, want [4 5]", animeTagIDs)
	}
}

// TestReadSonarrStateFromJSON_EmptyAnimeTags ensures an empty animeTags array
// results in an empty (non-null) list in state.
func TestReadSonarrStateFromJSON_EmptyAnimeTags(t *testing.T) {
	raw := `{
		"id":        1,
		"name":      "Sonarr",
		"tags":      [],
		"animeTags": []
	}`
	data := &SonarrServerModel{}
	if err := readSonarrStateFromJSON(context.Background(), []byte(raw), data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.Tags.IsNull() {
		t.Error("Tags should not be null for an explicit empty array")
	}
	if data.AnimeTags.IsNull() {
		t.Error("AnimeTags should not be null for an explicit empty array")
	}

	tagIDs, _ := listInt64(context.Background(), data.Tags)
	if len(tagIDs) != 0 {
		t.Errorf("expected empty tags, got %v", tagIDs)
	}
	animeTagIDs, _ := listInt64(context.Background(), data.AnimeTags)
	if len(animeTagIDs) != 0 {
		t.Errorf("expected empty animeTags, got %v", animeTagIDs)
	}
}

// TestReadSonarrStateFromJSON_InvalidJSON verifies an error is returned for
// malformed JSON.
func TestReadSonarrStateFromJSON_InvalidJSON(t *testing.T) {
	data := &SonarrServerModel{}
	if err := readSonarrStateFromJSON(context.Background(), []byte(`{not valid`), data); err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

// TestReadSonarrStateFromJSON_UserSuppliedFieldsUntouched verifies that
// api_key, url, and extra_payload_json are not overwritten by the Read helper.
func TestReadSonarrStateFromJSON_UserSuppliedFieldsUntouched(t *testing.T) {
	raw := `{"id": 1, "name": "Sonarr"}`
	data := &SonarrServerModel{
		APIKey:           types.StringValue("my-secret-key"),
		URL:              types.StringValue("http://sonarr:8989"),
		ExtraPayloadJSON: types.StringValue(`{"custom":"value"}`),
	}
	if err := readSonarrStateFromJSON(context.Background(), []byte(raw), data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.APIKey.ValueString() != "my-secret-key" {
		t.Errorf("APIKey should not be overwritten, got %q", data.APIKey.ValueString())
	}
	if data.URL.ValueString() != "http://sonarr:8989" {
		t.Errorf("URL should not be overwritten, got %q", data.URL.ValueString())
	}
	if data.ExtraPayloadJSON.ValueString() != `{"custom":"value"}` {
		t.Errorf("ExtraPayloadJSON should not be overwritten, got %q", data.ExtraPayloadJSON.ValueString())
	}
}

// TestParseSonarrURLIntoModel_HTTPS verifies that an https URL is decomposed
// into hostname, port, use_ssl, and base_url correctly.
func TestParseSonarrURLIntoModel_HTTPS(t *testing.T) {
	data := &SonarrServerModel{
		URL: types.StringValue("https://sonarr.example.com:9191/sonarr"),
	}
	parseSonarrURLIntoModel(data)

	if data.Hostname.ValueString() != "sonarr.example.com" {
		t.Errorf("Hostname: got %q", data.Hostname.ValueString())
	}
	if data.Port.ValueInt64() != 9191 {
		t.Errorf("Port: got %d", data.Port.ValueInt64())
	}
	if !data.UseSSL.ValueBool() {
		t.Error("UseSSL should be true for https URL")
	}
	if data.BaseURL.ValueString() != "/sonarr" {
		t.Errorf("BaseURL: got %q", data.BaseURL.ValueString())
	}
}

// TestParseSonarrURLIntoModel_NoURL verifies the default port when URL is omitted.
func TestParseSonarrURLIntoModel_NoURL(t *testing.T) {
	data := &SonarrServerModel{
		URL:  types.StringNull(),
		Port: types.Int64Null(),
	}
	parseSonarrURLIntoModel(data)
	if data.Port.ValueInt64() != 8989 {
		t.Errorf("Port default: got %d, want 8989", data.Port.ValueInt64())
	}
}
