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

func TestSonarrTagsDataSourceRead(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/tag" {
			t.Errorf("expected path /api/v3/tag, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{"id": 1, "label": "anime"},
			{"id": 2, "label": "english"}
		]`))
	}))
	defer srv.Close()

	d := NewSonarrTagsDataSource()

	req := datasource.ReadRequest{
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"url":      tftypes.String,
					"hostname": tftypes.String,
					"port":     tftypes.Number,
					"api_key":  tftypes.String,
					"use_ssl":  tftypes.Bool,
					"base_url": tftypes.String,
					"tags":     tftypes.List{ElementType: tftypes.Object{}},
				},
			}, map[string]tftypes.Value{
				"url":      tftypes.NewValue(tftypes.String, srv.URL),
				"hostname": tftypes.NewValue(tftypes.String, nil),
				"port":     tftypes.NewValue(tftypes.Number, nil),
				"api_key":  tftypes.NewValue(tftypes.String, "test-key"),
				"use_ssl":  tftypes.NewValue(tftypes.Bool, nil),
				"base_url": tftypes.NewValue(tftypes.String, nil),
				"tags":     tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{}}, nil),
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

	var data SonarrTagsDataSourceModel
	resp.State.Get(context.Background(), &data)

	if len(data.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(data.Tags))
	}
	if data.Tags[0].ID.ValueInt64() != 1 || data.Tags[0].Label.ValueString() != "anime" {
		t.Errorf("unexpected tag 1 data: %+v", data.Tags[0])
	}
}
