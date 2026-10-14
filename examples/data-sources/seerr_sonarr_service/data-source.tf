# Read all Sonarr servers reachable by Seerr
data "seerr_sonarr_service" "all" {}

# Read a specific Sonarr server by ID and its quality profiles
data "seerr_sonarr_service" "single" {
  server_id = 1
}
