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
	var overwrite bool

	fs := flag.NewFlagSet("seerr-import", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.StringVar(&serverURL, "url", defaultURL, "Base URL of the live Seerr/Jellyseerr/Overseerr instance (or SEERR_URL)")
	fs.StringVar(&apiKey, "api-key", defaultKey, "API key for authentication (or SEERR_API_KEY). Prefer the environment variable so the key is not stored in shell history.")
	fs.StringVar(&outDir, "out-dir", ".", "Output directory for generated files")
	fs.StringVar(&format, "format", "all", "Output format: 'all', 'hcl', 'imports', 'script'")
	fs.BoolVar(&providerHeader, "provider-header", true, "Include provider configuration block in generated HCL")
	fs.BoolVar(&overwrite, "overwrite", false, "Overwrite existing generated files in out-dir (refuses by default)")
	fs.BoolVar(&overwrite, "force", false, "Alias for -overwrite")
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

	type outFile struct {
		name    string
		content string
		mode    os.FileMode
	}
	var files []outFile
	variables := ""
	if format == "all" || format == "hcl" {
		files = append(files, outFile{"main.tf", GenerateHCL(resources, providerHeader), 0o600})
		if variables = GenerateVariables(resources); variables != "" {
			files = append(files, outFile{"variables.tf", variables, 0o600})
		}
	}
	if format == "all" || format == "imports" {
		files = append(files, outFile{"imports.tf", GenerateImportBlocks(resources), 0o600})
	}
	if format == "all" || format == "script" {
		files = append(files, outFile{"import.sh", GenerateImportScript(resources), 0o700})
	}

	// Generated files can reference credentials, so keep the directory and files owner-only.
	if err := os.MkdirAll(outDir, 0o700); err != nil {
		fmt.Fprintf(stderr, "Error creating output directory %s: %v\n", outDir, err)
		return 1
	}

	if !overwrite {
		for _, f := range files {
			path := filepath.Join(outDir, f.name)
			if _, err := os.Lstat(path); err == nil {
				fmt.Fprintf(stderr, "Error: refusing to overwrite existing file %s (use -overwrite to replace it)\n", path)
				return 1
			}
		}
	}

	for _, f := range files {
		path := filepath.Join(outDir, f.name)
		if err := os.WriteFile(path, []byte(f.content), f.mode); err != nil {
			fmt.Fprintf(stderr, "Error writing %s: %v\n", path, err)
			return 1
		}
		// WriteFile keeps the mode of an existing file and applies umask, so set it explicitly.
		if err := os.Chmod(path, f.mode); err != nil {
			fmt.Fprintf(stderr, "Error setting permissions on %s: %v\n", path, err)
			return 1
		}
		fmt.Fprintf(stdout, "📄 Wrote %s\n", path)
	}
	if variables != "" {
		fmt.Fprintln(stdout, "Sensitive values were replaced with variables. Supply them with TF_VAR_* or an uncommitted tfvars file before planning.")
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
