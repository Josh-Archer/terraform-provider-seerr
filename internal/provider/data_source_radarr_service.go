package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &RadarrServiceDataSource{}

type RadarrServiceDataSource struct {
	client *APIClient
}

type RadarrServiceDataSourceModel struct {
	ID       types.String                `tfsdk:"id"`
	ServerID types.Int64                 `tfsdk:"server_id"`
	Servers  []ServiceServerSummaryModel `tfsdk:"servers"`
	Profiles []ServiceProfileModel       `tfsdk:"profiles"`
}

func NewRadarrServiceDataSource() datasource.DataSource {
	return &RadarrServiceDataSource{}
}

func (d *RadarrServiceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_radarr_service"
}

func (d *RadarrServiceDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = serviceDataSourceSchema("Radarr")
}

func (d *RadarrServiceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RadarrServiceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured API Client", "API client was not provided during configuration.")
		return
	}

	var data RadarrServiceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, servers, profiles := readServiceReachability(ctx, d.client, "radarr", data.ServerID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(id)
	data.Servers = servers
	data.Profiles = profiles

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
