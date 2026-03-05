# OpenTofu Provider for Seerr

`seerr` is a community OpenTofu provider that manages Seerr through its API.

The provider is designed to support the full Seerr API surface by exposing:
- `seerr_api_object` resource for CRUD-style endpoint management.
- `seerr_api_request` data source for read/query and ad-hoc endpoint calls.

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
