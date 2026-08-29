# Publishing the reusable modules

The supported distribution is the versioned child modules in this repository. Users can consume them directly with a GitHub subdirectory source such as `github.com/Josh-Archer/terraform-provider-seerr//modules/family_media_server?ref=v0.41.0`; no separate repository is required for normal use.

Publishing them as independently searchable OpenTofu Registry packages is optional. The registry cannot point at a child directory in a provider repository, so that distribution model would require a separate public repository named `{owner}/terraform-{target}-{name}` for each package and submission through the registry's interactive GitHub issue form.

If independent registry packages are ever desired, use these mappings:

| Source directory | Publication repository | Module address |
| --- | --- | --- |
| `modules/family_media_server` | `Josh-Archer/terraform-seerr-family-media-server` | `Josh-Archer/family-media-server/seerr` |
| `modules/arr_stack_bootstrap` | `Josh-Archer/terraform-seerr-arr-stack-bootstrap` | `Josh-Archer/arr-stack-bootstrap/seerr` |
| `modules/monitoring` | `Josh-Archer/terraform-seerr-monitoring` | `Josh-Archer/monitoring/seerr` |

For each package:

1. Copy the module directory to the root of its public repository.
2. Add a semantic version tag such as `v1.0.0`.
3. Open the [OpenTofu Registry module submission form](https://github.com/opentofu/registry/issues/new?template=module.yml) in a signed-in browser and submit the repository path.

The OpenTofu registry explicitly rejects API- or CLI-created submission issues because its automation depends on the structured interactive form.
