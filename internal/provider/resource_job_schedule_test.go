package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestJobScheduleLifecycleRestoresOriginalSchedule(t *testing.T) {
	currentSchedule := "0 0 * * *"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			if req.URL.Path != "/api/v1/settings/jobs" {
				t.Fatalf("unexpected GET path %s", req.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]map[string]string{{"id": "plex-sync", "schedule": currentSchedule}})
		case http.MethodPost:
			if req.URL.Path != "/api/v1/settings/jobs/plex-sync/schedule" {
				t.Fatalf("unexpected POST path %s", req.URL.Path)
			}
			var payload map[string]string
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				t.Fatalf("decode schedule payload: %v", err)
			}
			currentSchedule = payload["schedule"]
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected method %s", req.Method)
		}
	}))
	defer srv.Close()
	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	r := &JobScheduleResource{client: NewClient(baseURL, "key", "test-agent", false, defaultRequestTimeout)}
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	model := JobScheduleModel{
		ID:               types.StringUnknown(),
		JobID:            types.StringValue("plex-sync"),
		Schedule:         types.StringValue("0 6 * * *"),
		PreviousSchedule: types.StringUnknown(),
	}
	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	if diags := plan.Set(context.Background(), &model); diags.HasError() {
		t.Fatalf("set create plan: %v", diags)
	}
	createResp := resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("create diagnostics: %v", createResp.Diagnostics)
	}
	if diags := createResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("get create state: %v", diags)
	}
	if model.PreviousSchedule.ValueString() != "0 0 * * *" || currentSchedule != "0 6 * * *" {
		t.Fatalf("create did not preserve/apply schedules: state=%#v current=%q", model, currentSchedule)
	}

	model.Schedule = types.StringValue("0 12 * * *")
	plan = tfsdk.Plan{Schema: schemaResp.Schema}
	if diags := plan.Set(context.Background(), &model); diags.HasError() {
		t.Fatalf("set update plan: %v", diags)
	}
	updateResp := resource.UpdateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Update(context.Background(), resource.UpdateRequest{Plan: plan}, &updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("update diagnostics: %v", updateResp.Diagnostics)
	}

	readResp := resource.ReadResponse{State: updateResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: updateResp.State}, &readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", readResp.Diagnostics)
	}
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("get read state: %v", diags)
	}
	if model.Schedule.ValueString() != "0 12 * * *" || model.PreviousSchedule.ValueString() != "0 0 * * *" {
		t.Fatalf("unexpected read state: %#v", model)
	}

	var deleteResp resource.DeleteResponse
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp.State}, &deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("delete diagnostics: %v", deleteResp.Diagnostics)
	}
	if currentSchedule != "0 0 * * *" {
		t.Fatalf("delete restored %q, want original schedule", currentSchedule)
	}
}

func TestJobScheduleRefreshSetsPreviousScheduleOnFirstRead(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/settings/jobs" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":"plex-sync","schedule":"0 0 * * *"}]`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	resource := &JobScheduleResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}
	data := JobScheduleModel{
		ID:    types.StringValue("plex-sync"),
		JobID: types.StringValue("plex-sync"),
	}

	if err := resource.refreshJobSchedule(context.Background(), &data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := data.Schedule.ValueString(); got != "0 0 * * *" {
		t.Fatalf("expected current schedule, got %q", got)
	}
	if got := data.PreviousSchedule.ValueString(); got != "0 0 * * *" {
		t.Fatalf("expected previous schedule to be initialized, got %q", got)
	}
}

func TestRestorePreviousSchedulePostsOriginalValue(t *testing.T) {
	var requestBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/settings/jobs/plex-sync/schedule" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		requestBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	resource := &JobScheduleResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}
	data := JobScheduleModel{
		JobID:            types.StringValue("plex-sync"),
		PreviousSchedule: types.StringValue("0 0 * * *"),
	}

	if err := resource.restorePreviousSchedule(context.Background(), &data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestBody != `{"schedule":"0 0 * * *"}` {
		t.Fatalf("unexpected restore payload %s", requestBody)
	}
}
