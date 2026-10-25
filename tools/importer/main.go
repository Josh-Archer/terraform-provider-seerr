// Copyright (c) Josh Archer
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	var (
		serverURL      string
		apiKey         string
		outDir         string
		format         string
		providerHeader bool
		timeoutSec     int
		allowPartial   bool
	)

	defaultURL := os.Getenv("SEERR_URL")
	if defaultURL == "" {
		defaultURL = "http://localhost:5055"
	}
	defaultKey := os.Getenv("SEERR_API_KEY")

	fs := flag.NewFlagSet("seerr-import", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.StringVar(&serverURL, "url", defaultURL, "Base URL of the live Seerr/Jellyseerr/Overseerr instance (or SEERR_URL)")
	fs.StringVar(&apiKey, "api-key", defaultKey, "API key for authentication (or SEERR_API_KEY). Prefer the environment variable so the key is not stored in shell history.")
	fs.StringVar(&outDir, "out-dir", ".", "Output directory for generated files")
	fs.StringVar(&format, "format", "all", "Output format: 'all', 'hcl', 'imports', 'script'")
	fs.BoolVar(&providerHeader, "provider-header", true, "Include provider configuration block in generated HCL")
	fs.IntVar(&timeoutSec, "timeout", 30, "HTTP request timeout in seconds")
	fs.BoolVar(&allowPartial, "allow-partial", false, "Write discovered resources even when authentication, transport, or server errors make discovery incomplete. HTTP 404 is skipped either way.")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "unexpected argument %q\n", fs.Arg(0))
		return 2
	}

	canonical, err := canonicalFormat(format)
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 2
	}
	format = canonical

	if apiKey == "" {
		fmt.Fprintln(stderr, "Warning: No API key provided. Set via -api-key or SEERR_API_KEY env var.")
	}

	fmt.Fprintf(stdout, "🔍 Connecting to Seerr at %s...\n", serverURL)

	imp := NewImporter(ImporterConfig{
		BaseURL:      serverURL,
		APIKey:       apiKey,
		Timeout:      time.Duration(timeoutSec) * time.Second,
		AllowPartial: allowPartial,
	})

	resources, err := imp.DiscoverAll(context.Background())
	if err != nil {
		fmt.Fprintf(stderr, "Error during discovery: %v\n", err)
		return 1
	}
	if failures := imp.Failures(); len(failures) > 0 {
		fmt.Fprintf(stderr, "Partial discovery: %d error(s) ignored because -allow-partial is set\n", len(failures))
		for _, failure := range failures {
			fmt.Fprintf(stderr, "  • %s\n", failure.Error())
		}
	}
	if skipped := imp.UnsupportedEndpoints(); len(skipped) > 0 {
		fmt.Fprintf(stdout, "Skipped %d unsupported endpoints (HTTP 404).\n", len(skipped))
	}
	if notes := imp.ManualAdoption(); len(notes) > 0 {
		fmt.Fprintln(stdout, "Manual adoption required for settings left unmanaged:")
		for _, note := range notes {
			fmt.Fprintf(stdout, "  • %s\n", note)
		}
	}

	if len(resources) == 0 {
		fmt.Fprintln(stdout, "⚠️  No manageable resources discovered. Verify your URL, API Key, and server permissions.")
		if len(imp.Failures()) > 0 {
			return 0
		}
		return 0
	}

	fmt.Fprintf(stdout, "✅ Discovered %d live resources across your Seerr instance!\n\n", len(resources))

	counts := map[string]int{}
	var countOrder []string
	for _, resource := range resources {
		if _, ok := counts[resource.ResourceType]; !ok {
			countOrder = append(countOrder, resource.ResourceType)
		}
		counts[resource.ResourceType]++
	}
	for _, resourceType := range countOrder {
		fmt.Fprintf(stdout, "  • %-30s : %d\n", resourceType, counts[resourceType])
	}
	fmt.Fprintln(stdout)

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintf(stderr, "Error creating output directory %s: %v\n", outDir, err)
		return 1
	}

	write := func(name, content string, mode os.FileMode) int {
		path := filepath.Join(outDir, name)
		if err := os.WriteFile(path, []byte(content), mode); err != nil {
			fmt.Fprintf(stderr, "Error writing %s: %v\n", path, err)
			return 1
		}
		fmt.Fprintf(stdout, "📄 Wrote %s\n", path)
		return 0
	}

	if format == "all" || format == "hcl" {
		if write("main.tf", GenerateHCL(resources, providerHeader), 0o644) != 0 {
			return 1
		}
		if variables := GenerateVariables(resources); variables != "" {
			if write("variables.tf", variables, 0o644) != 0 {
				return 1
			}
			fmt.Fprintln(stdout, "Sensitive values were replaced with variables. Supply them with TF_VAR_* or an uncommitted tfvars file before planning.")
		}
	}
	if format == "all" || format == "imports" {
		if write("imports.tf", GenerateImportBlocks(resources), 0o644) != 0 {
			return 1
		}
	}
	if format == "all" || format == "script" {
		if write("import.sh", GenerateImportScript(resources), 0o755) != 0 {
			return 1
		}
	}

	fmt.Fprintln(stdout, "\n🎉 Migration files generated successfully!")
	fmt.Fprintln(stdout, "Next steps:")
	fmt.Fprintln(stdout, "  1. Review main.tf, imports.tf, and variables.tf")
	fmt.Fprintln(stdout, "  2. Export TF_VAR_* values or add an uncommitted tfvars file for sensitive variables")
	fmt.Fprintln(stdout, "  3. Run 'tofu init' or 'terraform init'")
	fmt.Fprintln(stdout, "  4. Run 'tofu apply' to import, then 'tofu plan' and confirm no changes")
	fmt.Fprintln(stdout, "Settings for resource types this tool does not emit stay on the server. See docs/guides/migration.md.")
	return 0
}
