package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSetOptionalBoolOmitsNullAndUnknown(t *testing.T) {
	payload := map[string]any{}

	setOptionalBool(payload, "nullBool", types.BoolNull())
	setOptionalBool(payload, "unknownBool", types.BoolUnknown())

	if _, ok := payload["nullBool"]; ok {
		t.Fatal("null bool should be omitted from payload")
	}
	if _, ok := payload["unknownBool"]; ok {
		t.Fatal("unknown bool should be omitted from payload")
	}
}

func TestSetOptionalBoolIncludesKnownValue(t *testing.T) {
	payload := map[string]any{}

	setOptionalBool(payload, "enabled", types.BoolValue(false))

	got, ok := payload["enabled"]
	if !ok {
		t.Fatal("known bool should be included in payload")
	}
	if got != false {
		t.Fatalf("enabled: got %v, want false", got)
	}
}

func TestHandleAPIResponseErrors(t *testing.T) {
	tests := []struct {
		name   string
		resp   *APIResponse
		detail string
	}{
		{name: "nil response", resp: nil, detail: "no response"},
		{name: "message body", resp: &APIResponse{StatusCode: 400, Body: []byte(`{"message":"bad input"}`)}, detail: "bad input"},
		{name: "error body", resp: &APIResponse{StatusCode: 403, Body: []byte(`{"error":"forbidden"}`)}, detail: "forbidden"},
		{name: "raw body", resp: &APIResponse{StatusCode: 500, Body: []byte(`upstream unavailable`)}, detail: "upstream unavailable"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var diags diag.Diagnostics
			if HandleAPIResponse(context.Background(), test.resp, &diags, "Read") {
				t.Fatal("expected error response to be rejected")
			}
			if !diags.HasError() || !strings.Contains(diags.Errors()[0].Detail(), test.detail) {
				t.Fatalf("expected diagnostic containing %q, got %v", test.detail, diags)
			}
		})
	}
}

func TestHandleAPIResponseAcceptsSuccess(t *testing.T) {
	var diags diag.Diagnostics
	if !HandleAPIResponse(context.Background(), &APIResponse{StatusCode: 204}, &diags, "Delete") {
		t.Fatalf("expected 204 response to succeed, got %v", diags)
	}
}
