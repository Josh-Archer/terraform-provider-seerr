package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Josh-Archer/terraform-provider-seerr/internal/schemacmp"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type resourceAttributeContract struct {
	kind               string
	required           bool
	optional           bool
	computed           bool
	sensitive          bool
	useStateForUnknown bool
}

var (
	requiredString                = resourceAttributeContract{kind: "string", required: true}
	requiredInt64                 = resourceAttributeContract{kind: "int64", required: true}
	requiredSensitiveString       = resourceAttributeContract{kind: "string", required: true, sensitive: true}
	requiredSet                   = resourceAttributeContract{kind: "set", required: true}
	optionalString                = resourceAttributeContract{kind: "string", optional: true}
	optionalMap                   = resourceAttributeContract{kind: "map", optional: true}
	optionalSensitiveString       = resourceAttributeContract{kind: "string", optional: true, sensitive: true}
	optionalComputedString        = resourceAttributeContract{kind: "string", optional: true, computed: true}
	optionalComputedInt64         = resourceAttributeContract{kind: "int64", optional: true, computed: true}
	optionalComputedBool          = resourceAttributeContract{kind: "bool", optional: true, computed: true}
	optionalComputedList          = resourceAttributeContract{kind: "list", optional: true, computed: true}
	computedString                = resourceAttributeContract{kind: "string", computed: true}
	computedInt64                 = resourceAttributeContract{kind: "int64", computed: true}
	optionalComputedBoolKeepState = resourceAttributeContract{
		kind:               "bool",
		optional:           true,
		computed:           true,
		useStateForUnknown: true,
	}
	computedListNestedKeepState = resourceAttributeContract{
		kind:               "list_nested",
		computed:           true,
		useStateForUnknown: true,
	}
)

func assertResourceSchemaContract(
	t *testing.T,
	r resource.Resource,
	wantAttributes map[string]resourceAttributeContract,
	wantBlocks map[string][]string,
) {
	t.Helper()

	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() returned diagnostics errors: %v", resp.Diagnostics)
	}

	if len(resp.Schema.Attributes) != len(wantAttributes) {
		t.Errorf("Schema() returned %d attributes; want %d", len(resp.Schema.Attributes), len(wantAttributes))
	}
	for name, want := range wantAttributes {
		got, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Errorf("Schema() missing attribute %q", name)
			continue
		}

		if kind := resourceAttributeKind(got); kind != want.kind {
			t.Errorf("attribute %q kind = %q; want %q", name, kind, want.kind)
		}
		if got.IsRequired() != want.required {
			t.Errorf("attribute %q required = %t; want %t", name, got.IsRequired(), want.required)
		}
		if got.IsOptional() != want.optional {
			t.Errorf("attribute %q optional = %t; want %t", name, got.IsOptional(), want.optional)
		}
		if got.IsComputed() != want.computed {
			t.Errorf("attribute %q computed = %t; want %t", name, got.IsComputed(), want.computed)
		}
		if got.IsSensitive() != want.sensitive {
			t.Errorf("attribute %q sensitive = %t; want %t", name, got.IsSensitive(), want.sensitive)
		}
		if want.useStateForUnknown && !attributeKeepsStateForUnknown(got) {
			t.Errorf("attribute %q missing UseStateForUnknown plan modifier", name)
		}
	}

	if len(resp.Schema.Blocks) != len(wantBlocks) {
		t.Errorf("Schema() returned %d blocks; want %d", len(resp.Schema.Blocks), len(wantBlocks))
	}
	for name, wantNestedAttributes := range wantBlocks {
		got, ok := resp.Schema.Blocks[name]
		if !ok {
			t.Errorf("Schema() missing block %q", name)
			continue
		}

		single, ok := got.(schema.SingleNestedBlock)
		if !ok {
			t.Errorf("block %q type = %T; want schema.SingleNestedBlock", name, got)
			continue
		}
		if len(single.Attributes) != len(wantNestedAttributes) {
			t.Errorf("block %q returned %d attributes; want %d", name, len(single.Attributes), len(wantNestedAttributes))
		}
		for _, nestedName := range wantNestedAttributes {
			if _, ok := single.Attributes[nestedName]; !ok {
				t.Errorf("block %q missing attribute %q", name, nestedName)
			}
		}
	}
}

func resourceAttributeKind(attribute schema.Attribute) string {
	switch attribute.(type) {
	case schema.BoolAttribute:
		return "bool"
	case schema.Int64Attribute:
		return "int64"
	case schema.ListAttribute:
		return "list"
	case schema.ListNestedAttribute:
		return "list_nested"
	case schema.MapAttribute:
		return "map"
	case schema.SetAttribute:
		return "set"
	case schema.StringAttribute:
		return "string"
	default:
		return fmt.Sprintf("%T", attribute)
	}
}

func attributeKeepsStateForUnknown(attr schema.Attribute) bool {
	ctx := context.Background()
	var descriptions []string
	switch a := attr.(type) {
	case schema.BoolAttribute:
		for _, m := range a.PlanModifiers {
			descriptions = append(descriptions, m.Description(ctx))
		}
	case schema.Int64Attribute:
		for _, m := range a.PlanModifiers {
			descriptions = append(descriptions, m.Description(ctx))
		}
	case schema.ListAttribute:
		for _, m := range a.PlanModifiers {
			descriptions = append(descriptions, m.Description(ctx))
		}
	case schema.ListNestedAttribute:
		for _, m := range a.PlanModifiers {
			descriptions = append(descriptions, m.Description(ctx))
		}
	case schema.SetAttribute:
		for _, m := range a.PlanModifiers {
			descriptions = append(descriptions, m.Description(ctx))
		}
	case schema.StringAttribute:
		for _, m := range a.PlanModifiers {
			descriptions = append(descriptions, m.Description(ctx))
		}
	}
	for _, d := range descriptions {
		if strings.Contains(d, "state will not change") {
			return true
		}
	}
	return false
}

func TestReleaseToReleaseSchemaContract(t *testing.T) {
	candProvider := New("candidate")()
	candSchema, err := schemacmp.NormalizeProvider(candProvider)
	if err != nil {
		t.Fatalf("failed to normalize candidate provider schema: %v", err)
	}

	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("failed to locate repo root: %v", err)
	}
	baselineFile := filepath.Join(repoRoot, "testdata", "provider_schema_baseline.json")
	if custom := os.Getenv("SEERR_SCHEMA_BASELINE_FILE"); custom != "" {
		baselineFile = custom
	}

	resolveOpts := schemacmp.ResolveOptions{
		BaselineFile: baselineFile,
		AllowOffline: true,
	}

	prevSchema, prevVer, err := schemacmp.ResolveBaseline(context.Background(), resolveOpts)
	if err != nil {
		if errors.Is(err, schemacmp.ErrInitialBaseline) {
			t.Log("Initial baseline release: no previous release schema found; skipping comparison.")
			return
		}
		if errors.Is(err, schemacmp.ErrRegistryUnavailable) {
			t.Skipf("Public registry unavailable and no baseline file found (%s); skipping release comparison.", baselineFile)
			return
		}
		t.Fatalf("failed to resolve baseline provider schema: %v", err)
	}

	cmpOpts := schemacmp.CompareOptions{
		PreviousVersion:        prevVer,
		TargetVersion:          os.Getenv("SEERR_SCHEMA_TARGET_VERSION"),
		MigrationNotePath:      os.Getenv("SEERR_SCHEMA_MIGRATION_NOTE"),
		AllowBreaking:          os.Getenv("SEERR_SCHEMA_ALLOW_BREAKING") == "true",
		RequireStabilityPolicy: true,
	}

	report := schemacmp.Compare(prevSchema, candSchema, cmpOpts)
	if !report.IsCompatible() {
		t.Fatalf("Public provider schema release comparison failed (incompatible contract changes detected against %s):\n%s",
			prevVer, schemacmp.GenerateMarkdownReport(report))
	}
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find repo root from %s", dir)
		}
		dir = parent
	}
}
