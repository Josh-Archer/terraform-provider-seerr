run "test_users_data_source" {
  command = plan

  variables {
    # Provide placeholders logic isn't tied to active connection testing
  }

  assert {
    condition     = data.seerr_users.all.id != ""
    error_message = "Users data source ID should not be empty."
  }
}

run "test_jobs_data_source" {
  command = plan

  assert {
    condition     = data.seerr_jobs.all.id != ""
    error_message = "Jobs data source ID should not be empty."
  }
}

run "test_notification_agents_data_source" {
  command = plan

  assert {
    condition     = data.seerr_notification_agents.all.id != ""
    error_message = "Notification agents data source ID should not be empty."
  }
}
