variable "arr_api_key" {
  type      = string
  sensitive = true
}

resource "seerr_arr_connection_test" "radarr" {
  service  = "radarr"
  hostname = "radarr.example.net"
  port     = 7878
  api_key  = var.arr_api_key
  use_ssl  = true
  base_url = "/radarr"
}
