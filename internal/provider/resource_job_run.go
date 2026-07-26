package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &JobRunResource{}
var _ resource.ResourceWithImportState = &JobRunResource{}

type JobRunResource struct {
	client *APIClient
}

type JobRunModel struct {
	ID             types.String `tfsdk:"id"`
	JobID          types.String `tfsdk:"job_id"`
	Triggers       types.Map    `tfsdk:"triggers"`
	CancelOnDelete types.Bool   `tfsdk:"cancel_on_delete"`
}

func NewJobRunResource() resource.Resource {
	return &JobRunResource{}
}

func (r *JobRunResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job_run"
}

func (r *JobRunResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Trigger an operational job execution via `/api/v1/settings/jobs/{jobId}/run`. Optionally cancel in-flight job execution on destroy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"job_id": schema.StringAttribute{
				MarkdownDescription: "The ID string of the job to run (e.g. `plex-sync`, `user-import`, `radarr-sync`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"triggers": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Arbitrary map of values (such as timestamps or hashes) that, when changed, will re-trigger the job execution.",
				Optional:            true,
			},
			"cancel_on_delete": schema.BoolAttribute{
				MarkdownDescription: "If true, destroying this resource will call `/api/v1/settings/jobs/{jobId}/cancel` to attempt cancelling an in-flight job execution.",
				Optional:            true,
			},
		},
	}
}

func (r *JobRunResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *JobRunResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data JobRunModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.runJob(ctx, data.JobID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Job Run Failed", err.Error())
		return
	}

	data.ID = data.JobID
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JobRunResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data JobRunModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Job runs are operational action triggers; state is preserved across reads.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JobRunResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data JobRunModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.runJob(ctx, data.JobID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Job Run Failed", err.Error())
		return
	}

	data.ID = data.JobID
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JobRunResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data JobRunModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.CancelOnDelete.ValueBool() {
		if err := r.cancelJob(ctx, data.JobID.ValueString()); err != nil {
			resp.Diagnostics.AddWarning("Job Cancel Failed", fmt.Sprintf("Failed to cancel job %s on delete: %s", data.JobID.ValueString(), err.Error()))
		}
	}
}

func (r *JobRunResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("job_id"), req.ID)...)
}

func (r *JobRunResource) runJob(ctx context.Context, jobID string) error {
	apiPath := fmt.Sprintf("/api/v1/settings/jobs/%s/run", jobID)
	res, err := r.client.Request(ctx, "POST", apiPath, "", nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}
	return nil
}

func (r *JobRunResource) cancelJob(ctx context.Context, jobID string) error {
	apiPath := fmt.Sprintf("/api/v1/settings/jobs/%s/cancel", jobID)
	res, err := r.client.Request(ctx, "POST", apiPath, "", nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}
	return nil
}
