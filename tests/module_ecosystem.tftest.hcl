run "family_media_server_lifecycle" {
  command = apply

  variables {
    username = "tofu_module_family"
    email    = "tofu_module_family@example.com"
  }

  module {
    source = "./modules/family_media_server"
  }

  assert {
    condition     = output.user_id != ""
    error_message = "Family module did not create a user."
  }

  assert {
    condition     = output.permission_bitmask > 0
    error_message = "Family module did not resolve a permission set."
  }
}

run "arr_stack_bootstrap_lifecycle" {
  command = apply

  module {
    source = "./modules/arr_stack_bootstrap"
  }

  assert {
    condition     = output.radarr_resource_id != ""
    error_message = "ARR module did not register Radarr."
  }

  assert {
    condition     = output.sonarr_resource_id != ""
    error_message = "ARR module did not register Sonarr."
  }
}

run "monitoring_read" {
  command = plan

  module {
    source = "./modules/monitoring"
  }

  assert {
    condition     = output.metrics.server_version != ""
    error_message = "Monitoring module did not read the Seerr version."
  }

  assert {
    condition     = strcontains(output.prometheus_metrics, "seerr_requests_total")
    error_message = "Monitoring module did not emit Prometheus request metrics."
  }
}
