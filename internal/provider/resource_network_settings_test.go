package provider

import "testing"

func TestNetworkSettingsResourceSchema(t *testing.T) {
	t.Parallel()
	assertResourceSchemaContract(t, NewNetworkSettingsResource(), map[string]resourceAttributeContract{
		"id":                     computedString,
		"csrf_protection":        optionalComputedBool,
		"force_ipv4_first":       optionalComputedBool,
		"trust_proxy":            optionalComputedBool,
		"api_request_timeout_ms": optionalComputedInt64,
	}, map[string][]string{
		"proxy":     {"enabled", "hostname", "port", "use_ssl", "user", "password", "bypass_filter", "bypass_local_addresses"},
		"dns_cache": {"enabled", "force_min_ttl", "force_max_ttl"},
	})
}
