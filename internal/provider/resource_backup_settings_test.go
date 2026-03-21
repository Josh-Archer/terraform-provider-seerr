package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBackupSettingsApplyDecodedSettings(t *testing.T) {
	r := &BackupSettingsResource{}
	data := BackupSettingsModel{
		Schedule:    types.StringValue("0 0 * * *"),
		Retention:   types.Int64Value(10),
		StoragePath: types.StringValue("/backups"),
	}

	r.applyDecodedSettings(&data, map[string]any{
		"schedule":     "0 1 * * *",
		"retention":    float64(20),
		"storage_path": "/mnt/backups",
	})

	if got := data.Schedule.ValueString(); got != "0 1 * * *" {
		t.Fatalf("expected schedule 0 1 * * *, got %q", got)
	}
	if got := data.Retention.ValueInt64(); got != 20 {
		t.Fatalf("expected retention 20, got %d", got)
	}
	if got := data.StoragePath.ValueString(); got != "/mnt/backups" {
		t.Fatalf("expected storage_path /mnt/backups, got %q", got)
	}
}

func TestBackupSettingsRefreshState(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/settings/backups" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"schedule":"0 2 * * *",
			"retention":30,
			"storage_path":"/var/lib/os/backups"
		}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	r := &BackupSettingsResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}
	data := BackupSettingsModel{}

	if err := r.refreshState(context.Background(), &data); err != nil {
		t.Fatal(err)
	}

	if got := data.Schedule.ValueString(); got != "0 2 * * *" {
		t.Fatalf("expected schedule 0 2 * * *, got %q", got)
	}
	if got := data.Retention.ValueInt64(); got != 30 {
		t.Fatalf("expected retention 30, got %d", got)
	}
	if got := data.StoragePath.ValueString(); got != "/var/lib/os/backups" {
		t.Fatalf("expected storage_path /var/lib/os/backups, got %q", got)
	}
}
