package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestRadarrRootFoldersDataSourceRead(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/rootfolder" {
			t.Errorf("expected path /api/v3/rootfolder, got %s", r.URL.Path)
		}
		if r.Header.Get("X-Api-Key") != "test-key" {
			t.Errorf("expected api key test-key")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{"id": 1, "path": "/movies", "accessible": true, "freeSpace": 123456789},
			{"id": 2, "path": "/anime_movies", "accessible": false, "freeSpace": 987654321}
		]`))
	}))
	defer srv.Close()

	d := NewRadarrRootFoldersDataSource()

	req := datasource.ReadRequest{
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"url":          tftypes.String,
					"hostname":     tftypes.String,
					"port":         tftypes.Number,
					"api_key":      tftypes.String,
					"use_ssl":      tftypes.Bool,
					"base_url":     tftypes.String,
					"root_folders": tftypes.List{ElementType: tftypes.Object{}},
				},
			}, map[string]tftypes.Value{
				"url":          tftypes.NewValue(tftypes.String, srv.URL),
				"hostname":     tftypes.NewValue(tftypes.String, nil),
				"port":         tftypes.NewValue(tftypes.Number, nil),
				"api_key":      tftypes.NewValue(tftypes.String, "test-key"),
				"use_ssl":      tftypes.NewValue(tftypes.Bool, nil),
				"base_url":     tftypes.NewValue(tftypes.String, nil),
				"root_folders": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{}}, nil),
			}),
		},
	}

	schemaResp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, schemaResp)
	req.Config.Schema = schemaResp.Schema

	resp := &datasource.ReadResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
		},
	}

	d.Read(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics.Errors())
	}

	var data RadarrRootFoldersDataSourceModel
	resp.State.Get(context.Background(), &data)

	if len(data.RootFolders) != 2 {
		t.Fatalf("expected 2 root folders, got %d", len(data.RootFolders))
	}
	if data.RootFolders[0].ID.ValueInt64() != 1 || data.RootFolders[0].Path.ValueString() != "/movies" || data.RootFolders[0].Accessible.ValueBool() != true || data.RootFolders[0].FreeSpace.ValueInt64() != 123456789 {
		t.Errorf("unexpected folder 1 data: %+v", data.RootFolders[0])
	}
}
