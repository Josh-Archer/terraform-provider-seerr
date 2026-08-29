# Publishing the reusable modules

The reusable modules in this repository are tested as child modules and are ready to be extracted into registry packages. OpenTofu Registry module submissions cannot point at a child directory in a provider repository. Each published package must use a public repository named `{owner}/terraform-{target}-{name}` and must be submitted through the registry's interactive GitHub issue form.

Use these package mappings:

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

