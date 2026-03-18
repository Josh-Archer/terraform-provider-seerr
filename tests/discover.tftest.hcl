# Inherits provider "seerr" from provider.tf

run "apply_discover_sliders" {
  command = plan
  module {
    source = "./modules/discover_slider"
  }

}

/*
run "verify_reorder" {
  module {
    source = "./modules/discover_slider"
  }

  # In a real test we'd change order here, but for a single run we check stability
}
*/
