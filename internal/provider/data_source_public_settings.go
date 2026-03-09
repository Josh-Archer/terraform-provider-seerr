package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &PublicSettingsDataSource{}

type PublicSettingsDataSource struct {
	client *APIClient
}

type PublicSettingsDataSourceModel struct {
	Initialized           types.Bool   `tfsdk:"initialized"`
	AppTitle              types.String `tfsdk:"app_title"`
	ApplicationURL        types.String `tfsdk:"application_url"`
	EnableLocalLogin      types.Bool   `tfsdk:"enable_local_login"`
	EnablePlexLogin       types.Bool   `tfsdk:"enable_plex_login"`
	NewPlexLogin          types.Bool   `tfsdk:"new_plex_login"`
	MovieRequestsEnabled  types.Bool   `tfsdk:"movie_requests_enabled"`
	SeriesRequestsEnabled types.Bool   `tfsdk:"series_requests_enabled"`
	ResponseJSON          types.String `tfsdk:"response_json"`
	StatusCode            types.Int64  `tfsdk:"status_code"`
}

func NewPublicSettingsDataSource() datasource.DataSource { return &PublicSettingsDataSource{} }

func (d *PublicSettingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_public_settings"
}

func (d *PublicSettingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read Seerr public settings via /api/v1/settings/public.",
		Attributes: map[string]schema.Attribute{
			"initialized": schema.BoolAttribute{
				MarkdownDescription: "Whether the Seerr instance is initialized.",
				Computed:            true,
			},
			"app_title": schema.StringAttribute{
				MarkdownDescription: "The application title.",
				Computed:            true,
			},
			"application_url": schema.StringAttribute{
				MarkdownDescription: "The application URL.",
				Computed:            true,
			},
			"enable_local_login": schema.BoolAttribute{
				MarkdownDescription: "Whether local login is enabled.",
				Computed:            true,
			},
			"enable_plex_login": schema.BoolAttribute{
				MarkdownDescription: "Whether Plex login is enabled.",
				Computed:            true,
			},
			"new_plex_login": schema.BoolAttribute{
				MarkdownDescription: "Whether the new Plex login is enabled.",
				Computed:            true,
			},
			"movie_requests_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether movie requests are enabled.",
				Computed:            true,
			},
			"series_requests_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether series requests are enabled.",
				Computed:            true,
			},
			"response_json": schema.StringAttribute{
				MarkdownDescription: "Raw JSON response body.",
				Computed:            true,
			},
			"status_code": schema.Int64Attribute{
				MarkdownDescription: "HTTP status code.",
				Computed:            true,
			},
		},
	}
}

func (d *PublicSettingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Configure Type", fmt.Sprintf("Expected *APIClient, got %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *PublicSettingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PublicSettingsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.client.Request(ctx, "GET", "/api/v1/settings/public", "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	var decoded map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err != nil {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("failed to decode response: %s", err))
		return
	}

	data.StatusCode = types.Int64Value(int64(res.StatusCode))
	data.ResponseJSON = types.StringValue(string(res.Body))

	if v, ok := decoded["initialized"].(bool); ok {
		data.Initialized = types.BoolValue(v)
	}
	if v, ok := decoded["applicationTitle"].(string); ok {
		data.AppTitle = types.StringValue(v)
	}
	if v, ok := decoded["applicationUrl"].(string); ok {
		data.ApplicationURL = types.StringValue(v)
	}
	if v, ok := decoded["enableLocalLogin"].(bool); ok {
		data.EnableLocalLogin = types.BoolValue(v)
	}
	if v, ok := decoded["enablePlexLogin"].(bool); ok {
		data.EnablePlexLogin = types.BoolValue(v)
	}
	if v, ok := decoded["newPlexLogin"].(bool); ok {
		data.NewPlexLogin = types.BoolValue(v)
	}
	if v, ok := decoded["movieRequestsEnabled"].(bool); ok {
		data.MovieRequestsEnabled = types.BoolValue(v)
	}
	if v, ok := decoded["seriesRequestsEnabled"].(bool); ok {
		data.SeriesRequestsEnabled = types.BoolValue(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
