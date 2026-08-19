package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJellyfinSettingsDataSourceRead(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/settings/jellyfin", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"name": "Jellyfin",
			"ip": "192.168.1.100",
			"port": 8096,
			"useSsl": false,
			"urlBase": "/jellyfin",
			"externalHostname": "jellyfin.example.com",
			"jellyfinForgotPasswordUrl": "https://jellyfin.example.com/forgot",
			"apiKey": "mock-jellyfin-key-123",
			"serverId": "jf-server-uuid-123"
		}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	require.NoError(t, err)

	d := &JellyfinSettingsDataSource{
		client: NewClient(baseURL, "test-api-key", "test-agent", false, defaultRequestTimeout, 0, 0),
	}

	res, err := d.client.Request(context.Background(), "GET", "/api/v1/settings/jellyfin", "", nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	var decoded map[string]any
	err = json.Unmarshal(res.Body, &decoded)
	require.NoError(t, err)

	data := JellyfinSettingsDataSourceModel{}
	data.StatusCode = types.Int64Value(int64(res.StatusCode))
	data.ResponseJSON = types.StringValue(string(res.Body))

	if v, ok := decoded["name"].(string); ok {
		data.Name = types.StringValue(v)
	}
	if v, ok := decoded["ip"].(string); ok {
		data.IP = types.StringValue(v)
	}
	if v, ok := decoded["port"].(float64); ok {
		data.Port = types.Int64Value(int64(v))
	}
	if v, ok := decoded["useSsl"].(bool); ok {
		data.UseSSL = types.BoolValue(v)
	}
	if v, ok := decoded["urlBase"].(string); ok {
		data.URLBase = types.StringValue(v)
	}
	if v, ok := decoded["externalHostname"].(string); ok {
		data.ExternalHostname = types.StringValue(v)
	}
	if v, ok := decoded["jellyfinForgotPasswordUrl"].(string); ok {
		data.JellyfinForgotPasswordURL = types.StringValue(v)
	}
	if v, ok := decoded["apiKey"].(string); ok && v != "" {
		data.APIKey = types.StringValue(v)
	}
	if v, ok := decoded["serverId"].(string); ok {
		data.ServerID = types.StringValue(v)
	}
	data.ID = types.StringValue("jellyfin")

	assert.Equal(t, "Jellyfin", data.Name.ValueString())
	assert.Equal(t, "192.168.1.100", data.IP.ValueString())
	assert.Equal(t, int64(8096), data.Port.ValueInt64())
	assert.False(t, data.UseSSL.ValueBool())
	assert.Equal(t, "/jellyfin", data.URLBase.ValueString())
	assert.Equal(t, "jellyfin.example.com", data.ExternalHostname.ValueString())
	assert.Equal(t, "https://jellyfin.example.com/forgot", data.JellyfinForgotPasswordURL.ValueString())
	assert.Equal(t, "mock-jellyfin-key-123", data.APIKey.ValueString())
	assert.Equal(t, "jf-server-uuid-123", data.ServerID.ValueString())
	assert.Equal(t, "jellyfin", data.ID.ValueString())
}
