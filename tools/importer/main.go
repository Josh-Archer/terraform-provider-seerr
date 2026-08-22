// Copyright (c) Josh Archer
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	var (
		serverURL      string
		apiKey         string
		outDir         string
		format         string
		providerHeader bool
		timeoutSec     int
	)

	defaultURL := os.Getenv("SEERR_URL")
	if defaultURL == "" {
		defaultURL = "http://localhost:5055"
	}
	defaultKey := os.Getenv("SEERR_API_KEY")

	flag.StringVar(&serverURL, "url", defaultURL, "Base URL of the live Seerr/Jellyseerr/Overseerr instance (or SEERR_URL)")
	flag.StringVar(&apiKey, "api-key", defaultKey, "API key for authentication (or SEERR_API_KEY)")
	flag.StringVar(&outDir, "out-dir", ".", "Output directory for generated files")
	flag.StringVar(&format, "format", "all", "Output format: 'all', 'hcl', 'imports', 'script'")
	flag.BoolVar(&providerHeader, "provider-header", true, "Include provider configuration block in generated HCL")
	flag.IntVar(&timeoutSec, "timeout", 30, "HTTP request timeout in seconds")
	flag.Parse()

	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Warning: No API key provided. Set via -api-key or SEERR_API_KEY env var.")
	}

	fmt.Printf("🔍 Connecting to Seerr at %s...\n", serverURL)

	cfg := ImporterConfig{
		BaseURL: serverURL,
		APIKey:  apiKey,
		Timeout: time.Duration(timeoutSec) * time.Second,
	}

	imp := NewImporter(cfg)
	ctx := context.Background()

	resources, err := imp.DiscoverAll(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during discovery: %v\n", err)
		os.Exit(1)
	}

	if len(resources) == 0 {
		fmt.Println("⚠️  No manageable resources discovered. Verify your URL, API Key, and server permissions.")
		return
	}

	fmt.Printf("✅ Discovered %d live resources across your Seerr instance!\n\n", len(resources))

	// Group by resource type for reporting
	counts := map[string]int{}
	for _, r := range resources {
		counts[r.ResourceType]++
	}
	for rType, count := range counts {
		fmt.Printf("  • %-30s : %d\n", rType, count)
	}
	fmt.Println()

	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory %s: %v\n", outDir, err)
		os.Exit(1)
	}

	format = strings.ToLower(format)

	if format == "all" || format == "hcl" {
		hclContent := GenerateHCL(resources, providerHeader)
		hclPath := filepath.Join(outDir, "main.tf")
		if err := os.WriteFile(hclPath, []byte(hclContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", hclPath, err)
		} else {
			fmt.Printf("📄 Generated HCL resources: %s\n", hclPath)
		}
	}

	if format == "all" || format == "imports" {
		importContent := GenerateImportBlocks(resources)
		importPath := filepath.Join(outDir, "imports.tf")
		if err := os.WriteFile(importPath, []byte(importContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", importPath, err)
		} else {
			fmt.Printf("📦 Generated Import blocks:  %s\n", importPath)
		}
	}

	if format == "all" || format == "script" {
		scriptContent := GenerateImportScript(resources)
		scriptPath := filepath.Join(outDir, "import.sh")
		if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", scriptPath, err)
		} else {
			fmt.Printf("🚀 Generated Import script:  %s\n", scriptPath)
		}
	}

	fmt.Println("\n🎉 Migration files generated successfully!")
	fmt.Println("Next steps:")
	fmt.Println("  1. Review the generated main.tf and imports.tf")
	fmt.Println("  2. Run 'tofu init' or 'terraform init'")
	fmt.Println("  3. Run 'tofu plan' to verify 0 unexpected changes")
	fmt.Println("  4. Run 'tofu apply' to import state and manage your Seerr stack with IaC!")
}
