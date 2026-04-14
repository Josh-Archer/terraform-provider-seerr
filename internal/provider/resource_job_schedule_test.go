package provider

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
