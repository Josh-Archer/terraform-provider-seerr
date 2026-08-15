package provider

import (
	"testing"
)

func TestOverrideRuleResourceSchema(t *testing.T) {
	t.Parallel()
	assertResourceSchemaContract(t, NewOverrideRuleResource(), map[string]resourceAttributeContract{
		"id":                computedString,
		"users":             optionalComputedString,
		"genre":             optionalComputedString,
		"genres":            optionalComputedList,
		"language":          optionalComputedString,
		"languages":         optionalComputedList,
		"original_language": optionalComputedString,
		"keywords":          optionalComputedString,
		"profile_id":        optionalComputedInt64,
		"root_folder":       optionalComputedString,
		"tags":              optionalComputedString,
		"tag_ids":           optionalComputedList,
		"roles":             optionalComputedString,
		"user_roles":        optionalComputedList,
		"radarr_service_id": optionalComputedInt64,
		"sonarr_service_id": optionalComputedInt64,
		"created_at":        computedString,
		"updated_at":        computedString,
	}, nil)
}

func TestOverrideRuleApplyMapExpandedFields(t *testing.T) {
	var data OverrideRuleModel
	err := applyOverrideRuleBody(&data, []byte(`{
		"id": 10,
		"genre": "28,12",
		"tags": "1,5",
		"userRoles": "admin,user",
		"language": "en,ja",
		"originalLanguage": "ja",
		"createdAt": "2026-01-01T00:00:00.000Z",
		"updatedAt": "2026-01-02T00:00:00.000Z"
	}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := data.ID.ValueString(); got != "10" {
		t.Fatalf("expected id 10, got %q", got)
	}
	if got := data.OriginalLanguage.ValueString(); got != "ja" {
		t.Fatalf("expected original_language ja, got %q", got)
	}
	if got := data.Roles.ValueString(); got != "admin,user" {
		t.Fatalf("expected roles admin,user, got %q", got)
	}
	if data.Genres.IsNull() || len(data.Genres.Elements()) != 2 {
		t.Fatalf("expected 2 elements in genres, got %v", data.Genres)
	}
	if data.TagIDs.IsNull() || len(data.TagIDs.Elements()) != 2 {
		t.Fatalf("expected 2 elements in tag_ids, got %v", data.TagIDs)
	}
	if data.UserRoles.IsNull() || len(data.UserRoles.Elements()) != 2 {
		t.Fatalf("expected 2 elements in user_roles, got %v", data.UserRoles)
	}
	if data.Languages.IsNull() || len(data.Languages.Elements()) != 2 {
		t.Fatalf("expected 2 elements in languages, got %v", data.Languages)
	}
}
