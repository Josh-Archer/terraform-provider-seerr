run "radarr_lifecycle" {
  command = apply

  variables {
    name     = "tofu_radarr_test"
    hostname = "127.0.0.1"
    port     = 7878
    api_key  = "radarr_api_key"
  }

  module {
    source = "./modules/radarr_server"
  }

  assert {
    condition     = seerr_radarr_server.test.name == var.name
    error_message = "Radarr server name did not match expected value"
  }
}

run "sonarr_lifecycle" {
  command = apply

  variables {
    name     = "tofu_sonarr_test"
    hostname = "127.0.0.1"
    port     = 8989
    api_key  = "sonarr_api_key"
  }

  module {
    source = "./modules/sonarr_server"
  }

  assert {
    condition     = seerr_sonarr_server.test.name == var.name
    error_message = "Sonarr server name did not match expected value"
  }
}
