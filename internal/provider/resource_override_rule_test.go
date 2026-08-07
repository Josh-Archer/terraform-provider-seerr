package provider

import "testing"

func TestOverrideRuleResourceSchema(t *testing.T) {
	t.Parallel()
	assertResourceSchemaContract(t, NewOverrideRuleResource(), map[string]resourceAttributeContract{
		"id":                computedString,
		"users":             optionalComputedString,
		"genre":             optionalComputedString,
		"language":          optionalComputedString,
		"keywords":          optionalComputedString,
		"profile_id":        optionalComputedInt64,
		"root_folder":       optionalComputedString,
		"tags":              optionalComputedString,
		"radarr_service_id": optionalComputedInt64,
		"sonarr_service_id": optionalComputedInt64,
		"created_at":        computedString,
		"updated_at":        computedString,
	}, nil)
}
