package openapi

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenAPICoverageClassification(t *testing.T) {
	root, err := FindRepoRoot()
	if err != nil {
		t.Fatalf("failed to find repo root: %v", err)
	}

	specPath := filepath.Join(root, "tools", "openapi", "seerr-api.yml")
	spec, err := ParseOpenAPISpec(specPath)
	if err != nil {
		t.Fatalf("failed to parse openapi spec at %s: %v", specPath, err)
	}

	report, err := AnalyzeCoverage(spec, DefaultRules())
	if err != nil {
		t.Fatalf("failed to analyze coverage: %v", err)
	}

	if len(report.UnclassifiedPaths) > 0 {
		t.Errorf("Found %d unclassified OpenAPI path(s):\n%v\nAll paths must be classified in DefaultRules() as 'covered', 'intentionally-out-of-scope', or 'uncovered'.",
			len(report.UnclassifiedPaths), report.UnclassifiedPaths)
	}

	t.Logf("OpenAPI Coverage Summary: Total Paths=%d, Covered=%d, Intentionally Out of Scope=%d, Uncovered=%d",
		report.TotalPaths, report.CoveredPaths, report.IntentionallyOutOfScopePaths, report.UncoveredPaths)

	// Write / update docs/openapi-coverage.md
	docContent := GenerateMarkdownReport(report)
	docPath := filepath.Join(root, "docs", "openapi-coverage.md")
	if err := os.WriteFile(docPath, []byte(docContent), 0644); err != nil {
		t.Fatalf("failed to write openapi-coverage.md: %v", err)
	}
}
