# ARR stack bootstrap module

Registers keyed Radarr and Sonarr server sets, including separate quality-profile/root-folder routes, creates content override rules, and optionally configures Seerr's global webhook notification agent.

```hcl
module "arr" {
  source = "./modules/arr_stack_bootstrap"

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

API keys and the optional webhook configuration are treated as sensitive module inputs.

