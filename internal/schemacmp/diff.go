package schemacmp

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var semverPattern = regexp.MustCompile(`^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-([0-9A-Za-z.-]+))?$`)

// Compare evaluates a candidate schema against a previous baseline schema.
func Compare(prev, cand *NormalizedSchema, opts CompareOptions) *ComparisonReport {
	report := &ComparisonReport{
		PreviousVersion:  opts.PreviousVersion,
		CandidateVersion: opts.TargetVersion,
		PolicyPassed:     true,
	}

	if prev == nil {
		report.IsInitialBaseline = true
		return report
	}

	// 1. Compare Provider Schema
	if prev.Provider != nil && cand.Provider != nil {
		diffBlocks("provider", "provider", "", prev.Provider, cand.Provider, report)
	}

	// 2. Compare Resources
	var prevResNames []string
	for name := range prev.Resources {
		prevResNames = append(prevResNames, name)
	}
	sort.Strings(prevResNames)

	for _, name := range prevResNames {
		prevRes := prev.Resources[name]
		candRes, ok := cand.Resources[name]
		if !ok {
			report.BreakingChanges = append(report.BreakingChanges, SchemaChange{
				Risk:       RiskBreaking,
				Type:       ChangeResourceRemoved,
				TargetKind: "resource",
				TargetName: name,
				Path:       name,
				Previous:   "managed resource exists",
				Candidate:  "(removed)",
				Details:    fmt.Sprintf("Managed resource %q was removed from the public provider schema", name),
			})
			continue
		}
		diffBlocks("resource", name, "", prevRes, candRes, report)
	}

	var candResNames []string
	for name := range cand.Resources {
		candResNames = append(candResNames, name)
	}
	sort.Strings(candResNames)

	for _, name := range candResNames {
		if _, ok := prev.Resources[name]; !ok {
			report.CompatibleChanges = append(report.CompatibleChanges, SchemaChange{
				Risk:       RiskCompatible,
				Type:       ChangeResourceAdded,
				TargetKind: "resource",
				TargetName: name,
				Path:       name,
				Previous:   "(none)",
				Candidate:  "managed resource added",
				Details:    fmt.Sprintf("New managed resource %q added", name),
			})
		}
	}

	// 3. Compare Data Sources
	var prevDSNames []string
	for name := range prev.DataSources {
		prevDSNames = append(prevDSNames, name)
	}
	sort.Strings(prevDSNames)

	for _, name := range prevDSNames {
		prevDS := prev.DataSources[name]
		candDS, ok := cand.DataSources[name]
		if !ok {
			report.BreakingChanges = append(report.BreakingChanges, SchemaChange{
				Risk:       RiskBreaking,
				Type:       ChangeDataSourceRemoved,
				TargetKind: "data_source",
				TargetName: name,
				Path:       name,
				Previous:   "data source exists",
				Candidate:  "(removed)",
				Details:    fmt.Sprintf("Data source %q was removed from the public provider schema", name),
			})
			continue
		}
		diffBlocks("data_source", name, "", prevDS, candDS, report)
	}

	var candDSNames []string
	for name := range cand.DataSources {
		candDSNames = append(candDSNames, name)
	}
	sort.Strings(candDSNames)

	for _, name := range candDSNames {
		if _, ok := prev.DataSources[name]; !ok {
			report.CompatibleChanges = append(report.CompatibleChanges, SchemaChange{
				Risk:       RiskCompatible,
				Type:       ChangeDataSourceAdded,
				TargetKind: "data_source",
				TargetName: name,
				Path:       name,
				Previous:   "(none)",
				Candidate:  "data source added",
				Details:    fmt.Sprintf("New data source %q added", name),
			})
		}
	}

	report.HasBreakingChanges = len(report.BreakingChanges) > 0

	// 4. Validate Stability Policy and Reviewed Migration Note
	if report.HasBreakingChanges {
		validateStabilityPolicy(report, opts)
	}

	return report
}

func diffBlocks(targetKind, targetName, pathPrefix string, prevBlock, candBlock *NormalizedBlock, report *ComparisonReport) {
	// Attributes
	var prevAttrNames []string
	for name := range prevBlock.Attributes {
		prevAttrNames = append(prevAttrNames, name)
	}
	sort.Strings(prevAttrNames)

	for _, name := range prevAttrNames {
		prevAttr := prevBlock.Attributes[name]
		fullPath := joinPath(pathPrefix, name)

		candAttr, ok := candBlock.Attributes[name]
		if !ok {
			report.BreakingChanges = append(report.BreakingChanges, SchemaChange{
				Risk:       RiskBreaking,
				Type:       ChangeAttributeRemoved,
				TargetKind: targetKind,
				TargetName: targetName,
				Path:       fullPath,
				Previous:   prevAttr.Descriptor(),
				Candidate:  "(removed)",
				Details:    fmt.Sprintf("Attribute %q was removed; configurations referencing it will fail", fullPath),
			})
			continue
		}

		diffAttribute(targetKind, targetName, fullPath, prevAttr, candAttr, report)
	}

	var candAttrNames []string
	for name := range candBlock.Attributes {
		candAttrNames = append(candAttrNames, name)
	}
	sort.Strings(candAttrNames)

	for _, name := range candAttrNames {
		if _, ok := prevBlock.Attributes[name]; !ok {
			candAttr := candBlock.Attributes[name]
			fullPath := joinPath(pathPrefix, name)

			if candAttr.Required {
				// Newly required attribute on existing resource/block is BREAKING
				report.BreakingChanges = append(report.BreakingChanges, SchemaChange{
					Risk:       RiskBreaking,
					Type:       ChangeNewlyRequired,
					TargetKind: targetKind,
					TargetName: targetName,
					Path:       fullPath,
					Previous:   "(none)",
					Candidate:  candAttr.Descriptor(),
					Details:    fmt.Sprintf("New attribute %q is Required; existing configurations without it will fail validation", fullPath),
				})
			} else {
				report.CompatibleChanges = append(report.CompatibleChanges, SchemaChange{
					Risk:       RiskCompatible,
					Type:       ChangeAttributeAdded,
					TargetKind: targetKind,
					TargetName: targetName,
					Path:       fullPath,
					Previous:   "(none)",
					Candidate:  candAttr.Descriptor(),
					Details:    fmt.Sprintf("New backward-compatible attribute %q added", fullPath),
				})
			}
		}
	}

	// Blocks
	var prevBlockNames []string
	for name := range prevBlock.Blocks {
		prevBlockNames = append(prevBlockNames, name)
	}
	sort.Strings(prevBlockNames)

	for _, name := range prevBlockNames {
		prevChild := prevBlock.Blocks[name]
		fullPath := joinPath(pathPrefix, name)

		candChild, ok := candBlock.Blocks[name]
		if !ok {
			report.BreakingChanges = append(report.BreakingChanges, SchemaChange{
				Risk:       RiskBreaking,
				Type:       ChangeBlockRemoved,
				TargetKind: targetKind,
				TargetName: targetName,
				Path:       fullPath,
				Previous:   fmt.Sprintf("block %s", prevChild.NestingMode),
				Candidate:  "(removed)",
				Details:    fmt.Sprintf("Block %q was removed", fullPath),
			})
			continue
		}

		// Check if block was made newly required
		if prevChild.MinItems == 0 && candChild.MinItems > 0 {
			report.BreakingChanges = append(report.BreakingChanges, SchemaChange{
				Risk:       RiskBreaking,
				Type:       ChangeBlockModified,
				TargetKind: targetKind,
				TargetName: targetName,
				Path:       fullPath,
				Previous:   fmt.Sprintf("min_items=%d", prevChild.MinItems),
				Candidate:  fmt.Sprintf("min_items=%d", candChild.MinItems),
				Details:    fmt.Sprintf("Block %q now requires at least %d item(s)", fullPath, candChild.MinItems),
			})
		}

		if prevChild.Block != nil && candChild.Block != nil {
			diffBlocks(targetKind, targetName, fullPath, prevChild.Block, candChild.Block, report)
		}
	}

	var candBlockNames []string
	for name := range candBlock.Blocks {
		candBlockNames = append(candBlockNames, name)
	}
	sort.Strings(candBlockNames)

	for _, name := range candBlockNames {
		if _, ok := prevBlock.Blocks[name]; !ok {
			candChild := candBlock.Blocks[name]
			fullPath := joinPath(pathPrefix, name)

			if candChild.MinItems > 0 {
				report.BreakingChanges = append(report.BreakingChanges, SchemaChange{
					Risk:       RiskBreaking,
					Type:       ChangeBlockAdded,
					TargetKind: targetKind,
					TargetName: targetName,
					Path:       fullPath,
					Previous:   "(none)",
					Candidate:  fmt.Sprintf("block %s (min_items=%d)", candChild.NestingMode, candChild.MinItems),
					Details:    fmt.Sprintf("New block %q is required (min_items=%d)", fullPath, candChild.MinItems),
				})
			} else {
				report.CompatibleChanges = append(report.CompatibleChanges, SchemaChange{
					Risk:       RiskCompatible,
					Type:       ChangeBlockAdded,
					TargetKind: targetKind,
					TargetName: targetName,
					Path:       fullPath,
					Previous:   "(none)",
					Candidate:  fmt.Sprintf("block %s", candChild.NestingMode),
					Details:    fmt.Sprintf("New optional block %q added", fullPath),
				})
			}
		}
	}
}

func diffAttribute(targetKind, targetName, fullPath string, prevAttr, candAttr *NormalizedAttribute, report *ComparisonReport) {
	// 1. Type changed
	if prevAttr.Type != candAttr.Type {
		report.BreakingChanges = append(report.BreakingChanges, SchemaChange{
			Risk:       RiskBreaking,
			Type:       ChangeTypeModified,
			TargetKind: targetKind,
			TargetName: targetName,
			Path:       fullPath,
			Previous:   prevAttr.Descriptor(),
			Candidate:  candAttr.Descriptor(),
			Details:    fmt.Sprintf("Attribute %q type changed from %s to %s", fullPath, prevAttr.Type, candAttr.Type),
		})
	}

	// 2. Newly required
	if !prevAttr.Required && candAttr.Required {
		report.BreakingChanges = append(report.BreakingChanges, SchemaChange{
			Risk:       RiskBreaking,
			Type:       ChangeNewlyRequired,
			TargetKind: targetKind,
			TargetName: targetName,
			Path:       fullPath,
			Previous:   prevAttr.Descriptor(),
			Candidate:  candAttr.Descriptor(),
			Details:    fmt.Sprintf("Existing attribute %q was made Required", fullPath),
		})
	}

	// 3. Lost sensitive marking
	if prevAttr.Sensitive && !candAttr.Sensitive {
		report.BreakingChanges = append(report.BreakingChanges, SchemaChange{
			Risk:       RiskBreaking,
			Type:       ChangeLostSensitive,
			TargetKind: targetKind,
			TargetName: targetName,
			Path:       fullPath,
			Previous:   "sensitive=true",
			Candidate:  "sensitive=false",
			Details:    fmt.Sprintf("Attribute %q lost sensitive marking; values may be leaked in plaintext outputs", fullPath),
		})
	}

	// Compatible changes
	if !prevAttr.Sensitive && candAttr.Sensitive {
		report.CompatibleChanges = append(report.CompatibleChanges, SchemaChange{
			Risk:       RiskCompatible,
			Type:       ChangeSensitiveAdded,
			TargetKind: targetKind,
			TargetName: targetName,
			Path:       fullPath,
			Previous:   "sensitive=false",
			Candidate:  "sensitive=true",
			Details:    fmt.Sprintf("Attribute %q gained sensitive marking", fullPath),
		})
	}

	if !prevAttr.Deprecated && candAttr.Deprecated {
		report.CompatibleChanges = append(report.CompatibleChanges, SchemaChange{
			Risk:       RiskCompatible,
			Type:       ChangeDeprecated,
			TargetKind: targetKind,
			TargetName: targetName,
			Path:       fullPath,
			Previous:   "deprecated=false",
			Candidate:  "deprecated=true",
			Details:    fmt.Sprintf("Attribute %q marked as deprecated", fullPath),
		})
	}

	// Nested types diff
	if prevAttr.NestedType != nil && candAttr.NestedType != nil {
		diffNestedTypes(targetKind, targetName, fullPath, prevAttr.NestedType, candAttr.NestedType, report)
	} else if prevAttr.NestedType != nil && candAttr.NestedType == nil {
		report.BreakingChanges = append(report.BreakingChanges, SchemaChange{
			Risk:       RiskBreaking,
			Type:       ChangeTypeModified,
			TargetKind: targetKind,
			TargetName: targetName,
			Path:       fullPath,
			Previous:   fmt.Sprintf("nested_%s", prevAttr.NestedType.NestingMode),
			Candidate:  candAttr.Descriptor(),
			Details:    fmt.Sprintf("Attribute %q converted from nested type to primitive type", fullPath),
		})
	} else if prevAttr.NestedType == nil && candAttr.NestedType != nil {
		report.BreakingChanges = append(report.BreakingChanges, SchemaChange{
			Risk:       RiskBreaking,
			Type:       ChangeTypeModified,
			TargetKind: targetKind,
			TargetName: targetName,
			Path:       fullPath,
			Previous:   prevAttr.Descriptor(),
			Candidate:  fmt.Sprintf("nested_%s", candAttr.NestedType.NestingMode),
			Details:    fmt.Sprintf("Attribute %q converted from primitive type to nested type", fullPath),
		})
	}
}

func diffNestedTypes(targetKind, targetName, pathPrefix string, prevNT, candNT *NormalizedNestedType, report *ComparisonReport) {
	if prevNT.NestingMode != candNT.NestingMode {
		report.BreakingChanges = append(report.BreakingChanges, SchemaChange{
			Risk:       RiskBreaking,
			Type:       ChangeTypeModified,
			TargetKind: targetKind,
			TargetName: targetName,
			Path:       pathPrefix,
			Previous:   fmt.Sprintf("nesting_mode=%s", prevNT.NestingMode),
			Candidate:  fmt.Sprintf("nesting_mode=%s", candNT.NestingMode),
			Details:    fmt.Sprintf("Nested attribute %q nesting_mode changed from %s to %s", pathPrefix, prevNT.NestingMode, candNT.NestingMode),
		})
	}

	for name, prevAttr := range prevNT.Attributes {
		fullPath := joinPath(pathPrefix, name)
		candAttr, ok := candNT.Attributes[name]
		if !ok {
			report.BreakingChanges = append(report.BreakingChanges, SchemaChange{
				Risk:       RiskBreaking,
				Type:       ChangeAttributeRemoved,
				TargetKind: targetKind,
				TargetName: targetName,
				Path:       fullPath,
				Previous:   prevAttr.Descriptor(),
				Candidate:  "(removed)",
				Details:    fmt.Sprintf("Nested attribute %q was removed", fullPath),
			})
			continue
		}
		diffAttribute(targetKind, targetName, fullPath, prevAttr, candAttr, report)
	}

	for name, candAttr := range candNT.Attributes {
		if _, ok := prevNT.Attributes[name]; !ok {
			fullPath := joinPath(pathPrefix, name)
			if candAttr.Required {
				report.BreakingChanges = append(report.BreakingChanges, SchemaChange{
					Risk:       RiskBreaking,
					Type:       ChangeNewlyRequired,
					TargetKind: targetKind,
					TargetName: targetName,
					Path:       fullPath,
					Previous:   "(none)",
					Candidate:  candAttr.Descriptor(),
					Details:    fmt.Sprintf("New nested attribute %q is Required", fullPath),
				})
			} else {
				report.CompatibleChanges = append(report.CompatibleChanges, SchemaChange{
					Risk:       RiskCompatible,
					Type:       ChangeAttributeAdded,
					TargetKind: targetKind,
					TargetName: targetName,
					Path:       fullPath,
					Previous:   "(none)",
					Candidate:  candAttr.Descriptor(),
					Details:    fmt.Sprintf("New optional nested attribute %q added", fullPath),
				})
			}
		}
	}
}

func validateStabilityPolicy(report *ComparisonReport, opts CompareOptions) {
	// Policy rule 1: Breaking changes require an explicit major version bump
	prevMajor, errPrev := extractMajorVersion(opts.PreviousVersion)
	candMajor, errCand := extractMajorVersion(opts.TargetVersion)

	if errPrev == nil && errCand == nil {
		if candMajor <= prevMajor {
			msg := fmt.Sprintf("Stability Policy Violation: Breaking schema changes require a MAJOR version bump (previous version: %s, target version: %s).",
				opts.PreviousVersion, opts.TargetVersion)
			report.PolicyErrors = append(report.PolicyErrors, msg)
			report.PolicyPassed = false
		}
	} else if opts.TargetVersion != "" {
		// Target version was provided but not parseable as semver
		msg := fmt.Sprintf("Stability Policy Violation: Target version %q is not valid SemVer 2.0 format.", opts.TargetVersion)
		report.PolicyErrors = append(report.PolicyErrors, msg)
		report.PolicyPassed = false
	} else if !opts.AllowBreaking {
		// No target version provided and breaking not explicitly allowed
		msg := "Stability Policy Violation: Breaking schema changes detected without an approved major version decision."
		report.PolicyErrors = append(report.PolicyErrors, msg)
		report.PolicyPassed = false
	}

	// Policy rule 2: Breaking changes require an explicit reviewed migration note
	migrationNote := strings.TrimSpace(opts.MigrationNoteContent)
	if migrationNote == "" && opts.MigrationNotePath != "" {
		if data, err := os.ReadFile(opts.MigrationNotePath); err == nil {
			migrationNote = strings.TrimSpace(string(data))
		}
	}

	if migrationNote == "" {
		msg := "Stability Policy Violation: Breaking schema changes require an explicit reviewed migration note describing user impact and upgrade paths."
		report.PolicyErrors = append(report.PolicyErrors, msg)
		report.PolicyPassed = false
	}
}

func extractMajorVersion(v string) (int, error) {
	clean := strings.TrimPrefix(strings.TrimSpace(v), "v")
	matches := semverPattern.FindStringSubmatch(clean)
	if len(matches) < 2 {
		return 0, fmt.Errorf("invalid semver %q", v)
	}
	return strconv.Atoi(matches[1])
}

func joinPath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}
