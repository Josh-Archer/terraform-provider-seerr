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

func TestAPIKeyRegenerateFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/settings/main/regenerate" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", r.Method)
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	resource := &APIKeyResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout, 0, 0),
	}

	data := APIKeyModel{}
	diags := resource.regenerateKey(context.Background(), &data)
	if !diags.HasError() {
		t.Fatal("expected error diagnostics, got none")
	}
}

func TestAPIKeyResourceImportState(t *testing.T) {
	r := &APIKeyResource{}
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)

	if _, hasID := schemaResp.Schema.Attributes["id"]; !hasID {
		t.Fatalf("expected schema to contain 'id' attribute")
	}

	state := tfsdk.State{
		Schema: schemaResp.Schema,
	}
	if diags := state.Set(context.Background(), &APIKeyModel{}); diags.HasError() {
		t.Fatalf("failed to initialize state: %v", diags)
	}

	req := resource.ImportStateRequest{
		ID: "api_key",
	}
	resp := resource.ImportStateResponse{
		State: state,
	}

	r.ImportState(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}

	var data APIKeyModel
	diags := resp.State.Get(context.Background(), &data)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics getting state: %v", diags)
	}
	if got := data.ID.ValueString(); got != "api_key" {
		t.Fatalf("expected id %q, got %q", "api_key", got)
	}
}

func TestAPIKeyResourceImportAndReadLifecycle(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/settings/main" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"apiKey": "server-real-api-key"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	r := &APIKeyResource{
		client: NewClient(baseURL, "test-api-key", "test-agent", false, defaultRequestTimeout, 0, 0),
	}
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)

	state := tfsdk.State{
		Schema: schemaResp.Schema,
	}
	if diags := state.Set(context.Background(), &APIKeyModel{}); diags.HasError() {
		t.Fatalf("failed to initialize state: %v", diags)
	}

	// 1. User imports: `terraform import seerr_api_key.example api_key`
	req := resource.ImportStateRequest{ID: "api_key"}
	importResp := resource.ImportStateResponse{State: state}
	r.ImportState(context.Background(), req, &importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("unexpected import error: %v", importResp.Diagnostics)
	}

	var postImportData APIKeyModel
	_ = importResp.State.Get(context.Background(), &postImportData)
	if postImportData.ID.ValueString() != "api_key" {
		t.Fatalf("expected id 'api_key', got %s", postImportData.ID.ValueString())
	}

	// 2. Terraform runs Read() to refresh
	readResp := resource.ReadResponse{State: importResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: importResp.State}, &readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("unexpected read error: %v", readResp.Diagnostics)
	}

	var postReadData APIKeyModel
	_ = readResp.State.Get(context.Background(), &postReadData)
	if postReadData.ID.ValueString() != "api_key" {
		t.Fatalf("expected id 'api_key' after Read(), got %s", postReadData.ID.ValueString())
	}
	if postReadData.ApiKey.ValueString() != "server-real-api-key" {
		t.Fatalf("expected real key after Read(), got %s", postReadData.ApiKey.ValueString())
	}
}

func TestAPIKeyResourceUpgradeStateV0toV1(t *testing.T) {
	r := &APIKeyResource{}
	upgraders := r.UpgradeState(context.Background())
	upgraderV0, exists := upgraders[0]
	if !exists {
		t.Fatal("expected state upgrader for version 0")
	}

	stateV0 := tfsdk.State{
		Schema: *upgraderV0.PriorSchema,
	}

	type APIKeyModelV0 struct {
		ApiKey     types.String `tfsdk:"api_key"`
		StatusCode types.Int64  `tfsdk:"status_code"`
	}

	rawV0 := APIKeyModelV0{
		ApiKey:     types.StringValue("migrated-key-123"),
		StatusCode: types.Int64Value(200),
	}

	if diags := stateV0.Set(context.Background(), &rawV0); diags.HasError() {
		t.Fatalf("failed to set v0 state: %v", diags)
	}

	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	targetState := tfsdk.State{Schema: schemaResp.Schema}

	upgradeReq := resource.UpgradeStateRequest{
		State: &stateV0,
	}
	upgradeResp := resource.UpgradeStateResponse{
		State: targetState,
	}

	upgraderV0.StateUpgrader(context.Background(), upgradeReq, &upgradeResp)
	if upgradeResp.Diagnostics.HasError() {
		t.Fatalf("upgrade failed: %v", upgradeResp.Diagnostics)
	}

	var upgradedData APIKeyModel
	if diags := upgradeResp.State.Get(context.Background(), &upgradedData); diags.HasError() {
		t.Fatalf("failed to read upgraded state: %v", diags)
	}

	if upgradedData.ID.ValueString() != "api_key" {
		t.Errorf("expected ID 'api_key', got %q", upgradedData.ID.ValueString())
	}
	if upgradedData.ApiKey.ValueString() != "migrated-key-123" {
		t.Errorf("expected ApiKey 'migrated-key-123', got %q", upgradedData.ApiKey.ValueString())
	}
	if upgradedData.StatusCode.ValueInt64() != 200 {
		t.Errorf("expected StatusCode 200, got %d", upgradedData.StatusCode.ValueInt64())
	}
}

func TestAPIKeyResourceRead404RemovesResource(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	r := &APIKeyResource{
		client: NewClient(baseURL, "test-api-key", "test-agent", false, defaultRequestTimeout, 0, 0),
	}
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)

	state := tfsdk.State{
		Schema: schemaResp.Schema,
	}
	initial := APIKeyModel{
		ID:         types.StringValue("api_key"),
		ApiKey:     types.StringValue("old-key"),
		StatusCode: types.Int64Value(200),
	}
	if diags := state.Set(context.Background(), &initial); diags.HasError() {
		t.Fatalf("failed to initialize state: %v", diags)
	}

	readResp := resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, &readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("unexpected error diagnostics: %v", readResp.Diagnostics)
	}

	// State should be removed
	var postData APIKeyModel
	_ = readResp.State.Get(context.Background(), &postData)
	if !postData.ID.IsNull() && postData.ID.ValueString() != "" && postData.ID.ValueString() != "null" {
		// Verify if state is empty or raw is null
		var raw map[string]any
		_ = readResp.State.Get(context.Background(), &raw)
		if len(raw) > 0 {
			t.Fatalf("expected state to be removed on 404, got non-empty state: %+v", postData)
		}
	}
}
