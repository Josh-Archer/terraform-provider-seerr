package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &SonarrLanguageProfilesDataSource{}

type SonarrLanguageProfilesDataSource struct{}

type SonarrLanguageProfileModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type SonarrLanguageProfilesDataSourceModel struct {
	URL              types.String                 `tfsdk:"url"`
	Hostname         types.String                 `tfsdk:"hostname"`
	Port             types.Int64                  `tfsdk:"port"`
	APIKey           types.String                 `tfsdk:"api_key"`
	UseSSL           types.Bool                   `tfsdk:"use_ssl"`
	BaseURL          types.String                 `tfsdk:"base_url"`
	LanguageProfiles []SonarrLanguageProfileModel `tfsdk:"language_profiles"`
}

func NewSonarrLanguageProfilesDataSource() datasource.DataSource {
	return &SonarrLanguageProfilesDataSource{}
}

func (d *SonarrLanguageProfilesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sonarr_language_profiles"
}

func (d *SonarrLanguageProfilesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch Sonarr language profiles via /api/v3/languageprofile.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "Full Sonarr URL. If set, it takes precedence over hostname, port, use_ssl, and base_url.",
				Optional:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Sonarr hostname when url is not provided.",
				Optional:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Sonarr port when url is not provided.",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Sonarr API key.",
				Required:            true,
				Sensitive:           true,
			},
			"use_ssl": schema.BoolAttribute{
				MarkdownDescription: "Use HTTPS when url is not provided.",
				Optional:            true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Base path when Sonarr is served under a subpath.",
				Optional:            true,
			},
			"language_profiles": schema.ListNestedAttribute{
				MarkdownDescription: "List of language profiles from Sonarr.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "Language profile ID.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Language profile name.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *SonarrLanguageProfilesDataSource) Configure(_ context.Context, _ datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
}

func (d *SonarrLanguageProfilesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SonarrLanguageProfilesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hostname := data.Hostname.ValueString()
	if data.Hostname.IsNull() || data.Hostname.IsUnknown() || strings.TrimSpace(hostname) == "" {
		hostname = "sonarr-service"
	}
	port := data.Port.ValueInt64()
	if data.Port.IsNull() || data.Port.IsUnknown() || port == 0 {
		port = 8989
	}
	useSSL := false
	if !data.UseSSL.IsNull() && !data.UseSSL.IsUnknown() {
		useSSL = data.UseSSL.ValueBool()
	}
	baseURL := ""
	if !data.BaseURL.IsNull() && !data.BaseURL.IsUnknown() {
		baseURL = data.BaseURL.ValueString()
	}

	results, err := fetchArrEndpoint(
		ctx,
		data.URL.ValueString(),
		hostname,
		port,
		useSSL,
		baseURL,
		data.APIKey.ValueString(),
		"/api/v3/languageprofile",
	)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	var languageProfiles []SonarrLanguageProfileModel
	for _, item := range results {
		profile := SonarrLanguageProfileModel{}
		if id, ok := int64ValueFromAny(item["id"]); ok {
			profile.ID = types.Int64Value(id)
		}
		if name, ok := stringValueFromAny(item["name"]); ok {
			profile.Name = types.StringValue(name)
		}
		languageProfiles = append(languageProfiles, profile)
	}

	data.LanguageProfiles = languageProfiles
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
