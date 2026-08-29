output "user_ids" {
  description = "Seerr user IDs keyed by the input logical name."
  value       = { for key, user in seerr_user.this : key => user.id }
}

output "permission_bitmasks" {
  description = "Resolved permission bitmask for each managed user."
  value       = { for key, user in seerr_user.this : key => user.permissions }
}

output "discover_slider_id" {
  description = "Singleton discover slider resource ID, or null when sliders are unmanaged."
  value       = try(seerr_discover_slider.this[0].id, null)
}

