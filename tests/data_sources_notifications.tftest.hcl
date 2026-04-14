variables {
  notification_agents_supported = true
}

run "test_notification_agents_data_source" {
  command = plan

  variables {
    notification_agents_supported = var.notification_agents_supported
  }

  module {
    source = "./modules/notification_agents_data"
  }

  assert {
    condition     = output.supported == false || output.agent_count >= 1
    error_message = "Notification agents endpoint is supported but the data source returned no agents"
  }
}
