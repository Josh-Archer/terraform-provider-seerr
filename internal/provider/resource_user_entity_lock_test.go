package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// fakeSeerrUserRow emulates how Seerr persists a user: every write handler
// loads the whole user row, does some work, then saves the whole row back.
// Concurrent writes for the same user therefore lose each other's changes.
type fakeSeerrUserRow struct {
	mu  sync.Mutex
	row map[string]any
}

func (f *fakeSeerrUserRow) snapshot() map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make(map[string]any, len(f.row))
	for k, v := range f.row {
		out[k] = v
	}
	return out
}

func (f *fakeSeerrUserRow) save(row map[string]any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.row = row
}

// Regression for the arr-config consumer replay failure: seerr_user_quota and
// seerr_user_permissions for the same user applied in parallel, the quota
// write saved a stale user row, and permissions read back as 2 instead of the
// planned 26 ("Provider produced inconsistent result after apply").
func TestUserPermissionsAndQuotaSerializeSameUserWrites(t *testing.T) {
	fake := &fakeSeerrUserRow{row: map[string]any{"id": float64(1), "permissions": float64(2), "locale": "en"}}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/user/1" && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(fake.snapshot())
		case r.URL.Path == "/api/v1/user/1" && r.Method == http.MethodPut:
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			row := fake.snapshot()
			time.Sleep(50 * time.Millisecond)
			row["permissions"] = body["permissions"]
			fake.save(row)
			_ = json.NewEncoder(w).Encode(row)
		case r.URL.Path == "/api/v1/user/1/settings/main" && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(fake.snapshot())
		case r.URL.Path == "/api/v1/user/1/settings/main" && r.Method == http.MethodPost:
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			row := fake.snapshot()
			time.Sleep(150 * time.Millisecond)
			for _, k := range []string{"movieQuotaLimit", "movieQuotaDays", "tvQuotaLimit", "tvQuotaDays"} {
				row[k] = body[k]
			}
			fake.save(row)
			_ = json.NewEncoder(w).Encode(row)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	baseURL, _ := url.Parse(srv.URL)
	client := NewClient(baseURL, "tok", "agent", false, defaultRequestTimeout, 0, 0)
	ctx := context.Background()

	perms := &UserPermissionsResource{client: client}
	var schemaResp resource.SchemaResponse
	perms.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	if diags := plan.Set(ctx, &UserPermissionsModel{
		ID:           types.StringUnknown(),
		UserID:       types.Int64Value(1),
		Permissions:  types.Int64Value(26),
		ResponseJSON: types.StringUnknown(),
		StatusCode:   types.Int64Unknown(),
	}); diags.HasError() {
		t.Fatalf("plan: %v", diags)
	}

	quota := &UserQuotaResource{client: client}
	quotaModel := &UserQuotaModel{
		UserID:          types.Int64Value(1),
		MovieQuotaLimit: types.Int64Null(),
		MovieQuotaDays:  types.Int64Null(),
		TvQuotaLimit:    types.Int64Null(),
		TvQuotaDays:     types.Int64Null(),
	}

	var wg sync.WaitGroup
	createResp := resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	var quotaErr error
	wg.Add(2)
	go func() {
		defer wg.Done()
		quotaErr = quota.applyQuota(ctx, quotaModel)
	}()
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond) // let the quota write load the row first
		perms.Create(ctx, resource.CreateRequest{Plan: plan}, &createResp)
	}()
	wg.Wait()

	if quotaErr != nil {
		t.Fatalf("applyQuota: %v", quotaErr)
	}
	if createResp.Diagnostics.HasError() {
		t.Fatalf("create: %v", createResp.Diagnostics)
	}
	var got UserPermissionsModel
	if diags := createResp.State.Get(ctx, &got); diags.HasError() {
		t.Fatalf("state: %v", diags)
	}
	if got.Permissions.ValueInt64() != 26 {
		t.Fatalf("state permissions = %d, want 26", got.Permissions.ValueInt64())
	}
	if v := fake.snapshot()["permissions"]; v != float64(26) {
		t.Fatalf("server permissions = %v, want 26 (a concurrent same-user write reverted it)", v)
	}
}
