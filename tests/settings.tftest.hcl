run "main_settings_lifecycle" {
  command = apply

  variables {
    application_url   = "http://localhost:5055"
    application_title = "tofu_test_seerr"
  }

  module {
    source = "./modules/main_settings"
  }

  assert {
    condition     = seerr_main_settings.test.application_title == var.application_title
    error_message = "Main settings application_title did not match expected value"
  }
}

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
    error_message = "Plex settings ip did not match expected value"
  }
}

run "jellyfin_settings_lifecycle" {
  command = apply

  variables {
    ip      = "127.0.0.1"
    port    = 8096
    api_key = "abc123mockkey"
  }

  module {
    source = "./modules/jellyfin_settings"
  }

  assert {
    condition     = seerr_jellyfin_settings.test.ip == var.ip
    error_message = "Jellyfin settings ip did not match expected value"
  }
}


# Note: user_id 1 is usually the admin
run "watchlist_settings_lifecycle" {
  command = apply

  variables {
    user_id               = 1
    global_watchlist_sync = true
  }

  module {
    source = "./modules/watchlist_settings"
  }

  assert {
    condition     = seerr_user_watchlist_settings.test.global_watchlist_sync == var.global_watchlist_sync
    error_message = "Watchlist settings sync did not match expected value"
  }
}



