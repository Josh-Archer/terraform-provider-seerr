package openapi

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Classification string

const (
	ClassCovered                 Classification = "covered"
	ClassIntentionallyOutOfScope Classification = "intentionally-out-of-scope"
	ClassUncovered               Classification = "uncovered"
)

type PathRule struct {
	PathPattern    string         `json:"path_pattern"`
	Classification Classification `json:"classification"`
	Notes          string         `json:"notes"`
	MappedResource string         `json:"mapped_resource,omitempty"`
}

type OpenAPISpec struct {
	Paths map[string]map[string]any `yaml:"paths"`
}

type PathCoverageResult struct {
	Path           string         `json:"path"`
	Methods        []string       `json:"methods"`
	Classification Classification `json:"classification"`
	Notes          string         `json:"notes"`
	MappedResource string         `json:"mapped_resource,omitempty"`
}

type CoverageReport struct {
	TotalEndpoints               int                  `json:"total_endpoints"`
	TotalPaths                   int                  `json:"total_paths"`
	CoveredPaths                 int                  `json:"covered_paths"`
	IntentionallyOutOfScopePaths int                  `json:"intentionally_out_of_scope_paths"`
	UncoveredPaths               int                  `json:"uncovered_paths"`
	Results                      []PathCoverageResult `json:"results"`
	UnclassifiedPaths            []string             `json:"unclassified_paths"`
}

// DefaultRules defines the classification matrix for all Seerr OpenAPI paths.
func DefaultRules() []PathRule {
	return []PathRule{
		// --- Covered Settings Resources & Data Sources ---
		{PathPattern: "/settings/main", Classification: ClassCovered, MappedResource: "seerr_main_settings", Notes: "Main settings resource & data source"},
		{PathPattern: "/settings/network", Classification: ClassCovered, MappedResource: "seerr_network_settings", Notes: "Network settings resource & data source"},
		{PathPattern: "/settings/jellyfin", Classification: ClassCovered, MappedResource: "seerr_jellyfin_settings", Notes: "Jellyfin settings resource & data source"},
		{PathPattern: "/settings/plex", Classification: ClassCovered, MappedResource: "seerr_plex_settings", Notes: "Plex settings resource & data source"},
		{PathPattern: "/settings/metadatas", Classification: ClassCovered, MappedResource: "seerr_metadata_settings", Notes: "Metadata settings resource & data source"},
		{PathPattern: "/settings/tautulli", Classification: ClassCovered, MappedResource: "seerr_tautulli_settings", Notes: "Tautulli settings resource & data source"},
		{PathPattern: "/settings/radarr", Classification: ClassCovered, MappedResource: "seerr_radarr_server", Notes: "Radarr server resource & data source"},
		{PathPattern: "/settings/radarr/{radarrId}", Classification: ClassCovered, MappedResource: "seerr_radarr_server", Notes: "Radarr server instance resource"},
		{PathPattern: "/settings/radarr/{radarrId}/profiles", Classification: ClassCovered, MappedResource: "seerr_radarr_quality_profile", Notes: "Radarr quality profiles data source"},
		{PathPattern: "/settings/sonarr", Classification: ClassCovered, MappedResource: "seerr_sonarr_server", Notes: "Sonarr server resource & data source"},
		{PathPattern: "/settings/sonarr/{sonarrId}", Classification: ClassCovered, MappedResource: "seerr_sonarr_server", Notes: "Sonarr server instance resource"},
		{PathPattern: "/settings/sonarr/{sonarrId}/profiles", Classification: ClassCovered, MappedResource: "seerr_sonarr_quality_profile", Notes: "Sonarr quality profiles data source"},
		{PathPattern: "/settings/public", Classification: ClassCovered, MappedResource: "seerr_public_settings", Notes: "Public settings data source"},
		{PathPattern: "/settings/jobs", Classification: ClassCovered, MappedResource: "seerr_jobs", Notes: "Jobs data source"},
		{PathPattern: "/settings/jobs/{jobId}/schedule", Classification: ClassCovered, MappedResource: "seerr_job_schedule", Notes: "Job schedule resource"},
		{PathPattern: "/settings/notifications/email", Classification: ClassCovered, MappedResource: "seerr_notification_email", Notes: "Email notification agent resource & data source"},
		{PathPattern: "/settings/notifications/discord", Classification: ClassCovered, MappedResource: "seerr_notification_discord", Notes: "Discord notification agent resource & data source"},
		{PathPattern: "/settings/notifications/pushbullet", Classification: ClassCovered, MappedResource: "seerr_notification_pushbullet", Notes: "Pushbullet notification agent resource & data source"},
		{PathPattern: "/settings/notifications/pushover", Classification: ClassCovered, MappedResource: "seerr_notification_pushover", Notes: "Pushover notification agent resource & data source"},
		{PathPattern: "/settings/notifications/gotify", Classification: ClassCovered, MappedResource: "seerr_notification_gotify", Notes: "Gotify notification agent resource & data source"},
		{PathPattern: "/settings/notifications/ntfy", Classification: ClassCovered, MappedResource: "seerr_notification_ntfy", Notes: "Ntfy notification agent resource & data source"},
		{PathPattern: "/settings/notifications/slack", Classification: ClassCovered, MappedResource: "seerr_notification_slack", Notes: "Slack notification agent resource & data source"},
		{PathPattern: "/settings/notifications/telegram", Classification: ClassCovered, MappedResource: "seerr_notification_telegram", Notes: "Telegram notification agent resource & data source"},
		{PathPattern: "/settings/notifications/webhook", Classification: ClassCovered, MappedResource: "seerr_notification_webhook", Notes: "Webhook notification agent resource & data source"},
		{PathPattern: "/settings/notifications/webpush", Classification: ClassCovered, MappedResource: "seerr_notification_webpush", Notes: "Webpush notification agent resource & data source"},
		{PathPattern: "/settings/discover", Classification: ClassCovered, MappedResource: "seerr_discover_slider", Notes: "Discover slider resource & data source"},
		{PathPattern: "/settings/discover/reset", Classification: ClassCovered, MappedResource: "seerr_discover_slider", Notes: "Discover slider reset helper"},
		{PathPattern: "/settings/override-rules", Classification: ClassCovered, MappedResource: "seerr_override_rule", Notes: "Override rule resource & data source"},
		{PathPattern: "/settings/override-rules/{ruleId}", Classification: ClassCovered, MappedResource: "seerr_override_rule", Notes: "Override rule item resource"},

		// --- Covered Users, Permissions, Watchlist, Requests, Issues, Blocklist, Backup ---
		{PathPattern: "/user", Classification: ClassCovered, MappedResource: "seerr_user", Notes: "User list/create resource & data source"},
		{PathPattern: "/user/invitable-users", Classification: ClassCovered, MappedResource: "seerr_user_invitations", Notes: "User invitations data source"},
		{PathPattern: "/user/invitations", Classification: ClassCovered, MappedResource: "seerr_user_invitations", Notes: "User invitations data source"},
		{PathPattern: "/user/invitations/create", Classification: ClassCovered, MappedResource: "seerr_user_invitation", Notes: "User invitation resource"},
		{PathPattern: "/user/invitations/{invitationId}", Classification: ClassCovered, MappedResource: "seerr_user_invitation", Notes: "User invitation resource"},
		{PathPattern: "/user/{userId}", Classification: ClassCovered, MappedResource: "seerr_user", Notes: "User resource & data source"},
		{PathPattern: "/user/{userId}/settings/permissions", Classification: ClassCovered, MappedResource: "seerr_user_permissions", Notes: "User permissions resource & data source"},
		{PathPattern: "/user/{userId}/settings/main", Classification: ClassCovered, MappedResource: "seerr_user", Notes: "User main settings"},
		{PathPattern: "/user/{userId}/watchlist", Classification: ClassCovered, MappedResource: "seerr_user_watchlist_settings", Notes: "User watchlist settings resource & data source"},
		{PathPattern: "/request", Classification: ClassCovered, MappedResource: "seerr_request", Notes: "Media request resource & data source"},
		{PathPattern: "/request/count", Classification: ClassCovered, MappedResource: "seerr_requests", Notes: "Media request counts data source"},
		{PathPattern: "/request/{requestId}", Classification: ClassCovered, MappedResource: "seerr_request", Notes: "Media request item resource"},
		{PathPattern: "/request/{requestId}/retry", Classification: ClassCovered, MappedResource: "seerr_request_retry", Notes: "Request retry resource"},
		{PathPattern: "/issue", Classification: ClassCovered, MappedResource: "seerr_issue", Notes: "Issue resource & data source"},
		{PathPattern: "/issue/count", Classification: ClassCovered, MappedResource: "seerr_issues", Notes: "Issue counts data source"},
		{PathPattern: "/issue/{issueId}", Classification: ClassCovered, MappedResource: "seerr_issue", Notes: "Issue item resource"},
		{PathPattern: "/blocklist", Classification: ClassCovered, MappedResource: "seerr_blocklist", Notes: "Blocklist resource & data source"},
		{PathPattern: "/blocklist/{blocklistId}", Classification: ClassCovered, MappedResource: "seerr_blocklist", Notes: "Blocklist item resource"},
		{PathPattern: "/settings/backup", Classification: ClassCovered, MappedResource: "seerr_backup_settings", Notes: "Backup settings resource & data source"},
		{PathPattern: "/settings/backup/{backupId}", Classification: ClassCovered, MappedResource: "seerr_backup_settings", Notes: "Backup item resource"},
		{PathPattern: "/api-key", Classification: ClassCovered, MappedResource: "seerr_api_key", Notes: "API key resource & data source"},
		{PathPattern: "/api-key/regenerate", Classification: ClassCovered, MappedResource: "seerr_api_key", Notes: "API key regenerate action"},

		// --- Uncovered Settings / User Endpoints (Backlog Candidates) ---
		{PathPattern: "/settings/jobs/{jobId}/run", Classification: ClassUncovered, Notes: "Job trigger action (Issue #123 candidate)"},
		{PathPattern: "/settings/jobs/{jobId}/cancel", Classification: ClassUncovered, Notes: "Job cancel action (Issue #123 candidate)"},
		{PathPattern: "/settings/plex/library", Classification: ClassUncovered, Notes: "Plex library sync settings (Issue #121 candidate)"},
		{PathPattern: "/settings/plex/sync", Classification: ClassUncovered, Notes: "Plex library sync trigger (Issue #121 candidate)"},
		{PathPattern: "/settings/jellyfin/library", Classification: ClassUncovered, Notes: "Jellyfin library sync settings (Issue #121 candidate)"},
		{PathPattern: "/settings/jellyfin/sync", Classification: ClassUncovered, Notes: "Jellyfin library sync trigger (Issue #121 candidate)"},
		{PathPattern: "/user/import-from-plex", Classification: ClassUncovered, Notes: "Plex user import action (Issue #124 candidate)"},
		{PathPattern: "/user/import-from-jellyfin", Classification: ClassUncovered, Notes: "Jellyfin user import action (Issue #124 candidate)"},
		{PathPattern: "/user/{userId}/quota", Classification: ClassUncovered, Notes: "User quota settings (Issue #122 candidate)"},

		// --- Intentionally Out of Scope: Transient test actions, runtime flushes/logs, auth, status, search/tmdb ---
		{PathPattern: "/settings/main/regenerate", Classification: ClassIntentionallyOutOfScope, Notes: "API key regeneration helper on main settings endpoint"},
		{PathPattern: "/settings/jellyfin/users", Classification: ClassIntentionallyOutOfScope, Notes: "Transient user lookup for Jellyfin server configuration"},
		{PathPattern: "/settings/plex/devices/servers", Classification: ClassIntentionallyOutOfScope, Notes: "Transient device discovery endpoint for Plex"},
		{PathPattern: "/settings/plex/users", Classification: ClassIntentionallyOutOfScope, Notes: "Transient user lookup for Plex server configuration"},
		{PathPattern: "/settings/metadatas/test", Classification: ClassIntentionallyOutOfScope, Notes: "Transient test endpoint for metadata provider"},
		{PathPattern: "/settings/radarr/test", Classification: ClassIntentionallyOutOfScope, Notes: "Transient connection test action for Radarr"},
		{PathPattern: "/settings/sonarr/test", Classification: ClassIntentionallyOutOfScope, Notes: "Transient connection test action for Sonarr"},
		{PathPattern: "/settings/initialize", Classification: ClassIntentionallyOutOfScope, Notes: "First-run initial setup wizard endpoint"},
		{PathPattern: "/settings/cache", Classification: ClassIntentionallyOutOfScope, Notes: "Runtime cache status & flush controls"},
		{PathPattern: "/settings/cache/{cacheId}/flush", Classification: ClassIntentionallyOutOfScope, Notes: "Runtime cache flush endpoint"},
		{PathPattern: "/settings/cache/dns/{dnsEntry}/flush", Classification: ClassIntentionallyOutOfScope, Notes: "Runtime DNS cache flush endpoint"},
		{PathPattern: "/settings/logs", Classification: ClassIntentionallyOutOfScope, Notes: "Runtime application logs endpoint"},
		{PathPattern: "/settings/notifications/email/test", Classification: ClassIntentionallyOutOfScope, Notes: "Transient notification test action"},
		{PathPattern: "/settings/notifications/discord/test", Classification: ClassIntentionallyOutOfScope, Notes: "Transient notification test action"},
		{PathPattern: "/settings/notifications/pushbullet/test", Classification: ClassIntentionallyOutOfScope, Notes: "Transient notification test action"},
		{PathPattern: "/settings/notifications/pushover/test", Classification: ClassIntentionallyOutOfScope, Notes: "Transient notification test action"},
		{PathPattern: "/settings/notifications/pushover/sounds", Classification: ClassIntentionallyOutOfScope, Notes: "Static sound lookup helper for Pushover"},
		{PathPattern: "/settings/notifications/gotify/test", Classification: ClassIntentionallyOutOfScope, Notes: "Transient notification test action"},
		{PathPattern: "/settings/notifications/ntfy/test", Classification: ClassIntentionallyOutOfScope, Notes: "Transient notification test action"},
		{PathPattern: "/settings/notifications/slack/test", Classification: ClassIntentionallyOutOfScope, Notes: "Transient notification test action"},
		{PathPattern: "/settings/notifications/telegram/test", Classification: ClassIntentionallyOutOfScope, Notes: "Transient notification test action"},
		{PathPattern: "/settings/notifications/webhook/test", Classification: ClassIntentionallyOutOfScope, Notes: "Transient notification test action"},
		{PathPattern: "/settings/notifications/webpush/test", Classification: ClassIntentionallyOutOfScope, Notes: "Transient notification test action"},
		{PathPattern: "/settings/about", Classification: ClassIntentionallyOutOfScope, Notes: "System diagnostics and about endpoint"},
		{PathPattern: "/status", Classification: ClassIntentionallyOutOfScope, Notes: "Service health & version status endpoint"},
		{PathPattern: "/status/appdata", Classification: ClassIntentionallyOutOfScope, Notes: "App data path info endpoint"},
		{PathPattern: "/user/me", Classification: ClassIntentionallyOutOfScope, Notes: "Session profile endpoint (covered by data source seerr_current_user)"},
		{PathPattern: "/user/register", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive user registration endpoint"},
		{PathPattern: "/user/{userId}/settings/password", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive password update action"},
		{PathPattern: "/user/{userId}/requests", Classification: ClassIntentionallyOutOfScope, Notes: "Per-user request history query endpoint"},
		{PathPattern: "/auth/me", Classification: ClassIntentionallyOutOfScope, Notes: "Session authentication status"},
		{PathPattern: "/auth/local", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive login endpoint"},
		{PathPattern: "/auth/plex", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive Plex OAuth login endpoint"},
		{PathPattern: "/auth/logout", Classification: ClassIntentionallyOutOfScope, Notes: "Session logout action"},
		{PathPattern: "/request/{requestId}/{status}", Classification: ClassIntentionallyOutOfScope, Notes: "Request workflow status change (approve/decline/autoApprove)"},
		{PathPattern: "/request/{requestId}/send_retry", Classification: ClassIntentionallyOutOfScope, Notes: "Manual request retry trigger action"},
		{PathPattern: "/issue/{issueId}/comment", Classification: ClassIntentionallyOutOfScope, Notes: "Issue comment action"},
		{PathPattern: "/issue/{issueId}/comment/{commentId}", Classification: ClassIntentionallyOutOfScope, Notes: "Issue comment item resource"},
		{PathPattern: "/issue/{issueId}/{status}", Classification: ClassIntentionallyOutOfScope, Notes: "Issue status update action"},
		{PathPattern: "/issue/comment/{commentId}", Classification: ClassIntentionallyOutOfScope, Notes: "Issue comment item resource"},
		{PathPattern: "/search", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive search endpoint"},
		{PathPattern: "/search/company", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive TMDB company search"},
		{PathPattern: "/search/keyword", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive TMDB keyword search"},
		{PathPattern: "/discover/movies", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive movie discovery endpoint"},
		{PathPattern: "/discover/movies/upcoming", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive upcoming movies endpoint"},
		{PathPattern: "/discover/tv", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive TV discovery endpoint"},
		{PathPattern: "/discover/tv/upcoming", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive upcoming TV endpoint"},
		{PathPattern: "/discover/trending", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive trending media endpoint"},
		{PathPattern: "/discover/keyword/{keywordId}/movies", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive keyword movie lookup"},
		{PathPattern: "/discover/genres/movies", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive movie genres lookup"},
		{PathPattern: "/discover/genres/tv", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive TV genres lookup"},
		{PathPattern: "/discover/genres/languagelist", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive language list lookup"},
		{PathPattern: "/discover/watchproviders/movies", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive watch providers lookup"},
		{PathPattern: "/discover/watchproviders/tv", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive watch providers lookup"},
		{PathPattern: "/discover/watchproviders/regions", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive watch provider regions lookup"},
		{PathPattern: "/movie/{movieId}", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB movie details query"},
		{PathPattern: "/movie/{movieId}/ratings", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB movie ratings query"},
		{PathPattern: "/movie/{movieId}/ratingscombined", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB movie combined ratings query"},
		{PathPattern: "/movie/{movieId}/recommendations", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB movie recommendations query"},
		{PathPattern: "/movie/{movieId}/similar", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB movie similar query"},
		{PathPattern: "/tv/{tvId}", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB TV details query"},
		{PathPattern: "/tv/{tvId}/ratings", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB TV ratings query"},
		{PathPattern: "/tv/{tvId}/ratingscombined", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB TV combined ratings query"},
		{PathPattern: "/tv/{tvId}/recommendations", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB TV recommendations query"},
		{PathPattern: "/tv/{tvId}/similar", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB TV similar query"},
		{PathPattern: "/tv/{tvId}/season/{seasonId}", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB TV season details query"},
		{PathPattern: "/tv/{tvId}/season/{seasonId}/episode/{episodeId}", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB TV episode details query"},
		{PathPattern: "/person/{personId}", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB person details query"},
		{PathPattern: "/person/{personId}/combined_credits", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB person credits query"},

		// Additional Seerr API endpoints
		{PathPattern: "/media", Classification: ClassCovered, MappedResource: "seerr_media", Notes: "Media list data source"},
		{PathPattern: "/media/{mediaId}", Classification: ClassCovered, MappedResource: "seerr_media_item", Notes: "Media item data source"},
		{PathPattern: "/media/{mediaId}/file", Classification: ClassIntentionallyOutOfScope, Notes: "Media file info"},
		{PathPattern: "/media/{mediaId}/watch_data", Classification: ClassIntentionallyOutOfScope, Notes: "Media watch stats query"},
		{PathPattern: "/media/{mediaId}/{status}", Classification: ClassIntentionallyOutOfScope, Notes: "Media status update action (available/processing/pending)"},
		{PathPattern: "/media/{mediaId}/file/{fileId}", Classification: ClassIntentionallyOutOfScope, Notes: "Media file deletion action"},
		{PathPattern: "/collection/{collectionId}", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB collection details query"},
		{PathPattern: "/service/radarr", Classification: ClassIntentionallyOutOfScope, Notes: "Service health probe for Radarr"},
		{PathPattern: "/service/sonarr", Classification: ClassIntentionallyOutOfScope, Notes: "Service health probe for Sonarr"},
		{PathPattern: "/service/radarr/{radarrId}", Classification: ClassIntentionallyOutOfScope, Notes: "Service instance probe for Radarr"},
		{PathPattern: "/service/sonarr/{sonarrId}", Classification: ClassIntentionallyOutOfScope, Notes: "Service instance probe for Sonarr"},
		{PathPattern: "/service/sonarr/lookup/{tmdbId}", Classification: ClassIntentionallyOutOfScope, Notes: "Sonarr series lookup by TMDB ID"},
		{PathPattern: "/watchlist", Classification: ClassIntentionallyOutOfScope, Notes: "User watchlist items endpoint"},
		{PathPattern: "/watchlist/{tmdbId}", Classification: ClassIntentionallyOutOfScope, Notes: "Watchlist item query"},

		// Webpush & Push Subscription endpoints
		{PathPattern: "/webpush/vapidkey", Classification: ClassIntentionallyOutOfScope, Notes: "Webpush VAPID key endpoint"},
		{PathPattern: "/webpush/register", Classification: ClassIntentionallyOutOfScope, Notes: "Webpush browser registration endpoint"},
		{PathPattern: "/user/registerPushSubscription", Classification: ClassIntentionallyOutOfScope, Notes: "Browser push notification registration"},
		{PathPattern: "/user/{userId}/pushSubscription/{endpoint}", Classification: ClassIntentionallyOutOfScope, Notes: "Browser push notification item"},
		{PathPattern: "/user/{userId}/pushSubscriptions", Classification: ClassIntentionallyOutOfScope, Notes: "Browser push notifications list"},

		// Auth & Password Reset endpoints
		{PathPattern: "/auth/jellyfin", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive Jellyfin login"},
		{PathPattern: "/auth/jellyfin/quickconnect/authenticate", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive login"},
		{PathPattern: "/auth/jellyfin/quickconnect/check", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive login"},
		{PathPattern: "/auth/jellyfin/quickconnect/initiate", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive login"},
		{PathPattern: "/auth/reset-password", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive password reset"},
		{PathPattern: "/auth/reset-password/{guid}", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive password reset"},
		{PathPattern: "/user/jellyfin/{jellyfinUserId}", Classification: ClassIntentionallyOutOfScope, Notes: "Jellyfin user lookup"},
		{PathPattern: "/user/{userId}/settings/linked-accounts/jellyfin", Classification: ClassIntentionallyOutOfScope, Notes: "Jellyfin account linking"},
		{PathPattern: "/user/{userId}/settings/linked-accounts/jellyfin/quickconnect", Classification: ClassIntentionallyOutOfScope, Notes: "Jellyfin QuickConnect linking"},
		{PathPattern: "/user/{userId}/settings/linked-accounts/plex", Classification: ClassIntentionallyOutOfScope, Notes: "Plex account linking"},
		{PathPattern: "/user/{userId}/settings/notifications", Classification: ClassUncovered, Notes: "User notification settings (backlog candidate)"},
		{PathPattern: "/user/{userId}/watch_data", Classification: ClassIntentionallyOutOfScope, Notes: "Per-user watch data query"},

		// Additional Blocklist / Override Rule / Discover / Issue / Media endpoints
		{PathPattern: "/blacklist", Classification: ClassIntentionallyOutOfScope, Notes: "Legacy alias for blocklist"},
		{PathPattern: "/blacklist/{tmdbId}", Classification: ClassIntentionallyOutOfScope, Notes: "Legacy alias for blocklist"},
		{PathPattern: "/blocklist/collection/{collectionId}", Classification: ClassCovered, MappedResource: "seerr_blocklist", Notes: "Blocklist collection resource"},
		{PathPattern: "/blocklist/{tmdbId}", Classification: ClassCovered, MappedResource: "seerr_blocklist", Notes: "Blocklist item resource"},
		{PathPattern: "/overrideRule", Classification: ClassCovered, MappedResource: "seerr_override_rule", Notes: "Override rule resource"},
		{PathPattern: "/overrideRule/{ruleId}", Classification: ClassCovered, MappedResource: "seerr_override_rule", Notes: "Override rule item resource"},
		{PathPattern: "/settings/discover/add", Classification: ClassCovered, MappedResource: "seerr_discover_slider", Notes: "Discover slider add action"},
		{PathPattern: "/settings/discover/{sliderId}", Classification: ClassCovered, MappedResource: "seerr_discover_slider", Notes: "Discover slider item resource"},
		{PathPattern: "/issueComment/{commentId}", Classification: ClassIntentionallyOutOfScope, Notes: "Issue comment management"},

		// TMDB / Search / Discovery auxiliary endpoints
		{PathPattern: "/backdrops", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB/media UI backdrop lookup"},
		{PathPattern: "/certifications/movie", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB certifications lookup"},
		{PathPattern: "/certifications/tv", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB certifications lookup"},
		{PathPattern: "/discover/genreslider/movie", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive genre slider lookup"},
		{PathPattern: "/discover/genreslider/tv", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive genre slider lookup"},
		{PathPattern: "/discover/movies/genre/{genreId}", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive movie lookup"},
		{PathPattern: "/discover/movies/language/{language}", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive movie lookup"},
		{PathPattern: "/discover/movies/studio/{studioId}", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive movie lookup"},
		{PathPattern: "/discover/tv/genre/{genreId}", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive TV lookup"},
		{PathPattern: "/discover/tv/language/{language}", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive TV lookup"},
		{PathPattern: "/discover/tv/network/{networkId}", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive TV lookup"},
		{PathPattern: "/discover/watchlist", Classification: ClassIntentionallyOutOfScope, Notes: "Interactive watchlist discovery"},
		{PathPattern: "/genres/movie", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB genres lookup"},
		{PathPattern: "/genres/tv", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB genres lookup"},
		{PathPattern: "/keyword/{keywordId}", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB keyword details"},
		{PathPattern: "/languages", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB languages list"},
		{PathPattern: "/network/{networkId}", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB network details"},
		{PathPattern: "/regions", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB regions list"},
		{PathPattern: "/studio/{studioId}", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB studio details"},
		{PathPattern: "/tv/{tvId}/season/{seasonNumber}", Classification: ClassIntentionallyOutOfScope, Notes: "TMDB TV season details"},
		{PathPattern: "/watchproviders/movies", Classification: ClassIntentionallyOutOfScope, Notes: "Watch providers query"},
		{PathPattern: "/watchproviders/regions", Classification: ClassIntentionallyOutOfScope, Notes: "Watch provider regions query"},
		{PathPattern: "/watchproviders/tv", Classification: ClassIntentionallyOutOfScope, Notes: "Watch providers query"},
	}
}

// ParseOpenAPISpec loads and parses an OpenAPI 3.0 YAML file.
func ParseOpenAPISpec(filePath string) (*OpenAPISpec, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read openapi spec: %w", err)
	}

	var spec OpenAPISpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("unmarshal openapi spec: %w", err)
	}

	return &spec, nil
}

// AnalyzeCoverage checks all paths in spec against rules.
func AnalyzeCoverage(spec *OpenAPISpec, rules []PathRule) (*CoverageReport, error) {
	ruleMap := make(map[string]PathRule)
	for _, r := range rules {
		ruleMap[r.PathPattern] = r
	}

	report := &CoverageReport{}

	// Sort paths for deterministic output
	var sortedPaths []string
	for p := range spec.Paths {
		sortedPaths = append(sortedPaths, p)
	}
	sort.Strings(sortedPaths)

	report.TotalPaths = len(sortedPaths)

	for _, pathStr := range sortedPaths {
		methodsMap := spec.Paths[pathStr]
		var methods []string
		for m := range methodsMap {
			mUpper := strings.ToUpper(m)
			if mUpper == "GET" || mUpper == "POST" || mUpper == "PUT" || mUpper == "DELETE" || mUpper == "PATCH" {
				methods = append(methods, mUpper)
			}
		}
		sort.Strings(methods)
		report.TotalEndpoints += len(methods)

		rule, found := ruleMap[pathStr]
		if !found {
			report.UnclassifiedPaths = append(report.UnclassifiedPaths, pathStr)
			continue
		}

		res := PathCoverageResult{
			Path:           pathStr,
			Methods:        methods,
			Classification: rule.Classification,
			Notes:          rule.Notes,
			MappedResource: rule.MappedResource,
		}

		switch rule.Classification {
		case ClassCovered:
			report.CoveredPaths++
		case ClassIntentionallyOutOfScope:
			report.IntentionallyOutOfScopePaths++
		case ClassUncovered:
			report.UncoveredPaths++
		}

		report.Results = append(report.Results, res)
	}

	return report, nil
}

// GenerateMarkdownReport returns a formatted markdown document for openapi-coverage.md.
func GenerateMarkdownReport(report *CoverageReport) string {
	var sb strings.Builder

	sb.WriteString("# OpenAPI Coverage Matrix\n\n")
	sb.WriteString("Inventory of Seerr OpenAPI endpoints classified by Terraform provider support and scope.\n\n")

	sb.WriteString("## Coverage Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total OpenAPI Paths**: %d\n", report.TotalPaths))
	sb.WriteString(fmt.Sprintf("- **Total Endpoints (Methods)**: %d\n", report.TotalEndpoints))
	sb.WriteString(fmt.Sprintf("- **Covered Paths**: %d (%.1f%%)\n", report.CoveredPaths, float64(report.CoveredPaths)/float64(report.TotalPaths)*100))
	sb.WriteString(fmt.Sprintf("- **Intentionally Out of Scope**: %d (%.1f%%)\n", report.IntentionallyOutOfScopePaths, float64(report.IntentionallyOutOfScopePaths)/float64(report.TotalPaths)*100))
	sb.WriteString(fmt.Sprintf("- **Uncovered Settings/User Backlog**: %d (%.1f%%)\n\n", report.UncoveredPaths, float64(report.UncoveredPaths)/float64(report.TotalPaths)*100))

	sb.WriteString("## Classification Legend\n\n")
	sb.WriteString("- `covered`: Endpoint is fully mapped to a Terraform resource or data source.\n")
	sb.WriteString("- `uncovered`: Configuration or user endpoint suitable for future Terraform provider coverage (backlog candidate).\n")
	sb.WriteString("- `intentionally-out-of-scope`: Interactive search/browse, transient test action, session auth, or runtime log/cache flush endpoint out of scope for infrastructure-as-code management.\n\n")

	sb.WriteString("## Path Coverage Matrix\n\n")
	sb.WriteString("| Path | Methods | Classification | Mapped Resource / Notes |\n")
	sb.WriteString("|---|---|---|---|\n")

	for _, r := range report.Results {
		methodsStr := strings.Join(r.Methods, ", ")
		mappedStr := r.Notes
		if r.MappedResource != "" {
			mappedStr = fmt.Sprintf("`%s` - %s", r.MappedResource, r.Notes)
		}
		sb.WriteString(fmt.Sprintf("| `%s` | %s | `%s` | %s |\n", r.Path, methodsStr, r.Classification, mappedStr))
	}

	sb.WriteString("\n## Refreshing the Snapshot & Inventory\n\n")
	sb.WriteString("To update the OpenAPI snapshot from upstream Seerr and re-validate coverage:\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("# Download the latest seerr-api.yml snapshot\n")
	sb.WriteString("curl -fsSL https://raw.githubusercontent.com/seerr-team/seerr/develop/seerr-api.yml -o tools/openapi/seerr-api.yml\n\n")
	sb.WriteString("# Run the OpenAPI coverage audit and generate openapi-coverage.md\n")
	sb.WriteString("go test -v ./tools/openapi/...\n")
	sb.WriteString("```\n")

	return sb.String()
}

// FindRepoRoot returns the path to the workspace root directory.
func FindRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("go.mod not found in parent directories")
}
