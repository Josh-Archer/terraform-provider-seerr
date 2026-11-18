package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestJobScheduleReadOffline(t *testing.T) {
	for _, tc := range []struct {
		name               string
		status             int
		body               string
		jobID              string
		previous           types.String
		wantPrevious       string
		removed, wantError bool
	}{
		{"refresh_preserves_original", 200, `[{"id":"plex-sync","schedule":"new"}]`, "plex-sync", types.StringValue("original"), "original", false, false},
		{"import_fallback_initializes_original", 200, `[{"id":"plex-sync","schedule":"new"}]`, "", types.StringNull(), "new", false, false},
		{"unknown_original_initialized", 200, `[{"id":"plex-sync","schedule":"new"}]`, "plex-sync", types.StringUnknown(), "new", false, false},
		{"missing_job_removes_state", 200, `[{"id":"other","schedule":"new"}]`, "plex-sync", types.StringValue("original"), "", true, false},
		{"not_found_removes_state", 404, `{}`, "plex-sync", types.StringValue("original"), "", true, false},
		{"forbidden_retains_state", 403, `{"message":"forbidden"}`, "plex-sync", types.StringValue("original"), "", false, true},
		{"malformed_retains_state", 200, `not json`, "plex-sync", types.StringValue("original"), "", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.Method != http.MethodGet || req.URL.Path != "/api/v1/settings/jobs" {
					t.Errorf("unexpected request %s %s", req.Method, req.URL.Path)
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()
			base, err := url.Parse(srv.URL)
			if err != nil {
				t.Fatal(err)
			}
			r := &JobScheduleResource{client: NewClient(base, "dummy", "offline-test", false, defaultRequestTimeout, 0, 0)}
			ctx := context.Background()
			var schema resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schema)
			state := tfsdk.State{Schema: schema.Schema}
			model := JobScheduleModel{ID: types.StringValue("plex-sync"), JobID: types.StringValue(tc.jobID), Schedule: types.StringValue("old"), PreviousSchedule: tc.previous}
			if diags := state.Set(ctx, &model); diags.HasError() {
				t.Fatal(diags)
			}
			response := resource.ReadResponse{State: state}
			r.Read(ctx, resource.ReadRequest{State: state}, &response)
			if response.Diagnostics.HasError() != tc.wantError {
				t.Fatalf("unexpected diagnostics %v", response.Diagnostics)
			}
			if response.State.Raw.IsNull() != tc.removed {
				t.Fatalf("removed state = %v; want %v", response.State.Raw.IsNull(), tc.removed)
			}
			if tc.removed {
				return
			}
			if tc.wantError {
				if !response.State.Raw.Equal(state.Raw) {
					t.Fatal("failed read changed prior state")
				}
				return
			}
			var got JobScheduleModel
			if diags := response.State.Get(ctx, &got); diags.HasError() {
				t.Fatal(diags)
			}
			if got.ID.ValueString() != "plex-sync" || got.JobID.ValueString() != "plex-sync" || got.Schedule.ValueString() != "new" || got.PreviousSchedule.ValueString() != tc.wantPrevious {
				t.Fatalf("unexpected refreshed state %+v", got)
			}
		})
	}
}
