package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestJellyfinSettingsResourceSchema(t *testing.T) {
	t.Parallel()
	r := NewJellyfinSettingsResource()
	var req resource.SchemaRequest
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() returned diagnostics errors: %v", resp.Diagnostics)
	}
	if resp.Schema.Attributes == nil && resp.Schema.Blocks == nil {
		t.Fatal("Schema() returned empty attributes and blocks")
	}
}
