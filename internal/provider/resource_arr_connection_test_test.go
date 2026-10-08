package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestARRConnectionTestSendsEndpointPayloadAndCapturesResponse(t *testing.T) {
	tests := []struct{ service, path string }{
		{"radarr", "/api/v1/settings/radarr/test"},
		{"sonarr", "/api/v1/settings/sonarr/test"},
	}
	for _, tt := range tests {
		t.Run(tt.service, func(t *testing.T) {
			var body map[string]any
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				require.Equal(t, http.MethodPost, req.Method)
				require.Equal(t, tt.path, req.URL.Path)
				require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"profiles":[{"id":1}],"rootFolders":[],"tags":[]}`))
			}))
			defer srv.Close()
			baseURL, err := url.Parse(srv.URL)
			require.NoError(t, err)
			resource := &ARRConnectionTestResource{client: NewClient(baseURL, "test-key", "test-agent", false, defaultRequestTimeout, 0, 0)}
			data := ARRConnectionTestModel{
				Service: types.StringValue(tt.service), Hostname: types.StringValue("arr.local"), Port: types.Int64Value(7878),
				APIKey: types.StringValue("mock-arr-secret"), UseSSL: types.BoolValue(true), BaseURL: types.StringValue("/arr"),
			}
			require.NoError(t, resource.test(context.Background(), &data))
			require.Equal(t, map[string]any{"hostname": "arr.local", "port": float64(7878), "apiKey": "mock-arr-secret", "useSsl": true, "baseUrl": "/arr"}, body)
			require.Contains(t, data.ResponseJSON.ValueString(), `"profiles"`)
			require.Equal(t, tt.service, data.ID.ValueString())
		})
	}
}

func TestARRConnectionTestReturnsNonSuccessResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Unable to connect", http.StatusBadGateway)
	}))
	defer srv.Close()
	baseURL, err := url.Parse(srv.URL)
	require.NoError(t, err)
	r := &ARRConnectionTestResource{client: NewClient(baseURL, "test-key", "test-agent", false, defaultRequestTimeout, 0, 0)}
	data := ARRConnectionTestModel{Service: types.StringValue("sonarr"), Hostname: types.StringValue("sonarr.local"), Port: types.Int64Value(8989), APIKey: types.StringValue("mock-secret"), UseSSL: types.BoolValue(false)}
	require.ErrorContains(t, r.test(context.Background(), &data), "status 502")
}

func TestARRConnectionTestMarksCredentialsAndResultSensitive(t *testing.T) {
	var response resource.SchemaResponse
	NewARRConnectionTestResource().Schema(context.Background(), resource.SchemaRequest{}, &response)
	require.True(t, response.Schema.Attributes["api_key"].IsSensitive())
	require.True(t, response.Schema.Attributes["response_json"].IsSensitive())
}
