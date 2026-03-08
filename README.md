# OpenTofu Provider for Seerr

`seerr` is a community OpenTofu provider that manages Seerr via its API.

The provider supports two usage styles:
- Generic API primitives (`seerr_api_object` resource and `seerr_api_request` data source).
- Typed Seerr primitives for common integrations and settings.

## Typed resources

- `seerr_main_settings`
- `seerr_plex_settings`
- `seerr_notification_agent`
- `seerr_radarr_server`
- `seerr_sonarr_server`

## Reusable modules

- `modules/main_settings`
- `modules/plex_settings`
- `modules/notification_agent`
- `modules/radarr_server`
- `modules/sonarr_server`

## Object reference

Resources:
- `seerr_api_key`: regenerate and read the current Seerr API key.
- `seerr_api_object`: manage arbitrary Seerr endpoints with explicit HTTP methods.
- `seerr_main_settings`: manage core Seerr application settings.
- `seerr_notification_agent`: manage a named notification agent payload.
- `seerr_plex_settings`: manage Plex integration settings.
- `seerr_radarr_server`: manage a Radarr server integration in Seerr.
- `seerr_sonarr_server`: manage a Sonarr server integration in Seerr.
- `seerr_user_permissions`: manage a user's Seerr permissions bitmask.
- `seerr_user_watchlist_settings`: manage a user's Plex watchlist sync flags.

Data sources:
- `seerr_api_key`: read the current Seerr API key.
- `seerr_api_request`: issue an arbitrary read request to the Seerr API.
- `seerr_main_settings`: read current main settings.
- `seerr_notification_agent`: read a named notification agent payload.
- `seerr_plex_settings`: read current Plex settings.
- `seerr_radarr_quality_profile`: resolve a Radarr quality profile name to its numeric ID.
- `seerr_radarr_server`: read a configured Radarr server by Seerr server ID.
- `seerr_sonarr_quality_profile`: resolve a Sonarr quality profile name to its numeric ID.
- `seerr_sonarr_server`: read a configured Sonarr server by Seerr server ID.
- `seerr_user_permissions`: read a user's permissions bitmask.
- `seerr_user_watchlist_settings`: read a user's watchlist sync settings.

Each resource and data source has a dedicated page under [`docs/resources`](./docs/resources)
or [`docs/data-sources`](./docs/data-sources) with at least one example usage block.

## Requirements

- Go `1.25+`
- OpenTofu `1.8+`

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
resource "seerr_notification_agent" "ntfy" {
  agent        = "ntfy"
  enabled      = true
  embed_poster = true
  types        = 1023

  ntfy = {
    url               = "https://ntfy.example.com"
    topic             = "media"
    auth_method_token = true
    token             = var.ntfy_access_token
    priority          = 3
  }
}

resource "seerr_notification_agent" "pushover" {
  agent        = "pushover"
  enabled      = true
  embed_poster = true
  types        = 1023

  pushover = {
    access_token = var.pushover_access_token
    user_token   = var.pushover_user_token
    sound        = "pushover"
  }
}
```

## Example: main and plex settings

```hcl
resource "seerr_main_settings" "main" {
  payload_json = jsonencode({
    applicationTitle = "Seerr"
    locale           = "en"
  })
}

resource "seerr_plex_settings" "plex" {
  payload_json = jsonencode({
    hostname = "plex.media.svc.cluster.local"
    port     = 32400
    useSsl   = false
  })
}
```

## Build

```bash
go mod tidy
go test ./...
go build .
```

## Contributor workflow

Before pushing changes, keep generated docs and Go formatting current:

```bash
go generate ./...
gofmt -w .
```

This repo includes a tracked `pre-push` hook at `.githooks/pre-push`. Enable it once per clone:

```bash
git config core.hooksPath .githooks
```

The hook runs the same generated-file verification used by CI and fails the push if generated docs or formatting are stale.

## Release and publish

Releases are created from git tags matching `v*` by GitHub Actions.

Expected secrets for signed provider releases:
- `GPG_PRIVATE_KEY`
- `PASSPHRASE`

## OpenTofu registry naming

- Provider type: `seerr`
- Repository: `terraform-provider-seerr`
- Binary naming: `terraform-provider-seerr_vX.Y.Z`
- Suggested source address: `registry.opentofu.org/josh-archer/seerr`
