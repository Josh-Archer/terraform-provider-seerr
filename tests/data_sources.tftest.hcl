run "test_users_data_source" {
  command = apply
  module {
    source = "./modules/user_data"
  }
  variables {
    email = "ci-admin@example.invalid"
  }
}

# Skipping non-existent data sources for Jobs and Notification Agents
# as they are not currently implemented in the provider.
