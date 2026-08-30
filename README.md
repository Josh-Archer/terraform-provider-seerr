# OpenTofu & Terraform Provider for Seerr

[![CI](https://github.com/Josh-Archer/terraform-provider-seerr/actions/workflows/test.yml/badge.svg)](https://github.com/Josh-Archer/terraform-provider-seerr/actions/workflows/test.yml)
[![OpenTofu Registry](https://img.shields.io/badge/OpenTofu-Registry-FF5722?logo=opentofu&logoColor=white)](https://registry.opentofu.org/providers/josh-archer/seerr/latest)
[![Terraform Registry](https://img.shields.io/badge/Terraform-Registry-844FBA?logo=terraform&logoColor=white)](https://registry.terraform.io/providers/josh-archer/seerr/latest)
[![Roadmap](https://img.shields.io/badge/Roadmap-v1.0.0-2ea44f.svg)](#roadmap)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

`seerr` is a feature-complete OpenTofu and Terraform provider for managing [Overseerr](https://overseerr.dev) and [Jellyseerr](https://github.com/Fallenbagel/jellyseerr) instances via their REST APIs as Infrastructure as Code. Track ongoing milestones in the [roadmap](#roadmap) or join the conversation in [GitHub Discussion #161](https://github.com/Josh-Archer/terraform-provider-seerr/discussions/161).

---

## Capabilities & Highlights

- **Complete API Coverage**: 50 managed resources and 69 data sources covering 100% of applicable Seerr OpenAPI configuration endpoints (see [OpenAPI Coverage](docs/openapi-coverage.md)).
- **Typed Notification Integrations**: First-class support for 11 notification agents (Discord, Email, Gotify, LunaSea, Ntfy, Pushbullet, Pushover, Slack, Telegram, Webhook, Webpush) with live test triggers (`seerr_notification_agent_test`).
- **Dynamic Servarr Resolvers**: Real-time lookups for Radarr & Sonarr quality profiles, root folders, tags, and language profiles (`seerr_radarr_quality_profile`, `seerr_sonarr_quality_profile`, `seerr_radarr_root_folders`, etc.) eliminating hardcoded IDs.
- **Media Server Integrations**: Plex, Jellyfin, and Emby server configuration, library scan triggers (`seerr_plex_library_sync`, etc.), and batch user import workflows.
- **Filterable Queries & Bulk Data Sources**: Query and filter requests, media items, issues, user lists, invitations, and background job schedules.
- **Granular Permissions & Quotas**: Bitmask permission generator (`seerr_permission_set`), user quotas (`seerr_user_quota`), watchlist sync settings, and request override rules.
- **Drift Protection & Schema Fidelity**: Double-apply drift protection, schema normalization, and generic fallback primitives (`seerr_api_object`, `seerr_api_request`) for arbitrary endpoints.

---

## Roadmap

The latest stable release is [`v0.43.0`](https://github.com/Josh-Archer/terraform-provider-seerr/releases/tag/v0.43.0). The reusable module ecosystem is complete. The remaining planned milestone is `v1.0.0` general availability.

This section is the canonical roadmap. Use [GitHub Discussion #161](https://github.com/Josh-Archer/terraform-provider-seerr/discussions/161) for feedback, proposals, and status discussion.

| Milestone | Target | Outcome | Status |
| :--- | :--- | :--- | :--- |
| **Phases 1-8: Provider foundation** | `v0.31.0`-`v0.38.2` | Core quality, observability, identity and access, reference helpers, lifecycle hardening, Servarr resolvers, dual-registry publishing, and release automation. | ✅ Complete |
| **Phase 9: Community readiness** | `v0.39.0` | Contributor onboarding, issue and pull request templates, and reproducible development environments ([#183](https://github.com/Josh-Archer/terraform-provider-seerr/issues/183)). | ✅ Complete |
| **Phase 10: Import and migration tooling** | `v0.40.0` | Bulk importer that generates Terraform HCL and import blocks from a live Seerr instance ([#170](https://github.com/Josh-Archer/terraform-provider-seerr/issues/170)). | ✅ Complete |
| **Phase 11: State resilience** | `v0.41.0` | Graceful 404 removal, state schema resilience, and semantic diff normalization ([#169](https://github.com/Josh-Archer/terraform-provider-seerr/issues/169)). | ✅ Complete |
| **Phase 12: Observability and disaster recovery** | `v0.42.0` | Drift monitoring, Prometheus and Grafana assets, backup workflows, and recovery runbooks ([#184](https://github.com/Josh-Archer/terraform-provider-seerr/issues/184)). | ✅ Complete |
| **Phase 13: Module ecosystem** | `v0.43.0` | Reusable modules for family media servers, ARR stack bootstrapping, and monitoring ([#171](https://github.com/Josh-Archer/terraform-provider-seerr/issues/171)). | ✅ Complete |
| **Phase 14: General availability** | `v1.0.0` | Stable public contracts, a documented breaking-change policy, and full Terraform/OpenTofu acceptance coverage. | 📅 Planned |

### `v1.0.0` exit criteria

- Publish and enforce the compatibility and breaking-change policy for public provider schemas and import identifiers.
- Pass the supported Terraform and OpenTofu acceptance matrix against the documented Seerr, Jellyseerr, and Overseerr versions.
- Validate installation and signed release artifacts through both the Terraform and OpenTofu registries.
- Complete upgrade, migration, module, and recovery documentation for a stable release.

---

## Quick Start

### Provider Installation

#### OpenTofu
```hcl
terraform {
  required_providers {
    seerr = {
      source  = "registry.opentofu.org/josh-archer/seerr"
      version = "~> 0.42.0"
    }
  }
}
```

#### Terraform
```hcl
terraform {
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "~> 0.42.0"
    }
  }
}
```

### Provider Configuration

```hcl
# Standard setup using an API key
provider "seerr" {
  url                  = "https://seerr.example.com"
  api_key              = var.seerr_api_key
  insecure_skip_verify = false
}

# Optional: Initial setup bootstrap using a Plex admin token
# provider "seerr" {
#   url        = "https://seerr.example.com"
#   plex_token = var.plex_token
# }
```

| Attribute | Type | Default | Environment Variable | Description |
|---|---|---|---|---|
| `url` | String | *Required* | `SEERR_URL` | Base URL of Seerr instance (e.g. `https://seerr.example.com`). |
| `api_key` | String (Sensitive) | Optional* | `SEERR_API_KEY` | Seerr API Key for `X-Api-Key` header authentication. |
| `plex_token` | String (Sensitive) | Optional* | `SEERR_PLEX_TOKEN` | Plex admin token to bootstrap the initial API key on first run. |
| `insecure_skip_verify` | Boolean | `false` | `SEERR_INSECURE_SKIP_VERIFY` | Skip TLS certificate verification. |
| `max_retries` | Number | `3` | `SEERR_MAX_RETRIES` | Max retries for transient errors and rate limits (429/502/503/504). |
| `retry_backoff_seconds` | Number | `1` | `SEERR_RETRY_BACKOFF_SECONDS` | Base backoff delay in seconds between retries. |
| `request_timeout_seconds` | Number | `120` | `SEERR_REQUEST_TIMEOUT_SECONDS` | HTTP timeout in seconds for API calls and ARR lookups. |
| `user_agent` | String | Auto | — | Custom `User-Agent` header. |

*\*Either `api_key` or `plex_token` is required.*

---

## Usage Examples

### 1. Servarr Integrations with Dynamic Resolvers

Automatically resolve quality profile IDs by name directly from Radarr and Sonarr:

```hcl
# Query Radarr for quality profile ID
data "seerr_radarr_quality_profile" "hd_1080p" {
  url     = "http://radarr.media.svc.cluster.local:7878"
  api_key = var.radarr_api_key
  name    = "HD-1080p"
}

resource "seerr_radarr_server" "movies" {
  name                   = "Radarr Movies"
  hostname               = "radarr.media.svc.cluster.local"
  port                   = 7878
  api_key                = var.radarr_api_key
  use_ssl                = false
  quality_profile_id     = data.seerr_radarr_quality_profile.hd_1080p.quality_profile_id
  active_directory       = "/media/movies"
  is_default             = true
  enable_scan            = true
  tag_requests_with_user = true
}

# Query Sonarr for quality profile ID
data "seerr_sonarr_quality_profile" "hd_1080p" {
  url     = "http://sonarr.media.svc.cluster.local:8989"
  api_key = var.sonarr_api_key
  name    = "HD-1080p"
}

resource "seerr_sonarr_server" "tv" {
  name                   = "Sonarr Shows"
  hostname               = "sonarr.media.svc.cluster.local"
  port                   = 8989
  api_key                = var.sonarr_api_key
  use_ssl                = false
  quality_profile_id     = data.seerr_sonarr_quality_profile.hd_1080p.quality_profile_id
  active_directory       = "/media/tv"
  active_anime_directory = "/media/anime"
  is_default             = true
  enable_scan            = true
  tag_requests_with_user = true
}
```

### 2. Typed Notification Integrations

```hcl
resource "seerr_notification_discord" "alerts" {
  enabled      = true
  embed_poster = true
  notification_types = [
    "MEDIA_PENDING",
    "MEDIA_APPROVED",
    "MEDIA_AVAILABLE",
    "ISSUE_CREATED"
  ]

  discord {
    webhook_url = var.discord_webhook_url
    bot_username = "Seerr Notifier"
  }
}

resource "seerr_notification_ntfy" "push" {
  enabled      = true
  embed_poster = true
  notification_types = ["MEDIA_APPROVED", "MEDIA_AVAILABLE"]

  ntfy {
    url      = "https://ntfy.example.com"
    topic    = "media-alerts"
    token    = var.ntfy_access_token
    priority = 3
  }
}
```

### 3. User Management with Permission Bitmasks

```hcl
# Compute permission bitmask declaratively
data "seerr_permission_set" "standard_user" {
  request       = true
  request_movie = true
  request_tv    = true
  auto_approve  = false
  manage_users  = false
}

resource "seerr_user" "jane" {
  email       = "jane@example.com"
  username    = "jane"
  permissions = data.seerr_permission_set.standard_user.permissions
}

resource "seerr_user_quota" "jane_quota" {
  user_id           = seerr_user.jane.id
  movie_quota_limit = 5
  movie_quota_days  = 7
  tv_quota_limit    = 3
  tv_quota_days     = 7
}
```

### 4. Core & Media Server Settings

```hcl
resource "seerr_main_settings" "main" {
  app_title       = "My Homelab Seerr"
  application_url = "https://seerr.example.com"
  locale          = "en"
  hide_available  = true
}

resource "seerr_plex_settings" "plex" {
  ip      = "plex.media.svc.cluster.local"
  port    = 32400
  use_ssl = false
}
```

---

## Modules & Composite Examples

Modules are optional convenience layers over the provider's resources. Existing direct-resource configurations continue to work unchanged. For a remote module, pin the repository tag so its implementation stays aligned with a known provider release:

```hcl
module "family" {
  source = "github.com/Josh-Archer/terraform-provider-seerr//modules/family_media_server?ref=v0.42.1"

  users = {
    child = {
      username       = "Child"
      email          = "child@example.com"
      permission_set = "standard"
    }
  }
}
```

See the [module usage guide](docs/guides/modules.md) for simple family, ARR, and monitoring examples, provider inheritance, and state migration.

### Packaged Sub-Modules
- [`modules/main_settings`](modules/main_settings): Standard application branding, locale, and general settings.
- [`modules/plex_settings`](modules/plex_settings): Plex media server connection and library settings.
- [`modules/radarr_server`](modules/radarr_server): Radarr movie integration module.
- [`modules/sonarr_server`](modules/sonarr_server): Sonarr series integration module.
- [`modules/family_media_server`](modules/family_media_server): Family users, permission tiers, quotas, notification routing, and curated discover sliders.
- [`modules/arr_stack_bootstrap`](modules/arr_stack_bootstrap): Keyed Sonarr/Radarr quality routes, override rules, and webhook notifications.
- [`modules/monitoring`](modules/monitoring): Seerr server, request, and media metrics for Prometheus and Grafana consumption.

### Reference Architectures
- [**Full Media Stack Module** (`examples/modules/full-media-stack`)](examples/modules/full-media-stack): Complete composite module integrating Seerr with Plex/Jellyfin, Radarr, Sonarr, Anime routing rules, curated Discover sliders, and Discord notifications.
- [**ARR Integration Module** (`examples/modules/seerr_arr_integration`)](examples/modules/seerr_arr_integration): Turnkey ARR integration module with automatic quality profile resolution.
- [**Complete Media Stack Example** (`examples/complete_media_stack`)](examples/complete_media_stack): Standalone root module demonstrating full homelab automation.

---

## Resource & Data Source Inventory

### Resources (50)
| Category | Resources |
|---|---|
| **Core Settings** | `seerr_main_settings`, `seerr_network_settings`, `seerr_backup_settings`, `seerr_metadata_settings`, `seerr_api_key`, `seerr_job_schedule`, `seerr_job_run` |
| **Media Servers** | `seerr_plex_settings`, `seerr_plex_library_settings`, `seerr_plex_library_sync`, `seerr_jellyfin_settings`, `seerr_jellyfin_library_settings`, `seerr_jellyfin_library_sync`, `seerr_emby_settings`, `seerr_emby_library_settings`, `seerr_emby_library_sync`, `seerr_tautulli_settings` |
| **Servarr & Routing** | `seerr_radarr_server`, `seerr_sonarr_server`, `seerr_override_rule`, `seerr_blocklist` |
| **Notifications** | `seerr_notification_discord`, `seerr_notification_email`, `seerr_notification_gotify`, `seerr_notification_lunasea`, `seerr_notification_ntfy`, `seerr_notification_pushbullet`, `seerr_notification_pushover`, `seerr_notification_slack`, `seerr_notification_telegram`, `seerr_notification_webhook`, `seerr_notification_webpush`, `seerr_notification_agent_test` |
| **Users & Permissions** | `seerr_user`, `seerr_user_invitation`, `seerr_user_permissions`, `seerr_user_settings_permissions`, `seerr_user_quota`, `seerr_user_watchlist_settings`, `seerr_user_notification_settings`, `seerr_user_import_plex`, `seerr_user_import_jellyfin` |
| **Media & Requests** | `seerr_request`, `seerr_request_approval`, `seerr_request_retry`, `seerr_issue`, `seerr_issue_comment`, `seerr_watchlist`, `seerr_discover_slider` |
| **Generic Primitives** | `seerr_api_object` |

### Data Sources (69)
| Category | Data Sources |
|---|---|
| **System & Discovery** | `seerr_about`, `seerr_service_status`, `seerr_public_settings`, `seerr_main_settings`, `seerr_network_settings`, `seerr_backup_settings`, `seerr_metadata_settings`, `seerr_discover_slider`, `seerr_discover`, `seerr_genres`, `seerr_languages`, `seerr_regions`, `seerr_jobs` |
| **Dynamic Resolvers** | `seerr_radarr_quality_profile`, `seerr_radarr_root_folders`, `seerr_radarr_tags`, `seerr_radarr_server`, `seerr_sonarr_quality_profile`, `seerr_sonarr_root_folders`, `seerr_sonarr_tags`, `seerr_sonarr_language_profiles`, `seerr_sonarr_server` |
| **Media Server Lookups** | `seerr_plex_settings`, `seerr_plex_devices`, `seerr_plex_users`, `seerr_plex_library_settings`, `seerr_jellyfin_settings`, `seerr_jellyfin_users`, `seerr_jellyfin_library_settings`, `seerr_emby_settings`, `seerr_emby_library_settings`, `seerr_tautulli_settings`, `seerr_user_import_plex`, `seerr_user_import_jellyfin` |
| **Notifications** | `seerr_notification_agents`, `seerr_notification_discord`, `seerr_notification_email`, `seerr_notification_gotify`, `seerr_notification_lunasea`, `seerr_notification_ntfy`, `seerr_notification_pushbullet`, `seerr_notification_pushover`, `seerr_notification_slack`, `seerr_notification_telegram`, `seerr_notification_webhook`, `seerr_notification_webpush`, `seerr_email_settings`, `seerr_pushbullet_settings`, `seerr_pushover_sounds` |
| **Users & Permissions** | `seerr_user`, `seerr_users`, `seerr_current_user`, `seerr_user_permissions`, `seerr_permission_set`, `seerr_user_quota`, `seerr_user_watchlist_settings`, `seerr_user_notification_settings`, `seerr_user_invitations` |
| **Media, Requests & Issues** | `seerr_media`, `seerr_media_item`, `seerr_request`, `seerr_requests`, `seerr_issue`, `seerr_issues`, `seerr_blocklist`, `seerr_override_rule`, `seerr_watchlist` |
| **Generic Primitives** | `seerr_api_key`, `seerr_api_request` |

Detailed documentation and examples for every resource and data source are in [`docs/resources`](docs/resources) and [`docs/data-sources`](docs/data-sources).

---

## Compatibility & Version Support

The provider maintains continuous compatibility tracking with upstream media management servers:

| Target | Supported Versions | Tested CI Baseline |
| :--- | :--- | :--- |
| **Seerr (Unified)** | `v3.0.0` - `v3.4.1` | `seerr/seerr:v3.4.1` |
| **Jellyseerr (legacy)** | `v1.7.0` - `v2.7.3` | `fallenbagel/jellyseerr:2.7.3` |
| **Overseerr** | `v1.33.2` - `v1.35.0` | `sctx/overseerr:1.35.0` |
| **OpenTofu** | `>= 1.6.0` (tested `1.8.x` - `1.11.x`) | Protocol 6.0 |
| **Terraform** | `>= 1.5.0` (tested `1.5.x` - `1.11.x`) | Protocol 6.0 |

See the complete [Version Compatibility Guide](docs/guides/compatibility.md) for architectural details, dialect mappings, and automated drift monitoring.

---

## Bulk Import & Migration CLI

Adopting the provider on an **existing, live** Seerr instance? You don't need to write Terraform/OpenTofu HCL by hand.

Use the built-in bulk importer CLI to discover your live configuration and generate idiomatic HCL resources + `import { ... }` blocks:

```bash
# Point the importer at your live Seerr server
go run ./tools/importer \
  --url "http://localhost:5055" \
  --api-key "YOUR_SEERR_API_KEY" \
  --out-dir "./seerr-tf"
```

The importer inspects all settings, Radarr/Sonarr servers, notification agents, override rules, and users, generating ready-to-apply `main.tf` and `imports.tf` files.

Read the step-by-step [Migration & Adoption Guide](docs/guides/migration.md) for full instructions.

---

## Development & Verification

### Local Testing

Run the full local validation fast gate:

```bash
bash ./scripts/test-all-locally.sh
```

Run unit tests directly:

```bash
go test -v ./...
```

Run OpenAPI coverage and drift analysis:

```bash
# Verify coverage mapping
go test -v ./tools/openapi/...
go run ./tools/openapi

# Check schema drift against live upstream
go run ./tools/openapi diff
```

### Pre-Push Hook

Enable the tracked pre-push hook to enforce code generation and formatting before push:

```bash
git config core.hooksPath .githooks
```

### Release & Publish Automation

- **Automated Releases**: Version releases are managed automatically via [Release Please](https://github.com/googleapis/release-please) and GitHub Actions.
- **Release Reconciliation**: Orphaned or skipped release tags are automatically reconciled through `.github/workflows/reconcile-releases.yml`.
- **GPG Signing**: Official release binaries are signed with RSA 4096-bit keys and published across the OpenTofu Registry and HashiCorp Terraform Registry.

---

## Community & Contributing

We welcome contributions from the community!

- **Contributing Guide**: See [CONTRIBUTING.md](CONTRIBUTING.md) for a step-by-step "Your First PR" walkthrough, local development setup, and coding standards.
- **Community Discussions**: Ask questions or share configurations in [GitHub Discussions](https://github.com/Josh-Archer/terraform-provider-seerr/discussions).
- **Issue Templates**: Report bugs using [Bug Report](.github/ISSUE_TEMPLATE/bug_report.yml) or suggest features with [Feature Request](.github/ISSUE_TEMPLATE/feature_request.yml).
- **Security Policy**: See [SECURITY.md](SECURITY.md) for vulnerability reporting guidelines.
- **Code of Conduct**: All participants are expected to follow our [Code of Conduct](CODE_OF_CONDUCT.md).

---

## License

MIT License. See [LICENSE](LICENSE) for details.
