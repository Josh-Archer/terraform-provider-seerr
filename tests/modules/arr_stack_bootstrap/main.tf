terraform {
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "99.99.99"
    }
  }
}

module "arr" {
  source = "../../../modules/arr_stack_bootstrap"

  radarr_servers = {
    integration = {
      name               = "tofu_module_radarr"
      hostname           = "127.0.0.1"
      api_key            = "radarr_api_key"
      quality_profile_id = 1
      root_folder        = "/movies"
      is_default         = true
    }
  }

  sonarr_servers = {
    integration = {
      name               = "tofu_module_sonarr"
      hostname           = "127.0.0.1"
      api_key            = "sonarr_api_key"
      quality_profile_id = 1
      root_folder        = "/tv"
      is_default         = true
    }
  }
}

output "radarr_resource_id" {
  value = module.arr.radarr_resource_ids.integration
}

output "sonarr_resource_id" {
  value = module.arr.sonarr_resource_ids.integration
}
