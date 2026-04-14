run "public_settings_read" {
  command = plan

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

