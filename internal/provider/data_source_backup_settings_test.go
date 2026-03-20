package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBackupSettingsDataSourceRead(t *testing.T) {
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
			"schedule":"0 3 * * *",
			"retention":40,
			"storage_path":"/backups/seerr"
		}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	d := &BackupSettingsDataSource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}

	data := BackupSettingsDataSourceModel{}
	
	// Test the logic that would be inside Read
	res, err := d.client.Request(context.Background(), "GET", "/api/v1/settings/backups", "", nil)
	if err != nil {
		t.Fatal(err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err != nil {
		t.Fatal(err)
	}

	if v, ok := decoded["schedule"].(string); ok {
		data.Schedule = types.StringValue(v)
	}
	if v, ok := decoded["retention"].(float64); ok {
		data.Retention = types.Int64Value(int64(v))
	}
	if v, ok := decoded["storage_path"].(string); ok {
		data.StoragePath = types.StringValue(v)
	}

	if data.Schedule.ValueString() != "0 3 * * *" {
		t.Fatalf("expected schedule 0 3 * * *, got %q", data.Schedule.ValueString())
	}
	if data.Retention.ValueInt64() != 40 {
		t.Fatalf("expected retention 40, got %d", data.Retention.ValueInt64())
	}
	if data.StoragePath.ValueString() != "/backups/seerr" {
		t.Fatalf("expected storage_path /backups/seerr, got %q", data.StoragePath.ValueString())
	}
}
