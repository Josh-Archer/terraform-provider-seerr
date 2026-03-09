package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &TautulliSettingsDataSource{}

type TautulliSettingsDataSource struct {
	client *APIClient
}

type TautulliSettingsDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Hostname     types.String `tfsdk:"hostname"`
	Port         types.Int64  `tfsdk:"port"`
	UseSSL       types.Bool   `tfsdk:"use_ssl"`
	URLBase      types.String `tfsdk:"url_base"`
	APIKey       types.String `tfsdk:"api_key"`
	ExternalURL  types.String `tfsdk:"external_url"`
	ResponseJSON types.String `tfsdk:"response_json"`
	StatusCode   types.Int64  `tfsdk:"status_code"`
}

func NewTautulliSettingsDataSource() datasource.DataSource { return &TautulliSettingsDataSource{} }

func (d *TautulliSettingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tautulli_settings"
}

func (d *TautulliSettingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read Seerr Tautulli settings via /api/v1/settings/tautulli.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "The hostname of the Tautulli server.",
				Computed:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "The port of the Tautulli server.",
				Computed:            true,
			},
			"use_ssl": schema.BoolAttribute{
				MarkdownDescription: "Whether to use SSL for the connection.",
				Computed:            true,
			},
			"url_base": schema.StringAttribute{
				MarkdownDescription: "The base URL for the Tautulli server.",
				Computed:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The API key for the Tautulli server.",
				Computed:            true,
				Sensitive:           true,
			},
			"external_url": schema.StringAttribute{
				MarkdownDescription: "The external URL for the Tautulli server.",
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

func (d *TautulliSettingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TautulliSettingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TautulliSettingsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.client.Request(ctx, "GET", "/api/v1/settings/tautulli", "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	data.StatusCode = types.Int64Value(int64(res.StatusCode))
	data.ResponseJSON = types.StringValue(string(res.Body))

	// Refresh state from response
	var decoded map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err != nil {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("failed to decode response: %s", err))
		return
	}
	if v, ok := decoded["hostname"].(string); ok {
		data.Hostname = types.StringValue(v)
	}
	if v, ok := decoded["port"].(float64); ok {
		data.Port = types.Int64Value(int64(v))
	}
	if v, ok := decoded["useSsl"].(bool); ok {
		data.UseSSL = types.BoolValue(v)
	}
	if v, ok := decoded["urlBase"].(string); ok {
		data.URLBase = types.StringValue(v)
	}
	if v, ok := decoded["apiKey"].(string); ok && v != "" {
		data.APIKey = types.StringValue(v)
	}
	if v, ok := decoded["externalUrl"].(string); ok {
		data.ExternalURL = types.StringValue(v)
	}

	data.ID = types.StringValue("tautulli")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
