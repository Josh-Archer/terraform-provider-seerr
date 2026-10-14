# Read all Radarr servers reachable by Seerr
data "seerr_radarr_service" "all" {}

# Read a specific Radarr server by ID and its quality profiles
data "seerr_radarr_service" "single" {
  server_id = 1
}
