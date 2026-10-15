package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRadarrServiceDataSource_AllServers(t *testing.T) {
	ctx := context.Background()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, "/api/v1/service/radarr", req.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"id": 1,
				"name": "Radarr HD",
				"isDefault": true,
				"activeDirectory": "/movies/hd",
				"is4k": false
			},
			{
				"id": 2,
				"name": "Radarr 4K",
				"isDefault": false,
				"activeDirectory": "/movies/4k",
				"is4k": true
			}
		]`))
	}))
	defer srv.Close()

	d := &RadarrServiceDataSource{client: testAPIClient(t, srv.URL)}

	var schemaResp datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &schemaResp)
	typ, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	require.True(t, ok)

	vals := map[string]tftypes.Value{}
	for name, attrType := range typ.AttributeTypes {
		vals[name] = tftypes.NewValue(attrType, nil)
	}

	req := datasource.ReadRequest{Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: tftypes.NewValue(typ, vals)}}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}

	d.Read(ctx, req, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

	var got RadarrServiceDataSourceModel
	require.False(t, resp.State.Get(ctx, &got).HasError())

	assert.Equal(t, "radarr_service", got.ID.ValueString())
	assert.True(t, got.ServerID.IsNull())
	require.Len(t, got.Servers, 2)

	assert.Equal(t, int64(1), got.Servers[0].ID.ValueInt64())
	assert.Equal(t, "Radarr HD", got.Servers[0].Name.ValueString())
	assert.True(t, got.Servers[0].IsDefault.ValueBool())
	assert.Equal(t, "/movies/hd", got.Servers[0].ActiveDirectory.ValueString())
	assert.False(t, got.Servers[0].Is4K.ValueBool())

	assert.Equal(t, int64(2), got.Servers[1].ID.ValueInt64())
	assert.Equal(t, "Radarr 4K", got.Servers[1].Name.ValueString())
	assert.False(t, got.Servers[1].IsDefault.ValueBool())
	assert.Equal(t, "/movies/4k", got.Servers[1].ActiveDirectory.ValueString())
	assert.True(t, got.Servers[1].Is4K.ValueBool())

	assert.Empty(t, got.Profiles)
}

func TestRadarrServiceDataSource_SingleServer(t *testing.T) {
	ctx := context.Background()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, "/api/v1/service/radarr/1", req.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"server": {
				"id": 1,
				"name": "Radarr HD",
				"isDefault": true,
				"activeDirectory": "/movies/hd",
				"is4k": false
			},
			"profiles": [
				{
					"id": 10,
					"name": "HD-1080p"
				},
				{
					"id": 20,
					"name": "Ultra-HD"
				}
			]
		}`))
	}))
	defer srv.Close()

	d := &RadarrServiceDataSource{client: testAPIClient(t, srv.URL)}

	var schemaResp datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &schemaResp)
	typ, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	require.True(t, ok)

	vals := map[string]tftypes.Value{}
	for name, attrType := range typ.AttributeTypes {
		vals[name] = tftypes.NewValue(attrType, nil)
	}
	vals["server_id"] = tftypes.NewValue(tftypes.Number, 1)

	req := datasource.ReadRequest{Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: tftypes.NewValue(typ, vals)}}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}

	d.Read(ctx, req, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

	var got RadarrServiceDataSourceModel
	require.False(t, resp.State.Get(ctx, &got).HasError())

	assert.Equal(t, "1", got.ID.ValueString())
	assert.Equal(t, int64(1), got.ServerID.ValueInt64())
	require.Len(t, got.Servers, 1)
	assert.Equal(t, int64(1), got.Servers[0].ID.ValueInt64())
	assert.Equal(t, "Radarr HD", got.Servers[0].Name.ValueString())
	assert.True(t, got.Servers[0].IsDefault.ValueBool())
	assert.Equal(t, "/movies/hd", got.Servers[0].ActiveDirectory.ValueString())
	assert.False(t, got.Servers[0].Is4K.ValueBool())

	require.Len(t, got.Profiles, 2)
	assert.Equal(t, int64(10), got.Profiles[0].ID.ValueInt64())
	assert.Equal(t, "HD-1080p", got.Profiles[0].Name.ValueString())
	assert.Equal(t, int64(20), got.Profiles[1].ID.ValueInt64())
	assert.Equal(t, "Ultra-HD", got.Profiles[1].Name.ValueString())
}

func TestRadarrServiceDataSource_NotFound(t *testing.T) {
	ctx := context.Background()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/api/v1/service/radarr/99", req.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "Radarr server not found"}`))
	}))
	defer srv.Close()

	d := &RadarrServiceDataSource{client: testAPIClient(t, srv.URL)}

	var schemaResp datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &schemaResp)
	typ, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	require.True(t, ok)

	vals := map[string]tftypes.Value{}
	for name, attrType := range typ.AttributeTypes {
		vals[name] = tftypes.NewValue(attrType, nil)
	}
	vals["server_id"] = tftypes.NewValue(tftypes.Number, 99)

	req := datasource.ReadRequest{Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: tftypes.NewValue(typ, vals)}}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}

	d.Read(ctx, req, &resp)
	require.True(t, resp.Diagnostics.HasError())
}

func TestRadarrServiceDataSource_MetadataAndConfigure(t *testing.T) {
	d := NewRadarrServiceDataSource()

	var metaResp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "seerr"}, &metaResp)
	assert.Equal(t, "seerr_radarr_service", metaResp.TypeName)

	dWithConf, ok := d.(datasource.DataSourceWithConfigure)
	require.True(t, ok)

	var confResp datasource.ConfigureResponse
	dWithConf.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "invalid"}, &confResp)
	assert.True(t, confResp.Diagnostics.HasError())

	confResp = datasource.ConfigureResponse{}
	dWithConf.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: nil}, &confResp)
	assert.False(t, confResp.Diagnostics.HasError())
}
