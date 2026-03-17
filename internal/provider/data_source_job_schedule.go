package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &JobScheduleDataSource{}

type JobScheduleDataSource struct {
	client *APIClient
}

type JobScheduleDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	JobID    types.String `tfsdk:"job_id"`
	Schedule types.String `tfsdk:"schedule"`
}

func NewJobScheduleDataSource() datasource.DataSource { return &JobScheduleDataSource{} }

func (d *JobScheduleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job_schedule"
}

func (d *JobScheduleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read a background job schedule from Seerr via `/api/v1/settings/jobs`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"job_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the background job (e.g., `plex-sync`, `radarr-sync`).",
				Required:            true,
			},
			"schedule": schema.StringAttribute{
				MarkdownDescription: "The cron expression for the job schedule.",
				Computed:            true,
			},
		},
	}
}

func (d *JobScheduleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *JobScheduleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data JobScheduleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID := data.JobID.ValueString()

	res, err := d.client.Request(ctx, "GET", "/api/v1/settings/jobs", "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	var jobsList []map[string]any
	if err := json.Unmarshal(res.Body, &jobsList); err != nil {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("failed to decode response as list of jobs: %s\nBody: %s", err, string(res.Body)))
		return
	}

	foundJob := false
	for _, job := range jobsList {
		if jID, ok := job["id"].(string); ok && jID == jobID {
			foundJob = true
			if sched, ok := job["schedule"].(string); ok {
				data.Schedule = types.StringValue(sched)
			}
			break
		}
	}

	if !foundJob {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("could not find job with ID %q", jobID))
		return
	}

	data.ID = types.StringValue(jobID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
