terraform {
  required_providers {
    seerr = {
      source = "registry.opentofu.org/josh-archer/seerr"
    }
  }
}

locals {
  use_url = var.url != null && trimspace(var.url) != ""

  url_no_scheme = local.use_url ? replace(replace(var.url, "https://", ""), "http://", "") : ""
  host_and_path = local.use_url ? split("/", local.url_no_scheme) : []
  host_port     = local.use_url ? local.host_and_path[0] : ""
  host_parts    = local.use_url ? split(":", local.host_port) : []

  inferred_hostname = local.use_url ? local.host_parts[0] : null
  inferred_port = local.use_url ? (
    length(local.host_parts) > 1 ? tonumber(local.host_parts[1]) : (
      startswith(var.url, "https://") ? 443 : 80
    )
  ) : null
  inferred_base_url = local.use_url && length(local.host_and_path) > 1 ? "/${join("/", slice(local.host_and_path, 1, length(local.host_and_path)))}" : ""

  payload = merge({
    name                = var.name
    hostname            = local.use_url ? local.inferred_hostname : var.hostname
    port                = local.use_url ? local.inferred_port : var.port
    apiKey              = var.api_key
    useSsl              = local.use_url ? startswith(var.url, "https://") : var.use_ssl
    baseUrl             = local.use_url ? local.inferred_base_url : var.base_url
    activeProfileId     = var.active_profile_id
    activeDirectory     = var.active_directory
    is4k                = var.is_4k
    minimumAvailability = var.minimum_availability
    tags                = var.tags
    isDefault           = var.is_default
    syncEnabled         = var.sync_enabled
    preventSearch       = var.prevent_search
    tagRequests         = var.tag_requests
  }, var.extra_payload)
}

resource "seerr_api_object" "this" {
  path              = "/api/v1/settings/radarr"
  read_method       = "GET"
  create_method     = "POST"
  update_method     = "POST"
  delete_method     = "POST"
  skip_delete       = true
  request_body_json = jsonencode(local.payload)
}

output "resource_id" {
  value = seerr_api_object.this.id
}

