package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
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
	ID            types.String `tfsdk:"id"`
	IssueType     types.Int64  `tfsdk:"issue_type"`
	Message       types.String `tfsdk:"message"`
	MediaID       types.Int64  `tfsdk:"media_id"`
	Status        types.Int64  `tfsdk:"status"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
	CommentsCount types.Int64  `tfsdk:"comments_count"`
	CreatedBy     types.Object `tfsdk:"created_by"`
	ModifiedBy    types.Object `tfsdk:"modified_by"`
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
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"message": schema.StringAttribute{
				MarkdownDescription: "A message describing the issue.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"media_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the media associated with the issue.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"status": schema.Int64Attribute{
				MarkdownDescription: "The status of the issue (1: Open, 2: Resolved).",
				Optional:            true,
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Date and time the issue was created in ISO 8601 format.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Date and time the issue was last updated in ISO 8601 format.",
				Computed:            true,
			},
			"comments_count": schema.Int64Attribute{
				MarkdownDescription: "The number of comments on this issue.",
				Computed:            true,
			},
			"created_by": schema.SingleNestedAttribute{
				MarkdownDescription: "The user who created the issue.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"id": schema.Int64Attribute{
						MarkdownDescription: "User ID.",
						Computed:            true,
					},
					"email": schema.StringAttribute{
						MarkdownDescription: "User email.",
						Computed:            true,
					},
					"display_name": schema.StringAttribute{
						MarkdownDescription: "User display name.",
						Computed:            true,
					},
					"avatar": schema.StringAttribute{
						MarkdownDescription: "User avatar URL.",
						Computed:            true,
					},
				},
			},
			"modified_by": schema.SingleNestedAttribute{
				MarkdownDescription: "The user who last modified the issue.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"id": schema.Int64Attribute{
						MarkdownDescription: "User ID.",
						Computed:            true,
					},
					"email": schema.StringAttribute{
						MarkdownDescription: "User email.",
						Computed:            true,
					},
					"display_name": schema.StringAttribute{
						MarkdownDescription: "User display name.",
						Computed:            true,
					},
					"avatar": schema.StringAttribute{
						MarkdownDescription: "User avatar URL.",
						Computed:            true,
					},
				},
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
		if err := r.applyIssueStatus(ctx, extractedID, data.Status.ValueInt64()); err != nil {
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
	if data.ID.IsNull() || data.ID.IsUnknown() || data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

var issueUserAttrTypes = map[string]attr.Type{
	"id":           types.Int64Type,
	"email":        types.StringType,
	"display_name": types.StringType,
	"avatar":       types.StringType,
}

func parseIssueUserObject(v any) types.Object {
	m, ok := v.(map[string]any)
	if !ok || m == nil {
		return types.ObjectNull(issueUserAttrTypes)
	}
	id := types.Int64Null()
	if idVal, ok := int64ValueFromAny(m["id"]); ok {
		id = types.Int64Value(idVal)
	}
	email := types.StringNull()
	if emailVal, ok := stringValueFromAny(m["email"]); ok {
		email = types.StringValue(emailVal)
	}
	displayName := types.StringNull()
	if dn, ok := stringValueFromAny(m["displayName"]); ok {
		displayName = types.StringValue(dn)
	} else if un, ok := stringValueFromAny(m["username"]); ok {
		displayName = types.StringValue(un)
	}
	avatar := types.StringNull()
	if av, ok := stringValueFromAny(m["avatar"]); ok {
		avatar = types.StringValue(av)
	}
	obj, _ := types.ObjectValue(issueUserAttrTypes, map[string]attr.Value{
		"id":           id,
		"email":        email,
		"display_name": displayName,
		"avatar":       avatar,
	})
	return obj
}

func (r *IssueResource) readIssue(ctx context.Context, issueID string, data *IssueModel) diag.Diagnostics {
	var diags diag.Diagnostics

	res, err := r.client.Request(ctx, "GET", "/api/v1/issue/"+issueID, "", nil)
	if err != nil {
		diags.AddError("Read Failed", err.Error())
		return diags
	}
	if res.StatusCode == 404 {
		data.ID = types.StringNull()
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
	if media, ok := m["media"].(map[string]any); ok {
		if mediaID, ok := media["id"].(float64); ok {
			data.MediaID = types.Int64Value(int64(mediaID))
		}
	}

	if createdAt, ok := stringValueFromAny(m["createdAt"]); ok {
		data.CreatedAt = types.StringValue(createdAt)
	} else {
		data.CreatedAt = types.StringNull()
	}
	if updatedAt, ok := stringValueFromAny(m["updatedAt"]); ok {
		data.UpdatedAt = types.StringValue(updatedAt)
	} else {
		data.UpdatedAt = types.StringNull()
	}

	if comments, ok := m["comments"].([]any); ok {
		data.CommentsCount = types.Int64Value(int64(len(comments)))
	} else {
		data.CommentsCount = types.Int64Value(0)
	}

	data.CreatedBy = parseIssueUserObject(m["createdBy"])
	data.ModifiedBy = parseIssueUserObject(m["modifiedBy"])

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
		if err := r.applyIssueStatus(ctx, data.ID.ValueString(), data.Status.ValueInt64()); err != nil {
			resp.Diagnostics.AddError("Update Status Failed", err.Error())
			return
		}
	}

	diags := r.readIssue(ctx, data.ID.ValueString(), &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// applyIssueStatus posts a status transition to Seerr and fails on non-2xx responses.
func (r *IssueResource) applyIssueStatus(ctx context.Context, issueID string, status int64) error {
	statusPath, ok := issueStatusPath(status)
	if !ok {
		return fmt.Errorf("unsupported issue status %d; valid values are 1 (open) and 2 (resolved)", status)
	}
	res, err := r.client.Request(ctx, "POST", fmt.Sprintf("/api/v1/issue/%s/%s", issueID, statusPath), "", nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}
	return nil
}

func issueStatusPath(status int64) (string, bool) {
	switch status {
	case 1:
		return "open", true
	case 2:
		return "resolved", true
	default:
		return "", false
	}
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
