package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLibrarySettingsDataSourceReadOnSeerr341(t *testing.T) {
	for _, server := range []string{"jellyfin", "plex", "emby"} {
		for _, sync := range []bool{false, true} {
			for _, enabled := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/sync=%t/enabled=%t", server, sync, enabled), func(t *testing.T) {
					basePath := "/api/v1/settings/" + server + "/library"
					libs := []mockLibrary{{ID: "movies", Name: "Movies", Enabled: enabled}, {ID: "tv", Name: "TV", Enabled: false}}
					srv := newLibrarySettingsServerV341(t, basePath, libs)
					client := testAPIClient(t, srv.URL)
					var d datasource.DataSource
					switch server {
					case "jellyfin":
						d = &JellyfinLibrarySettingsDataSource{client: client}
					case "plex":
						d = &PlexLibrarySettingsDataSource{client: client}
					case "emby":
						d = &EmbyLibrarySettingsDataSource{client: client}
					}
					ctx := context.Background()
					schema := datasource.SchemaResponse{}
					d.Schema(ctx, datasource.SchemaRequest{}, &schema)
					typ := schema.Schema.Type().TerraformType(ctx).(tftypes.Object)
					vals := map[string]tftypes.Value{}
					for k, v := range typ.AttributeTypes {
						vals[k] = tftypes.NewValue(v, nil)
					}
					vals["sync_on_read"] = tftypes.NewValue(tftypes.Bool, sync)
					req := datasource.ReadRequest{Config: tfsdk.Config{Schema: schema.Schema, Raw: tftypes.NewValue(typ, vals)}}
					resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema.Schema}}
					d.Read(ctx, req, &resp)
					require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
					var discovered []struct {
						ID      string `tfsdk:"id"`
						Name    string `tfsdk:"name"`
						Enabled bool   `tfsdk:"enabled"`
					}
					require.False(t, resp.State.GetAttribute(ctx, path.Root("libraries"), &discovered).HasError())
					require.Len(t, discovered, len(libs))
					for i, lib := range libs {
						assert.Equal(t, lib.ID, discovered[i].ID)
						assert.Equal(t, lib.Name, discovered[i].Name)
						assert.Equal(t, lib.Enabled, discovered[i].Enabled)
					}
					var ids []string
					require.False(t, resp.State.GetAttribute(ctx, path.Root("enabled_libraries"), &ids).HasError())
					if enabled {
						assert.Equal(t, []string{"movies"}, ids)
					} else {
						assert.Empty(t, ids)
					}
					assert.Equal(t, libs, srv.snapshot(), "discovery must preserve enablement")
					var syncCalls int
					for _, call := range srv.recordedCalls() {
						if call.Path == basePath {
							syncCalls++
							assert.Equal(t, "true", call.Query.Get("sync"))
							if enabled {
								assert.Equal(t, "movies", call.Query.Get("enable"))
							} else {
								assert.NotContains(t, call.Query, "enable")
							}
						}
					}
					if sync {
						assert.Equal(t, 1, syncCalls)
					} else {
						assert.Zero(t, syncCalls)
					}
				})
			}
		}
	}
}
