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
	ID               types.String `tfsdk:"id"`
	JobID            types.String `tfsdk:"job_id"`
	Schedule         types.String `tfsdk:"schedule"`
	PreviousSchedule types.String `tfsdk:"previous_schedule"`
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
			"previous_schedule": schema.StringAttribute{
				MarkdownDescription: "The schedule observed before Terraform first managed this job. Used to restore the original schedule on delete.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
	previousSchedule, err := r.fetchCurrentSchedule(ctx, jobID)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}

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
	data.PreviousSchedule = types.StringValue(previousSchedule)

	if err := r.refreshJobSchedule(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}

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
			if sched, ok := jobScheduleFromJob(job); ok {
				data.Schedule = types.StringValue(sched)
				if data.PreviousSchedule.IsNull() || data.PreviousSchedule.IsUnknown() || data.PreviousSchedule.ValueString() == "" {
					data.PreviousSchedule = types.StringValue(sched)
				}
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

	if err := r.refreshJobSchedule(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JobScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data JobScheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.PreviousSchedule.IsNull() || data.PreviousSchedule.IsUnknown() || data.PreviousSchedule.ValueString() == "" {
		return
	}
	if err := r.restorePreviousSchedule(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Delete Failed", err.Error())
	}
}

func (r *JobScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *JobScheduleResource) refreshJobSchedule(ctx context.Context, data *JobScheduleModel) error {
	jobID := data.JobID.ValueString()
	if jobID == "" {
		jobID = data.ID.ValueString()
		data.JobID = types.StringValue(jobID)
	}

	schedule, err := r.fetchCurrentSchedule(ctx, jobID)
	if err != nil {
		return err
	}

	data.ID = types.StringValue(jobID)
	data.Schedule = types.StringValue(schedule)
	if data.PreviousSchedule.IsNull() || data.PreviousSchedule.IsUnknown() || data.PreviousSchedule.ValueString() == "" {
		data.PreviousSchedule = types.StringValue(schedule)
	}

	return nil
}

func (r *JobScheduleResource) fetchCurrentSchedule(ctx context.Context, jobID string) (string, error) {
	res, err := r.client.Request(ctx, "GET", "/api/v1/settings/jobs", "", nil)
	if err != nil {
		return "", err
	}

	if !StatusIsOK(res.StatusCode) {
		return "", fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}

	var jobsList []map[string]any
	if err := json.Unmarshal(res.Body, &jobsList); err != nil {
		return "", fmt.Errorf("failed to decode response as list of jobs: %s", err)
	}

	for _, job := range jobsList {
		if jID, ok := job["id"].(string); ok && jID == jobID {
			if schedule, ok := jobScheduleFromJob(job); ok {
				return schedule, nil
			}
			return "", fmt.Errorf("job %q did not include a schedule", jobID)
		}
	}

	return "", fmt.Errorf("job %q not found", jobID)
}

func jobScheduleFromJob(job map[string]any) (string, bool) {
	if sched, ok := job["schedule"].(string); ok && sched != "" {
		return sched, true
	}
	if sched, ok := job["cronSchedule"].(string); ok && sched != "" {
		return sched, true
	}
	return "", false
}

func (r *JobScheduleResource) restorePreviousSchedule(ctx context.Context, data *JobScheduleModel) error {
	payload := map[string]string{
		"schedule": data.PreviousSchedule.ValueString(),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %s", err)
	}

	endpoint := fmt.Sprintf("/api/v1/settings/jobs/%s/schedule", data.JobID.ValueString())
	res, err := r.client.Request(ctx, "POST", endpoint, string(body), nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}
	return nil
}
