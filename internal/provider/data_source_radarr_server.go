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
	ServerID     types.Int64  `tfsdk:"server_id"`
	ResponseJSON types.String `tfsdk:"response_json"`
	StatusCode   types.Int64  `tfsdk:"status_code"`
}

func NewRadarrServerDataSource() datasource.DataSource { return &RadarrServerDataSource{} }

func (d *RadarrServerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_radarr_server"
}

func (d *RadarrServerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read a Seerr Radarr server configuration via /api/v1/settings/radarr.",
		Attributes: map[string]schema.Attribute{
			"server_id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the Radarr server to look up.",
				Required:            true,
			},
			"response_json": schema.StringAttribute{
				MarkdownDescription: "Raw JSON response body for the matched server.",
				Computed:            true,
			},
			"status_code": schema.Int64Attribute{
				MarkdownDescription: "HTTP status code.",
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

	data.StatusCode = types.Int64Value(int64(res.StatusCode))
	data.ResponseJSON = types.StringValue(string(item))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
