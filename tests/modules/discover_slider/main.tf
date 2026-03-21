terraform {
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "99.99.99"
    }
  }
}
resource "seerr_discover_slider" "all" {
  sliders {
    type    = 1 # Recently Added TV
    enabled = true
  }
  sliders {
    type    = 2 # Recently Added Movies
    enabled = true
  }
  # Add a few more if known, or just these two for a basic test.
  # Based on defaultSliders research:
  # Trending: 4
  # Popular Movies: 5
  # Popular TV: 6
  sliders {
    type    = 4 # Trending
    enabled = true
  }
}

output "slider_id" {
  value = seerr_discover_slider.all.id
}
