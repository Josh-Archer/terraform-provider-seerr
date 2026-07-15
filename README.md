# OpenTofu Provider for Seerr

`seerr` is a community OpenTofu provider that manages Seerr via its API.

The provider supports two usage styles:
- Generic API primitives (`seerr_api_object` resource and `seerr_api_request` data source).
- Typed Seerr primitives for common integrations and settings.

## Typed resources

- `seerr_main_settings`
- `seerr_plex_settings`
- `seerr_notification_discord`, `seerr_notification_email`, `seerr_notification_gotify`, `seerr_notification_lunasea`, `seerr_notification_ntfy`, `seerr_notification_pushbullet`, `seerr_notification_pushover`, `seerr_notification_slack`, `seerr_notification_telegram`, `seerr_notification_webhook`, `seerr_notification_webpush`
- `seerr_radarr_server`
- `seerr_sonarr_server`
- `seerr_jellyfin_settings`
- `seerr_emby_settings`
- `seerr_tautulli_settings`
- `seerr_network_settings`
- `seerr_backup_settings`
- `seerr_job_schedule`
- `seerr_request`
- `seerr_issue`
- `seerr_blocklist`
- `seerr_override_rule`
- `seerr_user`
- `seerr_user_invitation`
- `seerr_user_permissions`
- `seerr_user_settings_permissions`
- `seerr_user_watchlist_settings`

## Reusable modules

- `modules/main_settings`
- `modules/plex_settings`
- `modules/radarr_server`
- `modules/sonarr_server`

## Object reference

Resources:
- `seerr_api_key`: regenerate and read the current Seerr API key.
- `seerr_api_object`: manage arbitrary Seerr endpoints with explicit HTTP methods.
- `seerr_backup_settings`: manage Seerr backup scheduling and retention settings.
- `seerr_blocklist`: manage manual Seerr blocklist entries.
- `seerr_discover_slider`: manage Seerr discover slider configuration.
- `seerr_emby_settings`: manage Emby integration settings.
- `seerr_issue`: manage Seerr issue records.
- `seerr_jellyfin_settings`: manage Jellyfin integration settings.
- `seerr_job_schedule`: manage Seerr background job schedules.
- `seerr_main_settings`: manage core Seerr application settings.
- `seerr_network_settings`: manage Seerr network settings.
- `seerr_notification_discord`, `seerr_notification_email`, `seerr_notification_gotify`, `seerr_notification_lunasea`, `seerr_notification_ntfy`, `seerr_notification_pushbullet`, `seerr_notification_pushover`, `seerr_notification_slack`, `seerr_notification_telegram`, `seerr_notification_webhook`, `seerr_notification_webpush`: manage typed notification integrations.
- `seerr_override_rule`: manage request override rules.
- `seerr_plex_settings`: manage Plex integration settings.
- `seerr_radarr_server`: manage a Radarr server integration in Seerr.
- `seerr_request`: manage Seerr media requests.
- `seerr_sonarr_server`: manage a Sonarr server integration in Seerr.
- `seerr_tautulli_settings`: manage Tautulli integration settings.
- `seerr_user`: manage Seerr users plus user-scoped settings and notification preferences.
- `seerr_user_invitation`: manage pending user invitations.
- `seerr_user_permissions`: manage a user's Seerr permissions bitmask.
- `seerr_user_settings_permissions`: manage the per-user settings permissions endpoint.
- `seerr_user_watchlist_settings`: manage a user's Plex watchlist sync flags.

Data sources:
- `seerr_api_key`: read the current Seerr API key.
- `seerr_api_request`: issue an arbitrary read request to the Seerr API.
- `seerr_backup_settings`: read current backup settings.
- `seerr_blocklist`: read a blocklist entry by TMDB ID and media type.
- `seerr_main_settings`: read current main settings.
- `seerr_current_user`: read the current authenticated Seerr user.
- `seerr_discover_slider`: read current discover slider settings.
- `seerr_emby_settings`: read current Emby settings.
- `seerr_issue`: look up a single issue by ID.
- `seerr_issues`: read issue lists.
- `seerr_jellyfin_settings`: read current Jellyfin settings.
- `seerr_jobs`: read background job status and schedules.
- `seerr_media`: read media lists.
- `seerr_media_item`: look up a single media record by ID.
- `seerr_network_settings`: read current network settings.
- `seerr_notification_agents`: read the configured notification integrations summary.
- `seerr_notification_discord`, `seerr_notification_email`, `seerr_notification_gotify`, `seerr_notification_lunasea`, `seerr_notification_ntfy`, `seerr_notification_pushbullet`, `seerr_notification_pushover`, `seerr_notification_slack`, `seerr_notification_telegram`, `seerr_notification_webhook`, `seerr_notification_webpush`: read typed notification integrations.
- `seerr_override_rule`: read an override rule by ID.
- `seerr_plex_settings`: read current Plex settings.
- `seerr_public_settings`: read current public settings.
- `seerr_radarr_quality_profile`: resolve a Radarr quality profile name to its numeric ID.
- `seerr_radarr_server`: read a configured Radarr server by Seerr server ID.
- `seerr_request`: look up a single media request by ID.
- `seerr_requests`: read request lists.
- `seerr_service_status`: read Seerr service status.
- `seerr_sonarr_quality_profile`: resolve a Sonarr quality profile name to its numeric ID.
- `seerr_sonarr_server`: read a configured Sonarr server by Seerr server ID.
- `seerr_tautulli_settings`: read current Tautulli settings.
- `seerr_user`: read a user by ID, username, or email.
- `seerr_user_invitations`: read pending user invitations.
- `seerr_user_permissions`: read a user's permissions bitmask.
- `seerr_users`: read paged user lists.
- `seerr_user_watchlist_settings`: read a user's watchlist sync settings.

Each resource and data source has a dedicated page under [`docs/resources`](./docs/resources)
or [`docs/data-sources`](./docs/data-sources) with at least one example usage block.

## Requirements

- Go `1.25+`
- OpenTofu `1.8.x+`

## Provider configuration

```hcl
provider "seerr" {
  url                  = "https://seerr.example.com"
  api_key              = var.seerr_api_key
  insecure_skip_verify = false
}
```

## Cross-provider references (Radarr/Sonarr)

You cannot pass another provider "object" directly. Terraform/OpenTofu providers exchange values through normal expression references. That means you pass URL, API key, and other attributes/outputs from Radarr/Sonarr resources or modules.

```hcl
# Example values from other providers/modules.
# Replace resource names with your actual radarr/sonarr resources.
data "seerr_radarr_quality_profile" "movies" {
  url     = radarr_system.this.url
  api_key = radarr_system.this.api_key
  name    = "HD-1080p"
}

resource "seerr_radarr_server" "movies" {
  url                    = radarr_system.this.url
  api_key                = radarr_system.this.api_key
  quality_profile_id     = data.seerr_radarr_quality_profile.movies.quality_profile_id
  active_directory       = "/media/movies"
  enable_scan            = true
  tag_requests_with_user = true
}

data "seerr_sonarr_quality_profile" "shows" {
  url     = sonarr_system.this.url
  api_key = sonarr_system.this.api_key
  name    = "HD-1080p"
}

resource "seerr_sonarr_server" "shows" {
  url                    = sonarr_system.this.url
  api_key                = sonarr_system.this.api_key
  quality_profile_id     = data.seerr_sonarr_quality_profile.shows.quality_profile_id
  active_directory       = "/media/tv"
  active_anime_directory = "/media/anime"
  enable_scan            = true
  tag_requests_with_user = true
}
```

### Quality profile resolution

Seerr server create/update calls require both `activeProfileId` and
`activeProfileName`.

- `quality_profile_id` is the resource input. The recommended way to get it is
  via `seerr_radarr_quality_profile` or `seerr_sonarr_quality_profile`, using
  the ARR quality-profile name as the lookup key.
- If `quality_profile_name` is set on `seerr_radarr_server` or
  `seerr_sonarr_server`, the provider uses that value directly in the Seerr
  payload.
- If `quality_profile_name` is omitted, the provider resolves the profile name
  by querying the target ARR API (`/api/v3/qualityprofile`) using the provided
  `url`/`hostname`, `port`, `use_ssl`, and `api_key`.

Operational note:

- OpenTofu execution context must be able to reach Radarr/Sonarr for automatic
  quality-profile lookup and automatic name resolution. If connectivity is
  limited, set `quality_profile_name` explicitly.

## Example: notification settings

```hcl
resource "seerr_notification_ntfy" "ntfy" {
  enabled      = true
  embed_poster = true

  notification_types = ["MEDIA_PENDING", "MEDIA_APPROVED"]

  ntfy {
    url    = "https://ntfy.example.com"
    topic  = "media"
    token  = var.ntfy_access_token
    priority = 3
  }
}

resource "seerr_notification_pushover" "pushover" {
  enabled      = true
  embed_poster = true
  notification_types = ["MEDIA_AVAILABLE", "ISSUE_CREATED"]

  pushover {
    access_token = var.pushover_access_token
    user_token   = var.pushover_user_key
    sound        = "pushover"
  }
}
```

## Example: main and plex settings

```hcl
resource "seerr_main_settings" "main" {
  app_title       = "Seerr"
  application_url = "https://seerr.example.com"
  locale          = "en"
}

resource "seerr_plex_settings" "plex" {
  ip      = "plex.media.svc.cluster.local"
  port    = 32400
  use_ssl = false
}

data "seerr_public_settings" "public" {}

output "seerr_initialized" {
  value = data.seerr_public_settings.public.initialized
}
```

## Build

```bash
bash ./scripts/test-all-locally.sh
```

## Contributor workflow

The repo-level validation entry points are:

```bash
bash ./scripts/test-all-locally.sh
SEERR_RUN_INTEGRATION=true bash ./scripts/test-all-locally.sh
SEERR_TEST_SUITE=stable bash ./scripts/test-integration.sh
SEERR_TEST_SUITE=all bash ./scripts/test-integration.sh
```

- `scripts/test-all-locally.sh`: generated-file check, `go build`, `go test`, and optional lint/integration.
- `scripts/test-integration.sh`: builds a provider mirror, boots a local Seerr target when `SEERR_URL` is not already set, and runs `tofu test`.
- The default integration suite is `stable` and mirrors the CI merge gate.
- Set `SEERR_TEST_SUITE=all` to run the broader compatibility/dependency-sensitive suite as well.

## CI model

- Pull requests run the fast gate (`scripts/test-all-locally.sh` without integration) plus the `stable` OpenTofu integration suite against an ephemeral local Seerr target. Pull requests never receive repository secrets and always use `ubuntu-latest`.
- Pushes to `main`, schedules, and manual runs may use the trusted runner selected by the `TRUSTED_RUNNER_LABEL` variable. Configure that label to a UECB or isolated self-hosted runner only after restricting the `integration` environment and keeping the runner off the pull-request path.
- Scheduled runs execute the broader `full` compatibility suite only. This keeps the merge gate deterministic while still exercising dependency-sensitive coverage regularly.
- Manual runs support three modes through the GitHub Actions `integration_mode` input: `stable`, `full`, or `both`.
- Stable- and full-suite artifacts upload only on failure, with a seven-day retention period.

Use the stable suite for merge confidence and the full suite for broader compatibility triage. If a change only fails in the full suite, treat it as an environment or unsupported-endpoint investigation unless the stable suite also regresses.

This repo includes a tracked `pre-push` hook at `.githooks/pre-push`. Enable it once per clone:

```bash
git config core.hooksPath .githooks
```

The hook runs the same generated-file verification used by CI and fails the push if generated docs or formatting are stale.

## Release and publish

The `Release` workflow builds an explicitly selected version tag. Auto-tagging and manual reconciliation dispatch the workflow from the default branch, so an old tag can never select an old workflow definition.

Auto-tagging reconciles orphaned stable version tags before creating a new version. Automatic reconciliation starts at the repository variable `RELEASE_RECONCILE_FROM_TAG` (default `v0.20.5`), which avoids publishing intentionally skipped legacy or prerelease tags. Every eligible orphan is published before the current commit is tagged. Maintainers can run the `Reconcile Releases` workflow manually to retry the oldest eligible tag without a new code push; its optional `tag` input explicitly backfills any valid historical version tag.

Release backfills always execute the current hardened workflow and publishing configuration from the default branch, then build the exact requested tag. The workflows serialize tag allocation and release dispatch, skip already-published releases, resume incomplete drafts, and fail closed when GitHub release state cannot be verified.

Expected secrets for signed provider releases:
- `GPG_PRIVATE_KEY`
- `PASSPHRASE`

Configure these as secrets on the protected `release` environment, not as repository secrets. Require maintainer approval for that environment. Configure `SEERR_URL` and `SEERR_API_KEY` on the protected `integration` environment if external integration testing is needed. Repository-level secrets with these names should be removed after the environment secrets are verified.

## OpenTofu registry naming

- Provider type: `seerr`
- Repository: `terraform-provider-seerr`
- Binary naming: `terraform-provider-seerr_vX.Y.Z`
- Suggested source address: `registry.opentofu.org/josh-archer/seerr`
