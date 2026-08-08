package provider

import "testing"

func TestUserPermissionsResourceSchema(t *testing.T) {
	t.Parallel()
	assertResourceSchemaContract(t, NewUserPermissionsResource(), map[string]resourceAttributeContract{
		"id":            computedString,
		"user_id":       requiredInt64,
		"permissions":   requiredInt64,
		"response_json": computedString,
		"status_code":   computedInt64,
	}, nil)
}
