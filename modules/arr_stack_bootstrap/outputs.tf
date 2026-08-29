output "radarr_server_ids" {
  description = "Registered Radarr IDs keyed by routing name."
  value       = { for key, server in seerr_radarr_server.this : key => server.server_id }
}

output "radarr_resource_ids" {
  description = "Terraform resource IDs for Radarr registrations keyed by routing name."
  value       = { for key, server in seerr_radarr_server.this : key => server.id }
}

output "sonarr_server_ids" {
  description = "Registered Sonarr IDs keyed by routing name."
  value       = { for key, server in seerr_sonarr_server.this : key => server.server_id }
}

output "sonarr_resource_ids" {
  description = "Terraform resource IDs for Sonarr registrations keyed by routing name."
  value       = { for key, server in seerr_sonarr_server.this : key => server.id }
}

output "override_rule_ids" {
  description = "Override rule resource IDs keyed by logical name."
  value       = { for key, rule in seerr_override_rule.this : key => rule.id }
}

output "notification_webhook_id" {
  description = "Webhook agent resource ID, or null when unmanaged."
  value       = try(seerr_notification_webhook.this[0].id, null)
}
