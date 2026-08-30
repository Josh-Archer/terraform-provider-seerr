package provider

import "testing"

func TestJellyfinSettingsResourceSchema(t *testing.T) {
	t.Parallel()
	assertResourceSchemaContract(t, NewJellyfinSettingsResource(), map[string]resourceAttributeContract{
		"id":                           computedString,
		"name":                         computedString,
		"ip":                           requiredString,
		"port":                         requiredInt64,
		"use_ssl":                      optionalComputedBool,
		"url_base":                     optionalComputedString,
		"external_hostname":            optionalComputedString,
		"jellyfin_forgot_password_url": optionalComputedString,
		"api_key":                      requiredSensitiveString,
		"server_id":                    computedString,
		"response_json":                computedString,
		"status_code":                  computedInt64,
	}, nil)
}
