# Stability and Breaking-Change Policy

This document defines the formal API stability, backward-compatibility, and breaking-change policy for the `seerr` OpenTofu and Terraform provider starting with `v1.0.0`.

---

## 🎯 Semantic Versioning (SemVer 2.0) Contract

Starting with **`v1.0.0`**, the `seerr` provider strictly follows [Semantic Versioning 2.0.0](https://semver.org/). Version numbers follow the `MAJOR.MINOR.PATCH` format:

```text
v<MAJOR>.<MINOR>.<PATCH>
```

| Component | Trigger | Compatibility Commitment |
| :--- | :--- | :--- |
| **`MAJOR`** | Breaking schema changes, removed resources, or backwards-incompatible state modifications. | Incompatible API changes; migration guide and state upgraders provided. |
| **`MINOR`** | New managed resources, new data sources, new optional attributes, or backward-compatible capabilities. | 100% backward compatible with prior minor releases in the same major line. |
| **`PATCH`** | Bug fixes, documentation updates, performance improvements, and security patches. | 100% backward compatible; drop-in replacement. |

---

## 🔍 Breaking vs. Non-Breaking Taxonomy

To ensure zero unexpected plan drift or apply failures across version upgrades, all changes are categorized against the following taxonomy:

### ❌ Breaking Changes (Requires `MAJOR` Version Bump)

The following changes are strictly prohibited in `MINOR` and `PATCH` releases and may only occur across major version increments:

1. **Resource & Data Source Removal**:
   - Deleting or renaming an existing resource (e.g. `seerr_user`) or data source (e.g. `data.seerr_about`).
2. **Schema Attribute Removal or Renaming**:
   - Removing any existing attribute from a resource or data source schema.
   - Renaming an attribute without retaining an alias or deprecated fallback.
3. **Attribute Type Modifications**:
   - Changing the underlying framework type of an existing attribute (e.g., changing a `types.StringType` to `types.Int64Type` or `types.ListType`).
4. **Tightening Validation Rules**:
   - Adding stricter regex constraints, lower upper bounds, or reducing allowable string sets that cause previously valid configurations to fail validation during `terraform plan`.
5. **Adding Required Attributes**:
   - Introducing a new attribute marked `Required: true` without an automatic default, computed fallback, or migration path.
6. **Altering `ImportState` Identifiers**:
   - Changing the format, delimiter, or parsing behavior of import IDs (e.g., changing composite ID delimiters from `/` to `:`).
7. **Retiring State Upgraders Prematurely**:
   - Removing any active state upgrader (`ResourceWithUpgradeState`) for prior state schema versions before a major release boundary.
8. **Plan Modifier Inversions**:
   - Adding `RequiresReplace()` to an attribute that previously updated in place, forcing resource recreation on existing infrastructure.

---

### ✅ Non-Breaking Changes (Permitted in `MINOR` and `PATCH` Releases)

The following changes are backward-compatible and permitted in minor or patch releases:

1. **New Primitives**:
   - Introducing new resources, data sources, or utility functions.
2. **New Optional Attributes**:
   - Adding new attributes marked `Optional: true` or `Computed: true`.
3. **Relaxing Constraints**:
   - Expanding allowed value ranges, widening regex validators, or supporting new enum values.
4. **State Schema Upgraders**:
   - Adding new `StateUpgrader` functions to migrate legacy state shapes transparently without user intervention.
5. **Deprecation Warnings**:
   - Emitting diagnostic warnings on attributes planned for eventual sunset via `DeprecationMessage`.
6. **Error Message & Diagnostic Clarifications**:
   - Improving contextual information in diagnostic errors returned to the user.

---

## ⏳ Deprecation Lifecycle & Sunset Policy

Features and attributes are never removed abruptly. Any deprecation follows a structured two-stage lifecycle:

```mermaid
flowchart LR
    A["Active & Supported<br/>(Normal Usage)"] -->|Stage 1: Minor Release| B["Deprecated<br/>(Emits Warning, Fully Functional)"]
    B -->|Stage 2: Next Major Release (v2.0.0)| C["Removed<br/>(Clean Sunset)"]
```

### Stage 1: Deprecation Notice (Minor Release)
- The attribute or resource is marked with `DeprecationMessage` in its Plugin Framework schema definition.
- Terraform and OpenTofu emit a non-blocking diagnostic warning during `plan` and `apply` explaining the replacement pattern.
- The feature remains **fully functional** and supported for **at least two minor versions** (`>= 2 minor versions`) prior to removal.

### Stage 2: Removal (Major Release)
- The deprecated item is removed only at the next `MAJOR` release boundary (e.g., `v2.0.0`).
- A comprehensive **Migration & Upgrade Guide** accompanies the release, detailing before-and-after HCL examples.

---

## 🔄 State Schema Migration Policy

State resilience is guaranteed across version upgrades through the following rules:

1. **Automatic In-Place Upgrades**:
   - When a resource's internal state schema changes, the provider increments its schema version and registers a state upgrader function implementing `resource.ResourceWithUpgradeState`.
   - Running `terraform plan` or `tofu plan` automatically migrates stored state to the latest schema version without requiring `terraform state rm` or manual JSON edits.
2. **Upgrader Retention**:
   - All state upgraders are retained across the entire major version lifecycle (`v1.x.x`). A user upgrading from `v1.0.0` to `v1.15.0` will experience smooth, chained state migrations.
3. **Semantic Diff Suppression**:
   - Computed attributes and list orderings are normalized during `Read` to prevent false drift when upstream APIs return unordered sets or omitted defaults.

---

## 🌐 Upstream Dialect & Drift Policy

The provider maintains compatibility with **Seerr**, **Jellyseerr**, and **Overseerr**:

1. **Dialect Graceful Degradation**:
   - If a resource target is unsupported by an upstream dialect (e.g., managing Jellyfin settings on an Overseerr instance), the provider returns a descriptive diagnostic error identifying the target dialect limitation rather than an unhandled panic.
2. **Upstream Schema Drift Guardrails**:
   - Weekly automated OpenAPI diff checks (`.github/workflows/openapi-drift.yml`) monitor upstream API changes.
   - Daily release watchers (`.github/workflows/upstream-watch.yml`) verify tag parity.
3. **Forward Compatibility Primitives**:
   - Generic fallback resources (`seerr_api_object`, `seerr_api_request`) remain available so users can automate new or experimental upstream endpoints before dedicated typed resources are released.

---

## 📋 Summary Checklist for Contributors

When submitting changes, ensure your PR complies with:

- [ ] Does this change remove, rename, or alter the type of any existing schema attribute? *(If yes, reject or re-architect as backward-compatible).*
- [ ] Does this change add a `Required: true` field to an existing resource? *(If yes, make it `Optional` with a sensible default).*
- [ ] Does this change modify `ImportState` parsing logic? *(If yes, preserve backward-compatible ID formats).*
- [ ] Are deprecations accompanied by a `DeprecationMessage` pointing to the replacement attribute?
- [ ] Have all unit tests and `test-all-locally.sh` checks passed with zero regressions?
