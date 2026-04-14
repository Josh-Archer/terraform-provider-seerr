run "plex_settings_lifecycle" {
  command = apply

  variables {
    enabled = false
    ip      = "plex-mock"
    port    = 32400
    use_ssl = false
  }

  module {
    source = "./modules/plex_settings"
  }

  assert {
    condition     = !var.enabled || seerr_plex_settings.test[0].ip == var.ip
    error_message = "Plex settings ip did not match expected value"
  }
}

run "jellyfin_settings_lifecycle" {
  command = apply

  variables {
    ip      = "jellyfin-mock"
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

run "notification_ntfy_lifecycle" {
  command = apply

  variables {
    ntfy_url = "http://notify-mock:8080"
    topic    = "tofu-test-notification"
    token    = "ntfy_mock_token"
    priority = 4
  }

  module {
    source = "./modules/notification_ntfy"
  }

  assert {
    condition     = seerr_notification_ntfy.test.ntfy.topic == var.topic
    error_message = "Ntfy notification topic did not match expected value"
  }
}

run "notification_pushover_lifecycle" {
  command = apply

  variables {
    access_token = "pushover_app_mock_token"
    user_token   = "pushover_user_mock_token"
    sound        = "bike"
  }

  module {
    source = "./modules/notification_pushover"
  }

  assert {
    condition     = seerr_notification_pushover.test.pushover.sound == var.sound
    error_message = "Pushover notification sound did not match expected value"
  }
}

run "watchlist_settings_lifecycle" {
  command = apply

  variables {
    user_id               = 1
    watchlist_sync_movies = true
    watchlist_sync_tv     = true
  }

  module {
    source = "./modules/watchlist_settings"
  }

  assert {
    condition     = seerr_user_watchlist_settings.test.watchlist_sync_movies == var.watchlist_sync_movies
    error_message = "Watchlist movie sync did not match expected value"
  }

  assert {
    condition     = seerr_user_watchlist_settings.test.watchlist_sync_tv == var.watchlist_sync_tv
    error_message = "Watchlist TV sync did not match expected value"
  }
}

run "tautulli_settings_lifecycle" {
  command = apply

  variables {
    hostname = "tautulli-mock"
    port     = 8181
    api_key  = "tautulli_mock_key"
  }

  module {
    source = "./modules/tautulli_settings"
  }

  assert {
    condition     = seerr_tautulli_settings.test.hostname == var.hostname
    error_message = "Tautulli settings hostname did not match expected value"
  }
}
