run "current_user_data_source" {
  command = apply

  module {
    source = "./modules/current_user"
  }

  assert {
    condition     = data.seerr_current_user.test.id != ""
    error_message = "Current user id is empty"
  }
}

