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

func TestTautulliSettingsDataSourceRead(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/settings/tautulli" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"hostname":"tautulli.local",
			"port":8181,
			"useSsl":true,
			"urlBase":"/tautulli",
			"apiKey":"mock-api-key-123",
			"externalUrl":"https://tautulli.example.com"
		}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	d := &TautulliSettingsDataSource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout, 0, 0),
	}

	res, err := d.client.Request(context.Background(), "GET", "/api/v1/settings/tautulli", "", nil)
	if err != nil {
		t.Fatal(err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err != nil {
		t.Fatal(err)
	}

	data := TautulliSettingsDataSourceModel{}
	data.StatusCode = types.Int64Value(int64(res.StatusCode))
	data.ResponseJSON = types.StringValue(string(res.Body))

	if v, ok := stringValueFromAny(decoded["hostname"]); ok {
		data.Hostname = types.StringValue(v)
		data.IP = types.StringValue(v)
	}
	if v, ok := int64ValueFromAny(decoded["port"]); ok {
		data.Port = types.Int64Value(v)
	}
	if v, ok := boolValueFromAny(decoded["useSsl"]); ok {
		data.UseSSL = types.BoolValue(v)
	}
	if v, ok := stringValueFromAny(decoded["urlBase"]); ok {
		data.URLBase = types.StringValue(v)
	}
	if v, ok := stringValueFromAny(decoded["apiKey"]); ok && v != "" {
		data.APIKey = types.StringValue(v)
	}
	if v, ok := stringValueFromAny(decoded["externalUrl"]); ok {
		data.ExternalURL = types.StringValue(v)
	}
	data.ID = types.StringValue("tautulli")

	if data.Hostname.ValueString() != "tautulli.local" {
		t.Fatalf("expected hostname tautulli.local, got %q", data.Hostname.ValueString())
	}
	if data.IP.ValueString() != "tautulli.local" {
		t.Fatalf("expected ip tautulli.local, got %q", data.IP.ValueString())
	}
	if data.Port.ValueInt64() != 8181 {
		t.Fatalf("expected port 8181, got %d", data.Port.ValueInt64())
	}
	if !data.UseSSL.ValueBool() {
		t.Fatalf("expected use_ssl true, got %v", data.UseSSL.ValueBool())
	}
	if data.URLBase.ValueString() != "/tautulli" {
		t.Fatalf("expected url_base /tautulli, got %q", data.URLBase.ValueString())
	}
	if data.APIKey.ValueString() != "mock-api-key-123" {
		t.Fatalf("expected api_key mock-api-key-123, got %q", data.APIKey.ValueString())
	}
	if data.ExternalURL.ValueString() != "https://tautulli.example.com" {
		t.Fatalf("expected external_url https://tautulli.example.com, got %q", data.ExternalURL.ValueString())
	}
}
