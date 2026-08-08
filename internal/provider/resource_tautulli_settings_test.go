package provider

import "testing"

func TestTautulliSettingsResourceSchema(t *testing.T) {
	t.Parallel()
	assertResourceSchemaContract(t, NewTautulliSettingsResource(), map[string]resourceAttributeContract{
		"id":            computedString,
		"hostname":      optionalComputedString,
		"ip":            optionalComputedString,
		"port":          optionalComputedInt64,
		"use_ssl":       optionalComputedBool,
		"url_base":      optionalComputedString,
		"api_key":       optionalSensitiveString,
		"external_url":  optionalComputedString,
		"response_json": computedString,
		"status_code":   computedInt64,
	}, nil)
}
