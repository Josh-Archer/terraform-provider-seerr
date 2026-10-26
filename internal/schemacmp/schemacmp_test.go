package schemacmp

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sampleBaseSchema() *NormalizedSchema {
	return &NormalizedSchema{
		FormatVersion: "1.0",
		Provider: &NormalizedBlock{
			Attributes: map[string]*NormalizedAttribute{
				"url":     {Type: "string", Optional: true},
				"api_key": {Type: "string", Optional: true, Sensitive: true},
			},
		},
		Resources: map[string]*NormalizedBlock{
			"seerr_user": {
				Attributes: map[string]*NormalizedAttribute{
					"id":          {Type: "string", Computed: true},
					"email":       {Type: "string", Required: true},
					"username":    {Type: "string", Optional: true, Computed: true},
					"permissions": {Type: "number", Optional: true, Computed: true},
					"password":    {Type: "string", Optional: true, Sensitive: true},
				},
				Blocks: map[string]*NormalizedBlockType{
					"quota": {
						NestingMode: "single",
						MinItems:    0,
						Block: &NormalizedBlock{
							Attributes: map[string]*NormalizedAttribute{
								"movie_quota_limit": {Type: "number", Optional: true},
							},
						},
					},
				},
			},
		},
		DataSources: map[string]*NormalizedBlock{
			"seerr_about": {
				Attributes: map[string]*NormalizedAttribute{
					"id":      {Type: "string", Computed: true},
					"version": {Type: "string", Computed: true},
				},
			},
		},
	}
}

func TestCompatibleAdditions(t *testing.T) {
	prev := sampleBaseSchema()

	// Candidate adds new optional attribute, new resource, and new data source
	cand := sampleBaseSchema()
	cand.Resources["seerr_user"].Attributes["avatar"] = &NormalizedAttribute{
		Type:     "string",
		Optional: true,
	}
	cand.Resources["seerr_new_res"] = &NormalizedBlock{
		Attributes: map[string]*NormalizedAttribute{
			"id":   {Type: "string", Computed: true},
			"name": {Type: "string", Required: true},
		},
	}
	cand.DataSources["seerr_new_ds"] = &NormalizedBlock{
		Attributes: map[string]*NormalizedAttribute{
			"id": {Type: "string", Computed: true},
		},
	}

	report := Compare(prev, cand, CompareOptions{
		PreviousVersion: "v2.2.0",
		TargetVersion:   "v2.3.0",
	})

	if report.HasBreakingChanges {
		t.Fatalf("expected no breaking changes, got: %v", report.BreakingChanges)
	}
	if !report.IsCompatible() {
		t.Fatalf("expected candidate to be compatible")
	}
	if len(report.CompatibleChanges) != 3 {
		t.Fatalf("expected 3 compatible changes, got %d: %v", len(report.CompatibleChanges), report.CompatibleChanges)
	}
}

func TestIncompatibleRemovals(t *testing.T) {
	t.Run("remove attribute", func(t *testing.T) {
		prev := sampleBaseSchema()
		cand := sampleBaseSchema()
		delete(cand.Resources["seerr_user"].Attributes, "email")

		report := Compare(prev, cand, CompareOptions{
			PreviousVersion: "v2.2.0",
			TargetVersion:   "v2.2.1",
		})

		if !report.HasBreakingChanges {
			t.Fatal("expected breaking change for removed attribute")
		}
		if report.IsCompatible() {
			t.Fatal("expected report to not be compatible")
		}
		found := false
		for _, c := range report.BreakingChanges {
			if c.Type == ChangeAttributeRemoved && c.Path == "email" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("did not find ChangeAttributeRemoved for email: %v", report.BreakingChanges)
		}
	})

	t.Run("remove resource", func(t *testing.T) {
		prev := sampleBaseSchema()
		cand := sampleBaseSchema()
		delete(cand.Resources, "seerr_user")

		report := Compare(prev, cand, CompareOptions{
			PreviousVersion: "v2.2.0",
			TargetVersion:   "v2.2.1",
		})

		if !report.HasBreakingChanges {
			t.Fatal("expected breaking change for removed resource")
		}
		found := false
		for _, c := range report.BreakingChanges {
			if c.Type == ChangeResourceRemoved && c.TargetName == "seerr_user" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("did not find ChangeResourceRemoved for seerr_user: %v", report.BreakingChanges)
		}
	})

	t.Run("remove data source", func(t *testing.T) {
		prev := sampleBaseSchema()
		cand := sampleBaseSchema()
		delete(cand.DataSources, "seerr_about")

		report := Compare(prev, cand, CompareOptions{
			PreviousVersion: "v2.2.0",
			TargetVersion:   "v2.2.1",
		})

		if !report.HasBreakingChanges {
			t.Fatal("expected breaking change for removed data source")
		}
	})
}

func TestIncompatibleTypeChanges(t *testing.T) {
	prev := sampleBaseSchema()
	cand := sampleBaseSchema()
	// Change permissions from number to string
	cand.Resources["seerr_user"].Attributes["permissions"] = &NormalizedAttribute{
		Type:     "string",
		Optional: true,
		Computed: true,
	}

	report := Compare(prev, cand, CompareOptions{
		PreviousVersion: "v2.2.0",
		TargetVersion:   "v2.3.0",
	})

	if !report.HasBreakingChanges {
		t.Fatal("expected breaking change for altered attribute type")
	}
	found := false
	for _, c := range report.BreakingChanges {
		if c.Type == ChangeTypeModified && c.Path == "permissions" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("did not find ChangeTypeModified for permissions: %v", report.BreakingChanges)
	}
}

func TestNewlyRequiredAttributes(t *testing.T) {
	t.Run("existing attribute becomes required", func(t *testing.T) {
		prev := sampleBaseSchema()
		cand := sampleBaseSchema()
		cand.Resources["seerr_user"].Attributes["username"] = &NormalizedAttribute{
			Type:     "string",
			Required: true,
		}

		report := Compare(prev, cand, CompareOptions{
			PreviousVersion: "v2.2.0",
			TargetVersion:   "v2.3.0",
		})

		if !report.HasBreakingChanges {
			t.Fatal("expected breaking change when attribute becomes required")
		}
		found := false
		for _, c := range report.BreakingChanges {
			if c.Type == ChangeNewlyRequired && c.Path == "username" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("did not find ChangeNewlyRequired for username: %v", report.BreakingChanges)
		}
	})

	t.Run("new attribute added as required", func(t *testing.T) {
		prev := sampleBaseSchema()
		cand := sampleBaseSchema()
		cand.Resources["seerr_user"].Attributes["new_mandatory_field"] = &NormalizedAttribute{
			Type:     "string",
			Required: true,
		}

		report := Compare(prev, cand, CompareOptions{
			PreviousVersion: "v2.2.0",
			TargetVersion:   "v2.3.0",
		})

		if !report.HasBreakingChanges {
			t.Fatal("expected breaking change when new required attribute is added")
		}
	})
}

func TestLostSensitiveMarking(t *testing.T) {
	prev := sampleBaseSchema()
	cand := sampleBaseSchema()
	// password loses sensitive marking
	cand.Resources["seerr_user"].Attributes["password"] = &NormalizedAttribute{
		Type:      "string",
		Optional:  true,
		Sensitive: false,
	}

	report := Compare(prev, cand, CompareOptions{
		PreviousVersion: "v2.2.0",
		TargetVersion:   "v2.3.0",
	})

	if !report.HasBreakingChanges {
		t.Fatal("expected breaking change when sensitive marking is lost")
	}
	found := false
	for _, c := range report.BreakingChanges {
		if c.Type == ChangeLostSensitive && c.Path == "password" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("did not find ChangeLostSensitive for password: %v", report.BreakingChanges)
	}
}

func TestInitialBaselineHandling(t *testing.T) {
	cand := sampleBaseSchema()
	report := Compare(nil, cand, CompareOptions{
		TargetVersion: "v1.0.0",
	})

	if !report.IsInitialBaseline {
		t.Fatal("expected report to be marked as initial baseline")
	}
	if report.HasBreakingChanges {
		t.Fatal("initial baseline should not report breaking changes")
	}
	if !report.IsCompatible() {
		t.Fatal("initial baseline must be deemed compatible")
	}

	md := GenerateMarkdownReport(report)
	if !strings.Contains(md, "Initial Baseline Release") {
		t.Fatalf("markdown report missing Initial Baseline banner: %s", md)
	}
}

func TestUnavailableRegistryHandling(t *testing.T) {
	tmpDir := t.TempDir()
	opts := ResolveOptions{
		PreviousVersion: "99.99.99",
		RegistryURL:     "invalid.registry.local/nonexistent/seerr",
		WorkDir:         tmpDir,
		AllowOffline:    true,
	}

	schema, ver, err := ResolveBaseline(context.Background(), opts)
	if schema != nil {
		t.Fatal("expected nil schema when registry is unreachable")
	}
	if ver != "99.99.99" {
		t.Fatalf("expected version 99.99.99, got %s", ver)
	}
	if !errors.Is(err, ErrRegistryUnavailable) {
		t.Fatalf("expected ErrRegistryUnavailable, got: %v", err)
	}
}

func TestStabilityPolicyValidation(t *testing.T) {
	prev := sampleBaseSchema()
	cand := sampleBaseSchema()
	delete(cand.Resources["seerr_user"].Attributes, "email") // breaking change

	t.Run("minor bump without migration note fails", func(t *testing.T) {
		report := Compare(prev, cand, CompareOptions{
			PreviousVersion: "v2.2.0",
			TargetVersion:   "v2.3.0",
		})
		if report.PolicyPassed {
			t.Fatal("expected stability policy to fail for minor bump with breaking changes")
		}
		if len(report.PolicyErrors) != 2 { // version error + note error
			t.Fatalf("expected 2 policy errors, got %d: %v", len(report.PolicyErrors), report.PolicyErrors)
		}
	})

	t.Run("major bump without migration note fails", func(t *testing.T) {
		report := Compare(prev, cand, CompareOptions{
			PreviousVersion: "v2.2.0",
			TargetVersion:   "v3.0.0",
		})
		if report.PolicyPassed {
			t.Fatal("expected policy to fail when migration note is missing")
		}
		if len(report.PolicyErrors) != 1 || !strings.Contains(report.PolicyErrors[0], "migration note") {
			t.Fatalf("expected migration note error, got: %v", report.PolicyErrors)
		}
	})

	t.Run("major bump with explicit migration note passes", func(t *testing.T) {
		tmpDir := t.TempDir()
		noteFile := filepath.Join(tmpDir, "MIGRATION.md")
		if err := os.WriteFile(noteFile, []byte("Email removed. Use username."), 0600); err != nil {
			t.Fatal(err)
		}

		report := Compare(prev, cand, CompareOptions{
			PreviousVersion:   "v2.2.0",
			TargetVersion:     "v3.0.0",
			MigrationNotePath: noteFile,
			AllowBreaking:     true,
		})
		if !report.PolicyPassed {
			t.Fatalf("expected policy to pass with major bump and note: %v", report.PolicyErrors)
		}
		if !report.IsCompatible() {
			t.Fatal("expected report to be approved compatible per policy")
		}
	})
}

func TestDocumentationAndOrderChangesIgnored(t *testing.T) {
	// Raw JSON dumps with identical schema definitions but differing documentation descriptions
	json1 := `{
		"format_version": "1.0",
		"provider_schemas": {
			"registry.opentofu.org/josh-archer/seerr": {
				"resource_schemas": {
					"seerr_user": {
						"block": {
							"description": "Original documentation",
							"description_kind": "markdown",
							"attributes": {
								"a": {"type": "string", "optional": true, "description": "Doc A"},
								"b": {"type": "number", "required": true, "description": "Doc B"}
							}
						}
					}
				}
			}
		}
	}`

	// json2 swaps order of attributes and changes all descriptions
	json2 := `{
		"format_version": "1.0",
		"provider_schemas": {
			"registry.opentofu.org/josh-archer/seerr": {
				"resource_schemas": {
					"seerr_user": {
						"block": {
							"description": "Completely updated description and guide",
							"description_kind": "plain",
							"attributes": {
								"b": {"type": "number", "required": true, "description": "Revamped Doc B"},
								"a": {"type": "string", "optional": true, "description": "Revamped Doc A"}
							}
						}
					}
				}
			}
		}
	}`

	schema1, err := NormalizeTFJSON([]byte(json1))
	if err != nil {
		t.Fatal(err)
	}
	schema2, err := NormalizeTFJSON([]byte(json2))
	if err != nil {
		t.Fatal(err)
	}

	report := Compare(schema1, schema2, CompareOptions{
		PreviousVersion: "v2.2.0",
		TargetVersion:   "v2.2.1",
	})

	if report.HasBreakingChanges {
		t.Fatalf("expected no breaking changes, got: %v", report.BreakingChanges)
	}
	if len(report.CompatibleChanges) > 0 {
		t.Fatalf("expected no compatible changes for doc edits, got: %v", report.CompatibleChanges)
	}
	if !report.IsCompatible() {
		t.Fatal("expected doc-only edits to be completely compatible")
	}
}

func TestMarkdownReportContent(t *testing.T) {
	prev := sampleBaseSchema()
	cand := sampleBaseSchema()
	delete(cand.Resources["seerr_user"].Attributes, "email")
	cand.Resources["seerr_user"].Attributes["new_opt"] = &NormalizedAttribute{
		Type:     "string",
		Optional: true,
	}

	report := Compare(prev, cand, CompareOptions{
		PreviousVersion: "v2.2.0",
		TargetVersion:   "v2.3.0",
	})

	md := GenerateMarkdownReport(report)

	// Check for required elements
	if !strings.Contains(md, "Public Provider Schema Compatibility Report") {
		t.Fatal("missing title")
	}
	if !strings.Contains(md, "Incompatible Breaking Changes") {
		t.Fatal("missing breaking changes section")
	}
	if !strings.Contains(md, "Backward-Compatible Additions") {
		t.Fatal("missing compatible additions section")
	}
	if !strings.Contains(md, "Stability Policy Requirements") {
		t.Fatal("missing stability policy requirements")
	}
	if !strings.Contains(md, "Semantic Review") || !strings.Contains(md, "Consumer Upgrade Replay") {
		t.Fatal("missing semantic review / upgrade replay complementary notice")
	}
}
