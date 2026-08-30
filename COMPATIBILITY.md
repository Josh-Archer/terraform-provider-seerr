# Compatibility Matrix

See the full [Version Compatibility Guide](docs/guides/compatibility.md) for detailed compatibility tables and version constraints, and the [Stability and Breaking-Change Policy](docs/guides/stability-and-breaking-changes.md) for SemVer 2.0 commitments.

## Quick Summary

- **Verified Targets**: Seerr `v3.4.1`, legacy Jellyseerr `v2.7.3`, Overseerr `v1.35.0`
- **OpenTofu**: `>= 1.6.0` (`registry.opentofu.org/josh-archer/seerr`)
- **Terraform**: `>= 1.5.0` (`registry.terraform.io/josh-archer/seerr`)
- **Protocol**: Terraform Plugin Framework Protocol 6.0
- **Automated Monitoring**: Weekly OpenAPI drift detection, daily release comparison, and a three-image pull-request compatibility matrix.
