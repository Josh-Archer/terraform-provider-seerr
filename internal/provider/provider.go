package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

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

type SeerrProviderModel struct {
	URL                types.String `tfsdk:"url"`
	APIKey             types.String `tfsdk:"api_key"`
	PlexToken          types.String `tfsdk:"plex_token"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
	UserAgent          types.String `tfsdk:"user_agent"`
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
				MarkdownDescription: "Base URL for Seerr, for example `https://seerr.example.com`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						urlRegex(),
						"url must start with http:// or https://",
					),
				},
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Seerr API key used as the `X-Api-Key` header. Required if `plex_token` is not set.",
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
		},
	}
}

func (p *SeerrProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data SeerrProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	baseURL := strings.TrimSpace(data.URL.ValueString())
	apiKey := strings.TrimSpace(data.APIKey.ValueString())
	plexToken := strings.TrimSpace(data.PlexToken.ValueString())

	if baseURL == "" {
		resp.Diagnostics.AddError("Missing Base URL", "Provider requires a 'url' to be set.")
		return
	}
	if apiKey == "" && plexToken == "" {
		resp.Diagnostics.AddError("Missing Authentication", "Provider requires either an 'api_key' or a 'plex_token' to be set.")
		return
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		resp.Diagnostics.AddError("Invalid URL", fmt.Sprintf("Cannot parse provider url %q: %s", baseURL, err))
		return
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		resp.Diagnostics.AddError("Invalid URL Scheme", "Provider url must use http or https.")
		return
	}

	userAgent := "terraform-provider-seerr/" + p.version
	if !data.UserAgent.IsNull() && !data.UserAgent.IsUnknown() && strings.TrimSpace(data.UserAgent.ValueString()) != "" {
		userAgent = strings.TrimSpace(data.UserAgent.ValueString())
	}

	insecure := false
	if !data.InsecureSkipVerify.IsNull() && !data.InsecureSkipVerify.IsUnknown() {
		insecure = data.InsecureSkipVerify.ValueBool()
	}

	client := NewClient(parsed, apiKey, userAgent, insecure)

	// Authentication flow logic
	if apiKey == "" && plexToken != "" {
		// 1. Authenticate with Plex token
		authPayload := map[string]string{"authToken": plexToken}
		authBody, _ := json.Marshal(authPayload)

		authRes, err := client.Request(ctx, "POST", "/api/v1/auth/plex", string(authBody), nil)
		if err != nil {
			resp.Diagnostics.AddError("Plex Auth Failed", err.Error())
			return
		}
		if !StatusIsOK(authRes.StatusCode) {
			resp.Diagnostics.AddError("Plex Auth Failed", fmt.Sprintf("status %d: %s", authRes.StatusCode, string(authRes.Body)))
			return
		}

		// 2. Extract session cookie (connect.sid)
		cookies := authRes.Headers.Values("Set-Cookie")
		var sessionCookie string
		for _, c := range cookies {
			if strings.HasPrefix(c, "connect.sid=") {
				// We just need the actual key=value part before the first semicolon
				sessionCookie = strings.SplitN(c, ";", 2)[0]
				break
			}
		}

		if sessionCookie == "" {
			resp.Diagnostics.AddError("Plex Auth Failed", "Did not receive a session cookie from Seerr.")
			return
		}

		client.SetSessionCookie(sessionCookie)

		// 3. Fetch settings to get API key
		settingsRes, err := client.Request(ctx, "GET", "/api/v1/settings/main", "", nil)
		if err != nil {
			resp.Diagnostics.AddError("API Key Fetch Failed", err.Error())
			return
		}
		if !StatusIsOK(settingsRes.StatusCode) {
			resp.Diagnostics.AddError("API Key Fetch Failed", fmt.Sprintf("status %d: %s", settingsRes.StatusCode, string(settingsRes.Body)))
			return
		}

		var settings map[string]any
		if err := json.Unmarshal(settingsRes.Body, &settings); err != nil {
			resp.Diagnostics.AddError("API Key Fetch Failed", fmt.Sprintf("failed to parse settings: %s", err))
			return
		}

		fetchedKey, ok := settings["apiKey"].(string)
		if !ok || fetchedKey == "" {
			resp.Diagnostics.AddError("API Key Fetch Failed", "apiKey not found in settings response. Ensure the Plex user is an admin.")
			return
		}

		// 4. Update the client with the new API key, discarding the session cookie
		client.SetAPIKey(fetchedKey)
		client.SetSessionCookie("") // We only need the API key going forward
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
	return regexp.MustCompile(`^https?://`)
}
