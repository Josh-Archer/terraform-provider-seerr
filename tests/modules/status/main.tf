terraform {
  required_providers {
    seerr = {
      source = "josh-archer/seerr"
    }
  }
}

data "seerr_service_status" "test" {}

output "version" {
  value = data.seerr_service_status.test.version
}
