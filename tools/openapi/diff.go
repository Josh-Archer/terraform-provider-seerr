package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// RawOpenAPISpec represents parsed OpenAPI 3.0 paths and operations.
type RawOpenAPISpec struct {
	OpenAPI string                             `json:"openapi" yaml:"openapi"`
	Info    map[string]any                     `json:"info" yaml:"info"`
	Paths   map[string]map[string]RawOperation `json:"paths" yaml:"paths"`
}

// RawOperation represents an OpenAPI operation (GET, POST, etc.)
type RawOperation struct {
	Summary     string         `json:"summary" yaml:"summary"`
	Description string         `json:"description" yaml:"description"`
	OperationID string         `json:"operationId" yaml:"operationId"`
	Deprecated  bool           `json:"deprecated" yaml:"deprecated"`
	Tags        []string       `json:"tags" yaml:"tags"`
	Parameters  []any          `json:"parameters" yaml:"parameters"`
	RequestBody map[string]any `json:"requestBody" yaml:"requestBody"`
	Responses   map[string]any `json:"responses" yaml:"responses"`
}

// EndpointID uniquely identifies a path and method.
type EndpointID struct {
	Path   string
	Method string
}

// DriftReport records all differences between a base and upstream OpenAPI spec.
type DriftReport struct {
	BaseSource             string           `json:"base_source"`
	UpstreamSource         string           `json:"upstream_source"`
	GeneratedAt            time.Time        `json:"generated_at"`
	TotalBaseEndpoints     int              `json:"total_base_endpoints"`
	TotalUpstreamEndpoints int              `json:"total_upstream_endpoints"`
	AddedEndpoints         []PathMethodInfo `json:"added_endpoints"`
	RemovedEndpoints       []PathMethodInfo `json:"removed_endpoints"`
	DeprecatedEndpoints    []PathMethodInfo `json:"deprecated_endpoints"`
	NewPaths               []string         `json:"new_paths"`
	RemovedPaths           []string         `json:"removed_paths"`
	HasDrift               bool             `json:"has_drift"`
}

// PathMethodInfo holds summary details for an endpoint.
type PathMethodInfo struct {
	Path        string `json:"path"`
	Method      string `json:"method"`
	Summary     string `json:"summary,omitempty"`
	OperationID string `json:"operation_id,omitempty"`
	Deprecated  bool   `json:"deprecated,omitempty"`
}

// LoadSpec reads and parses an OpenAPI spec from a local file or HTTP URL (JSON or YAML).
func LoadSpec(source string) (*RawOpenAPISpec, error) {
	var data []byte
	var err error

	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Get(source)
		if err != nil {
			return nil, fmt.Errorf("fetch remote spec from %s: %w", source, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("fetch remote spec %s returned HTTP %d", source, resp.StatusCode)
		}
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read response body: %w", err)
		}
	} else {
		data, err = os.ReadFile(source)
		if err != nil {
			return nil, fmt.Errorf("read local spec file %s: %w", source, err)
		}
	}

	var spec RawOpenAPISpec
	// Try JSON first if payload begins with '{'
	trimmed := strings.TrimSpace(string(data))
	if strings.HasPrefix(trimmed, "{") {
		if err := json.Unmarshal(data, &spec); err == nil && len(spec.Paths) > 0 {
			return &spec, nil
		}
	}

	// Fallback to YAML (handles both YAML and standard JSON)
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("unmarshal spec: %w", err)
	}

	return &spec, nil
}

// CompareSpecs compares a base OpenAPI spec against an upstream spec.
func CompareSpecs(base, upstream *RawOpenAPISpec, baseSrc, upstreamSrc string) *DriftReport {
	report := &DriftReport{
		BaseSource:     baseSrc,
		UpstreamSource: upstreamSrc,
		GeneratedAt:    time.Now().UTC(),
	}

	baseEndpoints := extractEndpoints(base)
	upstreamEndpoints := extractEndpoints(upstream)

	report.TotalBaseEndpoints = len(baseEndpoints)
	report.TotalUpstreamEndpoints = len(upstreamEndpoints)

	// Check added endpoints (in upstream but not base)
	for key, upOp := range upstreamEndpoints {
		if _, exists := baseEndpoints[key]; !exists {
			report.AddedEndpoints = append(report.AddedEndpoints, PathMethodInfo{
				Path:        key.Path,
				Method:      key.Method,
				Summary:     upOp.Summary,
				OperationID: upOp.OperationID,
				Deprecated:  upOp.Deprecated,
			})
		} else if upOp.Deprecated && !baseEndpoints[key].Deprecated {
			report.DeprecatedEndpoints = append(report.DeprecatedEndpoints, PathMethodInfo{
				Path:        key.Path,
				Method:      key.Method,
				Summary:     upOp.Summary,
				OperationID: upOp.OperationID,
				Deprecated:  true,
			})
		}
	}

	// Check removed endpoints (in base but not upstream)
	for key, baseOp := range baseEndpoints {
		if _, exists := upstreamEndpoints[key]; !exists {
			report.RemovedEndpoints = append(report.RemovedEndpoints, PathMethodInfo{
				Path:        key.Path,
				Method:      key.Method,
				Summary:     baseOp.Summary,
				OperationID: baseOp.OperationID,
				Deprecated:  baseOp.Deprecated,
			})
		}
	}

	// Sort for deterministic results
	sortInfoSlice(report.AddedEndpoints)
	sortInfoSlice(report.RemovedEndpoints)
	sortInfoSlice(report.DeprecatedEndpoints)

	// Detect new / removed paths
	basePaths := make(map[string]bool)
	for p := range base.Paths {
		basePaths[p] = true
	}
	upstreamPaths := make(map[string]bool)
	for p := range upstream.Paths {
		upstreamPaths[p] = true
	}

	for p := range upstreamPaths {
		if !basePaths[p] {
			report.NewPaths = append(report.NewPaths, p)
		}
	}
	for p := range basePaths {
		if !upstreamPaths[p] {
			report.RemovedPaths = append(report.RemovedPaths, p)
		}
	}
	sort.Strings(report.NewPaths)
	sort.Strings(report.RemovedPaths)

	report.HasDrift = len(report.AddedEndpoints) > 0 || len(report.RemovedEndpoints) > 0 || len(report.DeprecatedEndpoints) > 0

	return report
}

func extractEndpoints(spec *RawOpenAPISpec) map[EndpointID]RawOperation {
	m := make(map[EndpointID]RawOperation)
	if spec == nil {
		return m
	}
	for pathStr, ops := range spec.Paths {
		for method, op := range ops {
			mUpper := strings.ToUpper(method)
			if mUpper == "GET" || mUpper == "POST" || mUpper == "PUT" || mUpper == "DELETE" || mUpper == "PATCH" {
				m[EndpointID{Path: pathStr, Method: mUpper}] = op
			}
		}
	}
	return m
}

func sortInfoSlice(items []PathMethodInfo) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Path == items[j].Path {
			return items[i].Method < items[j].Method
		}
		return items[i].Path < items[j].Path
	})
}

// FormatMarkdown creates a markdown summary report of the drift comparison.
func (r *DriftReport) FormatMarkdown() string {
	var sb strings.Builder

	sb.WriteString("# Upstream OpenAPI Drift Report\n\n")
	sb.WriteString(fmt.Sprintf("**Generated At**: %s\n\n", r.GeneratedAt.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("- **Base Spec**: `%s` (%d endpoints)\n", r.BaseSource, r.TotalBaseEndpoints))
	sb.WriteString(fmt.Sprintf("- **Upstream Spec**: `%s` (%d endpoints)\n", r.UpstreamSource, r.TotalUpstreamEndpoints))
	sb.WriteString(fmt.Sprintf("- **Drift Detected**: `%t`\n\n", r.HasDrift))

	sb.WriteString("## Summary\n\n")
	sb.WriteString("| Metric | Count |\n")
	sb.WriteString("|---|---|\n")
	sb.WriteString(fmt.Sprintf("| New Endpoints Added | %d |\n", len(r.AddedEndpoints)))
	sb.WriteString(fmt.Sprintf("| Endpoints Removed | %d |\n", len(r.RemovedEndpoints)))
	sb.WriteString(fmt.Sprintf("| Newly Deprecated Endpoints | %d |\n", len(r.DeprecatedEndpoints)))
	sb.WriteString(fmt.Sprintf("| New Distinct Paths | %d |\n\n", len(r.NewPaths)))

	if len(r.AddedEndpoints) > 0 {
		sb.WriteString("### 🟢 Added Endpoints in Upstream\n\n")
		sb.WriteString("| Method | Path | Summary / Operation ID |\n")
		sb.WriteString("|---|---|---|\n")
		for _, ep := range r.AddedEndpoints {
			desc := ep.Summary
			if desc == "" {
				desc = ep.OperationID
			}
			sb.WriteString(fmt.Sprintf("| `%s` | `%s` | %s |\n", ep.Method, ep.Path, desc))
		}
		sb.WriteString("\n")
	}

	if len(r.RemovedEndpoints) > 0 {
		sb.WriteString("### 🔴 Removed Endpoints in Upstream\n\n")
		sb.WriteString("| Method | Path | Summary / Operation ID |\n")
		sb.WriteString("|---|---|---|\n")
		for _, ep := range r.RemovedEndpoints {
			desc := ep.Summary
			if desc == "" {
				desc = ep.OperationID
			}
			sb.WriteString(fmt.Sprintf("| `%s` | `%s` | %s |\n", ep.Method, ep.Path, desc))
		}
		sb.WriteString("\n")
	}

	if len(r.DeprecatedEndpoints) > 0 {
		sb.WriteString("### ⚠️ Newly Deprecated Endpoints\n\n")
		sb.WriteString("| Method | Path | Summary / Operation ID |\n")
		sb.WriteString("|---|---|---|\n")
		for _, ep := range r.DeprecatedEndpoints {
			desc := ep.Summary
			if desc == "" {
				desc = ep.OperationID
			}
			sb.WriteString(fmt.Sprintf("| `%s` | `%s` | %s |\n", ep.Method, ep.Path, desc))
		}
		sb.WriteString("\n")
	}

	if !r.HasDrift {
		sb.WriteString("✅ **No drift detected.** Local OpenAPI specification matches upstream perfectly.\n")
	}

	return sb.String()
}

func runDiffCLI() {
	var (
		baseFlag        = flag.String("base", "", "Path to local base OpenAPI spec (default: tools/openapi/seerr-api.yml)")
		upstreamFlag    = flag.String("upstream", "", "Path or URL to upstream OpenAPI spec")
		outputFlag      = flag.String("output", "", "Output file path (default: stdout)")
		formatFlag      = flag.String("format", "markdown", "Output format: markdown or json")
		failOnDriftFlag = flag.Bool("fail-on-drift", false, "Exit with non-zero code if drift is detected")
	)
	flag.Parse()

	basePath := *baseFlag
	if basePath == "" {
		root, err := FindRepoRoot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to find repo root: %v\n", err)
			os.Exit(1)
		}
		basePath = filepath.Join(root, "tools", "openapi", "seerr-api.yml")
	}

	upstreamPath := *upstreamFlag
	if upstreamPath == "" {
		upstreamPath = "https://raw.githubusercontent.com/seerr-team/seerr/develop/seerr-api.yml"
	}

	baseSpec, err := LoadSpec(basePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load base spec %s: %v\n", basePath, err)
		os.Exit(1)
	}

	upstreamSpec, err := LoadSpec(upstreamPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load upstream spec %s: %v\n", upstreamPath, err)
		os.Exit(1)
	}

	report := CompareSpecs(baseSpec, upstreamSpec, basePath, upstreamPath)

	var outputContent string
	if *formatFlag == "json" {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to marshal report json: %v\n", err)
			os.Exit(1)
		}
		outputContent = string(data)
	} else {
		outputContent = report.FormatMarkdown()
	}

	if *outputFlag != "" {
		if err := os.WriteFile(*outputFlag, []byte(outputContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Wrote drift report to %s (has_drift=%t)\n", *outputFlag, report.HasDrift)
	} else {
		fmt.Println(outputContent)
	}

	if *failOnDriftFlag && report.HasDrift {
		os.Exit(2)
	}
}
