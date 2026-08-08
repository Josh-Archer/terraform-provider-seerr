package provider

import "testing"

func TestEmbySettingsResourceSchema(t *testing.T) {
	t.Parallel()
	assertResourceSchemaContract(t, NewEmbySettingsResource(), map[string]resourceAttributeContract{
		"id":                       computedString,
		"name":                     optionalComputedString,
		"ip":                       requiredString,
		"port":                     requiredInt64,
		"use_ssl":                  optionalComputedBool,
		"url_base":                 optionalComputedString,
		"external_hostname":        optionalComputedString,
		"emby_forgot_password_url": optionalComputedString,
		"api_key":                  requiredSensitiveString,
		"server_id":                computedString,
		"response_json":            computedString,
		"status_code":              computedInt64,
	}, nil)
}
