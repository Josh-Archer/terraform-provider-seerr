run "public_settings_read" {
  command = apply

  module {
    source = "./modules/public_settings"
  }

  assert {
    condition     = data.seerr_public_settings.test.status_code == 200
    error_message = "Public settings status_code was not 200"
  }
}

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
    condition     = seerr_main_settings.test.app_title == var.application_title
    error_message = "Main settings app_title did not match expected value"
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

run "notification_ntfy_lifecycle" {
  command = apply

  variables {
    ntfy_url = "https://ntfy.example.com"
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

run "tautulli_settings_lifecycle" {
  command = apply

  variables {
    hostname = "127.0.0.1"
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



