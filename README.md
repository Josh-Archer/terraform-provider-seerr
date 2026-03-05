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
resource "seerr_radarr_server" "movies" {
  url               = radarr_system.this.url
  api_key           = radarr_system.this.api_key
  active_profile_id = 4
  active_directory  = "/media/movies"
}

resource "seerr_sonarr_server" "shows" {
  url                    = sonarr_system.this.url
  api_key                = sonarr_system.this.api_key
  active_profile_id      = 6
  active_directory       = "/media/tv"
  active_anime_directory = "/media/anime"
}
```

## Example: notification settings

```hcl
resource "seerr_notification_agent" "ntfy" {
  agent        = "ntfy"
  payload_json = jsonencode({
    enabled = true
    options = {
      serverUrl   = "https://ntfy.example.com"
      topic       = "media"
      accessToken = var.ntfy_access_token
      priority    = 3
    }
    types = {
      MEDIA_APPROVED  = true
      MEDIA_AVAILABLE = true
      MEDIA_FAILED    = true
    }
  })
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
