package schemacmp

import (
	"fmt"
	"strings"
)

// GenerateMarkdownReport creates a comprehensive, human-readable PR markdown artifact.
func GenerateMarkdownReport(r *ComparisonReport) string {
	var sb strings.Builder

	sb.WriteString("# Public Provider Schema Compatibility Report\n\n")

	// Baseline or Status Banner
	if r.IsInitialBaseline {
		sb.WriteString("> [!NOTE]\n")
		sb.WriteString("> **Initial Baseline Release**: No previous release was found to compare against. All exported provider schemas form the new baseline contract.\n\n")
		return sb.String()
	}

	prevVer := r.PreviousVersion
	if prevVer == "" {
		prevVer = "unspecified"
	}
	candVer := r.CandidateVersion
	if candVer == "" {
		candVer = "candidate"
	}

	sb.WriteString(fmt.Sprintf("- **Baseline Release**: `%s`\n", prevVer))
	sb.WriteString(fmt.Sprintf("- **Candidate Version**: `%s`\n\n", candVer))

	if r.HasBreakingChanges {
		if r.PolicyPassed {
			sb.WriteString("> [!WARNING]\n")
			sb.WriteString("> **Breaking Schema Changes Detected (Policy Approved)**: Backward-incompatible schema changes are present. An explicit reviewed migration note was verified, and the version decision aligns with the major-bump stability policy.\n\n")
		} else {
			sb.WriteString("> [!CAUTION]\n")
			sb.WriteString("> **Incompatible Public Schema Changes Detected**: One or more public schema modifications violate backward compatibility and stability policy requirements.\n\n")
		}
	} else {
		sb.WriteString("> [!TIP]\n")
		sb.WriteString("> **100% Backward Compatible**: All candidate provider schemas preserve existing public contracts. No breaking removals, type modifications, newly required attributes, or lost sensitive markings detected.\n\n")
	}

	// Metrics Summary
	sb.WriteString("## Summary Metrics\n\n")
	sb.WriteString("| Metric | Count | Status |\n")
	sb.WriteString("| :--- | :--- | :--- |\n")
	if len(r.BreakingChanges) == 0 {
		sb.WriteString(fmt.Sprintf("| Incompatible Breaking Changes | %d | ✅ Clean |\n", len(r.BreakingChanges)))
	} else {
		sb.WriteString(fmt.Sprintf("| Incompatible Breaking Changes | %d | ❌ Action Required |\n", len(r.BreakingChanges)))
	}
	sb.WriteString(fmt.Sprintf("| Backward-Compatible Additions | %d | ✅ Permitted |\n", len(r.CompatibleChanges)))
	if r.HasBreakingChanges {
		if r.PolicyPassed {
			sb.WriteString("| Stability Policy Compliance | Pass | ✅ Major Bump & Migration Note Verified |\n")
		} else {
			sb.WriteString("| Stability Policy Compliance | Fail | ❌ Incompatible SemVer or Missing Note |\n")
		}
	} else {
		sb.WriteString("| Stability Policy Compliance | Pass | ✅ No Major Bump Needed |\n")
	}
	sb.WriteString("\n")

	// Breaking changes table
	if len(r.BreakingChanges) > 0 {
		sb.WriteString("## ⚠️ Incompatible Breaking Changes\n\n")
		sb.WriteString("| Target | Component | Attribute / Path | Change Type | Baseline | Candidate | Risk & Impact |\n")
		sb.WriteString("| :--- | :--- | :--- | :--- | :--- | :--- | :--- |\n")
		for _, c := range r.BreakingChanges {
			sb.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | `%s` | `%s` | `%s` | %s |\n",
				c.TargetKind, c.TargetName, c.Path, c.Type, c.Previous, c.Candidate, c.Details))
		}
		sb.WriteString("\n")
	}

	// Compatible changes table
	if len(r.CompatibleChanges) > 0 {
		sb.WriteString("## ✅ Backward-Compatible Additions\n\n")
		sb.WriteString("| Target | Component | Attribute / Path | Addition Type | Candidate Descriptor |\n")
		sb.WriteString("| :--- | :--- | :--- | :--- | :--- |\n")
		for _, c := range r.CompatibleChanges {
			sb.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | `%s` | `%s` |\n",
				c.TargetKind, c.TargetName, c.Path, c.Type, c.Candidate))
		}
		sb.WriteString("\n")
	}

	// Stability Policy Errors
	if len(r.PolicyErrors) > 0 {
		sb.WriteString("## 📋 Stability Policy Requirements\n\n")
		for _, err := range r.PolicyErrors {
			sb.WriteString(fmt.Sprintf("- ❌ %s\n", err))
		}
		sb.WriteString("\nTo merge intentional breaking changes:\n")
		sb.WriteString("1. Bump the candidate version to the next `MAJOR` release line (e.g. `v3.0.0`) in according with SemVer 2.0.\n")
		sb.WriteString("2. Author an explicit migration guide and supply a reviewed migration note explaining user upgrade paths.\n\n")
	}

	// Semantic Review and Upgrade Replay complementary requirements
	sb.WriteString("---\n\n")
	sb.WriteString("> [!IMPORTANT]\n")
	sb.WriteString("> **Complementary Verification Requirements**:\n")
	sb.WriteString("> Schema comparison verifies structural syntactic contracts (detecting removed resources, altered types, newly required fields, and lost sensitive markings). **Schema comparison alone does NOT prove behavioral or operational compatibility.**\n")
	sb.WriteString("> \n")
	sb.WriteString("> In accordance with repository release-confidence standards:\n")
	sb.WriteString("> - **Semantic Review**: Changes to attribute interpretations or defaults (e.g., `0` vs `null` quota limits) must be reviewed semantically and documented in upgrade notes.\n")
	sb.WriteString("> - **Consumer Upgrade Replay**: Real-world consumer configurations must pass fresh apply, clean re-plan, and upgrade state replay from the previous release via `scripts/test-consumer-replay.sh`.\n")

	return sb.String()
}
