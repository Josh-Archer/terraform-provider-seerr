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

func TestBackupSettingsReadOffline(t *testing.T) {
	for _, tc := range []struct {
		name      string
		status    int
		body      string
		wantError bool
		populated bool
	}{
		{"success", 200, `{"schedule":"0 3 * * *","retention":40,"storage_path":"/fixture/backups"}`, false, true},
		{"missing_optional", 200, `{}`, false, false},
		{"wrong_optional_types", 200, `{"schedule":1,"retention":"forty","storage_path":false}`, false, false},
		{"forbidden", 403, `{"message":"Forbidden"}`, true, false},
		{"invalid_json", 200, `not json`, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/api/v1/settings/backups" {
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
			d := &BackupSettingsDataSource{client: NewClient(base, "fixture-key", "offline-test", false, defaultRequestTimeout, 0, 0)}
			ctx := context.Background()
			var schema datasource.SchemaResponse
			d.Schema(ctx, datasource.SchemaRequest{}, &schema)
			resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema.Schema}}
			configState := tfsdk.State{Schema: schema.Schema}
			if diags := configState.Set(context.Background(), &BackupSettingsDataSourceModel{}); diags.HasError() {
				t.Fatal(diags)
			}
			d.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: schema.Schema, Raw: configState.Raw}}, &resp)
			if resp.Diagnostics.HasError() != tc.wantError {
				t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
			}
			if tc.wantError {
				if !resp.State.Raw.IsNull() {
					t.Fatal("failure must not publish successful state")
				}
				return
			}
			var got BackupSettingsDataSourceModel
			if diags := resp.State.Get(ctx, &got); diags.HasError() {
				t.Fatal(diags)
			}
			if got.ID.ValueString() != "backups" || got.StatusCode.ValueInt64() != 200 || got.ResponseJSON.ValueString() != tc.body {
				t.Fatalf("unexpected response state: %+v", got)
			}
			if tc.populated {
				if got.Schedule.ValueString() != "0 3 * * *" || got.Retention.ValueInt64() != 40 || got.StoragePath.ValueString() != "/fixture/backups" {
					t.Fatalf("unexpected settings: %+v", got)
				}
			} else if !got.Schedule.IsNull() || !got.Retention.IsNull() || !got.StoragePath.IsNull() {
				t.Fatalf("optional fields should remain null: %+v", got)
			}
		})
	}
}

func TestBackupSettingsReadOfflineTransportFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("cancelled request reached server") }))
	defer srv.Close()
	base, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	d := &BackupSettingsDataSource{client: NewClient(base, "fixture-key", "offline-test", false, defaultRequestTimeout, 0, 0)}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var schema datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &schema)
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema.Schema}}
	configState := tfsdk.State{Schema: schema.Schema}
	if diags := configState.Set(context.Background(), &BackupSettingsDataSourceModel{}); diags.HasError() {
		t.Fatal(diags)
	}
	d.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: schema.Schema, Raw: configState.Raw}}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("cancelled transport must fail")
	}
	if !resp.State.Raw.IsNull() {
		t.Fatal("failure must not publish successful state")
	}
}
