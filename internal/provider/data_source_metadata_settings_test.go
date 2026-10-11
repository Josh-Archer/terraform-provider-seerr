package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestMetadataSettingsDataSourceReadUsesMetadatasEndpoint(t *testing.T) {
	ctx := context.Background()
	body := `{"tv":"tvdb","anime":"tmdb"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet || req.URL.Path != "/api/v1/settings/metadatas" {
			t.Errorf("unexpected request %s %s", req.Method, req.URL.Path)
			http.NotFound(w, req)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()
	d := &MetadataSettingsDataSource{client: testAPIClient(t, srv.URL)}
	var schemaResp datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &schemaResp)
	typ := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for name, attrType := range typ.AttributeTypes {
		vals[name] = tftypes.NewValue(attrType, nil)
	}
	req := datasource.ReadRequest{Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: tftypes.NewValue(typ, vals)}}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	d.Read(ctx, req, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	var got MetadataSettingsModel
	require.False(t, resp.State.Get(ctx, &got).HasError())
	require.Equal(t, "metadata", got.ID.ValueString())
	require.Equal(t, "tvdb", got.TV.ValueString())
	require.Equal(t, "tmdb", got.Anime.ValueString())
	require.JSONEq(t, body, got.Response.ValueString())
}
