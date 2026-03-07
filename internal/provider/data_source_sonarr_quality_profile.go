package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &SonarrQualityProfileDataSource{}

type SonarrQualityProfileDataSource struct{}

type SonarrQualityProfileDataSourceModel struct {
	Name             types.String `tfsdk:"name"`
	URL              types.String `tfsdk:"url"`
	Hostname         types.String `tfsdk:"hostname"`
	Port             types.Int64  `tfsdk:"port"`
	APIKey           types.String `tfsdk:"api_key"`
	UseSSL           types.Bool   `tfsdk:"use_ssl"`
	BaseURL          types.String `tfsdk:"base_url"`
	QualityProfileID types.Int64  `tfsdk:"quality_profile_id"`
	StatusCode       types.Int64  `tfsdk:"status_code"`
	ResponseJSON     types.String `tfsdk:"response_json"`
}

func NewSonarrQualityProfileDataSource() datasource.DataSource {
	return &SonarrQualityProfileDataSource{}
}

func (d *SonarrQualityProfileDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sonarr_quality_profile"
}

func (d *SonarrQualityProfileDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up a Sonarr quality profile by name via /api/v3/qualityprofile.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The Sonarr quality profile name to match exactly.",
				Required:            true,
			},
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
				MarkdownDescription: "Sonarr API key used to read quality profiles.",
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
			"quality_profile_id": schema.Int64Attribute{
				MarkdownDescription: "The numeric quality profile ID matching name.",
				Computed:            true,
			},
			"status_code": schema.Int64Attribute{
				MarkdownDescription: "HTTP status code for the quality profile lookup.",
				Computed:            true,
			},
			"response_json": schema.StringAttribute{
				MarkdownDescription: "Raw JSON for the matched quality profile.",
				Computed:            true,
			},
		},
	}
}

func (d *SonarrQualityProfileDataSource) Configure(_ context.Context, _ datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
}

func (d *SonarrQualityProfileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SonarrQualityProfileDataSourceModel
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

	profiles, _, err := fetchArrProfiles(
		ctx,
		data.URL.ValueString(),
		hostname,
		port,
		useSSL,
		baseURL,
		data.APIKey.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	needle := strings.TrimSpace(data.Name.ValueString())
	for _, profile := range profiles {
		name, ok := profile["name"].(string)
		if !ok || strings.TrimSpace(name) != needle {
			continue
		}
		rawID, ok := profile["id"]
		if !ok {
			resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("quality profile %q has no id", needle))
			return
		}

		var id int64
		switch v := rawID.(type) {
		case float64:
			id = int64(v)
		case int64:
			id = v
		default:
			resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("quality profile %q returned unsupported id type", needle))
			return
		}

		body, err := json.Marshal(profile)
		if err != nil {
			resp.Diagnostics.AddError("Read Failed", err.Error())
			return
		}

		data.QualityProfileID = types.Int64Value(id)
		data.StatusCode = types.Int64Value(200)
		data.ResponseJSON = types.StringValue(string(body))
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("Sonarr quality profile %q not found", needle))
}
