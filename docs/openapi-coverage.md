# OpenAPI Coverage Matrix

Inventory of Seerr OpenAPI endpoints classified by Terraform provider support and scope.

## Coverage Summary

- **Total OpenAPI Paths**: 163
- **Total Endpoints (Methods)**: 212
- **Covered Paths**: 47 (28.8%)
- **Intentionally Out of Scope**: 106 (65.0%)
- **Uncovered Settings/User Backlog**: 10 (6.1%)

## Classification Legend

- `covered`: Endpoint is fully mapped to a Terraform resource or data source.
- `uncovered`: Configuration or user endpoint suitable for future Terraform provider coverage (backlog candidate).
- `intentionally-out-of-scope`: Interactive search/browse, transient test action, session auth, or runtime log/cache flush endpoint out of scope for infrastructure-as-code management.

## Path Coverage Matrix

| Path | Methods | Classification | Mapped Resource / Notes |
|---|---|---|---|
| `/auth/jellyfin` | POST | `intentionally-out-of-scope` | Interactive Jellyfin login |
| `/auth/jellyfin/quickconnect/authenticate` | POST | `intentionally-out-of-scope` | Interactive login |
| `/auth/jellyfin/quickconnect/check` | GET | `intentionally-out-of-scope` | Interactive login |
| `/auth/jellyfin/quickconnect/initiate` | POST | `intentionally-out-of-scope` | Interactive login |
| `/auth/local` | POST | `intentionally-out-of-scope` | Interactive login endpoint |
| `/auth/logout` | POST | `intentionally-out-of-scope` | Session logout action |
| `/auth/me` | GET | `intentionally-out-of-scope` | Session authentication status |
| `/auth/plex` | POST | `intentionally-out-of-scope` | Interactive Plex OAuth login endpoint |
| `/auth/reset-password` | POST | `intentionally-out-of-scope` | Interactive password reset |
| `/auth/reset-password/{guid}` | POST | `intentionally-out-of-scope` | Interactive password reset |
| `/backdrops` | GET | `intentionally-out-of-scope` | TMDB/media UI backdrop lookup |
| `/blacklist` | GET, POST | `intentionally-out-of-scope` | Legacy alias for blocklist |
| `/blacklist/{tmdbId}` | DELETE, GET | `intentionally-out-of-scope` | Legacy alias for blocklist |
| `/blocklist` | GET, POST | `covered` | `seerr_blocklist` - Blocklist resource & data source |
| `/blocklist/collection/{collectionId}` | DELETE, POST | `covered` | `seerr_blocklist` - Blocklist collection resource |
| `/blocklist/{tmdbId}` | DELETE, GET | `covered` | `seerr_blocklist` - Blocklist item resource |
| `/certifications/movie` | GET | `intentionally-out-of-scope` | TMDB certifications lookup |
| `/certifications/tv` | GET | `intentionally-out-of-scope` | TMDB certifications lookup |
| `/collection/{collectionId}` | GET | `intentionally-out-of-scope` | TMDB collection details query |
| `/discover/genreslider/movie` | GET | `intentionally-out-of-scope` | Interactive genre slider lookup |
| `/discover/genreslider/tv` | GET | `intentionally-out-of-scope` | Interactive genre slider lookup |
| `/discover/keyword/{keywordId}/movies` | GET | `intentionally-out-of-scope` | Interactive keyword movie lookup |
| `/discover/movies` | GET | `intentionally-out-of-scope` | Interactive movie discovery endpoint |
| `/discover/movies/genre/{genreId}` | GET | `intentionally-out-of-scope` | Interactive movie lookup |
| `/discover/movies/language/{language}` | GET | `intentionally-out-of-scope` | Interactive movie lookup |
| `/discover/movies/studio/{studioId}` | GET | `intentionally-out-of-scope` | Interactive movie lookup |
| `/discover/movies/upcoming` | GET | `intentionally-out-of-scope` | Interactive upcoming movies endpoint |
| `/discover/trending` | GET | `intentionally-out-of-scope` | Interactive trending media endpoint |
| `/discover/tv` | GET | `intentionally-out-of-scope` | Interactive TV discovery endpoint |
| `/discover/tv/genre/{genreId}` | GET | `intentionally-out-of-scope` | Interactive TV lookup |
| `/discover/tv/language/{language}` | GET | `intentionally-out-of-scope` | Interactive TV lookup |
| `/discover/tv/network/{networkId}` | GET | `intentionally-out-of-scope` | Interactive TV lookup |
| `/discover/tv/upcoming` | GET | `intentionally-out-of-scope` | Interactive upcoming TV endpoint |
| `/discover/watchlist` | GET | `intentionally-out-of-scope` | Interactive watchlist discovery |
| `/genres/movie` | GET | `intentionally-out-of-scope` | TMDB genres lookup |
| `/genres/tv` | GET | `intentionally-out-of-scope` | TMDB genres lookup |
| `/issue` | GET, POST | `covered` | `seerr_issue` - Issue resource & data source |
| `/issue/count` | GET | `covered` | `seerr_issues` - Issue counts data source |
| `/issue/{issueId}` | DELETE, GET | `covered` | `seerr_issue` - Issue item resource |
| `/issue/{issueId}/comment` | POST | `intentionally-out-of-scope` | Issue comment action |
| `/issue/{issueId}/{status}` | POST | `intentionally-out-of-scope` | Issue status update action |
| `/issueComment/{commentId}` | DELETE, GET, PUT | `intentionally-out-of-scope` | Issue comment management |
| `/keyword/{keywordId}` | GET | `intentionally-out-of-scope` | TMDB keyword details |
| `/languages` | GET | `intentionally-out-of-scope` | TMDB languages list |
| `/media` | GET | `covered` | `seerr_media` - Media list data source |
| `/media/{mediaId}` | DELETE | `covered` | `seerr_media_item` - Media item data source |
| `/media/{mediaId}/file` | DELETE | `intentionally-out-of-scope` | Media file info |
| `/media/{mediaId}/watch_data` | GET | `intentionally-out-of-scope` | Media watch stats query |
| `/media/{mediaId}/{status}` | POST | `intentionally-out-of-scope` | Media status update action (available/processing/pending) |
| `/movie/{movieId}` | GET | `intentionally-out-of-scope` | TMDB movie details query |
| `/movie/{movieId}/ratings` | GET | `intentionally-out-of-scope` | TMDB movie ratings query |
| `/movie/{movieId}/ratingscombined` | GET | `intentionally-out-of-scope` | TMDB movie combined ratings query |
| `/movie/{movieId}/recommendations` | GET | `intentionally-out-of-scope` | TMDB movie recommendations query |
| `/movie/{movieId}/similar` | GET | `intentionally-out-of-scope` | TMDB movie similar query |
| `/network/{networkId}` | GET | `intentionally-out-of-scope` | TMDB network details |
| `/overrideRule` | GET, POST | `covered` | `seerr_override_rule` - Override rule resource |
| `/overrideRule/{ruleId}` | DELETE, PUT | `covered` | `seerr_override_rule` - Override rule item resource |
| `/person/{personId}` | GET | `intentionally-out-of-scope` | TMDB person details query |
| `/person/{personId}/combined_credits` | GET | `intentionally-out-of-scope` | TMDB person credits query |
| `/regions` | GET | `intentionally-out-of-scope` | TMDB regions list |
| `/request` | GET, POST | `covered` | `seerr_request` - Media request resource & data source |
| `/request/count` | GET | `covered` | `seerr_requests` - Media request counts data source |
| `/request/{requestId}` | DELETE, GET, PUT | `covered` | `seerr_request` - Media request item resource |
| `/request/{requestId}/retry` | POST | `covered` | `seerr_request_retry` - Request retry resource |
| `/request/{requestId}/{status}` | POST | `intentionally-out-of-scope` | Request workflow status change (approve/decline/autoApprove) |
| `/search` | GET | `intentionally-out-of-scope` | Interactive search endpoint |
| `/search/company` | GET | `intentionally-out-of-scope` | Interactive TMDB company search |
| `/search/keyword` | GET | `intentionally-out-of-scope` | Interactive TMDB keyword search |
| `/service/radarr` | GET | `intentionally-out-of-scope` | Service health probe for Radarr |
| `/service/radarr/{radarrId}` | GET | `intentionally-out-of-scope` | Service instance probe for Radarr |
| `/service/sonarr` | GET | `intentionally-out-of-scope` | Service health probe for Sonarr |
| `/service/sonarr/lookup/{tmdbId}` | GET | `intentionally-out-of-scope` | Sonarr series lookup by TMDB ID |
| `/service/sonarr/{sonarrId}` | GET | `intentionally-out-of-scope` | Service instance probe for Sonarr |
| `/settings/about` | GET | `intentionally-out-of-scope` | System diagnostics and about endpoint |
| `/settings/cache` | GET | `intentionally-out-of-scope` | Runtime cache status & flush controls |
| `/settings/cache/dns/{dnsEntry}/flush` | POST | `intentionally-out-of-scope` | Runtime DNS cache flush endpoint |
| `/settings/cache/{cacheId}/flush` | POST | `intentionally-out-of-scope` | Runtime cache flush endpoint |
| `/settings/discover` | GET, POST | `covered` | `seerr_discover_slider` - Discover slider resource & data source |
| `/settings/discover/add` | POST | `covered` | `seerr_discover_slider` - Discover slider add action |
| `/settings/discover/reset` | GET | `covered` | `seerr_discover_slider` - Discover slider reset helper |
| `/settings/discover/{sliderId}` | DELETE, PUT | `covered` | `seerr_discover_slider` - Discover slider item resource |
| `/settings/initialize` | POST | `intentionally-out-of-scope` | First-run initial setup wizard endpoint |
| `/settings/jellyfin` | GET, POST | `covered` | `seerr_jellyfin_settings` - Jellyfin settings resource & data source |
| `/settings/jellyfin/library` | GET | `uncovered` | Jellyfin library sync settings (Issue #121 candidate) |
| `/settings/jellyfin/sync` | GET, POST | `uncovered` | Jellyfin library sync trigger (Issue #121 candidate) |
| `/settings/jellyfin/users` | GET | `intentionally-out-of-scope` | Transient user lookup for Jellyfin server configuration |
| `/settings/jobs` | GET | `covered` | `seerr_jobs` - Jobs data source |
| `/settings/jobs/{jobId}/cancel` | POST | `uncovered` | Job cancel action (Issue #123 candidate) |
| `/settings/jobs/{jobId}/run` | POST | `uncovered` | Job trigger action (Issue #123 candidate) |
| `/settings/jobs/{jobId}/schedule` | POST | `covered` | `seerr_job_schedule` - Job schedule resource |
| `/settings/logs` | GET | `intentionally-out-of-scope` | Runtime application logs endpoint |
| `/settings/main` | GET, POST | `covered` | `seerr_main_settings` - Main settings resource & data source |
| `/settings/main/regenerate` | POST | `intentionally-out-of-scope` | API key regeneration helper on main settings endpoint |
| `/settings/metadatas` | GET, PUT | `covered` | `seerr_metadata_settings` - Metadata settings resource & data source |
| `/settings/metadatas/test` | POST | `intentionally-out-of-scope` | Transient test endpoint for metadata provider |
| `/settings/network` | GET, POST | `covered` | `seerr_network_settings` - Network settings resource & data source |
| `/settings/notifications/discord` | GET, POST | `covered` | `seerr_notification_discord` - Discord notification agent resource & data source |
| `/settings/notifications/discord/test` | POST | `intentionally-out-of-scope` | Transient notification test action |
| `/settings/notifications/email` | GET, POST | `covered` | `seerr_notification_email` - Email notification agent resource & data source |
| `/settings/notifications/email/test` | POST | `intentionally-out-of-scope` | Transient notification test action |
| `/settings/notifications/gotify` | GET, POST | `covered` | `seerr_notification_gotify` - Gotify notification agent resource & data source |
| `/settings/notifications/gotify/test` | POST | `intentionally-out-of-scope` | Transient notification test action |
| `/settings/notifications/ntfy` | GET, POST | `covered` | `seerr_notification_ntfy` - Ntfy notification agent resource & data source |
| `/settings/notifications/ntfy/test` | POST | `intentionally-out-of-scope` | Transient notification test action |
| `/settings/notifications/pushbullet` | GET, POST | `covered` | `seerr_notification_pushbullet` - Pushbullet notification agent resource & data source |
| `/settings/notifications/pushbullet/test` | POST | `intentionally-out-of-scope` | Transient notification test action |
| `/settings/notifications/pushover` | GET, POST | `covered` | `seerr_notification_pushover` - Pushover notification agent resource & data source |
| `/settings/notifications/pushover/sounds` | GET | `intentionally-out-of-scope` | Static sound lookup helper for Pushover |
| `/settings/notifications/pushover/test` | POST | `intentionally-out-of-scope` | Transient notification test action |
| `/settings/notifications/slack` | GET, POST | `covered` | `seerr_notification_slack` - Slack notification agent resource & data source |
| `/settings/notifications/slack/test` | POST | `intentionally-out-of-scope` | Transient notification test action |
| `/settings/notifications/telegram` | GET, POST | `covered` | `seerr_notification_telegram` - Telegram notification agent resource & data source |
| `/settings/notifications/telegram/test` | POST | `intentionally-out-of-scope` | Transient notification test action |
| `/settings/notifications/webhook` | GET, POST | `covered` | `seerr_notification_webhook` - Webhook notification agent resource & data source |
| `/settings/notifications/webhook/test` | POST | `intentionally-out-of-scope` | Transient notification test action |
| `/settings/notifications/webpush` | GET, POST | `covered` | `seerr_notification_webpush` - Webpush notification agent resource & data source |
| `/settings/notifications/webpush/test` | POST | `intentionally-out-of-scope` | Transient notification test action |
| `/settings/plex` | GET, POST | `covered` | `seerr_plex_settings` - Plex settings resource & data source |
| `/settings/plex/devices/servers` | GET | `intentionally-out-of-scope` | Transient device discovery endpoint for Plex |
| `/settings/plex/library` | GET | `uncovered` | Plex library sync settings (Issue #121 candidate) |
| `/settings/plex/sync` | GET, POST | `uncovered` | Plex library sync trigger (Issue #121 candidate) |
| `/settings/plex/users` | GET | `intentionally-out-of-scope` | Transient user lookup for Plex server configuration |
| `/settings/public` | GET | `covered` | `seerr_public_settings` - Public settings data source |
| `/settings/radarr` | GET, POST | `covered` | `seerr_radarr_server` - Radarr server resource & data source |
| `/settings/radarr/test` | POST | `intentionally-out-of-scope` | Transient connection test action for Radarr |
| `/settings/radarr/{radarrId}` | DELETE, PUT | `covered` | `seerr_radarr_server` - Radarr server instance resource |
| `/settings/radarr/{radarrId}/profiles` | GET | `covered` | `seerr_radarr_quality_profile` - Radarr quality profiles data source |
| `/settings/sonarr` | GET, POST | `covered` | `seerr_sonarr_server` - Sonarr server resource & data source |
| `/settings/sonarr/test` | POST | `intentionally-out-of-scope` | Transient connection test action for Sonarr |
| `/settings/sonarr/{sonarrId}` | DELETE, PUT | `covered` | `seerr_sonarr_server` - Sonarr server instance resource |
| `/settings/tautulli` | GET, POST | `covered` | `seerr_tautulli_settings` - Tautulli settings resource & data source |
| `/status` | GET | `intentionally-out-of-scope` | Service health & version status endpoint |
| `/status/appdata` | GET | `intentionally-out-of-scope` | App data path info endpoint |
| `/studio/{studioId}` | GET | `intentionally-out-of-scope` | TMDB studio details |
| `/tv/{tvId}` | GET | `intentionally-out-of-scope` | TMDB TV details query |
| `/tv/{tvId}/ratings` | GET | `intentionally-out-of-scope` | TMDB TV ratings query |
| `/tv/{tvId}/recommendations` | GET | `intentionally-out-of-scope` | TMDB TV recommendations query |
| `/tv/{tvId}/season/{seasonNumber}` | GET | `intentionally-out-of-scope` | TMDB TV season details |
| `/tv/{tvId}/similar` | GET | `intentionally-out-of-scope` | TMDB TV similar query |
| `/user` | GET, POST, PUT | `covered` | `seerr_user` - User list/create resource & data source |
| `/user/import-from-jellyfin` | POST | `uncovered` | Jellyfin user import action (Issue #124 candidate) |
| `/user/import-from-plex` | POST | `uncovered` | Plex user import action (Issue #124 candidate) |
| `/user/jellyfin/{jellyfinUserId}` | GET | `intentionally-out-of-scope` | Jellyfin user lookup |
| `/user/registerPushSubscription` | POST | `intentionally-out-of-scope` | Browser push notification registration |
| `/user/{userId}` | DELETE, GET, PUT | `covered` | `seerr_user` - User resource & data source |
| `/user/{userId}/pushSubscription/{endpoint}` | DELETE, GET | `intentionally-out-of-scope` | Browser push notification item |
| `/user/{userId}/pushSubscriptions` | GET | `intentionally-out-of-scope` | Browser push notifications list |
| `/user/{userId}/quota` | GET | `uncovered` | User quota settings (Issue #122 candidate) |
| `/user/{userId}/requests` | GET | `intentionally-out-of-scope` | Per-user request history query endpoint |
| `/user/{userId}/settings/linked-accounts/jellyfin` | DELETE, POST | `intentionally-out-of-scope` | Jellyfin account linking |
| `/user/{userId}/settings/linked-accounts/jellyfin/quickconnect` | POST | `intentionally-out-of-scope` | Jellyfin QuickConnect linking |
| `/user/{userId}/settings/linked-accounts/plex` | DELETE, POST | `intentionally-out-of-scope` | Plex account linking |
| `/user/{userId}/settings/main` | GET, POST | `covered` | `seerr_user` - User main settings |
| `/user/{userId}/settings/notifications` | GET, POST | `uncovered` | User notification settings (backlog candidate) |
| `/user/{userId}/settings/password` | GET, POST | `intentionally-out-of-scope` | Interactive password update action |
| `/user/{userId}/settings/permissions` | GET, POST | `covered` | `seerr_user_permissions` - User permissions resource & data source |
| `/user/{userId}/watch_data` | GET | `intentionally-out-of-scope` | Per-user watch data query |
| `/user/{userId}/watchlist` | GET | `covered` | `seerr_user_watchlist_settings` - User watchlist settings resource & data source |
| `/watchlist` | POST | `intentionally-out-of-scope` | User watchlist items endpoint |
| `/watchlist/{tmdbId}` | DELETE | `intentionally-out-of-scope` | Watchlist item query |
| `/watchproviders/movies` | GET | `intentionally-out-of-scope` | Watch providers query |
| `/watchproviders/regions` | GET | `intentionally-out-of-scope` | Watch provider regions query |
| `/watchproviders/tv` | GET | `intentionally-out-of-scope` | Watch providers query |

## Refreshing the Snapshot & Inventory

To update the OpenAPI snapshot from upstream Seerr and re-validate coverage:

```bash
# Download the latest seerr-api.yml snapshot
curl -fsSL https://raw.githubusercontent.com/seerr-team/seerr/develop/seerr-api.yml -o tools/openapi/seerr-api.yml

# Run the OpenAPI coverage audit and generate openapi-coverage.md
go test -v ./tools/openapi/...
```
