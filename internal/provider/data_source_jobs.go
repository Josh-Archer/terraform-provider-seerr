package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &JobsDataSource{}

type JobsDataSource struct {
	client *APIClient
}

type JobSummaryModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Interval types.String `tfsdk:"interval"`
	LastRun  types.String `tfsdk:"last_run"`
	NextRun  types.String `tfsdk:"next_run"`
	Schedule types.String `tfsdk:"schedule"`
	Enabled  types.Bool   `tfsdk:"enabled"`
}

type JobsDataSourceModel struct {
	ID   types.String      `tfsdk:"id"`
	Jobs []JobSummaryModel `tfsdk:"jobs"`
}

func NewJobsDataSource() datasource.DataSource {
	return &JobsDataSource{}
}

func (d *JobsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jobs"
}

func (d *JobsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get information about all background jobs in Seerr.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder ID for the data source.",
				Computed:            true,
			},
			"jobs": schema.ListNestedAttribute{
				MarkdownDescription: "List of jobs.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The ID of the job.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the job.",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of job.",
							Computed:            true,
						},
						"interval": schema.StringAttribute{
							MarkdownDescription: "Job interval.",
							Computed:            true,
						},
						"last_run": schema.StringAttribute{
							MarkdownDescription: "Last run time.",
							Computed:            true,
						},
						"next_run": schema.StringAttribute{
							MarkdownDescription: "Next run time.",
							Computed:            true,
						},
						"schedule": schema.StringAttribute{
							MarkdownDescription: "Cron schedule.",
							Computed:            true,
						},
						"enabled": schema.BoolAttribute{
							MarkdownDescription: "Whether the job is enabled.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *JobsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *JobsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data JobsDataSourceModel

	// Fetch jobs
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
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("failed to decode response: %s", err))
		return
	}

	for _, j := range jobsList {
		job := JobSummaryModel{}

		if id, ok := j["id"].(string); ok {
			job.ID = types.StringValue(id)
		}
		if name, ok := j["name"].(string); ok {
			job.Name = types.StringValue(name)
		}
		if t, ok := j["type"].(string); ok {
			job.Type = types.StringValue(t)
		}
		if interval, ok := j["interval"].(string); ok {
			job.Interval = types.StringValue(interval)
		}
		if last, ok := j["lastRun"].(string); ok {
			job.LastRun = types.StringValue(last)
		}
		if next, ok := j["nextRun"].(string); ok {
			job.NextRun = types.StringValue(next)
		}
		if sched, ok := j["schedule"].(string); ok {
			job.Schedule = types.StringValue(sched)
		}
		if enabled, ok := j["enabled"].(bool); ok {
			job.Enabled = types.BoolValue(enabled)
		} else if enabledStr, ok := j["enabled"].(string); ok {
			b, _ := strconv.ParseBool(enabledStr)
			job.Enabled = types.BoolValue(b)
		}

		data.Jobs = append(data.Jobs, job)
	}

	data.ID = types.StringValue("all_jobs")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
