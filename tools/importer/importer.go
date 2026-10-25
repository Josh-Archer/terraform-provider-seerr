// Copyright (c) Josh Archer
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	defaultUserPageSize = 100
	defaultMaxUserPages = 1000
	maxResponseBytes    = 16 << 20
)

const (
	manualNetworkNote   = "seerr_network_settings is populated and was left unmanaged so a partial resource cannot null live proxy, DNS, or timeout settings. Import it by hand to manage it."
	manualDiscoverNote  = "seerr_discover_slider is populated and was left unmanaged. Import the singleton by hand; built-in sliders are filtered on read."
	manualOverrideNote  = "seerr_override_rule entries were left unmanaged. Import each rule by id and adopt paired genre, language, and tag attributes together."
	manualMainAliasNote = "seerr_main_settings deprecated aliases (region, image_proxy, plex_login) were not emitted because they conflict with streaming_region, discover_region, cache_images, and media_server_login."
)

// ImporterConfig holds configuration for the live Seerr importer.
type ImporterConfig struct {
	BaseURL      string
	APIKey       string
	Timeout      time.Duration
	HTTPClient   *http.Client
	AllowPartial bool
	// UserPageSize is the /api/v1/user take size. Zero uses defaultUserPageSize.
	UserPageSize int
	// MaxUserPages caps user pagination. Zero uses defaultMaxUserPages.
	MaxUserPages int
}

// DiscoveredResource represents a live Seerr resource ready to be converted into HCL and import blocks.
type DiscoveredResource struct {
	ResourceType string
	ResourceName string
	ImportID     string
	Attributes   map[string]any
	Comments     []string
}

// Importer fetches configuration from a live Seerr instance and generates Terraform/OpenTofu code.
type Importer struct {
	cfg         ImporterConfig
	httpClient  *http.Client
	failures    []error
	unsupported []string
	manual      []string
}

// NewImporter creates a new instance of Importer.
func NewImporter(cfg ImporterConfig) *Importer {
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: cfg.Timeout}
		if client.Timeout == 0 {
			client.Timeout = 30 * time.Second
		}
	}
	return &Importer{cfg: cfg, httpClient: client}
}

// Failures returns authentication, transport, server, and pagination errors from the last discovery.
func (imp *Importer) Failures() []error {
	return append([]error(nil), imp.failures...)
}

// UnsupportedEndpoints returns endpoints that responded 404 during the last discovery.
func (imp *Importer) UnsupportedEndpoints() []string {
	return append([]string(nil), imp.unsupported...)
}

// ManualAdoption returns resources that were seen but intentionally not emitted.
func (imp *Importer) ManualAdoption() []string {
	return append([]string(nil), imp.manual...)
}

type unsupportedEndpointError struct {
	Path string
}

func (e *unsupportedEndpointError) Error() string {
	return "unsupported endpoint " + e.Path
}

type discoveryIncompleteError struct {
	Failures []error
}

func (e *discoveryIncompleteError) Error() string {
	var sb strings.Builder
	sb.WriteString("incomplete discovery:")
	for _, err := range e.Failures {
		sb.WriteString("\n- ")
		sb.WriteString(err.Error())
	}
	sb.WriteString("\nre-run with -allow-partial to write the resources discovered so far")
	return sb.String()
}

// sensitiveValue marks an attribute whose live value must not be written.
// bindSensitiveVars replaces it with a varRef after resource names are final.
type sensitiveValue struct {
	Attr string
}

// varRef is a Terraform variable reference. It never carries the secret value.
type varRef struct {
	Name string
}

// nestedObject is a Terraform nested attribute object, rendered as `name = { ... }`.
type nestedObject struct {
	Attrs map[string]any
}

// stringSet is rendered as toset([...]) so it matches set attributes such as
// notification_types. A bare list plans a type conversion on every apply.
type stringSet []string

// sanitizeName transforms arbitrary names/titles into valid HCL identifiers.
func sanitizeName(name string) string {
	if name == "" {
		return "main"
	}
	var sb strings.Builder
	lastWasUnderscore := false
	for _, ch := range strings.ToLower(name) {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			sb.WriteRune(ch)
			lastWasUnderscore = false
		} else if ch == '_' || ch == ' ' || ch == '-' || ch == '.' || ch == '(' || ch == ')' {
			if !lastWasUnderscore && sb.Len() > 0 {
				sb.WriteRune('_')
				lastWasUnderscore = true
			}
		}
	}
	res := strings.Trim(sb.String(), "_")
	if res == "" {
		return "main"
	}
	if res[0] >= '0' && res[0] <= '9' {
		res = "item_" + res
	}
	return res
}

func (imp *Importer) userPageSize() int {
	if imp.cfg.UserPageSize > 0 {
		return imp.cfg.UserPageSize
	}
	return defaultUserPageSize
}

func (imp *Importer) maxUserPages() int {
	if imp.cfg.MaxUserPages > 0 {
		return imp.cfg.MaxUserPages
	}
	return defaultMaxUserPages
}

func (imp *Importer) fetch(ctx context.Context, apiPath string) (any, error) {
	baseURL := strings.TrimRight(imp.cfg.BaseURL, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+apiPath, nil)
	if err != nil {
		return nil, fmt.Errorf("transport error from %s: %w", apiPath, err)
	}
	req.Header.Set("Accept", "application/json")
	if imp.cfg.APIKey != "" {
		req.Header.Set("X-Api-Key", imp.cfg.APIKey)
	}

	resp, err := imp.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("transport error from %s: %w", apiPath, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &unsupportedEndpointError{Path: apiPath}
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("transport error from %s: %w", apiPath, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		kind := "server"
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			kind = "authentication"
		} else if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			kind = "request"
		}
		// Response bodies are omitted: error payloads can echo credentials.
		return nil, fmt.Errorf("%s error from %s (HTTP %d)", kind, apiPath, resp.StatusCode)
	}

	if len(strings.TrimSpace(string(body))) == 0 {
		return map[string]any{}, nil
	}

	var result any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode error from %s: %w", apiPath, err)
	}
	return result, nil
}

// take records unsupported and fatal responses. ok is true only when a payload was read.
func (imp *Importer) take(ctx context.Context, apiPath string) (any, bool) {
	payload, err := imp.fetch(ctx, apiPath)
	if err == nil {
		return payload, true
	}
	var unsupported *unsupportedEndpointError
	if errors.As(err, &unsupported) {
		imp.unsupported = append(imp.unsupported, unsupported.Path)
		return nil, false
	}
	imp.failures = append(imp.failures, err)
	return nil, false
}

func (imp *Importer) noteIfPopulated(ctx context.Context, apiPath, note string) {
	payload, err := imp.fetch(ctx, apiPath)
	if err != nil {
		var unsupported *unsupportedEndpointError
		if errors.As(err, &unsupported) {
			imp.unsupported = append(imp.unsupported, unsupported.Path)
			return
		}
		imp.failures = append(imp.failures, err)
		return
	}
	if payloadPopulated(payload) {
		imp.manual = append(imp.manual, note)
	}
}

func payloadPopulated(v any) bool {
	switch t := v.(type) {
	case map[string]any:
		return len(t) > 0
	case []any:
		return len(t) > 0
	default:
		return false
	}
}

// DiscoverAll retrieves manageable resources from the live instance.
// Unsupported HTTP 404s are skipped. Authentication, transport, decode, and server
// errors fail discovery unless AllowPartial is set.
func (imp *Importer) DiscoverAll(ctx context.Context) ([]DiscoveredResource, error) {
	imp.failures = nil
	imp.unsupported = nil
	imp.manual = nil

	var discovered []DiscoveredResource
	discovered = append(discovered, imp.discoverMain(ctx)...)
	discovered = append(discovered, imp.discoverPlex(ctx)...)
	discovered = append(discovered, imp.discoverMediaServer(ctx, "/api/v1/settings/jellyfin", "seerr_jellyfin_settings", "jellyfin", "jellyfinForgotPasswordUrl", "jellyfin_forgot_password_url")...)
	discovered = append(discovered, imp.discoverMediaServer(ctx, "/api/v1/settings/emby", "seerr_emby_settings", "emby", "embyForgotPasswordUrl", "emby_forgot_password_url")...)
	discovered = append(discovered, imp.discoverTautulli(ctx)...)
	discovered = append(discovered, imp.discoverArr(ctx, "/api/v1/settings/radarr", "seerr_radarr_server", "radarr", false)...)
	discovered = append(discovered, imp.discoverArr(ctx, "/api/v1/settings/sonarr", "seerr_sonarr_server", "sonarr", true)...)
	discovered = append(discovered, imp.discoverNotifications(ctx)...)
	discovered = append(discovered, imp.discoverUsers(ctx)...)

	// These endpoints are probed so errors still fail a complete discovery, but their
	// payloads are not turned into HCL. A partial resource would null live settings
	// the provider mirrors across deprecated or paired attributes.
	imp.noteIfPopulated(ctx, "/api/v1/settings/network", manualNetworkNote)
	imp.noteIfPopulated(ctx, "/api/v1/settings/discover", manualDiscoverNote)
	imp.noteIfPopulated(ctx, "/api/v1/overrideRule", manualOverrideNote)

	if len(imp.failures) > 0 && !imp.cfg.AllowPartial {
		return nil, &discoveryIncompleteError{Failures: imp.failures}
	}

	assignUniqueResourceNames(discovered)
	bindSensitiveVars(discovered)
	sort.SliceStable(discovered, func(i, j int) bool {
		if discovered[i].ResourceType != discovered[j].ResourceType {
			return discovered[i].ResourceType < discovered[j].ResourceType
		}
		if c := compareImportID(discovered[i].ImportID, discovered[j].ImportID); c != 0 {
			return c < 0
		}
		return discovered[i].ResourceName < discovered[j].ResourceName
	})
	return discovered, nil
}

func (imp *Importer) discoverMain(ctx context.Context) []DiscoveredResource {
	payload, ok := imp.take(ctx, "/api/v1/settings/main")
	if !ok {
		return nil
	}
	m, ok := payload.(map[string]any)
	if !ok {
		imp.failures = append(imp.failures, fmt.Errorf("decode error from /api/v1/settings/main: expected a JSON object"))
		return nil
	}
	attrs := map[string]any{}
	if _, ok := m["applicationTitle"]; ok {
		putString(attrs, m, "applicationTitle", "app_title")
	} else {
		putString(attrs, m, "appTitle", "app_title")
	}
	putString(attrs, m, "applicationUrl", "application_url")
	putMediaServerType(attrs, m)
	putBool(attrs, m, "trustProxy", "trust_proxy")
	putBool(attrs, m, "csrfProtection", "csrf_protection")
	putString(attrs, m, "locale", "locale")
	putString(attrs, m, "originalLanguage", "original_language")
	putBool(attrs, m, "hideAvailable", "hide_available")
	putBool(attrs, m, "hideRequested", "hide_requested")
	if _, ok := m["partialRequestsEnabled"]; ok {
		putBool(attrs, m, "partialRequestsEnabled", "partial_requests")
	} else {
		putBool(attrs, m, "partialRequests", "partial_requests")
	}
	putBool(attrs, m, "localLogin", "local_login")
	putBool(attrs, m, "newPlexLogin", "new_plex_login")
	putBool(attrs, m, "movieRequestsEnabled", "movie_requests_enabled")
	putBool(attrs, m, "seriesRequestsEnabled", "series_requests_enabled")
	putBool(attrs, m, "enableReportAnIssue", "enable_report_an_issue")
	putInt(attrs, m, "movieRequestLimit", "movie_request_limit")
	putInt(attrs, m, "seriesRequestLimit", "series_request_limit")

	// cacheImages also fills deprecated image_proxy, and region fields are mirrored
	// onto conflicting attributes. Emitting either side makes the configuration invalid
	// or plans a null-out of the alias the provider still stores.
	if hasAnyKey(m, "cacheImages", "imageProxy", "streamingRegion", "discoverRegion", "region", "mediaServerLogin", "plexLogin") {
		imp.manual = append(imp.manual, manualMainAliasNote)
	}

	return []DiscoveredResource{{
		ResourceType: "seerr_main_settings",
		ResourceName: "main",
		ImportID:     "main",
		Attributes:   attrs,
	}}
}

func (imp *Importer) discoverPlex(ctx context.Context) []DiscoveredResource {
	payload, ok := imp.take(ctx, "/api/v1/settings/plex")
	if !ok {
		return nil
	}
	m, ok := payload.(map[string]any)
	if !ok {
		imp.failures = append(imp.failures, fmt.Errorf("decode error from /api/v1/settings/plex: expected a JSON object"))
		return nil
	}
	ip := getString(m, "ip")
	if ip == "" {
		ip = getString(m, "hostname")
	}
	port, hasPort := intIfPresent(m, "port")
	if ip == "" || !hasPort || port <= 0 {
		return nil
	}
	attrs := map[string]any{
		"ip":   ip,
		"port": port,
	}
	putBool(attrs, m, "useSsl", "use_ssl")
	return []DiscoveredResource{{
		ResourceType: "seerr_plex_settings",
		ResourceName: "plex",
		ImportID:     "plex",
		Attributes:   attrs,
	}}
}

func (imp *Importer) discoverMediaServer(ctx context.Context, apiPath, resourceType, importID, forgotAPI, forgotHCL string) []DiscoveredResource {
	payload, ok := imp.take(ctx, apiPath)
	if !ok {
		return nil
	}
	m, ok := payload.(map[string]any)
	if !ok {
		imp.failures = append(imp.failures, fmt.Errorf("decode error from %s: expected a JSON object", apiPath))
		return nil
	}
	ip := getString(m, "ip")
	if ip == "" {
		ip = getString(m, "hostname")
	}
	port, hasPort := intIfPresent(m, "port")
	if ip == "" || !hasPort || port <= 0 {
		return nil
	}
	attrs := map[string]any{
		"ip":   ip,
		"port": port,
	}
	putBool(attrs, m, "useSsl", "use_ssl")
	putString(attrs, m, "urlBase", "url_base")
	putString(attrs, m, "externalHostname", "external_hostname")
	putString(attrs, m, forgotAPI, forgotHCL)
	putSensitive(attrs, m, "apiKey", "api_key", true)
	return []DiscoveredResource{{
		ResourceType: resourceType,
		ResourceName: importID,
		ImportID:     importID,
		Attributes:   attrs,
	}}
}

func (imp *Importer) discoverTautulli(ctx context.Context) []DiscoveredResource {
	payload, ok := imp.take(ctx, "/api/v1/settings/tautulli")
	if !ok {
		return nil
	}
	m, ok := payload.(map[string]any)
	if !ok {
		imp.failures = append(imp.failures, fmt.Errorf("decode error from /api/v1/settings/tautulli: expected a JSON object"))
		return nil
	}
	attrs := map[string]any{}
	// The provider mirrors hostname and ip onto each other. Emit both so a plan
	// does not try to null the one left out of configuration.
	if host := getString(m, "hostname"); host != "" {
		attrs["hostname"] = host
		attrs["ip"] = host
	} else if ip := getString(m, "ip"); ip != "" {
		attrs["hostname"] = ip
		attrs["ip"] = ip
	}
	putInt(attrs, m, "port", "port")
	putBool(attrs, m, "useSsl", "use_ssl")
	putString(attrs, m, "urlBase", "url_base")
	putString(attrs, m, "externalUrl", "external_url")
	putSensitive(attrs, m, "apiKey", "api_key", true)
	if len(attrs) == 0 {
		return nil
	}
	return []DiscoveredResource{{
		ResourceType: "seerr_tautulli_settings",
		ResourceName: "tautulli",
		ImportID:     "tautulli",
		Attributes:   attrs,
	}}
}

func (imp *Importer) discoverArr(ctx context.Context, apiPath, resourceType, namePrefix string, sonarr bool) []DiscoveredResource {
	payload, ok := imp.take(ctx, apiPath)
	if !ok {
		return nil
	}
	list, ok := asObjectList(payload)
	if !ok {
		imp.failures = append(imp.failures, fmt.Errorf("decode error from %s: expected a JSON array", apiPath))
		return nil
	}
	var found []DiscoveredResource
	for _, server := range list {
		idStr, ok := objectID(server)
		if !ok {
			imp.failures = append(imp.failures, fmt.Errorf("decode error from %s: server id missing", apiPath))
			continue
		}
		display := getString(server, "name")
		base := display
		if base == "" {
			base = fmt.Sprintf("%s_%s", namePrefix, idStr)
		}
		attrs := map[string]any{}
		if _, ok := server["name"]; ok {
			attrs["name"] = display
		}
		putString(attrs, server, "hostname", "hostname")
		putInt(attrs, server, "port", "port")
		putBool(attrs, server, "useSsl", "use_ssl")
		putString(attrs, server, "baseUrl", "base_url")
		putInt(attrs, server, "activeProfileId", "quality_profile_id")
		putString(attrs, server, "activeProfileName", "quality_profile_name")
		putString(attrs, server, "activeDirectory", "active_directory")
		putBool(attrs, server, "is4k", "is_4k")
		putBool(attrs, server, "isDefault", "is_default")
		putBool(attrs, server, "enableScan", "enable_scan")
		putBool(attrs, server, "syncEnabled", "sync_enabled")
		putBool(attrs, server, "preventSearch", "prevent_search")
		putBool(attrs, server, "tagRequests", "tag_requests_with_user")
		putIntList(attrs, server, "tags", "tags")
		putSensitive(attrs, server, "apiKey", "api_key", false)
		if sonarr {
			putString(attrs, server, "activeAnimeDirectory", "active_anime_directory")
			putBool(attrs, server, "enableSeasonFolders", "enable_season_folders")
			putIntList(attrs, server, "animeTags", "anime_tags")
		} else {
			putString(attrs, server, "minimumAvailability", "minimum_availability")
		}
		found = append(found, DiscoveredResource{
			ResourceType: resourceType,
			ResourceName: sanitizeName(base),
			ImportID:     idStr,
			Attributes:   attrs,
		})
	}
	return found
}

type notificationField struct {
	API          string
	HCL          string
	Kind         string
	NonEmptyOnly bool
}

func notificationOptionFields(agent string) []notificationField {
	switch agent {
	case "discord":
		return []notificationField{
			{API: "botUsername", HCL: "bot_username", Kind: "string"},
			{API: "botAvatarUrl", HCL: "bot_avatar_url", Kind: "string"},
			{API: "webhookUrl", HCL: "webhook_url", Kind: "string"},
			{API: "enableMentions", HCL: "enable_mentions", Kind: "bool"},
		}
	case "slack":
		return []notificationField{{API: "webhookUrl", HCL: "webhook_url", Kind: "string"}}
	case "email":
		return []notificationField{
			{API: "emailFrom", HCL: "email_from", Kind: "string"},
			{API: "smtpHost", HCL: "smtp_host", Kind: "string"},
			{API: "smtpPort", HCL: "smtp_port", Kind: "int"},
			{API: "secure", HCL: "secure", Kind: "bool"},
			{API: "ignoreTls", HCL: "ignore_tls", Kind: "bool"},
			{API: "requireTls", HCL: "require_tls", Kind: "bool"},
			{API: "authUser", HCL: "auth_user", Kind: "string"},
			{API: "authPass", HCL: "auth_pass", Kind: "sensitive"},
			{API: "allowSelfSigned", HCL: "allow_self_signed", Kind: "bool"},
			{API: "senderName", HCL: "sender_name", Kind: "string"},
			{API: "pgpPrivateKey", HCL: "pgp_private_key", Kind: "sensitive"},
			{API: "pgpPassword", HCL: "pgp_password", Kind: "sensitive"},
		}
	case "lunasea":
		return []notificationField{
			{API: "webhookUrl", HCL: "webhook_url", Kind: "string"},
			{API: "profileName", HCL: "profile_name", Kind: "string"},
		}
	case "telegram":
		return []notificationField{
			{API: "botUsername", HCL: "bot_username", Kind: "string"},
			{API: "botAPI", HCL: "bot_api", Kind: "sensitive"},
			{API: "chatId", HCL: "chat_id", Kind: "string"},
			{API: "sendSilently", HCL: "send_silently", Kind: "bool"},
		}
	case "pushbullet":
		return []notificationField{
			{API: "accessToken", HCL: "access_token", Kind: "sensitive"},
			{API: "channelTag", HCL: "channel_tag", Kind: "string"},
		}
	case "pushover":
		return []notificationField{
			{API: "accessToken", HCL: "access_token", Kind: "sensitive"},
			{API: "userToken", HCL: "user_token", Kind: "sensitive"},
			{API: "sound", HCL: "sound", Kind: "string"},
		}
	case "ntfy":
		return []notificationField{
			{API: "url", HCL: "url", Kind: "string"},
			{API: "topic", HCL: "topic", Kind: "string"},
			{API: "authMethodUsernamePassword", HCL: "auth_method_username_password", Kind: "bool"},
			{API: "username", HCL: "username", Kind: "string"},
			{API: "password", HCL: "password", Kind: "sensitive"},
			{API: "authMethodToken", HCL: "auth_method_token", Kind: "bool"},
			{API: "token", HCL: "token", Kind: "sensitive"},
			{API: "priority", HCL: "priority", Kind: "int"},
			{API: "tags", HCL: "tags", Kind: "string"},
		}
	case "webhook":
		return []notificationField{
			{API: "webhookUrl", HCL: "webhook_url", Kind: "string"},
			{API: "jsonPayload", HCL: "json_payload", Kind: "string"},
			{API: "authHeader", HCL: "auth_header", Kind: "sensitive"},
		}
	case "gotify":
		return []notificationField{
			{API: "url", HCL: "url", Kind: "string"},
			{API: "token", HCL: "token", Kind: "sensitive"},
		}
	default:
		return nil
	}
}

func (imp *Importer) discoverNotifications(ctx context.Context) []DiscoveredResource {
	agents := []string{
		"discord", "email", "gotify", "lunasea", "ntfy", "pushbullet",
		"pushover", "slack", "telegram", "webhook", "webpush",
	}
	var found []DiscoveredResource
	for _, agent := range agents {
		path := "/api/v1/settings/notifications/" + agent
		payload, ok := imp.take(ctx, path)
		if !ok {
			continue
		}
		m, ok := payload.(map[string]any)
		if !ok {
			imp.failures = append(imp.failures, fmt.Errorf("decode error from %s: expected a JSON object", path))
			continue
		}
		enabled := getBool(m, "enabled")
		typesVal, _ := intIfPresent(m, "types")
		if !enabled && typesVal == 0 {
			continue
		}
		// The provider's JSON decoder treats omitted enabled/embedPoster as false,
		// so the imported state is false rather than null.
		embedPoster := false
		if _, ok := m["embedPoster"]; ok {
			embedPoster = getBool(m, "embedPoster")
		}
		attrs := map[string]any{
			"enabled":            enabled,
			"embed_poster":       embedPoster,
			"notification_types": stringSet(notificationEventNames(typesVal)),
		}
		options, _ := m["options"].(map[string]any)
		nested := map[string]any{}
		for _, field := range notificationOptionFields(agent) {
			switch field.Kind {
			case "string":
				putString(nested, options, field.API, field.HCL)
			case "bool":
				putBool(nested, options, field.API, field.HCL)
			case "int":
				putInt(nested, options, field.API, field.HCL)
			case "sensitive":
				putSensitive(nested, options, field.API, field.HCL, field.NonEmptyOnly)
			}
		}
		if len(nested) > 0 || agent == "webpush" {
			attrs[agent] = nestedObject{Attrs: nested}
		}
		found = append(found, DiscoveredResource{
			ResourceType: "seerr_notification_" + agent,
			ResourceName: agent,
			ImportID:     agent,
			Attributes:   attrs,
		})
	}
	return found
}

func notificationEventNames(mask int64) []string {
	bits := []struct {
		bit  int64
		name string
	}{
		{2, "MEDIA_PENDING"},
		{4, "MEDIA_APPROVED"},
		{8, "MEDIA_AVAILABLE"},
		{16, "MEDIA_FAILED"},
		{64, "MEDIA_DECLINED"},
		{128, "MEDIA_AUTO_APPROVED"},
		{256, "ISSUE_CREATED"},
		{512, "ISSUE_COMMENT"},
		{1024, "ISSUE_RESOLVED"},
		{2048, "ISSUE_REOPENED"},
		{4096, "MEDIA_AUTO_REQUESTED"},
	}
	names := make([]string, 0)
	for _, bit := range bits {
		if mask&bit.bit != 0 {
			names = append(names, bit.name)
		}
	}
	sort.Strings(names)
	return names
}

type pageMeta struct {
	total    int
	hasTotal bool
	page     int
	pages    int
	hasPage  bool
}

func (imp *Importer) discoverUsers(ctx context.Context) []DiscoveredResource {
	pageSize := imp.userPageSize()
	maxPages := imp.maxUserPages()
	seen := map[string]struct{}{}
	var collected []map[string]any
	skip := 0

	for page := 0; page < maxPages; page++ {
		query := url.Values{}
		query.Set("skip", strconv.Itoa(skip))
		query.Set("take", strconv.Itoa(pageSize))
		endpoint := "/api/v1/user?" + query.Encode()

		payload, err := imp.fetch(ctx, endpoint)
		if err != nil {
			var unsupported *unsupportedEndpointError
			if errors.As(err, &unsupported) {
				if page == 0 {
					imp.unsupported = append(imp.unsupported, unsupported.Path)
				}
				break
			}
			imp.failures = append(imp.failures, err)
			break
		}

		results, info, err := parseUserPage(payload)
		if err != nil {
			imp.failures = append(imp.failures, fmt.Errorf("decode error from %s: %w", endpoint, err))
			break
		}

		added := 0
		for _, user := range results {
			idStr, ok := objectID(user)
			if !ok {
				continue
			}
			if _, dup := seen[idStr]; dup {
				continue
			}
			seen[idStr] = struct{}{}
			collected = append(collected, user)
			added++
		}

		fullPage := len(results) >= pageSize
		reachedTotal := info.hasTotal && len(seen) >= info.total
		reachedLast := info.hasPage && info.pages > 0 && info.page >= info.pages
		if len(results) == 0 || added == 0 || !fullPage || reachedTotal || reachedLast {
			break
		}
		if page == maxPages-1 {
			imp.failures = append(imp.failures, fmt.Errorf("incomplete user pagination from /api/v1/user: stopped after %d pages of %d", maxPages, pageSize))
			break
		}
		skip += pageSize
	}

	sort.SliceStable(collected, func(i, j int) bool {
		left, _ := objectID(collected[i])
		right, _ := objectID(collected[j])
		return compareImportID(left, right) < 0
	})

	found := make([]DiscoveredResource, 0, len(collected))
	for _, user := range collected {
		idStr, ok := objectID(user)
		if !ok {
			continue
		}
		username := getString(user, "username")
		base := username
		if strings.TrimSpace(base) == "" {
			base = "user_" + idStr
		}
		attrs := map[string]any{}
		if _, ok := user["email"]; ok {
			attrs["email"] = getString(user, "email")
		}
		if _, ok := user["username"]; ok {
			attrs["username"] = username
		}
		putInt(attrs, user, "permissions", "permissions")

		settingsPath := "/api/v1/user/" + url.PathEscape(idStr) + "/settings/main"
		if settings, ok := imp.take(ctx, settingsPath); ok {
			if settingsMap, ok := settings.(map[string]any); ok {
				putString(attrs, settingsMap, "locale", "locale")
				putString(attrs, settingsMap, "discoverRegion", "discover_region")
				putString(attrs, settingsMap, "streamingRegion", "streaming_region")
				putString(attrs, settingsMap, "originalLanguage", "original_language")
				putInt(attrs, settingsMap, "movieQuotaLimit", "movie_quota_limit")
				putInt(attrs, settingsMap, "movieQuotaDays", "movie_quota_days")
				putInt(attrs, settingsMap, "tvQuotaLimit", "tv_quota_limit")
				putInt(attrs, settingsMap, "tvQuotaDays", "tv_quota_days")
				putBool(attrs, settingsMap, "watchlistSyncMovies", "watchlist_sync_movies")
				putBool(attrs, settingsMap, "watchlistSyncTv", "watchlist_sync_tv")
			} else {
				imp.failures = append(imp.failures, fmt.Errorf("decode error from %s: expected a JSON object", settingsPath))
			}
		}

		found = append(found, DiscoveredResource{
			ResourceType: "seerr_user",
			ResourceName: sanitizeName(base),
			ImportID:     idStr,
			Attributes:   attrs,
		})
	}
	return found
}

func parseUserPage(payload any) ([]map[string]any, pageMeta, error) {
	switch body := payload.(type) {
	case []any:
		users, err := objectsFromList(body)
		return users, pageMeta{}, err
	case map[string]any:
		raw, ok := body["results"]
		if !ok || raw == nil {
			return nil, parsePageInfo(body["pageInfo"]), nil
		}
		list, ok := raw.([]any)
		if !ok {
			return nil, pageMeta{}, fmt.Errorf("results is not an array")
		}
		users, err := objectsFromList(list)
		return users, parsePageInfo(body["pageInfo"]), err
	default:
		return nil, pageMeta{}, fmt.Errorf("unexpected user payload")
	}
}

func parsePageInfo(raw any) pageMeta {
	infoMap, ok := raw.(map[string]any)
	if !ok {
		return pageMeta{}
	}
	var info pageMeta
	if total, ok := intIfPresent(infoMap, "total"); ok {
		info.total = int(total)
		info.hasTotal = true
	}
	if page, ok := intIfPresent(infoMap, "page"); ok {
		info.page = int(page)
		info.hasPage = true
	}
	if pages, ok := intIfPresent(infoMap, "pages"); ok {
		info.pages = int(pages)
	}
	return info
}

func objectsFromList(list []any) ([]map[string]any, error) {
	out := make([]map[string]any, 0, len(list))
	for _, item := range list {
		object, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected a JSON object in the list")
		}
		out = append(out, object)
	}
	return out, nil
}

func asObjectList(payload any) ([]map[string]any, bool) {
	list, ok := payload.([]any)
	if !ok {
		return nil, false
	}
	objects, err := objectsFromList(list)
	if err != nil {
		return nil, false
	}
	return objects, true
}

func assignUniqueResourceNames(resources []DiscoveredResource) {
	order := make([]int, len(resources))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool {
		left, right := resources[order[a]], resources[order[b]]
		if left.ResourceType != right.ResourceType {
			return left.ResourceType < right.ResourceType
		}
		if c := compareImportID(left.ImportID, right.ImportID); c != 0 {
			return c < 0
		}
		if left.ResourceName != right.ResourceName {
			return left.ResourceName < right.ResourceName
		}
		return order[a] < order[b]
	})

	used := map[string]struct{}{}
	for _, idx := range order {
		base := resources[idx].ResourceName
		if base == "" {
			base = "main"
		}
		candidate := base
		for n := 2; ; n++ {
			key := resources[idx].ResourceType + "\x00" + candidate
			if _, exists := used[key]; !exists {
				used[key] = struct{}{}
				resources[idx].ResourceName = candidate
				break
			}
			candidate = fmt.Sprintf("%s_%d", base, n)
		}
	}
}

func compareImportID(left, right string) int {
	leftNum, leftErr := strconv.ParseInt(left, 10, 64)
	rightNum, rightErr := strconv.ParseInt(right, 10, 64)
	if leftErr == nil && rightErr == nil {
		switch {
		case leftNum < rightNum:
			return -1
		case leftNum > rightNum:
			return 1
		default:
			return 0
		}
	}
	return strings.Compare(left, right)
}

func bindSensitiveVars(resources []DiscoveredResource) {
	for i := range resources {
		resources[i].Attributes = bindAttrMap(resources[i].ResourceType, resources[i].ResourceName, resources[i].Attributes)
	}
}

func bindAttrMap(resourceType, resourceName string, attrs map[string]any) map[string]any {
	for key, value := range attrs {
		switch typed := value.(type) {
		case sensitiveValue:
			attr := typed.Attr
			if attr == "" {
				attr = key
			}
			attrs[key] = varRef{Name: sensitiveVarName(resourceType, resourceName, attr)}
		case nestedObject:
			typed.Attrs = bindAttrMap(resourceType, resourceName, typed.Attrs)
			attrs[key] = typed
		}
	}
	return attrs
}

func sensitiveVarName(resourceType, resourceName, attr string) string {
	return sanitizeName(resourceType + "_" + resourceName + "_" + attr)
}

type sensitiveDecl struct {
	Name        string
	Description string
}

func collectSensitiveVars(resources []DiscoveredResource) []sensitiveDecl {
	var decls []sensitiveDecl
	seen := map[string]struct{}{}
	var walk func(res DiscoveredResource, prefix string, attrs map[string]any)
	walk = func(res DiscoveredResource, prefix string, attrs map[string]any) {
		keys := make([]string, 0, len(attrs))
		for key := range attrs {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			path := key
			if prefix != "" {
				path = prefix + "." + key
			}
			switch typed := attrs[key].(type) {
			case varRef:
				if _, ok := seen[typed.Name]; ok {
					continue
				}
				seen[typed.Name] = struct{}{}
				decls = append(decls, sensitiveDecl{
					Name: typed.Name,
					Description: fmt.Sprintf(
						"Sensitive value for %s.%s %s. Set TF_VAR_%s or an uncommitted tfvars file. State stores this value in plaintext; use an encrypted backend.",
						res.ResourceType, res.ResourceName, path, typed.Name,
					),
				})
			case nestedObject:
				walk(res, path, typed.Attrs)
			}
		}
	}
	for _, res := range resources {
		walk(res, "", res.Attributes)
	}
	sort.Slice(decls, func(i, j int) bool { return decls[i].Name < decls[j].Name })
	return decls
}

// GenerateVariables writes sensitive variable declarations with no defaults and no secret values.
func GenerateVariables(resources []DiscoveredResource) string {
	decls := collectSensitiveVars(resources)
	if len(decls) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(`# Sensitive credentials discovered on the live instance.
# Values are intentionally not written here. Supply them at plan time.
# Terraform and OpenTofu still store sensitive resource attributes in state.
# Use an encrypted remote backend, or OpenTofu state encryption.

`)
	for _, decl := range decls {
		sb.WriteString("variable ")
		sb.WriteString(strconv.Quote(decl.Name))
		sb.WriteString(" {\n")
		sb.WriteString("  type        = string\n")
		sb.WriteString("  sensitive   = true\n")
		sb.WriteString("  description = ")
		sb.WriteString(strconv.Quote(decl.Description))
		sb.WriteString("\n}\n\n")
	}
	return sb.String()
}

// GenerateHCL produces Terraform/OpenTofu HCL for discovered resources.
func GenerateHCL(resources []DiscoveredResource, providerHeader bool) string {
	var sb strings.Builder
	if providerHeader {
		sb.WriteString(`# --------------------------------------------------------------------------
# Terraform & OpenTofu Provider Configuration for Seerr
# Generated by seerr-import CLI
# --------------------------------------------------------------------------

terraform {
  required_version = ">= 1.5.0"
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "~> 2.0"
    }
  }
}

provider "seerr" {
  # Set via SEERR_URL and SEERR_API_KEY environment variables or specify below:
  # url     = "http://localhost:5055"
  # api_key = "YOUR_API_KEY"
}

`)
	}

	for _, res := range resources {
		for _, comment := range res.Comments {
			sb.WriteString("# ")
			sb.WriteString(comment)
			sb.WriteByte('\n')
		}
		sb.WriteString(fmt.Sprintf("resource \"%s\" \"%s\" {\n", res.ResourceType, res.ResourceName))
		writeAttrs(&sb, "  ", res.Attributes)
		sb.WriteString("}\n\n")
	}
	return sb.String()
}

func quoteStrings(values []string) string {
	quoted := make([]string, len(values))
	for i, item := range values {
		quoted[i] = strconv.Quote(item)
	}
	return strings.Join(quoted, ", ")
}

func writeAttrs(sb *strings.Builder, indent string, attrs map[string]any) {
	keys := make([]string, 0, len(attrs))
	for key := range attrs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		writeAttribute(sb, indent, key, attrs[key])
	}
}

func writeAttribute(sb *strings.Builder, indent, key string, value any) {
	switch typed := value.(type) {
	case nestedObject:
		if len(typed.Attrs) == 0 {
			fmt.Fprintf(sb, "%s%s = {}\n", indent, key)
			return
		}
		fmt.Fprintf(sb, "%s%s = {\n", indent, key)
		writeAttrs(sb, indent+"  ", typed.Attrs)
		fmt.Fprintf(sb, "%s}\n", indent)
	case varRef:
		fmt.Fprintf(sb, "%s%-25s = var.%s\n", indent, key, typed.Name)
	case string:
		fmt.Fprintf(sb, "%s%-25s = %s\n", indent, key, strconv.Quote(typed))
	case bool:
		fmt.Fprintf(sb, "%s%-25s = %t\n", indent, key, typed)
	case int64:
		fmt.Fprintf(sb, "%s%-25s = %d\n", indent, key, typed)
	case int:
		fmt.Fprintf(sb, "%s%-25s = %d\n", indent, key, typed)
	case []string:
		fmt.Fprintf(sb, "%s%-25s = [%s]\n", indent, key, quoteStrings(typed))
	case stringSet:
		fmt.Fprintf(sb, "%s%-25s = toset([%s])\n", indent, key, quoteStrings([]string(typed)))
	case []int64:
		parts := make([]string, len(typed))
		for i, item := range typed {
			parts[i] = strconv.FormatInt(item, 10)
		}
		fmt.Fprintf(sb, "%s%-25s = [%s]\n", indent, key, strings.Join(parts, ", "))
	case sensitiveValue:
		// A secret placeholder that was not bound must not be rendered as a value.
	}
}

// GenerateImportBlocks produces modern Terraform 1.5+ import blocks for discovered resources.
func GenerateImportBlocks(resources []DiscoveredResource) string {
	var sb strings.Builder
	sb.WriteString(`# --------------------------------------------------------------------------
# Terraform 1.5+ / OpenTofu Import Blocks
# Run: tofu plan -generate-config-out=generated.tf
#   or tofu plan to verify clean adoption
# --------------------------------------------------------------------------

`)
	for _, res := range resources {
		sb.WriteString("import {\n")
		sb.WriteString(fmt.Sprintf("  to = %s.%s\n", res.ResourceType, res.ResourceName))
		sb.WriteString(fmt.Sprintf("  id = \"%s\"\n", res.ImportID))
		sb.WriteString("}\n\n")
	}
	return sb.String()
}

// GenerateImportScript produces legacy shell import commands.
func GenerateImportScript(resources []DiscoveredResource) string {
	var sb strings.Builder
	sb.WriteString("#!/usr/bin/env bash\n")
	sb.WriteString("# Legacy Terraform Import Script\n")
	sb.WriteString("# Prefer SEERR_URL and SEERR_API_KEY over putting the API key on the command line.\n")
	sb.WriteString("set -euo pipefail\n\n")

	for _, res := range resources {
		sb.WriteString(fmt.Sprintf("terraform import %s.%s %s\n", res.ResourceType, res.ResourceName, res.ImportID))
	}
	return sb.String()
}

func canonicalFormat(format string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "all", "hcl", "imports", "script":
		return strings.ToLower(strings.TrimSpace(format)), nil
	default:
		return "", fmt.Errorf("invalid output format %q: expected all, hcl, imports, or script", format)
	}
}

func putString(dst, src map[string]any, srcKey, dstKey string) {
	raw, ok := src[srcKey]
	if !ok || raw == nil {
		return
	}
	value, ok := raw.(string)
	if !ok {
		return
	}
	dst[dstKey] = value
}

func putBool(dst, src map[string]any, srcKey, dstKey string) {
	raw, ok := src[srcKey]
	if !ok || raw == nil {
		return
	}
	value, ok := raw.(bool)
	if !ok {
		return
	}
	dst[dstKey] = value
}

func putInt(dst, src map[string]any, srcKey, dstKey string) {
	value, ok := intIfPresent(src, srcKey)
	if !ok {
		return
	}
	dst[dstKey] = value
}

func putIntList(dst, src map[string]any, srcKey, dstKey string) {
	raw, ok := src[srcKey]
	if !ok || raw == nil {
		return
	}
	list, ok := raw.([]any)
	if !ok {
		return
	}
	values := make([]int64, 0, len(list))
	for _, item := range list {
		number, ok := intFromAny(item)
		if ok {
			values = append(values, number)
		}
	}
	dst[dstKey] = values
}

func putSensitive(dst, src map[string]any, srcKey, dstKey string, nonEmptyOnly bool) {
	raw, ok := src[srcKey]
	if !ok || raw == nil {
		return
	}
	value, ok := raw.(string)
	if !ok {
		return
	}
	if nonEmptyOnly && value == "" {
		return
	}
	dst[dstKey] = sensitiveValue{Attr: dstKey}
}

func putMediaServerType(dst, src map[string]any) {
	raw, ok := src["mediaServerType"]
	if !ok || raw == nil {
		return
	}
	if number, ok := intFromAny(raw); ok {
		switch number {
		case 1:
			dst["media_server_type"] = "plex"
		case 2:
			dst["media_server_type"] = "jellyfin"
		case 3:
			dst["media_server_type"] = "emby"
		}
		return
	}
	if text, ok := raw.(string); ok {
		switch strings.ToLower(strings.TrimSpace(text)) {
		case "plex", "jellyfin", "emby":
			dst["media_server_type"] = strings.ToLower(strings.TrimSpace(text))
		}
	}
}

func hasAnyKey(src map[string]any, keys ...string) bool {
	for _, key := range keys {
		if _, ok := src[key]; ok {
			return true
		}
	}
	return false
}

func getString(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if value, ok := m[key].(string); ok {
		return value
	}
	return ""
}

func getBool(m map[string]any, key string) bool {
	if m == nil {
		return false
	}
	if value, ok := m[key].(bool); ok {
		return value
	}
	return false
}

func intIfPresent(m map[string]any, key string) (int64, bool) {
	if m == nil {
		return 0, false
	}
	raw, ok := m[key]
	if !ok || raw == nil {
		return 0, false
	}
	return intFromAny(raw)
}

func intFromAny(value any) (int64, bool) {
	switch typed := value.(type) {
	case float64:
		return int64(typed), true
	case int64:
		return typed, true
	case int:
		return int64(typed), true
	case json.Number:
		number, err := typed.Int64()
		return number, err == nil
	default:
		return 0, false
	}
}

func objectID(m map[string]any) (string, bool) {
	raw, ok := m["id"]
	if !ok || raw == nil {
		return "", false
	}
	if number, ok := intFromAny(raw); ok {
		return strconv.FormatInt(number, 10), true
	}
	text, ok := raw.(string)
	if !ok || text == "" {
		return "", false
	}
	return text, true
}
