data "seerr_discover_slider" "all" {}

output "sliders_count" {
  value = length(data.seerr_discover_slider.all.sliders)
}
