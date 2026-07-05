package provider

import (
	"testing"

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
