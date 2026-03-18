variables {
  username = "test_user_perms"
  email    = "test_user_perms@example.com"
}

run "setup_user" {
  command = apply
  module {
    source = "./modules/user"
  }
  variables {
    username    = "test_user_perms"
    email       = "test_user_perms@example.com"
    permissions = 0
  }
}

run "user_permissions_lifecycle" {
  command = apply

  variables {
    user_id             = run.setup_user.id
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
    user_id             = run.setup_user.id
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

