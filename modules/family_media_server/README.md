# Family media server module

Creates family users from stable map keys, resolves `standard`, `power`, and `admin` permission sets, applies per-user request quotas, configures per-user notification routing, and optionally owns the ordered discover-slider singleton.

```hcl
module "family" {
  source = "./modules/family_media_server"

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
  }
}
```

Only one configuration should manage `discover_sliders`, because Seerr exposes it as a singleton.

