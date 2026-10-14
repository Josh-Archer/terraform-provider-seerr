package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSonarrServerTestExecute_Success(t *testing.T) {
	testCalled := false

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/settings/sonarr/test" && r.Method == "POST" {
			testCalled = true
			var body map[string]any
			err := json.NewDecoder(r.Body).Decode(&body)
			require.NoError(t, err)
			assert.Equal(t, "127.0.0.1", body["hostname"])
			assert.Equal(t, float64(8989), body["port"])
			assert.Equal(t, "test-api-key", body["apiKey"])
			assert.Equal(t, false, body["useSsl"])

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"profiles": [{"id": 1, "name": "HD-1080p"}, {"id": 2, "name": "Ultra-HD"}]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	baseURL, err := url.Parse(mockServer.URL)
	require.NoError(t, err)

	client := NewClient(baseURL, "test-api-key", "test-agent", false, defaultRequestTimeout, 0, 0)
	res := &SonarrServerTestResource{client: client}

	ctx := context.Background()
	model := &SonarrServerTestModel{
		Hostname: types.StringValue("127.0.0.1"),
		Port:     types.Int64Value(8989),
		APIKey:   types.StringValue("test-api-key"),
		UseSSL:   types.BoolValue(false),
	}

	err = res.testSonarrServer(ctx, model)
	require.NoError(t, err)
	assert.True(t, testCalled)
	assert.Equal(t, int64(2), model.ProfilesCount.ValueInt64())
}

func TestSonarrServerTestExecute_WithBaseURL(t *testing.T) {
	testCalled := false

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/settings/sonarr/test" && r.Method == "POST" {
			testCalled = true
			var body map[string]any
			err := json.NewDecoder(r.Body).Decode(&body)
			require.NoError(t, err)
			assert.Equal(t, "sonarr.internal", body["hostname"])
			assert.Equal(t, float64(443), body["port"])
			assert.Equal(t, "key-456", body["apiKey"])
			assert.Equal(t, true, body["useSsl"])
			assert.Equal(t, "/sonarr", body["baseUrl"])

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"profiles": [{"id": 1, "name": "Standard"}]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	baseURL, err := url.Parse(mockServer.URL)
	require.NoError(t, err)

	client := NewClient(baseURL, "key-456", "test-agent", false, defaultRequestTimeout, 0, 0)
	res := &SonarrServerTestResource{client: client}

	ctx := context.Background()
	model := &SonarrServerTestModel{
		Hostname: types.StringValue("sonarr.internal"),
		Port:     types.Int64Value(443),
		APIKey:   types.StringValue("key-456"),
		UseSSL:   types.BoolValue(true),
		BaseURL:  types.StringValue("/sonarr"),
	}

	err = res.testSonarrServer(ctx, model)
	require.NoError(t, err)
	assert.True(t, testCalled)
	assert.Equal(t, int64(1), model.ProfilesCount.ValueInt64())
}

func TestSonarrServerTestExecute_Failure(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message": "Failed to connect to Sonarr"}`))
	}))
	defer mockServer.Close()

	baseURL, err := url.Parse(mockServer.URL)
	require.NoError(t, err)

	client := NewClient(baseURL, "test-api-key", "test-agent", false, defaultRequestTimeout, 0, 0)
	res := &SonarrServerTestResource{client: client}

	ctx := context.Background()
	model := &SonarrServerTestModel{
		Hostname: types.StringValue("127.0.0.1"),
		Port:     types.Int64Value(8989),
		APIKey:   types.StringValue("test-api-key"),
		UseSSL:   types.BoolValue(false),
	}

	err = res.testSonarrServer(ctx, model)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 400")
}

func TestSonarrServerTest_MetadataAndSchema(t *testing.T) {
	res := NewSonarrServerTestResource()

	var metaResp resource.MetadataResponse
	res.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "seerr"}, &metaResp)
	assert.Equal(t, "seerr_sonarr_server_test", metaResp.TypeName)

	var schemaResp resource.SchemaResponse
	res.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	attrs := schemaResp.Schema.Attributes
	assert.Equal(t, 8, len(attrs))
	assert.True(t, attrs["id"].IsComputed())
	assert.True(t, attrs["hostname"].IsRequired())
	assert.True(t, attrs["port"].IsRequired())
	assert.True(t, attrs["api_key"].IsRequired())
	assert.True(t, attrs["api_key"].IsSensitive())
	assert.True(t, attrs["use_ssl"].IsRequired())
	assert.True(t, attrs["base_url"].IsOptional())
	assert.True(t, attrs["triggers"].IsOptional())
	assert.True(t, attrs["profiles_count"].IsComputed())
}

func TestSonarrServerTest_ImportState(t *testing.T) {
	res := NewSonarrServerTestResource()
	var schemaResp resource.SchemaResponse
	res.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	state := tfsdk.State{
		Schema: schemaResp.Schema,
	}
	initialModel := SonarrServerTestModel{
		Triggers: types.MapNull(types.StringType),
	}
	require.False(t, state.Set(context.Background(), &initialModel).HasError())

	req := resource.ImportStateRequest{ID: "127.0.0.1:8989"}
	resp := resource.ImportStateResponse{State: state}

	importRes, ok := res.(resource.ResourceWithImportState)
	require.True(t, ok)
	importRes.ImportState(context.Background(), req, &resp)
	require.False(t, resp.Diagnostics.HasError())

	var data SonarrServerTestModel
	diags := resp.State.Get(context.Background(), &data)
	require.False(t, diags.HasError())
	assert.Equal(t, "127.0.0.1:8989", data.ID.ValueString())
}
