variable "name" { type = string }
variable "hostname" { type = string }
variable "port" { type = number }
variable "api_key" { type = string }

resource "seerr_sonarr_server" "test" {
  name     = var.name
  hostname = var.hostname
  port     = var.port
  api_key  = var.api_key
}

output "id" {
  value = seerr_sonarr_server.test.id
}
