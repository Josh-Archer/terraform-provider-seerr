package provider

import (
	"context"
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
				MarkdownDescription: "Seerr API key used as the `X-Api-Key` header.",
				Required:            true,
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
	if baseURL == "" || apiKey == "" {
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
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *SeerrProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAPIObjectResource,
		NewMainSettingsResource,
		NewPlexSettingsResource,
		NewNotificationAgentResource,
		NewRadarrServerResource,
		NewSonarrServerResource,
		NewUserPermissionsResource,
		NewUserWatchlistSettingsResource,
	}
}

func (p *SeerrProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAPIRequestDataSource,
		NewMainSettingsDataSource,
		NewPlexSettingsDataSource,
		NewNotificationAgentDataSource,
		NewRadarrServerDataSource,
		NewSonarrServerDataSource,
		NewUserPermissionsDataSource,
		NewUserWatchlistSettingsDataSource,
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
