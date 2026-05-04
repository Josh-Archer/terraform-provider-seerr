package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	stringvalidator "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &SeerrProvider{}

type SeerrProvider struct {
	version string
}

type providerConfigValues struct {
	BaseURL        string
	APIKey         string
	PlexToken      string
	UserAgent      string
	Insecure       bool
	RequestTimeout time.Duration
}

type SeerrProviderModel struct {
	URL                types.String `tfsdk:"url"`
	APIKey             types.String `tfsdk:"api_key"`
	PlexToken          types.String `tfsdk:"plex_token"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
	UserAgent          types.String `tfsdk:"user_agent"`
	RequestTimeout     types.Int64  `tfsdk:"request_timeout_seconds"`
}

func (p *SeerrProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "seerr"
	resp.Version = p.version
}

func (p *SeerrProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provider for Seerr APIs.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "Base URL for Seerr, for example `https://seerr.example.com`. Can also be configured via the `SEERR_URL` environment variable.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						urlRegex(),
						"url must start with http:// or https:// and must not have a trailing slash",
					),
				},
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Seerr API key used as the `X-Api-Key` header. Required if `plex_token` is not set. Can also be configured via the `SEERR_API_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"plex_token": schema.StringAttribute{
				MarkdownDescription: "Plex token used as the `X-Plex-Token` header for authentication. This token must belong to a server admin user in order to be used for the setup flow. Required if `api_key` is not set.",
				Optional:            true,
				Sensitive:           true,
			},
			"insecure_skip_verify": schema.BoolAttribute{
				MarkdownDescription: "Skip TLS certificate verification.",
				Optional:            true,
			},
			"user_agent": schema.StringAttribute{
				MarkdownDescription: "Optional custom User-Agent header.",
				Optional:            true,
			},
			"request_timeout_seconds": schema.Int64Attribute{
				MarkdownDescription: "HTTP request timeout in seconds for Seerr API calls and ARR quality-profile lookups. Defaults to 120 seconds. Can also be configured via the `SEERR_REQUEST_TIMEOUT_SECONDS` environment variable.",
				Optional:            true,
			},
		},
	}
}

func (p *SeerrProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data SeerrProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := resolveProviderConfigValues(data, p.version, os.Getenv)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Request Timeout", err.Error())
		return
	}

	if config.BaseURL == "" {
		resp.Diagnostics.AddError("Missing Base URL", "Provider requires a 'url' to be set.")
		return
	}
	if config.APIKey == "" && config.PlexToken == "" {
		resp.Diagnostics.AddError("Missing Authentication", "Provider requires either an 'api_key' or a 'plex_token' to be set.")
		return
	}

	if !urlRegex().MatchString(config.BaseURL) {
		resp.Diagnostics.AddError("Invalid Base URL", "Provider url must start with http:// or https:// and must not have a trailing slash.")
		return
	}

	parsed, err := url.Parse(config.BaseURL)
	if err != nil {
		resp.Diagnostics.AddError("Invalid URL", fmt.Sprintf("Cannot parse provider url %q: %s", config.BaseURL, err))
		return
	}

	client := NewClient(parsed, config.APIKey, config.UserAgent, config.Insecure, config.RequestTimeout)

	// Authentication flow logic
	if config.APIKey == "" && config.PlexToken != "" {
		fetchedKey, err := bootstrapAPIKeyFromPlexToken(ctx, client, config.PlexToken)
		if err != nil {
			resp.Diagnostics.AddError("Plex Auth Failed", err.Error())
			return
		}
		client.SetAPIKey(fetchedKey)
		client.SetSessionCookie("")
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *SeerrProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAPIObjectResource,
		NewJobScheduleResource,
		NewMainSettingsResource,
		NewPlexSettingsResource,
		NewJellyfinSettingsResource,
		NewEmbySettingsResource,
		NewNotificationDiscordResource,
		NewNotificationSlackResource,
		NewNotificationEmailResource,
		NewNotificationLunaSeaResource,
		NewNotificationTelegramResource,
		NewNotificationPushbulletResource,
		NewNotificationPushoverResource,
		NewNotificationNtfyResource,
		NewNotificationWebhookResource,
		NewNotificationGotifyResource,
		NewNotificationWebpushResource,
		NewRadarrServerResource,
		NewSonarrServerResource,
		NewUserPermissionsResource,
		NewUserWatchlistSettingsResource,
		NewAPIKeyResource,
		NewUserResource,
		NewDiscoverSliderResource,
		NewTautulliSettingsResource,
		NewUserSettingsPermissionsResource,
		NewRequestResource,
		NewRequestRetryResource,
		NewIssueResource,
		NewBackupSettingsResource,
		NewUserInvitationResource,
		NewNetworkSettingsResource,
		NewMetadataSettingsResource,
		NewOverrideRuleResource,
		NewBlocklistResource,
	}
}

func (p *SeerrProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAPIRequestDataSource,
		NewPublicSettingsDataSource,
		NewMainSettingsDataSource,
		NewPlexSettingsDataSource,
		NewJellyfinSettingsDataSource,
		NewEmbySettingsDataSource,
		NewNotificationDiscordDataSource,
		NewNotificationSlackDataSource,
		NewNotificationEmailDataSource,
		NewNotificationLunaSeaDataSource,
		NewNotificationTelegramDataSource,
		NewNotificationPushbulletDataSource,
		NewNotificationPushoverDataSource,
		NewNotificationNtfyDataSource,
		NewNotificationWebhookDataSource,
		NewNotificationGotifyDataSource,
		NewNotificationWebpushDataSource,
		NewRadarrQualityProfileDataSource,
		NewRadarrServerDataSource,
		NewSonarrQualityProfileDataSource,
		NewSonarrServerDataSource,
		NewUserPermissionsDataSource,
		NewUserWatchlistSettingsDataSource,
		NewAPIKeyDataSource,
		NewUserDataSource,
		NewUsersDataSource,
		NewJobsDataSource,
		NewNotificationAgentsDataSource,
		NewTautulliSettingsDataSource,
		NewIssuesDataSource,
		NewRequestsDataSource,
		NewServiceStatusDataSource,
		NewCurrentUserDataSource,
		NewDiscoverSliderDataSource,
		NewBackupSettingsDataSource,
		NewUserInvitationsDataSource,
		NewNetworkSettingsDataSource,
		NewMetadataSettingsDataSource,
		NewOverrideRuleDataSource,
		NewBlocklistDataSource,
		NewMediaDataSource,
		NewDiscoverDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SeerrProvider{
			version: version,
		}
	}
}

func urlRegex() *regexp.Regexp {
	return regexp.MustCompile(`^https?://[^/](.*[^/])?$`)
}

func resolveProviderConfigValues(data SeerrProviderModel, version string, getenv func(string) string) (providerConfigValues, error) {
	config := providerConfigValues{
		BaseURL:        strings.TrimSpace(data.URL.ValueString()),
		APIKey:         strings.TrimSpace(data.APIKey.ValueString()),
		PlexToken:      strings.TrimSpace(data.PlexToken.ValueString()),
		UserAgent:      "terraform-provider-seerr/" + version,
		RequestTimeout: defaultRequestTimeout,
	}

	if config.BaseURL == "" {
		config.BaseURL = strings.TrimSpace(getenv("SEERR_URL"))
	}
	if config.APIKey == "" {
		config.APIKey = strings.TrimSpace(getenv("SEERR_API_KEY"))
	}
	if !data.UserAgent.IsNull() && !data.UserAgent.IsUnknown() && strings.TrimSpace(data.UserAgent.ValueString()) != "" {
		config.UserAgent = strings.TrimSpace(data.UserAgent.ValueString())
	}
	if !data.InsecureSkipVerify.IsNull() && !data.InsecureSkipVerify.IsUnknown() {
		config.Insecure = data.InsecureSkipVerify.ValueBool()
	}
	if !data.RequestTimeout.IsNull() && !data.RequestTimeout.IsUnknown() {
		config.RequestTimeout = normalizeRequestTimeout(time.Duration(data.RequestTimeout.ValueInt64()) * time.Second)
		return config, nil
	}

	if rawTimeout := strings.TrimSpace(getenv("SEERR_REQUEST_TIMEOUT_SECONDS")); rawTimeout != "" {
		timeoutSeconds, err := strconv.ParseInt(rawTimeout, 10, 64)
		if err != nil {
			return providerConfigValues{}, fmt.Errorf("cannot parse SEERR_REQUEST_TIMEOUT_SECONDS %q: %s", rawTimeout, err)
		}
		config.RequestTimeout = normalizeRequestTimeout(time.Duration(timeoutSeconds) * time.Second)
	}

	return config, nil
}

func bootstrapAPIKeyFromPlexToken(ctx context.Context, client *APIClient, plexToken string) (string, error) {
	authPayload := map[string]string{"authToken": plexToken}
	authBody, _ := json.Marshal(authPayload)

	authRes, err := client.Request(ctx, "POST", "/api/v1/auth/plex", string(authBody), nil)
	if err != nil {
		return "", err
	}
	if !StatusIsOK(authRes.StatusCode) {
		return "", fmt.Errorf("status %d: %s", authRes.StatusCode, string(authRes.Body))
	}

	cookies := authRes.Headers.Values("Set-Cookie")
	var sessionCookie string
	for _, c := range cookies {
		if strings.HasPrefix(c, "connect.sid=") {
			sessionCookie = strings.SplitN(c, ";", 2)[0]
			break
		}
	}
	if sessionCookie == "" {
		return "", fmt.Errorf("did not receive a session cookie from Seerr")
	}

	client.SetSessionCookie(sessionCookie)
	settingsRes, err := client.Request(ctx, "GET", "/api/v1/settings/main", "", nil)
	if err != nil {
		return "", err
	}
	if !StatusIsOK(settingsRes.StatusCode) {
		return "", fmt.Errorf("status %d: %s", settingsRes.StatusCode, string(settingsRes.Body))
	}

	var settings map[string]any
	if err := json.Unmarshal(settingsRes.Body, &settings); err != nil {
		return "", fmt.Errorf("failed to parse settings: %s", err)
	}

	fetchedKey, ok := settings["apiKey"].(string)
	if !ok || fetchedKey == "" {
		return "", fmt.Errorf("apiKey not found in settings response; ensure the Plex user is an admin")
	}

	return fetchedKey, nil
}
