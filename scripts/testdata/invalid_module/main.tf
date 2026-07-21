terraform {
  required_providers {
    seerr = {
      source = "josh-archer/seerr"
    }
  }
}

resource "seerr_main_settings" "invalid" {
  app_title              = "Test"
  application_url        = "http://localhost:5055"
  this_argument_does_not_exist = true # This should trigger a schema validation error
}
