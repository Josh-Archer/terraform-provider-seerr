package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildNetworkPayloadPreservesExistingNestedValues(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/settings/network" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"csrfProtection":false,
			"proxy":{"enabled":true,"hostname":"proxy.internal","port":8080,"password":"existing"},
			"dnsCache":{"enabled":true,"forceMinTtl":10,"forceMaxTtl":20}
		}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	resource := &NetworkSettingsResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}
	payload, err := resource.buildNetworkPayload(context.Background(), &NetworkSettingsModel{
		ForceIPv4First:      types.BoolValue(true),
		APIRequestTimeoutMS: types.Int64Value(15000),
		Proxy: &NetworkProxyModel{
			Hostname: types.StringValue("override.proxy"),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := payload["forceIpv4First"]; got != true {
		t.Fatalf("expected forceIpv4First override, got %#v", got)
	}
	proxy, ok := payload["proxy"].(map[string]any)
	if !ok {
		t.Fatalf("expected proxy map, got %#v", payload["proxy"])
	}
	if got := proxy["hostname"]; got != "override.proxy" {
		t.Fatalf("expected proxy hostname override, got %#v", got)
	}
	if got := proxy["password"]; got != "existing" {
		t.Fatalf("expected existing proxy password to be preserved, got %#v", got)
	}
}

func TestRefreshNetworkSettingsLeavesOmittedNestedBlocksUnmanaged(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/settings/network" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"csrfProtection":false,
			"forceIpv4First":false,
			"trustProxy":true,
			"apiRequestTimeout":15000,
			"proxy":{
				"enabled":false,
				"hostname":"",
				"port":8080,
				"useSsl":false,
				"user":"",
				"password":"",
				"bypassFilter":"",
				"bypassLocalAddresses":true
			},
			"dnsCache":{
				"enabled":false,
				"forceMinTtl":0,
				"forceMaxTtl":-1
			}
		}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	resource := &NetworkSettingsResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}
	data := &NetworkSettingsModel{
		TrustProxy:          types.BoolValue(true),
		APIRequestTimeoutMS: types.Int64Value(15000),
	}

	if err := resource.refreshNetworkSettings(context.Background(), data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.Proxy != nil {
		t.Fatalf("expected omitted proxy block to remain unmanaged, got %#v", data.Proxy)
	}
	if data.DNSCache != nil {
		t.Fatalf("expected omitted dns_cache block to remain unmanaged, got %#v", data.DNSCache)
	}
	if got := data.TrustProxy.ValueBool(); !got {
		t.Fatalf("expected trust_proxy true, got %v", got)
	}
	if got := data.APIRequestTimeoutMS.ValueInt64(); got != 15000 {
		t.Fatalf("expected timeout 15000, got %d", got)
	}
}

func TestRefreshNetworkSettingsPreservesManagedNestedBlocksAndProxyPassword(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/settings/network" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"proxy":{
				"enabled":true,
				"hostname":"proxy.internal",
				"port":8080,
				"useSsl":true,
				"user":"proxy-user",
				"bypassFilter":"localhost",
				"bypassLocalAddresses":true
			},
			"dnsCache":{
				"enabled":true,
				"forceMinTtl":60,
				"forceMaxTtl":600
			}
		}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	resource := &NetworkSettingsResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}
	data := &NetworkSettingsModel{
		Proxy: &NetworkProxyModel{
			Password: types.StringValue("existing-secret"),
		},
		DNSCache: &NetworkDNSCacheModel{},
	}

	if err := resource.refreshNetworkSettings(context.Background(), data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.Proxy == nil {
		t.Fatal("expected managed proxy block to be refreshed")
	}
	if got := data.Proxy.Hostname.ValueString(); got != "proxy.internal" {
		t.Fatalf("expected proxy hostname proxy.internal, got %q", got)
	}
	if got := data.Proxy.Password.ValueString(); got != "existing-secret" {
		t.Fatalf("expected existing proxy password to be preserved, got %q", got)
	}
	if data.DNSCache == nil {
		t.Fatal("expected managed dns_cache block to be refreshed")
	}
	if got := data.DNSCache.ForceMaxTTL.ValueInt64(); got != 600 {
		t.Fatalf("expected force_max_ttl 600, got %d", got)
	}
}

func TestApplyOverrideRuleBodyMapsFields(t *testing.T) {
	var data OverrideRuleModel
	err := applyOverrideRuleBody(&data, []byte(`{
		"id":3,
		"users":"1,2",
		"genre":"action",
		"profileId":7,
		"radarrServiceId":4,
		"createdAt":"2026-01-01T00:00:00.000Z",
		"updatedAt":"2026-01-02T00:00:00.000Z"
	}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := data.ID.ValueString(); got != "3" {
		t.Fatalf("expected id 3, got %q", got)
	}
	if got := data.ProfileID.ValueInt64(); got != 7 {
		t.Fatalf("expected profile id 7, got %d", got)
	}
	if got := data.RadarrServiceID.ValueInt64(); got != 4 {
		t.Fatalf("expected radarr service id 4, got %d", got)
	}
}

func TestApplyBlocklistMapMapsFields(t *testing.T) {
	var decoded map[string]any
	if err := json.Unmarshal([]byte(`{
		"tmdbId":438631,
		"mediaType":"movie",
		"title":"Dune",
		"user":{"id":1},
		"blocklistedTags":",123,",
		"createdAt":"2026-01-01T00:00:00.000Z"
	}`), &decoded); err != nil {
		t.Fatal(err)
	}

	var data BlocklistModel
	if err := applyBlocklistMap(&data, decoded); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := data.ID.ValueString(); got != "movie:438631" {
		t.Fatalf("expected composite id, got %q", got)
	}
	if got := data.UserID.ValueInt64(); got != 1 {
		t.Fatalf("expected user id 1, got %d", got)
	}
	if got := data.BlocklistedTags.ValueString(); got != ",123," {
		t.Fatalf("expected blocklisted tags, got %q", got)
	}
}
