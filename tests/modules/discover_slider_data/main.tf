terraform {
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "99.99.99"
    }
  }
}
data "seerr_discover_slider" "all" {}

output "sliders_count" {
  value = length(data.seerr_discover_slider.all.sliders)
}
