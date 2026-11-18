package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func TestAboutReadOffline(t *testing.T) {
	for _, tc := range []struct {
		name      string
		status    int
		body      string
		wantError bool
	}{
		{"success", 200, `{"version":"3.5.0","totalRequests":17,"totalMediaItems":23,"tz":"UTC","appDataPath":"/test"}`, false},
		{"forbidden", 403, `{"message":"Forbidden"}`, true},
		{"invalid_json", 200, `not json`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" || r.URL.Path != "/api/v1/settings/about" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()
			base, err := url.Parse(srv.URL)
			if err != nil {
				t.Fatal(err)
			}
			d := &AboutDataSource{client: NewClient(base, "dummy-api-key", "unit-test", false, defaultRequestTimeout, 0, 0)}
			ctx := context.Background()
			var schema datasource.SchemaResponse
			d.Schema(ctx, datasource.SchemaRequest{}, &schema)
			resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema.Schema}}
			d.Read(ctx, datasource.ReadRequest{}, &resp)
			if resp.Diagnostics.HasError() != tc.wantError {
				t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
			}
			if !tc.wantError {
				var data AboutDataSourceModel
				if diags := resp.State.Get(ctx, &data); diags.HasError() {
					t.Fatal(diags)
				}
				if data.ID.ValueString() != "seerr_about" || data.Version.ValueString() != "3.5.0" || data.TotalRequests.ValueInt64() != 17 || data.TotalMediaItems.ValueInt64() != 23 || data.TZ.ValueString() != "UTC" || data.AppDataPath.ValueString() != "/test" {
					t.Fatalf("unexpected state %+v", data)
				}
			}
		})
	}
}
