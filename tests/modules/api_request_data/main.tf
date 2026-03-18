terraform {
  required_providers {
    seerr = {
      source = "josh-archer/seerr"
    }
  }
}

variable "endpoint" { type = string }

data "seerr_api_request" "test" {
  path = var.endpoint
}

output "response" {
  value = data.seerr_api_request.test.response_body_json
}
