package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &RadarrServerDataSource{}

type RadarrServerDataSource struct {
	client *APIClient
}

type RadarrServerDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	ServerID            types.Int64  `tfsdk:"server_id"`
	Name                types.String `tfsdk:"name"`
	Hostname            types.String `tfsdk:"hostname"`
	Port                types.Int64  `tfsdk:"port"`
	APIKey              types.String `tfsdk:"api_key"`
	UseSSL              types.Bool   `tfsdk:"use_ssl"`
	BaseURL             types.String `tfsdk:"base_url"`
	QualityProfileID    types.Int64  `tfsdk:"quality_profile_id"`
	QualityProfileName  types.String `tfsdk:"quality_profile_name"`
	ActiveDirectory     types.String `tfsdk:"active_directory"`
	Is4K                types.Bool   `tfsdk:"is_4k"`
	MinimumAvailability types.String `tfsdk:"minimum_availability"`
	Tags                types.List   `tfsdk:"tags"`
	IsDefault           types.Bool   `tfsdk:"is_default"`
	EnableScan          types.Bool   `tfsdk:"enable_scan"`
	SyncEnabled         types.Bool   `tfsdk:"sync_enabled"`
	PreventSearch       types.Bool   `tfsdk:"prevent_search"`
	TagRequestsWithUser types.Bool   `tfsdk:"tag_requests_with_user"`
}

func NewRadarrServerDataSource() datasource.DataSource { return &RadarrServerDataSource{} }

func (d *RadarrServerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_radarr_server"
}

func (d *RadarrServerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read a Seerr Radarr server configuration via /api/v1/settings/radarr.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the matched Radarr server.",
				Computed:            true,
			},
			"server_id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the Radarr server to look up.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The Radarr server name reported by Seerr.",
				Computed:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "The Radarr hostname reported by Seerr.",
				Computed:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "The Radarr port reported by Seerr.",
				Computed:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The Radarr API key reported by Seerr.",
				Computed:            true,
				Sensitive:           true,
			},
			"use_ssl": schema.BoolAttribute{
				MarkdownDescription: "Whether Radarr uses HTTPS.",
				Computed:            true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "The Radarr base URL reported by Seerr.",
				Computed:            true,
			},
			"quality_profile_id": schema.Int64Attribute{
				MarkdownDescription: "The active Radarr quality profile ID.",
				Computed:            true,
			},
			"quality_profile_name": schema.StringAttribute{
				MarkdownDescription: "The active Radarr quality profile name.",
				Computed:            true,
			},
			"active_directory": schema.StringAttribute{
				MarkdownDescription: "The active Radarr download directory.",
				Computed:            true,
			},
			"is_4k": schema.BoolAttribute{
				MarkdownDescription: "Whether the Radarr server is configured for 4K.",
				Computed:            true,
			},
			"minimum_availability": schema.StringAttribute{
				MarkdownDescription: "The Radarr minimum availability setting.",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "The Radarr tag IDs attached to the server.",
				Computed:            true,
				ElementType:         types.Int64Type,
			},
			"is_default": schema.BoolAttribute{
				MarkdownDescription: "Whether this is the default Radarr server.",
				Computed:            true,
			},
			"enable_scan": schema.BoolAttribute{
				MarkdownDescription: "Whether scan is enabled for the Radarr server.",
				Computed:            true,
			},
			"sync_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether sync is enabled for the Radarr server.",
				Computed:            true,
			},
			"prevent_search": schema.BoolAttribute{
				MarkdownDescription: "Whether search is prevented for the Radarr server.",
				Computed:            true,
			},
			"tag_requests_with_user": schema.BoolAttribute{
				MarkdownDescription: "Whether requests are tagged with the user.",
				Computed:            true,
			},
		},
	}
}

func (d *RadarrServerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RadarrServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RadarrServerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.client.Request(ctx, "GET", "/api/v1/settings/radarr", "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	item, found, err := findByIDInJSONArray(res.Body, data.ServerID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !found {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("Radarr server with id %d not found", data.ServerID.ValueInt64()))
		return
	}

	var state RadarrServerModel
	if err := readRadarrStateFromJSON(ctx, item, &state); err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%d", data.ServerID.ValueInt64()))
	data.Name = state.Name
	data.Hostname = state.Hostname
	data.Port = state.Port
	data.APIKey = state.APIKey
	data.UseSSL = state.UseSSL
	data.BaseURL = state.BaseURL
	data.QualityProfileID = state.QualityProfileID
	data.QualityProfileName = state.QualityProfileName
	data.ActiveDirectory = state.ActiveDirectory
	data.Is4K = state.Is4K
	data.MinimumAvailability = state.MinimumAvailability
	data.Tags = state.Tags
	data.IsDefault = state.IsDefault
	data.EnableScan = state.EnableScan
	data.SyncEnabled = state.SyncEnabled
	data.PreventSearch = state.PreventSearch
	data.TagRequestsWithUser = state.TagRequestsWithUser

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
