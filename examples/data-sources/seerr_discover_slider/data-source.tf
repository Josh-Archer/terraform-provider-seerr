data "seerr_discover_slider" "all" {}

output "first_slider_title" {
  value = data.seerr_discover_slider.all.sliders[0].title
}
