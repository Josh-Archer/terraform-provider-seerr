# ARR stack bootstrap module

Registers keyed Radarr and Sonarr server sets, including separate quality-profile/root-folder routes, creates content override rules, and optionally configures Seerr's global webhook notification agent.

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
}
```

When calling the module from a checkout of this repository, use `source = "./modules/arr_stack_bootstrap"` instead.

Treat API keys and the optional webhook authorization header as sensitive values in the calling configuration and its state.
