terraform {
  required_providers {
    seerr = {
      source = "registry.opentofu.org/josh-archer/seerr"
    }
  }
}

resource "seerr_main_settings" "this" {
  payload_json = jsonencode(var.payload)
}

output "resource_id" {
  value = seerr_main_settings.this.id
}
