terraform {
  required_providers {
    seerr = {
      source = "registry.opentofu.org/josh-archer/seerr"
    }
  }
}

locals {
  read_path = "/api/v1/settings/notifications/${var.agent}"
}

resource "seerr_api_object" "this" {
  path              = local.read_path
  read_method       = "GET"
  create_method     = "POST"
  update_method     = "POST"
  delete_method     = "POST"
  skip_delete       = var.skip_delete
  delete_body_json  = var.delete_body_json
  request_body_json = jsonencode(var.payload)
}

output "resource_id" {
  value = seerr_api_object.this.id
}

