data "seerr_notification_agents" "example" {}

output "agents" {
  value = data.seerr_notification_agents.example.agents
}
