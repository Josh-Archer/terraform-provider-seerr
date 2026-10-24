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
	"sort"
	"strings"
	"time"
)

// OutputFile represents a file to be generated.
type OutputFile struct {
	Path        string
	Content     []byte
	Permissions os.FileMode
	Description string
}

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
		overwrite      bool
	)

	fs := flag.NewFlagSet("seerr-import", flag.ContinueOnError)
	fs.SetOutput(stderr)

	defaultURL := os.Getenv("SEERR_URL")
	if defaultURL == "" {
		defaultURL = "http://localhost:5055"
	}
	defaultKey := os.Getenv("SEERR_API_KEY")

	fs.StringVar(&serverURL, "url", defaultURL, "Base URL of the live Seerr/Jellyseerr/Overseerr instance (or SEERR_URL)")
	fs.StringVar(&apiKey, "api-key", defaultKey, "API key for authentication. Setting SEERR_API_KEY environment variable is preferred over command-line arguments to avoid exposing credentials in process listings.")
	fs.StringVar(&outDir, "out-dir", ".", "Output directory for generated files")
	fs.StringVar(&format, "format", "all", "Output format: 'all', 'hcl', 'variables', 'imports', 'script'")
	fs.BoolVar(&providerHeader, "provider-header", true, "Include provider configuration block in generated HCL")
	fs.IntVar(&timeoutSec, "timeout", 30, "HTTP request timeout in seconds")
	fs.BoolVar(&overwrite, "overwrite", false, "Overwrite existing configuration files in out-dir if they exist (default: false)")
	fs.BoolVar(&overwrite, "force", false, "Alias for -overwrite")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	// Security check: if -api-key was passed on command line rather than env var, advise preferring env var.
	cliPassedKey := false
	for _, arg := range args {
		if strings.HasPrefix(arg, "-api-key") || strings.HasPrefix(arg, "--api-key") {
			cliPassedKey = true
			break
		}
	}
	if cliPassedKey {
		fmt.Fprintln(stderr, "Notice: -api-key was provided via command-line argument. Setting SEERR_API_KEY environment variable is preferred for security.")
	} else if apiKey == "" {
		fmt.Fprintln(stderr, "Warning: No API key provided. Set via SEERR_API_KEY environment variable or -api-key.")
	}

	fmt.Fprintf(stdout, "🔍 Connecting to Seerr at %s...\n", redactURL(serverURL))

	cfg := ImporterConfig{
		BaseURL: serverURL,
		APIKey:  apiKey,
		Timeout: time.Duration(timeoutSec) * time.Second,
	}

	imp := NewImporter(cfg)
	ctx := context.Background()

	resources, err := imp.DiscoverAll(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "Error during discovery: %v\n", redactSensitiveText(err.Error(), apiKey))
		return 1
	}

	if len(resources) == 0 {
		fmt.Fprintln(stdout, "⚠️  No manageable resources discovered. Verify your URL, API Key, and server permissions.")
		return 0
	}

	fmt.Fprintf(stdout, "✅ Discovered %d live resources across your Seerr instance!\n\n", len(resources))

	// Group by resource type for reporting.
	counts := map[string]int{}
	for _, r := range resources {
		counts[r.ResourceType]++
	}
	rTypes := make([]string, 0, len(counts))
	for rType := range counts {
		rTypes = append(rTypes, rType)
	}
	sort.Strings(rTypes)
	for _, rType := range rTypes {
		fmt.Fprintf(stdout, "  • %-30s : %d\n", rType, counts[rType])
	}
	fmt.Fprintln(stdout)

	// Ensure output directory exists with restrictive permissions (0700).
	if err := os.MkdirAll(outDir, 0700); err != nil {
		fmt.Fprintf(stderr, "Error creating output directory %s: %v\n", outDir, err)
		return 1
	}
	// Chmod to guarantee restrictive permissions even if directory pre-existed.
	_ = os.Chmod(outDir, 0700)

	format = strings.ToLower(format)

	var filesToWrite []OutputFile

	if format == "all" || format == "hcl" {
		hclContent := GenerateHCL(resources, providerHeader)
		filesToWrite = append(filesToWrite, OutputFile{
			Path:        filepath.Join(outDir, "main.tf"),
			Content:     []byte(hclContent),
			Permissions: 0600,
			Description: "📄 Generated HCL resources",
		})
	}

	sensitiveVarsContent := GenerateVariables(resources)
	if (format == "all" || format == "hcl" || format == "variables") && sensitiveVarsContent != "" {
		filesToWrite = append(filesToWrite, OutputFile{
			Path:        filepath.Join(outDir, "variables.tf"),
			Content:     []byte(sensitiveVarsContent),
			Permissions: 0600,
			Description: "🔒 Generated Sensitive variables",
		})
	}

	if format == "all" || format == "imports" {
		importContent := GenerateImportBlocks(resources)
		filesToWrite = append(filesToWrite, OutputFile{
			Path:        filepath.Join(outDir, "imports.tf"),
			Content:     []byte(importContent),
			Permissions: 0600,
			Description: "📦 Generated Import blocks",
		})
	}

	if format == "all" || format == "script" {
		scriptContent := GenerateImportScript(resources)
		filesToWrite = append(filesToWrite, OutputFile{
			Path:        filepath.Join(outDir, "import.sh"),
			Content:     []byte(scriptContent),
			Permissions: 0700,
			Description: "🚀 Generated Import script",
		})
	}

	// Refuse overwriting existing files by default.
	if !overwrite {
		for _, f := range filesToWrite {
			if _, err := os.Stat(f.Path); err == nil {
				fmt.Fprintf(stderr, "Error: refusing to overwrite existing file %s (use -overwrite to force)\n", f.Path)
				return 1
			}
		}
	}

	// Write files with restrictive permissions.
	for _, f := range filesToWrite {
		if err := os.WriteFile(f.Path, f.Content, f.Permissions); err != nil {
			fmt.Fprintf(stderr, "Error writing %s: %v\n", f.Path, err)
			return 1
		}
		// Enforce permissions strictly in case umask altered them.
		_ = os.Chmod(f.Path, f.Permissions)
		fmt.Fprintf(stdout, "%s: %s\n", f.Description, f.Path)
	}

	fmt.Fprintln(stdout, "\n🎉 Migration files generated successfully!")
	fmt.Fprintln(stdout, "Next steps:")
	if sensitiveVarsContent != "" {
		fmt.Fprintln(stdout, "  1. Review the generated main.tf, variables.tf, and imports.tf")
		fmt.Fprintln(stdout, "  2. Supply values for sensitive variables via terraform.tfvars or TF_VAR_<name> environment variables")
		fmt.Fprintln(stdout, "  3. SECURITY NOTICE: Terraform/OpenTofu state stores sensitive values in plaintext.")
		fmt.Fprintln(stdout, "     Configure an encrypted remote backend or OpenTofu state encryption. Never commit .tfstate!")
		fmt.Fprintln(stdout, "  4. Run 'tofu init' or 'terraform init'")
		fmt.Fprintln(stdout, "  5. Run 'tofu plan' to verify 0 unexpected changes")
		fmt.Fprintln(stdout, "  6. Run 'tofu apply' to import state and manage your Seerr stack with IaC!")
	} else {
		fmt.Fprintln(stdout, "  1. Review the generated main.tf and imports.tf")
		fmt.Fprintln(stdout, "  2. SECURITY NOTICE: Terraform/OpenTofu state stores resource attributes.")
		fmt.Fprintln(stdout, "     Configure a secure remote backend. Never commit .tfstate!")
		fmt.Fprintln(stdout, "  3. Run 'tofu init' or 'terraform init'")
		fmt.Fprintln(stdout, "  4. Run 'tofu plan' to verify 0 unexpected changes")
		fmt.Fprintln(stdout, "  5. Run 'tofu apply' to import state and manage your Seerr stack with IaC!")
	}

	return 0
}
