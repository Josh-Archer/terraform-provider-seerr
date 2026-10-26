package schemacmp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	ErrInitialBaseline     = errors.New("initial baseline: no previous release found")
	ErrRegistryUnavailable = errors.New("registry unavailable: failed to install baseline provider from public registry")
)

// ResolveOptions controls baseline retrieval.
type ResolveOptions struct {
	PreviousVersion string
	BaselineFile    string
	RegistryURL     string // default: registry.opentofu.org/josh-archer/seerr
	WorkDir         string
	AllowOffline    bool
}

// ResolveBaseline locates and loads the baseline NormalizedSchema for the previous stable release.
func ResolveBaseline(ctx context.Context, opts ResolveOptions) (*NormalizedSchema, string, error) {
	// 1. If explicit baseline file is provided, load directly
	if opts.BaselineFile != "" {
		data, err := os.ReadFile(opts.BaselineFile)
		if err != nil {
			return nil, "", fmt.Errorf("failed to read baseline file %s: %w", opts.BaselineFile, err)
		}
		schema, err := LoadNormalizedJSON(data)
		if err != nil {
			// Try as raw TF JSON
			var tfErr error
			schema, tfErr = NormalizeTFJSON(data)
			if tfErr != nil {
				return nil, "", fmt.Errorf("failed to parse baseline file %s: %w (tf parse err: %v)", opts.BaselineFile, err, tfErr)
			}
		}
		ver := opts.PreviousVersion
		if ver == "" {
			ver = extractVersionFromFilename(opts.BaselineFile)
		}
		return schema, ver, nil
	}

	// 2. Resolve previous version tag
	prevVersion := opts.PreviousVersion
	if prevVersion == "" {
		prevVersion = os.Getenv("SEERR_SCHEMA_PREVIOUS_VERSION")
	}
	if prevVersion == "" {
		var err error
		prevVersion, err = findLatestStableGitTag()
		if err != nil || prevVersion == "" {
			return nil, "", ErrInitialBaseline
		}
	}
	prevVersion = strings.TrimPrefix(prevVersion, "v")

	// 3. Check for committed baseline file in well-known locations
	knownPaths := []string{
		filepath.Join("testdata", fmt.Sprintf("provider_schema_baseline_v%s.json", prevVersion)),
		filepath.Join("testdata", "provider_schema_baseline.json"),
		filepath.Join("internal", "provider", "testdata", "provider_schema_baseline.json"),
	}
	for _, p := range knownPaths {
		if data, err := os.ReadFile(p); err == nil {
			schema, err := LoadNormalizedJSON(data)
			if err == nil {
				return schema, prevVersion, nil
			}
		}
	}

	// 4. Download previous release from registry via tofu / terraform
	schema, err := fetchSchemaFromRegistry(ctx, prevVersion, opts)
	if err != nil {
		if errors.Is(err, ErrRegistryUnavailable) && opts.AllowOffline {
			return nil, prevVersion, ErrRegistryUnavailable
		}
		return nil, prevVersion, err
	}

	return schema, prevVersion, nil
}

func findLatestStableGitTag() (string, error) {
	cmd := exec.Command("git", "tag", "--list", "v*", "--sort=-v:refname")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	stablePattern := regexp.MustCompile(`^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$`)
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if stablePattern.MatchString(line) {
			return line, nil
		}
	}

	return "", nil
}

func fetchSchemaFromRegistry(ctx context.Context, version string, opts ResolveOptions) (*NormalizedSchema, error) {
	tool := "tofu"
	if _, err := exec.LookPath("tofu"); err != nil {
		if _, err := exec.LookPath("terraform"); err == nil {
			tool = "terraform"
		} else {
			return nil, fmt.Errorf("%w: neither tofu nor terraform is available to fetch registry schema", ErrRegistryUnavailable)
		}
	}

	tmpDir, err := os.MkdirTemp(opts.WorkDir, "seerr-schema-fetch-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	source := opts.RegistryURL
	if source == "" {
		if tool == "tofu" {
			source = "registry.opentofu.org/josh-archer/seerr"
		} else {
			source = "registry.terraform.io/josh-archer/seerr"
		}
	}

	mainTF := fmt.Sprintf(`terraform {
  required_providers {
    seerr = {
      source  = %q
      version = %q
    }
  }
}
`, source, version)

	if err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(mainTF), 0600); err != nil {
		return nil, fmt.Errorf("failed to write main.tf: %w", err)
	}

	initCmd := exec.CommandContext(ctx, tool, "init", "-input=false", "-no-color")
	initCmd.Dir = tmpDir
	var initErrBuf bytes.Buffer
	initCmd.Stderr = &initErrBuf
	if err := initCmd.Run(); err != nil {
		return nil, fmt.Errorf("%w: %s init failed for version %s: %s (%v)",
			ErrRegistryUnavailable, tool, version, initErrBuf.String(), err)
	}

	schemaCmd := exec.CommandContext(ctx, tool, "providers", "schema", "-json")
	schemaCmd.Dir = tmpDir
	out, err := schemaCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run %s providers schema -json: %w", tool, err)
	}

	return NormalizeTFJSON(out)
}

func extractVersionFromFilename(p string) string {
	base := filepath.Base(p)
	re := regexp.MustCompile(`v?([0-9]+\.[0-9]+\.[0-9]+)`)
	matches := re.FindStringSubmatch(base)
	if len(matches) > 1 {
		return matches[1]
	}
	return "baseline"
}
