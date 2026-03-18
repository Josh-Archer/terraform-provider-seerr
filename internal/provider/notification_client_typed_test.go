package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestNotificationClientResourceMetadataAndSchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		agent       string
		expectedKey string
		absentKey   string
	}{
		{agent: "ntfy", expectedKey: "ntfy", absentKey: "pushover"},
		{agent: "pushover", expectedKey: "pushover", absentKey: "ntfy"},
	}

	for _, tt := range tests {
		t.Run(tt.agent, func(t *testing.T) {
			t.Parallel()

			r := &NotificationClientResource{agent: tt.agent}
			metaResp := &resource.MetadataResponse{}
			r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "seerr"}, metaResp)

			if got, want := metaResp.TypeName, "seerr_notification_"+tt.agent; got != want {
				t.Fatalf("Metadata type name mismatch: got %q want %q", got, want)
			}

			schemaResp := &resource.SchemaResponse{}
			r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)
			if schemaResp.Diagnostics.HasError() {
				t.Fatalf("Schema returned diagnostics: %v", schemaResp.Diagnostics)
			}

			assertResourceSchemaHasKey(t, schemaResp.Schema, tt.expectedKey)
			assertResourceSchemaLacksKey(t, schemaResp.Schema, tt.absentKey)
			assertResourceSchemaLacksKey(t, schemaResp.Schema, "agent")
			assertResourceSchemaHasKey(t, schemaResp.Schema, "enabled")
			assertResourceSchemaHasKey(t, schemaResp.Schema, "notification_types")
			assertResourceSchemaLacksKey(t, schemaResp.Schema, "types")
		})
	}
}

func TestNotificationClientDataSourceMetadataAndSchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		agent       string
		expectedKey string
		absentKey   string
	}{
		{agent: "ntfy", expectedKey: "ntfy", absentKey: "pushover"},
		{agent: "pushover", expectedKey: "pushover", absentKey: "ntfy"},
	}

	for _, tt := range tests {
		t.Run(tt.agent, func(t *testing.T) {
			t.Parallel()

			d := &NotificationClientDataSource{agent: tt.agent}
			metaResp := &datasource.MetadataResponse{}
			d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "seerr"}, metaResp)

			if got, want := metaResp.TypeName, "seerr_notification_"+tt.agent; got != want {
				t.Fatalf("Metadata type name mismatch: got %q want %q", got, want)
			}

			schemaResp := &datasource.SchemaResponse{}
			d.Schema(context.Background(), datasource.SchemaRequest{}, schemaResp)
			if schemaResp.Diagnostics.HasError() {
				t.Fatalf("Schema returned diagnostics: %v", schemaResp.Diagnostics)
			}

			assertDataSourceSchemaHasKey(t, schemaResp.Schema, tt.expectedKey)
			assertDataSourceSchemaLacksKey(t, schemaResp.Schema, tt.absentKey)
			assertDataSourceSchemaLacksKey(t, schemaResp.Schema, "agent")
			assertDataSourceSchemaHasKey(t, schemaResp.Schema, "enabled")
			assertDataSourceSchemaHasKey(t, schemaResp.Schema, "notification_types")
		})
	}
}

func assertResourceSchemaHasKey(t *testing.T, s rschema.Schema, key string) {
	t.Helper()
	if _, ok := s.Attributes[key]; !ok {
		t.Fatalf("expected resource schema to contain %q", key)
	}
}

func assertResourceSchemaLacksKey(t *testing.T, s rschema.Schema, key string) {
	t.Helper()
	if _, ok := s.Attributes[key]; ok {
		t.Fatalf("expected resource schema to omit %q", key)
	}
}

func assertResourceInt64AttributeHasNoDefault(t *testing.T, s rschema.Schema, key string) {
	t.Helper()
	attr, ok := s.Attributes[key]
	if !ok {
		t.Fatalf("expected resource schema to contain %q", key)
	}

	intAttr, ok := attr.(rschema.Int64Attribute)
	if !ok {
		t.Fatalf("expected %q to be an Int64Attribute, got %T", key, attr)
	}
	if intAttr.Default != nil {
		t.Fatalf("expected %q to have no default", key)
	}
}

func assertDataSourceSchemaHasKey(t *testing.T, s dschema.Schema, key string) {
	t.Helper()
	if _, ok := s.Attributes[key]; !ok {
		t.Fatalf("expected data source schema to contain %q", key)
	}
}

func assertDataSourceSchemaLacksKey(t *testing.T, s dschema.Schema, key string) {
	t.Helper()
	if _, ok := s.Attributes[key]; ok {
		t.Fatalf("expected data source schema to omit %q", key)
	}
}
