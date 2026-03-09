package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &MainSettingsDataSource{}

type MainSettingsDataSource struct {
	client *APIClient
}

type MainSettingsDataSourceModel struct {
	AppTitle              types.String `tfsdk:"app_title"`
	ApplicationURL        types.String `tfsdk:"application_url"`
	TrustProxy            types.Bool   `tfsdk:"trust_proxy"`
	CSRFProtection        types.Bool   `tfsdk:"csrf_protection"`
	ImageProxy            types.Bool   `tfsdk:"image_proxy"`
	Locale                types.String `tfsdk:"locale"`
	Region                types.String `tfsdk:"region"`
	OriginalLanguage      types.String `tfsdk:"original_language"`
	HideAvailable         types.Bool   `tfsdk:"hide_available"`
	PartialRequests       types.Bool   `tfsdk:"partial_requests"`
	LocalLogin            types.Bool   `tfsdk:"local_login"`
	NewPlexLogin          types.Bool   `tfsdk:"new_plex_login"`
	PlexLogin             types.Bool   `tfsdk:"plex_login"`
	MovieRequestsEnabled  types.Bool   `tfsdk:"movie_requests_enabled"`
	SeriesRequestsEnabled types.Bool   `tfsdk:"series_requests_enabled"`
	EnableReportAnIssue   types.Bool   `tfsdk:"enable_report_an_issue"`
	MovieRequestLimit     types.Int64  `tfsdk:"movie_request_limit"`
	SeriesRequestLimit    types.Int64  `tfsdk:"series_request_limit"`
	ResponseJSON          types.String `tfsdk:"response_json"`
	StatusCode            types.Int64  `tfsdk:"status_code"`
}

func NewMainSettingsDataSource() datasource.DataSource { return &MainSettingsDataSource{} }

func (d *MainSettingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_main_settings"
}

func (d *MainSettingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read Seerr main settings via /api/v1/settings/main.",
		Attributes: map[string]schema.Attribute{
			"app_title": schema.StringAttribute{
				MarkdownDescription: "The application title.",
				Computed:            true,
			},
			"application_url": schema.StringAttribute{
				MarkdownDescription: "The application URL.",
				Computed:            true,
			},
			"trust_proxy": schema.BoolAttribute{
				MarkdownDescription: "Whether to trust the proxy.",
				Computed:            true,
			},
			"csrf_protection": schema.BoolAttribute{
				MarkdownDescription: "Whether CSRF protection is enabled.",
				Computed:            true,
			},
			"image_proxy": schema.BoolAttribute{
				MarkdownDescription: "Whether the image proxy is enabled.",
				Computed:            true,
			},
			"locale": schema.StringAttribute{
				MarkdownDescription: "The application locale.",
				Computed:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "The application region.",
				Computed:            true,
			},
			"original_language": schema.StringAttribute{
				MarkdownDescription: "The original language.",
				Computed:            true,
			},
			"hide_available": schema.BoolAttribute{
				MarkdownDescription: "Whether to hide available media.",
				Computed:            true,
			},
			"partial_requests": schema.BoolAttribute{
				MarkdownDescription: "Whether partial requests are allowed.",
				Computed:            true,
			},
			"local_login": schema.BoolAttribute{
				MarkdownDescription: "Whether local login is enabled.",
				Computed:            true,
			},
			"new_plex_login": schema.BoolAttribute{
				MarkdownDescription: "Whether the new Plex login is enabled.",
				Computed:            true,
			},
			"plex_login": schema.BoolAttribute{
				MarkdownDescription: "Whether Plex login is enabled.",
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
			"enable_report_an_issue": schema.BoolAttribute{
				MarkdownDescription: "Whether the 'Report an Issue' feature is enabled.",
				Computed:            true,
			},
			"movie_request_limit": schema.Int64Attribute{
				MarkdownDescription: "The movie request limit.",
				Computed:            true,
			},
			"series_request_limit": schema.Int64Attribute{
				MarkdownDescription: "The series request limit.",
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

func (d *MainSettingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MainSettingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MainSettingsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.client.Request(ctx, "GET", "/api/v1/settings/main", "", nil)
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

	if v, ok := decoded["appTitle"].(string); ok {
		data.AppTitle = types.StringValue(v)
	}
	if v, ok := decoded["applicationUrl"].(string); ok {
		data.ApplicationURL = types.StringValue(v)
	}
	if v, ok := decoded["trustProxy"].(bool); ok {
		data.TrustProxy = types.BoolValue(v)
	}
	if v, ok := decoded["csrfProtection"].(bool); ok {
		data.CSRFProtection = types.BoolValue(v)
	}
	if v, ok := decoded["imageProxy"].(bool); ok {
		data.ImageProxy = types.BoolValue(v)
	}
	if v, ok := decoded["locale"].(string); ok {
		data.Locale = types.StringValue(v)
	}
	if v, ok := decoded["region"].(string); ok {
		data.Region = types.StringValue(v)
	}
	if v, ok := decoded["originalLanguage"].(string); ok {
		data.OriginalLanguage = types.StringValue(v)
	}
	if v, ok := decoded["hideAvailable"].(bool); ok {
		data.HideAvailable = types.BoolValue(v)
	}
	if v, ok := decoded["partialRequests"].(bool); ok {
		data.PartialRequests = types.BoolValue(v)
	}
	if v, ok := decoded["localLogin"].(bool); ok {
		data.LocalLogin = types.BoolValue(v)
	}
	if v, ok := decoded["newPlexLogin"].(bool); ok {
		data.NewPlexLogin = types.BoolValue(v)
	}
	if v, ok := decoded["plexLogin"].(bool); ok {
		data.PlexLogin = types.BoolValue(v)
	}
	if v, ok := decoded["movieRequestsEnabled"].(bool); ok {
		data.MovieRequestsEnabled = types.BoolValue(v)
	}
	if v, ok := decoded["seriesRequestsEnabled"].(bool); ok {
		data.SeriesRequestsEnabled = types.BoolValue(v)
	}
	if v, ok := decoded["enableReportAnIssue"].(bool); ok {
		data.EnableReportAnIssue = types.BoolValue(v)
	}
	if v, ok := decoded["movieRequestLimit"].(float64); ok {
		data.MovieRequestLimit = types.Int64Value(int64(v))
	}
	if v, ok := decoded["seriesRequestLimit"].(float64); ok {
		data.SeriesRequestLimit = types.Int64Value(int64(v))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
