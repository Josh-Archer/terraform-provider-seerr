# Full Media Stack Composite Module

This example module demonstrates how to configure a complete, production-grade media automation environment using `terraform-provider-seerr`.

It shows how to connect Seerr with:
- **Media Servers**: Plex or Jellyfin
- **Servarr Integrations**: Sonarr and Radarr using direct quality profile and root folder resolvers (`seerr_radarr_quality_profile`, `seerr_radarr_root_folders`, `seerr_sonarr_quality_profile`, `seerr_sonarr_root_folders`)
- **Content Routing**: `seerr_override_rule` to automatically route Anime to dedicated root folders
- **Discover Customization**: `seerr_discover_slider` to curate home page sections
- **Notifications**: Discord and Ntfy alerting channels

## Architecture Diagram

```
                 +-------------------+
                 |    OpenTofu /     |
                 |     Terraform     |
                 +--------+----------+
                          |
              +-----------+-----------+
              |                       |
              v                       v
     +-----------------+     +-----------------+
     |  Direct Resolvers|    | Seerr Provider  |
     | (Profiles/Roots)|     |   Management    |
     +--------+--------+     +--------+--------+
              |                       |
      +-------+-------+               v
      |               |      +-----------------+
      v               v      |  Seerr Instance |
+-----------+   +-----------+|  (Jellyseerr /  |
|  Radarr   |   |  Sonarr   ||   Overseerr)    |
+-----------+   +-----------++-----------------+
```

## Usage Example

```hcl
module "media_stack" {
  source = "./examples/modules/full-media-stack"

  seerr_url = "https://seerr.homelab.local"

  # Media Server
  enable_plex = true
  plex_host   = "plex.homelab.local"
  plex_port   = 32400

  # Servarr Applications
  enable_radarr               = true
  radarr_url                  = "http://radarr.homelab.local:7878"
  radarr_host                 = "radarr.homelab.local"
  radarr_api_key              = var.radarr_api_key
  radarr_quality_profile_name = "HD-1080p"

  enable_sonarr               = true
  sonarr_url                  = "http://sonarr.homelab.local:8989"
  sonarr_host                 = "sonarr.homelab.local"
  sonarr_api_key              = var.sonarr_api_key
  sonarr_quality_profile_name = "HD-1080p"
  anime_root_folder           = "/media/anime"

  # Notifications
  discord_webhook_url = var.discord_webhook_url
}
```
