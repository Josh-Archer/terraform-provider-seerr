run "plex_settings_lifecycle" {
  command = apply

  variables {
    ip      = "127.0.0.1"
    port    = 32400
    use_ssl = false
  }

  module {
    source = "./modules/plex_settings"
  }

  assert {
    condition     = seerr_plex_settings.test.ip == var.ip
    error_message = "Plex settings IP did not match expected value"
  }

  assert {
    condition     = seerr_plex_settings.test.port == var.port
    error_message = "Plex settings port did not match expected value"
  }

  assert {
    condition     = seerr_plex_settings.test.use_ssl == var.use_ssl
    error_message = "Plex settings use_ssl did not match expected value"
  }
}

run "plex_settings_data_source" {
  command = plan

  module {
    source = "./modules/plex_settings_data"
  }

  assert {
    condition     = data.seerr_plex_settings.test.ip == "127.0.0.1"
    error_message = "Plex settings data source IP did not match expected value"
  }
}
