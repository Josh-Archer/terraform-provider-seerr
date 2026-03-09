terraform {
  required_providers {
    seerr = {
      source = "josh-archer/seerr"
    }
  }
}

variable "path" { type = string }
variable "request_body" { type = string }

resource "seerr_api_object" "test" {
  path              = var.path
  request_body_json = var.request_body
}

output "id" {
  value = seerr_api_object.test.id
}
