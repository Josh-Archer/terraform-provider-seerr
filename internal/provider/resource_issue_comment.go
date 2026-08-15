package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &IssueCommentResource{}
var _ resource.ResourceWithImportState = &IssueCommentResource{}

type IssueCommentResource struct {
	client *APIClient
}

type IssueCommentModel struct {
	ID        types.String `tfsdk:"id"`
	IssueID   types.Int64  `tfsdk:"issue_id"`
	Message   types.String `tfsdk:"message"`
	UserID    types.Int64  `tfsdk:"user_id"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func NewIssueCommentResource() resource.Resource {
	return &IssueCommentResource{}
}

func (r *IssueCommentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_issue_comment"
}

func (r *IssueCommentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Seerr issue comments via `/api/v1/issue/{issueId}/comment` and `/api/v1/issueComment/{commentId}`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The comment ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"issue_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the issue to comment on.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"message": schema.StringAttribute{
				MarkdownDescription: "The comment message text.",
				Required:            true,
			},
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the user who authored the comment.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Date and time the comment was created in ISO 8601 format.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Date and time the comment was last updated in ISO 8601 format.",
				Computed:            true,
			},
		},
	}
}

func (r *IssueCommentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IssueCommentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IssueCommentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	issueID := data.IssueID.ValueInt64()
	payload := map[string]any{
		"message": data.Message.ValueString(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Create Issue Comment Failed", err.Error())
		return
	}

	endpoint := fmt.Sprintf("/api/v1/issue/%d/comment", issueID)
	res, err := r.client.Request(ctx, "POST", endpoint, string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Create Issue Comment Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Create Issue Comment Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	var decoded map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err != nil {
		resp.Diagnostics.AddError("Create Issue Comment Failed", fmt.Sprintf("failed to parse response: %v", err))
		return
	}

	var commentID int64
	// The POST endpoint may return the full Issue object containing a comments array, or a direct IssueComment object.
	if comments, ok := decoded["comments"].([]any); ok && len(comments) > 0 {
		// Prefer the last comment in the list
		if lastComment, ok := comments[len(comments)-1].(map[string]any); ok {
			if id, ok := int64ValueFromAny(lastComment["id"]); ok {
				commentID = id
			}
		}
	} else if id, ok := int64ValueFromAny(decoded["id"]); ok {
		commentID = id
	}

	if commentID == 0 {
		resp.Diagnostics.AddError("Create Issue Comment Failed", "could not determine comment ID from API response")
		return
	}

	data.ID = types.StringValue(strconv.FormatInt(commentID, 10))

	if err := r.readCommentDetails(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Read Issue Comment Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IssueCommentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IssueCommentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readCommentDetails(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Read Issue Comment Failed", err.Error())
		return
	}

	if data.ID.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IssueCommentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data IssueCommentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]any{
		"message": data.Message.ValueString(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Update Issue Comment Failed", err.Error())
		return
	}

	endpoint := fmt.Sprintf("/api/v1/issueComment/%s", data.ID.ValueString())
	res, err := r.client.Request(ctx, "PUT", endpoint, string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Update Issue Comment Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Update Issue Comment Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	if err := r.readCommentDetails(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Read Issue Comment Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IssueCommentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IssueCommentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/api/v1/issueComment/%s", data.ID.ValueString())
	res, err := r.client.Request(ctx, "DELETE", endpoint, "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Delete Issue Comment Failed", err.Error())
		return
	}
	if res.StatusCode != 404 && res.StatusCode != 204 && !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Delete Issue Comment Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
	}
}

func (r *IssueCommentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *IssueCommentResource) readCommentDetails(ctx context.Context, data *IssueCommentModel) error {
	commentID := data.ID.ValueString()
	endpoint := fmt.Sprintf("/api/v1/issueComment/%s", commentID)
	res, err := r.client.Request(ctx, "GET", endpoint, "", nil)
	if err != nil {
		return err
	}
	if res.StatusCode == 404 {
		data.ID = types.StringNull()
		return nil
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}

	var m map[string]any
	if err := json.Unmarshal(res.Body, &m); err != nil {
		return fmt.Errorf("failed to parse issue comment JSON: %w", err)
	}

	if id, ok := int64ValueFromAny(m["id"]); ok {
		data.ID = types.StringValue(strconv.FormatInt(id, 10))
	}
	if msg, ok := stringValueFromAny(m["message"]); ok {
		data.Message = types.StringValue(msg)
	}
	if createdAt, ok := stringValueFromAny(m["createdAt"]); ok {
		data.CreatedAt = types.StringValue(createdAt)
	}
	if updatedAt, ok := stringValueFromAny(m["updatedAt"]); ok {
		data.UpdatedAt = types.StringValue(updatedAt)
	}

	if user, ok := m["user"].(map[string]any); ok {
		if uid, ok := int64ValueFromAny(user["id"]); ok {
			data.UserID = types.Int64Value(uid)
		}
	} else if uid, ok := int64ValueFromAny(m["userId"]); ok {
		data.UserID = types.Int64Value(uid)
	}

	// Try to populate issue_id if present in comment payload
	if issue, ok := m["issue"].(map[string]any); ok {
		if iid, ok := int64ValueFromAny(issue["id"]); ok {
			data.IssueID = types.Int64Value(iid)
		}
	} else if iid, ok := int64ValueFromAny(m["issueId"]); ok {
		data.IssueID = types.Int64Value(iid)
	}

	return nil
}
