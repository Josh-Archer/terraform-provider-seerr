# Using the reusable modules

The modules are optional convenience layers over the provider's existing resources. Direct `seerr_*` resources remain fully supported, and adopting a module does not change how the provider is configured.

## Configure the provider once

The root configuration installs and configures the provider. Child modules inherit this provider configuration automatically.

```hcl
terraform {
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "~> 0.41"
    }
  }
}

provider "seerr" {
  url     = var.seerr_url
  api_key = var.seerr_api_key
}
```

Pin module sources with `?ref=`. This makes module changes deliberate and repeatable instead of following the repository's moving default branch.

## Family users and quotas

```hcl
module "family" {
  source = "github.com/Josh-Archer/terraform-provider-seerr//modules/family_media_server?ref=v0.41.0"

  users = {
    child = {
      username       = "Child"
      email          = "child@example.com"
      permission_set = "standard"

      quota = {
        movie_limit = 3
        movie_days  = 7
        tv_limit    = 2
        tv_days     = 7
      }
    }

    parent = {
      username       = "Parent"
      email          = "parent@example.com"
      permission_set = "power"
    }
  }
}
```

The module resolves the named permission tiers and creates the associated users and quota resources. Discover sliders are optional and should be managed by only one configuration because Seerr exposes them as a singleton.

## Radarr and Sonarr

```hcl
module "arr" {
  source = "github.com/Josh-Archer/terraform-provider-seerr//modules/arr_stack_bootstrap?ref=v0.41.0"

  radarr_servers = {
    hd = {
      name               = "Radarr HD"
      hostname           = "radarr"
      api_key            = var.radarr_api_key
      quality_profile_id = 1
      root_folder        = "/movies"
      is_default         = true
    }
  }

  sonarr_servers = {
    hd = {
      name               = "Sonarr HD"
      hostname           = "sonarr"
      api_key            = var.sonarr_api_key
      quality_profile_id = 1
      root_folder        = "/tv"
      is_default         = true
    }
  }
}
```

Additional map entries can represent 4K, anime, or other quality-profile and root-folder routes.

## Monitoring outputs

```hcl
module "seerr_monitoring" {
  source = "github.com/Josh-Archer/terraform-provider-seerr//modules/monitoring?ref=v0.41.0"
}

output "seerr_metrics" {
  value = module.seerr_monitoring.metrics
}

output "prometheus_text" {
  value = module.seerr_monitoring.prometheus_metrics
}
```

The structured output is suitable for Grafana provisioning or another exporter. The text output follows the Prometheus exposition format.

## Using a local checkout

When the calling configuration is at the repository root, replace the remote source with a local path:

```hcl
module "family" {
  source = "./modules/family_media_server"
  users  = {}
}
```

## Migrating existing resources

Moving existing resources behind a module changes their state addresses. Declare `moved` blocks so OpenTofu transfers ownership instead of planning deletion and recreation:

```hcl
moved {
  from = seerr_user.child
  to   = module.family.seerr_user.this["child"]
}

moved {
  from = seerr_user_quota.child
  to   = module.family.seerr_user_quota.this["child"]
}
```

Run `tofu plan` and confirm that it reports address moves without replacement before applying. Resource keys in the `moved` destination must match the keys supplied to the module.

