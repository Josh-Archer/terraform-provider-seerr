terraform {
  required_providers {
    seerr = {
      source = "josh-archer/seerr"
    }
  }
}

data "seerr_current_user" "test" {}

output "id" {
  value = data.seerr_current_user.test.id
}
