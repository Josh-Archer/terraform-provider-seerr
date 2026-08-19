package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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

	state := tfsdk.State{
		Schema: schemaResp.Schema,
	}
	if diags := state.Set(context.Background(), &APIKeyModel{}); diags.HasError() {
		t.Fatalf("failed to initialize state: %v", diags)
	}

	req := resource.ImportStateRequest{
		ID: "test-api-key-12345",
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
	if got := data.ApiKey.ValueString(); got != "test-api-key-12345" {
		t.Fatalf("expected api_key %q, got %q", "test-api-key-12345", got)
	}
}
