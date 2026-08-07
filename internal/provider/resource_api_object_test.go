package provider

import "testing"

func TestAPIObjectResourceSchema(t *testing.T) {
	t.Parallel()
	assertResourceSchemaContract(t, NewAPIObjectResource(), map[string]resourceAttributeContract{
		"id":                 computedString,
		"path":               requiredString,
		"headers":            optionalMap,
		"request_body_json":  optionalString,
		"delete_body_json":   optionalString,
		"read_method":        optionalComputedString,
		"create_method":      optionalComputedString,
		"update_method":      optionalComputedString,
		"delete_method":      optionalComputedString,
		"skip_delete":        optionalComputedBool,
		"suppress_not_found": optionalComputedBool,
		"response_body_json": computedString,
		"status_code":        computedInt64,
	}, nil)
}
