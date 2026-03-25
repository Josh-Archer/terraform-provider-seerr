package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ServiceStatusDataSource{}

type ServiceStatusDataSource struct {
	client *APIClient
}

type ServiceStatusDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Version         types.String `tfsdk:"version"`
	CommitTag       types.String `tfsdk:"commit_tag"`
	UpdateAvailable types.Bool   `tfsdk:"update_available"`
	CommitsBehind   types.Int64  `tfsdk:"commits_behind"`
	RestartRequired types.Bool   `tfsdk:"restart_required"`
}

func NewServiceStatusDataSource() datasource.DataSource {
	return &ServiceStatusDataSource{}
}

func (d *ServiceStatusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_status"
}

func (d *ServiceStatusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get Seerr service status and version information.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "The current version of Seerr.",
				Computed:            true,
			},
			"commit_tag": schema.StringAttribute{
				MarkdownDescription: "The current commit tag of Seerr.",
				Computed:            true,
			},
			"update_available": schema.BoolAttribute{
				MarkdownDescription: "Whether an update is available.",
				Computed:            true,
			},
			"commits_behind": schema.Int64Attribute{
				MarkdownDescription: "Number of commits behind the latest version.",
				Computed:            true,
			},
			"restart_required": schema.BoolAttribute{
				MarkdownDescription: "Whether a restart is required.",
				Computed:            true,
			},
		},
	}
}

func (d *ServiceStatusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServiceStatusDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ServiceStatusDataSourceModel

	res, err := d.client.Request(ctx, "GET", "/api/v1/status", "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !HandleAPIResponse(ctx, res, &resp.Diagnostics, "Read") {
		return
	}

	var m map[string]any
	if err := json.Unmarshal(res.Body, &m); err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	if v, ok := m["version"].(string); ok {
		data.Version = types.StringValue(v)
	}
	if v, ok := m["commitTag"].(string); ok {
		data.CommitTag = types.StringValue(v)
	}
	if v, ok := m["updateAvailable"].(bool); ok {
		data.UpdateAvailable = types.BoolValue(v)
	}
	if v, ok := m["commitsBehind"].(float64); ok {
		data.CommitsBehind = types.Int64Value(int64(v))
	}
	if v, ok := m["restartRequired"].(bool); ok {
		data.RestartRequired = types.BoolValue(v)
	}

	data.ID = types.StringValue("seerr_status")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
