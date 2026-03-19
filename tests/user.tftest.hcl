run "user_lifecycle" {
  command = apply

  variables {
    username    = "tofu_test"
    email       = "tofu_test@example.com"
    permissions = 32
  }

  module {
    source = "./modules/user"
  }

  assert {
    condition     = seerr_user.test.username == var.username
    error_message = "User username did not match expected value"
  }

  assert {
    condition     = seerr_user.test.email == var.email
    error_message = "User email did not match expected value"
  }
}

run "user_data_source" {
  command = plan

  variables {
    email = "tofu_test@example.com"
  }

  module {
    source = "./modules/user_data"
  }

  assert {
    condition     = data.seerr_user.test.email == var.email
    error_message = "User data source email did not match expected value"
  }
}




