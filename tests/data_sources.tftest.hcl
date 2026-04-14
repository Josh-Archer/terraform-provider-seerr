run "test_users_data_source" {
  command = apply
  module {
    source = "./modules/user_data"
  }
  variables {
    email = "ci-admin@example.invalid"
  }
}

run "test_jobs_data_source" {
  command = plan
  module {
    source = "./modules/jobs_data"
  }

  assert {
    condition     = length(data.seerr_jobs.test.jobs) > 0
    error_message = "Expected jobs data source to return at least one job"
  }
}
