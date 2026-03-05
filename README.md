# OpenTofu Provider for Seerr

`seerr` is a community OpenTofu provider that manages Seerr through its API.

The provider is designed to support the full Seerr API surface by exposing:
- `seerr_api_object` resource for CRUD-style endpoint management.
- `seerr_api_request` data source for read/query and ad-hoc endpoint calls.
- Reusable modules for common Seerr setup:
  - `modules/radarr_server`
  - `modules/sonarr_server`
  - `modules/notification_agent`

This repository is self-contained and includes:
- Provider source code
- GitHub Actions CI
- GitHub Actions release/publish pipeline
- OpenTofu registry manifest release artifacts

## Requirements

- Go `1.25+`
- OpenTofu `1.8+` (or Terraform-compatible CLI for local provider testing)

## Provider Configuration

```hcl
provider "seerr" {
  url                 = "https://seerr.example.com"
  api_key             = var.seerr_api_key
  insecure_skip_verify = false
}
```

## Example: Manage Main Settings

```hcl
resource "seerr_api_object" "main_settings" {
  path              = "/api/v1/settings/main"
  read_method       = "GET"
  create_method     = "POST"
  update_method     = "POST"
  delete_method     = "POST"
  request_body_json = jsonencode({
    applicationTitle = "Seerr"
    locale           = "en"
  })
}
```

## Example: Read Current Seerr Status

```hcl
data "seerr_api_request" "status" {
  path   = "/api/v1/status"
  method = "GET"
}

## Modules: ARR and Notifications

You can pass OpenTofu references directly into module inputs.

```hcl
module "radarr_server" {
  source = "github.com/Josh-Archer/terraform-provider-seerr//modules/radarr_server?ref=v0.1.1"

  # This can come from another module/output.
  url               = var.radarr_url
  api_key           = var.radarr_api_key
  active_profile_id = 4
  active_directory  = "/media/movies"
}

module "sonarr_server" {
  source = "github.com/Josh-Archer/terraform-provider-seerr//modules/sonarr_server?ref=v0.1.1"

  url                    = var.sonarr_url
  api_key                = var.sonarr_api_key
  active_profile_id      = 4
  active_directory       = "/media/tv"
  active_anime_directory = "/media/tv"
}

module "ntfy_notification" {
  source = "github.com/Josh-Archer/terraform-provider-seerr//modules/notification_agent?ref=v0.1.1"

  agent = "ntfy"
  payload = {
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
  }
}
```
```

## Build

```bash
go mod tidy
go test ./...
go build .
```

## Release and Publish

Releases are created from git tags matching `v*` and built by GitHub Actions.

Expected secrets for signed provider releases:
- `GPG_PRIVATE_KEY`
- `PASSPHRASE`

Release artifacts include:
- Provider archives for supported OS/arch targets
- `terraform-provider-seerr_<version>_SHA256SUMS`
- `terraform-provider-seerr_<version>_SHA256SUMS.sig`
- `terraform-provider-seerr_<version>_manifest.json`

## OpenTofu Registry Naming

- Provider type: `seerr`
- Repository: `terraform-provider-seerr`
- Binary naming: `terraform-provider-seerr_vX.Y.Z`
- Suggested source address:
  - `registry.opentofu.org/josh-archer/seerr`

## Notes

- Seerr API compatibility can evolve across releases. Pin provider and Seerr versions in production.
- The generic API model is intentional so new Seerr endpoints can be managed without waiting for provider schema updates.
- For settings endpoints that do not implement HTTP DELETE, use `skip_delete = true` (default in modules) or set `delete_body_json`.
