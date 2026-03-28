variables {
  username = "test_user_perms"
  email    = "test_user_perms@example.com"
}

run "user_permissions_lifecycle" {
  command = apply

  variables {
    username            = "test_user_perms"
    email               = "test_user_perms@example.com"
    auto_approve_movies = true
    auto_approve_tv     = false
  }

  module {
    source = "./modules/user_settings_permissions"
  }

  assert {
    condition     = seerr_user_settings_permissions.test.auto_approve_movies == true
    error_message = "auto_approve_movies did not match expected value"
  }

  assert {
    condition     = seerr_user_settings_permissions.test.auto_approve_tv == false
    error_message = "auto_approve_tv did not match expected value"
  }
}

run "user_permissions_update" {
  command = apply

  variables {
    username            = "test_user_perms"
    email               = "test_user_perms@example.com"
    auto_approve_movies = false
    auto_approve_tv     = true
  }

  module {
    source = "./modules/user_settings_permissions"
  }

  assert {
    condition     = seerr_user_settings_permissions.test.auto_approve_movies == false
    error_message = "auto_approve_movies did not match expected update"
  }

  assert {
    condition     = seerr_user_settings_permissions.test.auto_approve_tv == true
    error_message = "auto_approve_tv did not match expected update"
  }
}

