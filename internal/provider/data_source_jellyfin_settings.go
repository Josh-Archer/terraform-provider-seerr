package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &JellyfinSettingsDataSource{}

type JellyfinSettingsDataSource struct {
	client *APIClient
}

type JellyfinSettingsDataSourceModel struct {
	ID                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	IP                        types.String `tfsdk:"ip"`
	Port                      types.Int64  `tfsdk:"port"`
	UseSSL                    types.Bool   `tfsdk:"use_ssl"`
	URLBase                   types.String `tfsdk:"url_base"`
	ExternalHostname          types.String `tfsdk:"external_hostname"`
	JellyfinForgotPasswordURL types.String `tfsdk:"jellyfin_forgot_password_url"`
	ServerID                  types.String `tfsdk:"server_id"`
	ResponseJSON              types.String `tfsdk:"response_json"`
	StatusCode                types.Int64  `tfsdk:"status_code"`
}

func NewJellyfinSettingsDataSource() datasource.DataSource { return &JellyfinSettingsDataSource{} }

func (d *JellyfinSettingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jellyfin_settings"
}

func (d *JellyfinSettingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read Seerr Jellyfin settings via /api/v1/settings/jellyfin.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Jellyfin server.",
				Computed:            true,
			},
			"ip": schema.StringAttribute{
				MarkdownDescription: "The IP address or hostname of the Jellyfin server.",
				Computed:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "The port of the Jellyfin server.",
				Computed:            true,
			},
			"use_ssl": schema.BoolAttribute{
				MarkdownDescription: "Whether to use SSL for the connection.",
				Computed:            true,
			},
			"url_base": schema.StringAttribute{
				MarkdownDescription: "The base URL for the Jellyfin server.",
				Computed:            true,
			},
			"external_hostname": schema.StringAttribute{
				MarkdownDescription: "The external hostname for the Jellyfin server.",
				Computed:            true,
			},
			"jellyfin_forgot_password_url": schema.StringAttribute{
				MarkdownDescription: "The URL for forgotten passwords on the Jellyfin server.",
				Computed:            true,
			},
			"server_id": schema.StringAttribute{
				MarkdownDescription: "The unique server ID of the connected Jellyfin server.",
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

func (d *JellyfinSettingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *JellyfinSettingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data JellyfinSettingsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.client.Request(ctx, "GET", "/api/v1/settings/jellyfin", "", nil)
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
	if v, ok := decoded["name"].(string); ok {
		data.Name = types.StringValue(v)
	}
	if v, ok := decoded["ip"].(string); ok {
		data.IP = types.StringValue(v)
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
	if v, ok := decoded["externalHostname"].(string); ok {
		data.ExternalHostname = types.StringValue(v)
	}
	if v, ok := decoded["jellyfinForgotPasswordUrl"].(string); ok {
		data.JellyfinForgotPasswordURL = types.StringValue(v)
	}
	if v, ok := decoded["serverId"].(string); ok {
		data.ServerID = types.StringValue(v)
	}

	data.ID = types.StringValue("jellyfin")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
