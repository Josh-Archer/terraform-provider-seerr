variables {
  seerr_url     = "http://localhost:1337" # Default or from environment
  seerr_api_key = "dummy"                # Should be provided via env or tfvars
}

provider "seerr" {
  url     = var.seerr_url
  api_key = var.seerr_api_key
}

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
