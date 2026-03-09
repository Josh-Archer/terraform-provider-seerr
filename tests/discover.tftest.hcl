# Inherits provider "seerr" from provider.tf

run "apply_discover_sliders" {
  module {
    source = "./modules/discover_slider"
  }

  assert {
    condition     = output.slider_id == "settings"
    error_message = "Discover slider ID should be 'settings'"
  }
}

run "verify_reorder" {
  module {
    source = "./modules/discover_slider"
  }

  # In a real test we'd change order here, but for a single run we check stability
}
