package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenAPIDiffLocalIdentity(t *testing.T) {
	root, err := FindRepoRoot()
	if err != nil {
		t.Fatalf("failed to find repo root: %v", err)
	}

	specPath := filepath.Join(root, "tools", "openapi", "seerr-api.yml")
	baseSpec, err := LoadSpec(specPath)
	if err != nil {
		t.Fatalf("failed to load base spec %s: %v", specPath, err)
	}

	// Comparing spec against itself should result in 0 drift
	report := CompareSpecs(baseSpec, baseSpec, specPath, specPath)
	if report.HasDrift {
		t.Errorf("expected no drift when comparing spec to itself, got drift: %+v", report)
	}
	if len(report.AddedEndpoints) != 0 || len(report.RemovedEndpoints) != 0 {
		t.Errorf("expected 0 added and 0 removed endpoints, got added=%d removed=%d",
			len(report.AddedEndpoints), len(report.RemovedEndpoints))
	}
	if report.TotalBaseEndpoints != report.TotalUpstreamEndpoints {
		t.Errorf("expected endpoint counts to match, base=%d upstream=%d",
			report.TotalBaseEndpoints, report.TotalUpstreamEndpoints)
	}
}

func TestOpenAPIDiffSyntheticDrift(t *testing.T) {
	baseYAML := `
openapi: 3.0.0
paths:
  /settings/main:
    get:
      summary: Get main settings
      deprecated: false
    post:
      summary: Update main settings
  /user:
    get:
      summary: List users
`

	upstreamYAML := `
openapi: 3.0.0
paths:
  /settings/main:
    get:
      summary: Get main settings
      deprecated: true
  /user:
    get:
      summary: List users
  /settings/ai:
    get:
      summary: New AI feature
      operationId: getAISettings
`

	baseTmp := filepath.Join(t.TempDir(), "base.yml")
	upstreamTmp := filepath.Join(t.TempDir(), "upstream.yml")

	if err := os.WriteFile(baseTmp, []byte(baseYAML), 0644); err != nil {
		t.Fatalf("write base tmp: %v", err)
	}
	if err := os.WriteFile(upstreamTmp, []byte(upstreamYAML), 0644); err != nil {
		t.Fatalf("write upstream tmp: %v", err)
	}

	baseSpec, err := LoadSpec(baseTmp)
	if err != nil {
		t.Fatalf("load base spec: %v", err)
	}
	upstreamSpec, err := LoadSpec(upstreamTmp)
	if err != nil {
		t.Fatalf("load upstream spec: %v", err)
	}

	report := CompareSpecs(baseSpec, upstreamSpec, baseTmp, upstreamTmp)
	if !report.HasDrift {
		t.Errorf("expected drift to be detected")
	}
	if len(report.AddedEndpoints) != 1 || report.AddedEndpoints[0].Path != "/settings/ai" {
		t.Errorf("expected 1 added endpoint at /settings/ai, got: %+v", report.AddedEndpoints)
	}
	if len(report.RemovedEndpoints) != 1 || report.RemovedEndpoints[0].Path != "/settings/main" || report.RemovedEndpoints[0].Method != "POST" {
		t.Errorf("expected 1 removed endpoint (POST /settings/main), got: %+v", report.RemovedEndpoints)
	}
	if len(report.DeprecatedEndpoints) != 1 || report.DeprecatedEndpoints[0].Path != "/settings/main" {
		t.Errorf("expected 1 deprecated endpoint, got: %+v", report.DeprecatedEndpoints)
	}

	md := report.FormatMarkdown()
	if len(md) == 0 {
		t.Errorf("markdown report should not be empty")
	}
}
