# Version Compatibility Matrix

This document outlines the tested and supported versions of upstream media management servers (Seerr, Jellyseerr, Overseerr) and Infrastructure-as-Code engines (OpenTofu, Terraform).

---

## 📺 Upstream Server Compatibility

The `seerr` Terraform provider targets the Seerr / Jellyseerr / Overseerr REST API (`/api/v1`).

| Server Application | Supported Versions | Tested CI Baseline | Status |
| :--- | :--- | :--- | :--- |
| **Seerr (Unified)** | `v3.0.0` - `v3.4.1` | `seerr/seerr:v3.4.1` | **Tier 1 (Primary Target)** ✅ |
| **Jellyseerr (legacy)** | `v1.7.0` - `v2.7.3` | `fallenbagel/jellyseerr:2.7.3` | **Tier 1 (Legacy Target)** ✅ |
| **Overseerr** | `v1.33.2` - `v1.35.0` | `sctx/overseerr:1.35.0` | **Tier 2 (Supported)** ✅ |

### Verified Release Snapshot (2026-08-29)

Pull requests run OpenTofu integration coverage against all three pinned images above. Seerr and Jellyseerr run the stable provider suite; Overseerr runs the common Plex, ARR, request, issue, job, settings, user, and discovery subset because it does not implement Jellyfin/Emby-specific endpoints.

The Seerr `v3.4.1` release OpenAPI specification was compared with `tools/openapi/seerr-api.yml`. It adds no endpoints missing from the provider snapshot. The provider snapshot contains four forward-development library sync endpoints that are not present in the release specification, so no release-spec rollback is required.

Jellyseerr's GitHub repository was consolidated into Seerr after `v2.7.3`. GitHub's redirected `releases/latest` response therefore reports the current Seerr release and must not be interpreted as a new Jellyseerr release. The pinned `fallenbagel/jellyseerr:2.7.3` image is the legacy compatibility baseline.

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
      version = "~> 0.40.0"
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
      version = "~> 0.40.0"
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
   - Compares Seerr and Overseerr releases with `.github/upstream-compatibility.json` while retaining Jellyseerr `v2.7.3` as the final legacy baseline.
   - Opens a compatibility issue only when a release differs from the verified baseline.

3. **OpenAPI Coverage Audit (`tools/openapi/coverage.go`)**:
   - Validates that 100% of applicable configuration, user management, and service endpoints are mapped to provider resources and data sources.
   - Generated matrix is tracked in [`docs/openapi-coverage.md`](../openapi-coverage.md).

4. **Multi-Version Integration Test Matrix (`.github/workflows/test.yml`)**:
   - Executes dialect-aware integration tests against pinned Seerr, Jellyseerr, and Overseerr containers on every pull request.
