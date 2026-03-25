package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &IssueResource{}
var _ resource.ResourceWithImportState = &IssueResource{}

type IssueResource struct {
	client *APIClient
}

type IssueModel struct {
	ID        types.String `tfsdk:"id"`
	IssueType types.Int64  `tfsdk:"issue_type"`
	Message   types.String `tfsdk:"message"`
	MediaID   types.Int64  `tfsdk:"media_id"`
	Status    types.Int64  `tfsdk:"status"`
}

func NewIssueResource() resource.Resource {
	return &IssueResource{}
}

func (r *IssueResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_issue"
}

func (r *IssueResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Seerr media issues.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"issue_type": schema.Int64Attribute{
				MarkdownDescription: "The type of the issue (1: Video, 2: Audio, 3: Subtitle, 4: Other).",
				Required:            true,
			},
			"message": schema.StringAttribute{
				MarkdownDescription: "A message describing the issue.",
				Optional:            true,
			},
			"media_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the media associated with the issue.",
				Required:            true,
			},
			"status": schema.Int64Attribute{
				MarkdownDescription: "The status of the issue (1: Open, 2: Resolved).",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *IssueResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IssueResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IssueModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]any{
		"issueType": data.IssueType.ValueInt64(),
		"mediaId":   data.MediaID.ValueInt64(),
	}
	if !data.Message.IsNull() && !data.Message.IsUnknown() {
		payload["message"] = data.Message.ValueString()
	}

	body, _ := json.Marshal(payload)
	res, err := r.client.Request(ctx, "POST", "/api/v1/issue", string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}
	if !HandleAPIResponse(ctx, res, &resp.Diagnostics, "Create") {
		return
	}

	extractedID, ok := ExtractIDFromJSON(res.Body)
	if !ok {
		resp.Diagnostics.AddError("Create Failed", "Could not extract issue ID from response")
		return
	}

	data.ID = types.StringValue(extractedID)

	// If status was provided as Resolved (2) in plan, update it
	if !data.Status.IsNull() && !data.Status.IsUnknown() && data.Status.ValueInt64() == 2 {
		_, err := r.client.Request(ctx, "POST", fmt.Sprintf("/api/v1/issue/%s/resolved", extractedID), "", nil)
		if err != nil {
			resp.Diagnostics.AddError("Update Status Failed", err.Error())
			return
		}
	}

	diags := r.readIssue(ctx, extractedID, &data)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IssueResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IssueModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := r.readIssue(ctx, data.ID.ValueString(), &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IssueResource) readIssue(ctx context.Context, issueID string, data *IssueModel) diag.Diagnostics {
	var diags diag.Diagnostics

	res, err := r.client.Request(ctx, "GET", "/api/v1/issue/"+issueID, "", nil)
	if err != nil {
		diags.AddError("Read Failed", err.Error())
		return diags
	}
	if res.StatusCode == 404 {
		return diags
	}
	if !HandleAPIResponse(ctx, res, &diags, "Read") {
		return diags
	}

	var m map[string]any
	if err := json.Unmarshal(res.Body, &m); err != nil {
		diags.AddError("Read Failed", err.Error())
		return diags
	}

	if it, ok := m["issueType"].(float64); ok {
		data.IssueType = types.Int64Value(int64(it))
	}
	if status, ok := m["status"].(float64); ok {
		data.Status = types.Int64Value(int64(status))
	}

	// Message is in an array of comments or similar in Overseerr issues,
	// but here we just simplify or keep as provided if we can't easily fetch the original message.

	return diags
}

func (r *IssueResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state IssueModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var data IssueModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update status if it changed
	if !data.Status.IsNull() && !data.Status.IsUnknown() && data.Status.ValueInt64() != state.Status.ValueInt64() {
		statusPath := "open"
		if data.Status.ValueInt64() == 2 {
			statusPath = "resolved"
		}
		_, err := r.client.Request(ctx, "POST", fmt.Sprintf("/api/v1/issue/%s/%s", data.ID.ValueString(), statusPath), "", nil)
		if err != nil {
			resp.Diagnostics.AddError("Update Status Failed", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IssueResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IssueModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.Request(ctx, "DELETE", "/api/v1/issue/"+data.ID.ValueString(), "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Delete Failed", err.Error())
		return
	}
	if res.StatusCode != 404 && !HandleAPIResponse(ctx, res, &resp.Diagnostics, "Delete") {
		return
	}
}

func (r *IssueResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
