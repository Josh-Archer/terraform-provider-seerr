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

func TestSonarrLanguageProfilesDataSourceRead(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/languageprofile" {
			t.Errorf("expected path /api/v3/languageprofile, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{"id": 1, "name": "English"},
			{"id": 2, "name": "Any"}
		]`))
	}))
	defer srv.Close()

	d := NewSonarrLanguageProfilesDataSource()

	req := datasource.ReadRequest{
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"url":               tftypes.String,
					"hostname":          tftypes.String,
					"port":              tftypes.Number,
					"api_key":           tftypes.String,
					"use_ssl":           tftypes.Bool,
					"base_url":          tftypes.String,
					"language_profiles": tftypes.List{ElementType: tftypes.Object{}},
				},
			}, map[string]tftypes.Value{
				"url":               tftypes.NewValue(tftypes.String, srv.URL),
				"hostname":          tftypes.NewValue(tftypes.String, nil),
				"port":              tftypes.NewValue(tftypes.Number, nil),
				"api_key":           tftypes.NewValue(tftypes.String, "test-key"),
				"use_ssl":           tftypes.NewValue(tftypes.Bool, nil),
				"base_url":          tftypes.NewValue(tftypes.String, nil),
				"language_profiles": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{}}, nil),
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

	var data SonarrLanguageProfilesDataSourceModel
	resp.State.Get(context.Background(), &data)

	if len(data.LanguageProfiles) != 2 {
		t.Fatalf("expected 2 language profiles, got %d", len(data.LanguageProfiles))
	}
	if data.LanguageProfiles[0].ID.ValueInt64() != 1 || data.LanguageProfiles[0].Name.ValueString() != "English" {
		t.Errorf("unexpected profile 1 data: %+v", data.LanguageProfiles[0])
	}
}
