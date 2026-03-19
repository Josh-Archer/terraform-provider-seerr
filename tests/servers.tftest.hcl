run "radarr_lifecycle" {
  command = plan

  variables {
    name               = "tofu_radarr_test"
    hostname           = "127.0.0.1"
    port               = 7878
    api_key            = "radarr_api_key"
    quality_profile_id = 1
    active_directory   = "/movies"
  }

  module {
    source = "./modules/radarr_server"
  }

  assert {
    condition     = seerr_radarr_server.test.name == var.name
    error_message = "Radarr server name did not match expected value"
  }

  assert {
    condition     = seerr_radarr_server.test.hostname == var.hostname
    error_message = "Radarr hostname was not synced correctly"
  }

  assert {
    condition     = seerr_radarr_server.test.port == var.port
    error_message = "Radarr port was not synced correctly"
  }
}

run "radarr_no_drift" {
  command = plan
}

run "sonarr_lifecycle" {
  command = plan

  variables {
    name               = "tofu_sonarr_test"
    hostname           = "127.0.0.1"
    port               = 8989
    api_key            = "sonarr_api_key"
    quality_profile_id = 1
    active_directory   = "/tv"
  }

  module {
    source = "./modules/sonarr_server"
  }

  assert {
    condition     = seerr_sonarr_server.test.name == var.name
    error_message = "Sonarr server name did not match expected value"
  }

  assert {
    condition     = seerr_sonarr_server.test.hostname == var.hostname
    error_message = "Sonarr hostname was not synced correctly"
  }

  assert {
    condition     = seerr_sonarr_server.test.port == var.port
    error_message = "Sonarr port was not synced correctly"
  }
}

run "sonarr_no_drift" {
  command = plan
}





