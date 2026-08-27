# Version Compatibility Matrix

This document outlines the tested and supported versions of upstream media management servers (Seerr, Jellyseerr, Overseerr) and Infrastructure-as-Code engines (OpenTofu, Terraform).

---

## 📺 Upstream Server Compatibility

The `seerr` Terraform provider targets the Seerr / Jellyseerr / Overseerr REST API (`/api/v1`).

| Server Application | Supported Versions | Tested CI Baseline | Status |
| :--- | :--- | :--- | :--- |
| **Seerr (Unified)** | `v3.0.0` - `v3.4.1`+ | `seerr/seerr:v3.1.1` (verified `v3.4.1`) | **Tier 1 (Primary Target)** ✅ |
| **Jellyseerr** | `v1.7.0` - `v3.4.1`+ | `fallenbagel/jellyseerr:latest` (`v3.4.1`) | **Tier 1 (Fully Supported)** ✅ |
| **Overseerr** | `v1.33.2` - `v1.35.0`+ | `sct/overseerr:latest` (`v1.35.0`) | **Tier 2 (Supported)** ✅ |

### Upstream Feature Mapping & API Dialects

1. **Jellyfin & Emby Integration**:
   - Supported on **Seerr** and **Jellyseerr**.
   - Overseerr does not support Jellyfin/Emby settings endpoints (`/settings/jellyfin`, `/settings/emby`); attempting to manage these resources against an Overseerr instance will return an upstream 404.
2. **Plex Integration**:
   - Supported across all three variants (**Seerr**, **Jellyseerr**, **Overseerr**).
3. **Override Rules & Discover Sliders**:
   - Advanced routing by user roles, tags, and languages is supported on Seerr and Jellyseerr.

---

## 🛠️ Infrastructure as Code (IaC) Engine Compatibility

The provider is built on the official **Terraform Plugin Framework (Protocol v6)**.

| IaC Tool | Supported Versions | Protocol Version | Registry Namespace |
| :--- | :--- | :--- | :--- |
| **OpenTofu** | `>= 1.6.0` (tested `1.8.x` - `1.11.x`) | `Protocol 6.0` | `registry.opentofu.org/josh-archer/seerr` |
| **HashiCorp Terraform** | `>= 1.5.0` (tested `1.5.x` - `1.11.x`) | `Protocol 6.0` | `registry.terraform.io/josh-archer/seerr` |

### OpenTofu Configuration
```hcl
terraform {
  required_version = ">= 1.6.0"
  required_providers {
    seerr = {
      source  = "registry.opentofu.org/josh-archer/seerr"
      version = "~> 0.38.0"
    }
  }
}
```

### Terraform Configuration
```hcl
terraform {
  required_version = ">= 1.5.0"
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "~> 0.38.0"
    }
  }
}
```

---

## 🤖 Automated Upstream Compatibility Guardrails

To prevent silent schema drift and ensure continuous compatibility as upstream projects release new features:

1. **Scheduled OpenAPI Drift Detector (`.github/workflows/openapi-drift.yml`)**:
   - Runs every Monday at 04:00 UTC.
   - Compares the provider's local OpenAPI schema (`tools/openapi/seerr-api.yml`) against upstream `seerr-team/seerr/develop/seerr-api.yml`.
   - Automatically opens a categorized GitHub tracking issue whenever new endpoints or schema modifications are detected.

2. **Upstream Release Watcher (`.github/workflows/upstream-watch.yml`)**:
   - Runs daily at 06:00 UTC.
   - Monitors upstream GitHub releases from `seerr-team/seerr`, `Fallenbagel/jellyseerr`, and `sct/overseerr`.
   - Flags new upstream versions for automated regression and integration testing.

3. **OpenAPI Coverage Audit (`tools/openapi/coverage.go`)**:
   - Validates that 100% of applicable configuration, user management, and service endpoints are mapped to provider resources and data sources.
   - Generated matrix is tracked in [`docs/openapi-coverage.md`](../openapi-coverage.md).

4. **Multi-Version Integration Test Matrix (`.github/workflows/test.yml`)**:
   - Executes integration tests against containerized upstream servers on every push and pull request.
