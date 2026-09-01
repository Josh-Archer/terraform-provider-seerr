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

func TestFormatAPIErrorBody(t *testing.T) {
	t.Run("nested structured errors array", func(t *testing.T) {
		raw := []byte(`{
			"message": "Validation failed",
			"errors": [
				{"field": "apiKey", "message": "Invalid API key"},
				{"field": "port", "message": "Port must be positive"}
			]
		}`)
		got := formatAPIErrorBody(raw)
		want := "Validation failed\n• apiKey: Invalid API key\n• port: Port must be positive"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("nested string errors array", func(t *testing.T) {
		raw := []byte(`{
			"message": "Validation failed",
			"errors": ["First error", "Second error"]
		}`)
		got := formatAPIErrorBody(raw)
		want := "Validation failed\n• First error\n• Second error"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("single error message", func(t *testing.T) {
		raw := []byte(`{"message": "Not found"}`)
		got := formatAPIErrorBody(raw)
		if got != "Not found" {
			t.Fatalf("got %q, want Not found", got)
		}
	})

	t.Run("plain text fallback", func(t *testing.T) {
		raw := []byte(`Internal Server Error`)
		got := formatAPIErrorBody(raw)
		if got != "Internal Server Error" {
			t.Fatalf("got %q, want Internal Server Error", got)
		}
	})
}
