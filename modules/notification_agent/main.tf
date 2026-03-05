terraform {
  required_providers {
    seerr = {
      source = "registry.opentofu.org/josh-archer/seerr"
    }
  }
}

resource "seerr_notification_agent" "this" {
  agent               = var.agent
  payload_json        = jsonencode(var.payload)
  delete_payload_json = var.delete_body_json
  disable_on_delete   = !var.skip_delete
}

output "resource_id" {
  value = seerr_notification_agent.this.id
}
