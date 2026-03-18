terraform {
  required_providers {
    seerr = {
      source = "josh-archer/seerr"
    }
  }
}

variable "path" { type = string }
variable "request_body" { type = string }
variable "create_method" {
  type    = string
  default = "PUT"
}
variable "update_method" {
  type    = string
  default = "PUT"
}
variable "delete_method" {
  type    = string
  default = "DELETE"
}
variable "skip_delete" {
  type    = bool
  default = false
}

resource "seerr_api_object" "test" {
  path              = var.path
  request_body_json = var.request_body
  create_method     = var.create_method
  update_method     = var.update_method
  delete_method     = var.delete_method
  skip_delete       = var.skip_delete
}

output "id" {
  value = seerr_api_object.test.id
}
