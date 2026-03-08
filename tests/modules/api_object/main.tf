terraform {
  required_providers {
    seerr = {
      source = "josh-archer/seerr"
    }
  }
}

variable "name" { type = string }
variable "endpoint" { type = string }
variable "payload" { type = string }

resource "seerr_api_object" "test" {
  name     = var.name
  endpoint = var.endpoint
  payload  = var.payload
}

output "id" {
  value = seerr_api_object.test.id
}

