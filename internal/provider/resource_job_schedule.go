package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &JobScheduleResource{}
var _ resource.ResourceWithImportState = &JobScheduleResource{}

type JobScheduleResource struct {
	client *APIClient
}

type JobScheduleModel struct {
	ID       types.String `tfsdk:"id"`
	JobID    types.String `tfsdk:"job_id"`
	Schedule types.String `tfsdk:"schedule"`
}

func NewJobScheduleResource() resource.Resource { return &JobScheduleResource{} }

func (r *JobScheduleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job_schedule"
}

func (r *JobScheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage a background job schedule in Seerr via `/api/v1/settings/jobs/{job_id}/schedule`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"job_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the background job (e.g., `plex-sync`, `radarr-sync`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"schedule": schema.StringAttribute{
				MarkdownDescription: "The cron expression for the job schedule.",
				Required:            true,
			},
		},
	}
}

func (r *JobScheduleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Configure Type", fmt.Sprintf("Expected *APIClient, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *JobScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data JobScheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID := data.JobID.ValueString()
	schedule := data.Schedule.ValueString()

	payload := map[string]string{
		"schedule": schedule,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("failed to marshal payload: %s", err))
		return
	}

	endpoint := fmt.Sprintf("/api/v1/settings/jobs/%s/schedule", jobID)
	res, err := r.client.Request(ctx, "POST", endpoint, string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	// We set ID to JobID
	data.ID = types.StringValue(jobID)

	// Since POST is typically fire-and-forget for schedules and returns simple confirmation or the job object itself,
	// let's do a Read to ensure state matches exactly what Seerr returns.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JobScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data JobScheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID := data.JobID.ValueString()
	if jobID == "" {
		// fallback to ID if JobID somehow empty in migration or import
		jobID = data.ID.ValueString()
		data.JobID = data.ID
	}

	res, err := r.client.Request(ctx, "GET", "/api/v1/settings/jobs", "", nil)
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
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JobScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Re-use Create logic for POST /api/v1/settings/jobs/{job_id}/schedule since it just sets the schedule
	var data JobScheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID := data.JobID.ValueString()
	schedule := data.Schedule.ValueString()

	payload := map[string]string{
		"schedule": schedule,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", fmt.Sprintf("failed to marshal payload: %s", err))
		return
	}

	endpoint := fmt.Sprintf("/api/v1/settings/jobs/%s/schedule", jobID)
	res, err := r.client.Request(ctx, "POST", endpoint, string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Update Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JobScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// There is no concept of "deleting" a job schedule in Seerr, it just resets to default if empty or we keep it.
	// Since we can't easily fetch default cleanly without a fresh instance, we will just leave the job schedule as-is on Seerr's side.
	// The resource will be removed from Terraform state automatically.
}

func (r *JobScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
