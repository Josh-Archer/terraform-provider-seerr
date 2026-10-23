package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Josh-Archer/terraform-provider-seerr/internal/provider"
	"github.com/Josh-Archer/terraform-provider-seerr/internal/schemacmp"
)

func main() {
	var (
		candidatePath        string
		previousVersion      string
		baselineFilePath     string
		reportPath           string
		outCandidateSchema   string
		outPreviousSchema    string
		exportBaselineOnly   string
		targetVersion        string
		migrationNotePath    string
		migrationNoteContent string
		allowBreaking        bool
		allowOffline         bool
		quiet                bool
	)

	flag.StringVar(&candidatePath, "candidate", "", "Path to candidate schema JSON (if omitted, extracted from provider in Go)")
	flag.StringVar(&previousVersion, "previous", "", "Baseline version tag or number (e.g. '2.2.2' or 'latest')")
	flag.StringVar(&baselineFilePath, "baseline-file", "", "Path to explicit baseline schema JSON file")
	flag.StringVar(&reportPath, "report", "", "Path to write output markdown PR report artifact")
	flag.StringVar(&outCandidateSchema, "output-candidate-schema", "", "Path to write normalized candidate schema JSON")
	flag.StringVar(&outPreviousSchema, "output-previous-schema", "", "Path to write normalized previous schema JSON")
	flag.StringVar(&exportBaselineOnly, "export-baseline-only", "", "Export baseline schema to given path and exit")
	flag.StringVar(&targetVersion, "target-version", "", "Target candidate release version (e.g. 'v2.3.0')")
	flag.StringVar(&migrationNotePath, "migration-note", "", "Path to reviewed migration note file for breaking changes")
	flag.StringVar(&migrationNoteContent, "migration-note-content", "", "Inline reviewed migration note content")
	flag.BoolVar(&allowBreaking, "allow-breaking", false, "Permit breaking schema changes if stability policy and migration note pass")
	flag.BoolVar(&allowOffline, "allow-offline", false, "Do not fail if registry is unreachable and no baseline file exists")
	flag.BoolVar(&quiet, "quiet", false, "Suppress report stdout printing")
	flag.Parse()

	ctx := context.Background()

	// 1. Resolve baseline schema
	resolveOpts := schemacmp.ResolveOptions{
		PreviousVersion: previousVersion,
		BaselineFile:    baselineFilePath,
		AllowOffline:    allowOffline,
	}
	if previousVersion == "latest" {
		resolveOpts.PreviousVersion = ""
	}

	prevSchema, resolvedPrevVer, err := schemacmp.ResolveBaseline(ctx, resolveOpts)
	if err != nil {
		if errors.Is(err, schemacmp.ErrInitialBaseline) {
			prevSchema = nil
		} else if errors.Is(err, schemacmp.ErrRegistryUnavailable) {
			if allowOffline {
				fmt.Fprintf(os.Stderr, "Warning: public registry unavailable and no baseline file found; skipping schema comparison (--allow-offline).\n")
				os.Exit(0)
			}
			fmt.Fprintf(os.Stderr, "Error: registry unavailable to fetch baseline provider: %v\n", err)
			os.Exit(1)
		} else {
			fmt.Fprintf(os.Stderr, "Error resolving baseline: %v\n", err)
			os.Exit(1)
		}
	}

	if exportBaselineOnly != "" {
		if prevSchema == nil {
			fmt.Fprintf(os.Stderr, "Error: cannot export baseline; no baseline schema resolved.\n")
			os.Exit(1)
		}
		data, err := schemacmp.ExportJSON(prevSchema)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting baseline JSON: %v\n", err)
			os.Exit(1)
		}
		if err := os.MkdirAll(filepath.Dir(exportBaselineOnly), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output dir: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(exportBaselineOnly, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing baseline file %s: %v\n", exportBaselineOnly, err)
			os.Exit(1)
		}
		fmt.Printf("Baseline schema (v%s) successfully exported to %s\n", resolvedPrevVer, exportBaselineOnly)
		os.Exit(0)
	}

	if outPreviousSchema != "" && prevSchema != nil {
		data, err := schemacmp.ExportJSON(prevSchema)
		if err == nil {
			_ = os.MkdirAll(filepath.Dir(outPreviousSchema), 0755)
			_ = os.WriteFile(outPreviousSchema, data, 0644)
		}
	}

	// 2. Resolve candidate schema
	var candSchema *schemacmp.NormalizedSchema
	if candidatePath != "" {
		data, err := os.ReadFile(candidatePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading candidate schema %s: %v\n", candidatePath, err)
			os.Exit(1)
		}
		candSchema, err = schemacmp.LoadNormalizedJSON(data)
		if err != nil {
			candSchema, err = schemacmp.NormalizeTFJSON(data)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing candidate schema: %v\n", err)
				os.Exit(1)
			}
		}
	} else {
		// Extract directly from Go provider
		candProvider := provider.New("candidate")()
		var err error
		candSchema, err = schemacmp.NormalizeProvider(candProvider)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error extracting provider schema from Go codebase: %v\n", err)
			os.Exit(1)
		}
	}

	if outCandidateSchema != "" && candSchema != nil {
		data, err := schemacmp.ExportJSON(candSchema)
		if err == nil {
			_ = os.MkdirAll(filepath.Dir(outCandidateSchema), 0755)
			_ = os.WriteFile(outCandidateSchema, data, 0644)
		}
	}

	// 3. Compare schemas
	cmpOpts := schemacmp.CompareOptions{
		AllowBreaking:          allowBreaking,
		TargetVersion:          targetVersion,
		PreviousVersion:        resolvedPrevVer,
		MigrationNotePath:      migrationNotePath,
		MigrationNoteContent:   migrationNoteContent,
		RequireStabilityPolicy: true,
	}

	report := schemacmp.Compare(prevSchema, candSchema, cmpOpts)
	reportMarkdown := schemacmp.GenerateMarkdownReport(report)

	if reportPath != "" {
		if err := os.MkdirAll(filepath.Dir(reportPath), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating report directory: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(reportPath, []byte(reportMarkdown), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing report artifact %s: %v\n", reportPath, err)
			os.Exit(1)
		}
	}

	if !quiet {
		fmt.Println(reportMarkdown)
	}

	if !report.IsCompatible() {
		fmt.Fprintf(os.Stderr, "\nSchema comparison FAILED: incompatible public schema changes detected without policy approval.\n")
		for _, err := range report.PolicyErrors {
			fmt.Fprintf(os.Stderr, "  - %s\n", err)
		}
		os.Exit(1)
	}

	if report.HasBreakingChanges {
		fmt.Printf("\nSchema comparison PASSED with approved breaking changes (major version bump + reviewed migration note verified).\n")
	} else {
		fmt.Printf("\nSchema comparison PASSED: fully backward compatible.\n")
	}
}
