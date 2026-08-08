package provider

import "testing"

func TestBlocklistResourceSchema(t *testing.T) {
	t.Parallel()
	assertResourceSchemaContract(t, NewBlocklistResource(), map[string]resourceAttributeContract{
		"id":               computedString,
		"tmdb_id":          requiredInt64,
		"media_type":       requiredString,
		"title":            optionalComputedString,
		"user_id":          requiredInt64,
		"blocklisted_tags": optionalComputedString,
		"created_at":       computedString,
	}, nil)
}
