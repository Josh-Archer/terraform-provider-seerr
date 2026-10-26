package schemacmp

import (
	"fmt"
	"strings"
)

// ChangeRisk classifies the backward-compatibility risk level of a schema change.
type ChangeRisk string

const (
	RiskNone       ChangeRisk = "NONE"
	RiskCompatible ChangeRisk = "COMPATIBLE"
	RiskBreaking   ChangeRisk = "BREAKING"
)

// ChangeType specifies the precise nature of the schema modification.
type ChangeType string

const (
	ChangeResourceRemoved   ChangeType = "RESOURCE_REMOVED"
	ChangeResourceAdded     ChangeType = "RESOURCE_ADDED"
	ChangeDataSourceRemoved ChangeType = "DATA_SOURCE_REMOVED"
	ChangeDataSourceAdded   ChangeType = "DATA_SOURCE_ADDED"
	ChangeAttributeRemoved  ChangeType = "ATTRIBUTE_REMOVED"
	ChangeAttributeAdded    ChangeType = "ATTRIBUTE_ADDED"
	ChangeTypeModified      ChangeType = "TYPE_MODIFIED"
	ChangeNewlyRequired     ChangeType = "NEWLY_REQUIRED"
	ChangeLostSensitive     ChangeType = "LOST_SENSITIVE"
	ChangeSensitiveAdded    ChangeType = "SENSITIVE_ADDED"
	ChangeDeprecated        ChangeType = "DEPRECATED"
	ChangeBlockRemoved      ChangeType = "BLOCK_REMOVED"
	ChangeBlockAdded        ChangeType = "BLOCK_ADDED"
	ChangeBlockModified     ChangeType = "BLOCK_MODIFIED"
)

// NormalizedSchema represents a deterministic, documentation-stripped provider schema.
type NormalizedSchema struct {
	FormatVersion string                      `json:"format_version"`
	Provider      *NormalizedBlock            `json:"provider,omitempty"`
	Resources     map[string]*NormalizedBlock `json:"resources"`
	DataSources   map[string]*NormalizedBlock `json:"data_sources"`
}

// NormalizedBlock represents the attributes and child blocks of a resource, data source, or provider.
type NormalizedBlock struct {
	Attributes map[string]*NormalizedAttribute `json:"attributes"`
	Blocks     map[string]*NormalizedBlockType `json:"blocks,omitempty"`
}

// NormalizedAttribute defines a schema attribute without documentation or ordering variance.
type NormalizedAttribute struct {
	Type       string                `json:"type"`
	Required   bool                  `json:"required"`
	Optional   bool                  `json:"optional"`
	Computed   bool                  `json:"computed"`
	Sensitive  bool                  `json:"sensitive"`
	Deprecated bool                  `json:"deprecated,omitempty"`
	NestedType *NormalizedNestedType `json:"nested_type,omitempty"`
}

// NormalizedNestedType describes framework-nested attribute structures.
type NormalizedNestedType struct {
	NestingMode string                          `json:"nesting_mode"`
	Attributes  map[string]*NormalizedAttribute `json:"attributes"`
}

// NormalizedBlockType describes child configuration blocks.
type NormalizedBlockType struct {
	NestingMode string           `json:"nesting_mode"`
	MinItems    int64            `json:"min_items,omitempty"`
	MaxItems    int64            `json:"max_items,omitempty"`
	Block       *NormalizedBlock `json:"block"`
}

// SchemaChange records a detected schema variance between two versions.
type SchemaChange struct {
	Risk       ChangeRisk `json:"risk"`
	Type       ChangeType `json:"type"`
	TargetKind string     `json:"target_kind"` // "provider", "resource", or "data_source"
	TargetName string     `json:"target_name"` // e.g. "seerr_user"
	Path       string     `json:"path"`        // e.g. "email" or "server.port"
	Previous   string     `json:"previous"`    // Human-readable old descriptor
	Candidate  string     `json:"candidate"`   // Human-readable new descriptor
	Details    string     `json:"details"`     // Contextual rationale / impact
}

func (c SchemaChange) String() string {
	return fmt.Sprintf("[%s] %s %s (%s): %s -> %s (%s)",
		c.Risk, c.TargetKind, c.TargetName, c.Path, c.Previous, c.Candidate, c.Details)
}

// CompareOptions configures schema comparison and policy enforcement.
type CompareOptions struct {
	AllowBreaking          bool
	TargetVersion          string
	PreviousVersion        string
	MigrationNotePath      string
	MigrationNoteContent   string
	RequireStabilityPolicy bool
}

// ComparisonReport captures the complete result of comparing two provider schemas.
type ComparisonReport struct {
	PreviousVersion     string         `json:"previous_version"`
	CandidateVersion    string         `json:"candidate_version"`
	IsInitialBaseline   bool           `json:"is_initial_baseline"`
	HasBreakingChanges  bool           `json:"has_breaking_changes"`
	BreakingChanges     []SchemaChange `json:"breaking_changes,omitempty"`
	CompatibleChanges   []SchemaChange `json:"compatible_changes,omitempty"`
	IgnoredChangesCount int            `json:"ignored_changes_count"`
	PolicyPassed        bool           `json:"policy_passed"`
	PolicyErrors        []string       `json:"policy_errors,omitempty"`
}

// IsCompatible returns true if no breaking changes are present or if breaking changes are approved per policy.
func (r *ComparisonReport) IsCompatible() bool {
	if r.IsInitialBaseline {
		return true
	}
	if !r.HasBreakingChanges {
		return true
	}
	return r.PolicyPassed
}

func (a *NormalizedAttribute) Descriptor() string {
	var flags []string
	if a.Required {
		flags = append(flags, "required")
	}
	if a.Optional {
		flags = append(flags, "optional")
	}
	if a.Computed {
		flags = append(flags, "computed")
	}
	if a.Sensitive {
		flags = append(flags, "sensitive")
	}
	if a.Deprecated {
		flags = append(flags, "deprecated")
	}
	t := a.Type
	if t == "" && a.NestedType != nil {
		t = fmt.Sprintf("nested_%s", a.NestedType.NestingMode)
	}
	if len(flags) > 0 {
		return fmt.Sprintf("%s (%s)", t, strings.Join(flags, ", "))
	}
	return t
}
