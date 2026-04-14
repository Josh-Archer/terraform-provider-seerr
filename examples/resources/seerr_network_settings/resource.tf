resource "seerr_network_settings" "example" {
  trust_proxy            = true
  force_ipv4_first       = false
  api_request_timeout_ms = 15000

  dns_cache {
    enabled       = true
    force_min_ttl = 60
    force_max_ttl = 600
  }

  proxy {
    enabled                = true
    hostname               = "proxy.internal"
    port                   = 8080
    use_ssl                = false
    user                   = "proxy-user"
    password               = var.proxy_password
    bypass_filter          = "localhost,*.svc.cluster.local"
    bypass_local_addresses = true
  }
}
