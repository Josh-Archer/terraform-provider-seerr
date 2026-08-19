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

func TestNetworkSettingsDataSourceRead(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/settings/network" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"csrfProtection": true,
			"forceIpv4First": false,
			"trustProxy": true,
			"apiRequestTimeout": 5000,
			"proxy": {
				"enabled": true,
				"hostname": "10.0.0.1",
				"port": 8080,
				"useSsl": false,
				"user": "proxyuser",
				"password": "proxypassword",
				"bypassFilter": "localhost",
				"bypassLocalAddresses": true
			},
			"dnsCache": {
				"enabled": true,
				"forceMinTtl": 60,
				"forceMaxTtl": 3600
			}
		}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	d := &NetworkSettingsDataSource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout, 0, 0),
	}

	var data NetworkSettingsModel
	res, err := d.client.Request(context.Background(), "GET", "/api/v1/settings/network", "", nil)
	if err != nil {
		t.Fatal(err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err != nil {
		t.Fatal(err)
	}

	resource := &NetworkSettingsResource{}
	resource.applyNetworkSettingsMap(&data, decoded)
	data.ID = types.StringValue("network")

	if !data.CSRFProtection.ValueBool() {
		t.Fatalf("expected csrf_protection true, got %v", data.CSRFProtection.ValueBool())
	}
	if data.ForceIPv4First.ValueBool() {
		t.Fatalf("expected force_ipv4_first false, got %v", data.ForceIPv4First.ValueBool())
	}
	if !data.TrustProxy.ValueBool() {
		t.Fatalf("expected trust_proxy true, got %v", data.TrustProxy.ValueBool())
	}
	if data.APIRequestTimeoutMS.ValueInt64() != 5000 {
		t.Fatalf("expected api_request_timeout_ms 5000, got %d", data.APIRequestTimeoutMS.ValueInt64())
	}
	if data.Proxy == nil || !data.Proxy.Enabled.ValueBool() || data.Proxy.Hostname.ValueString() != "10.0.0.1" || data.Proxy.Port.ValueInt64() != 8080 {
		t.Fatalf("unexpected proxy data: %+v", data.Proxy)
	}
	if data.DNSCache == nil || !data.DNSCache.Enabled.ValueBool() || data.DNSCache.ForceMinTTL.ValueInt64() != 60 || data.DNSCache.ForceMaxTTL.ValueInt64() != 3600 {
		t.Fatalf("unexpected dns_cache data: %+v", data.DNSCache)
	}
}
