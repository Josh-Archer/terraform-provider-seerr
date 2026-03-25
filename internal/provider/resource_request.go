package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &RequestResource{}
var _ resource.ResourceWithImportState = &RequestResource{}

type RequestResource struct {
	client *APIClient
}

type RequestModel struct {
	ID         types.String `tfsdk:"id"`
	MediaType  types.String `tfsdk:"media_type"`
	MediaID    types.Int64  `tfsdk:"media_id"`
	Seasons    types.List   `tfsdk:"seasons"`
	Is4K       types.Bool   `tfsdk:"is_4k"`
	ServerID   types.Int64  `tfsdk:"server_id"`
	ProfileID  types.Int64  `tfsdk:"profile_id"`
	RootFolder types.String `tfsdk:"root_folder"`
	UserID     types.Int64  `tfsdk:"user_id"`
	Status     types.Int64  `tfsdk:"status"`
}

func NewRequestResource() resource.Resource {
	return &RequestResource{}
}

func (r *RequestResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_request"
}

func (r *RequestResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Seerr media requests.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"media_type": schema.StringAttribute{
				MarkdownDescription: "The type of media (movie or tv).",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("movie", "tv"),
				},
			},
			"media_id": schema.Int64Attribute{
				MarkdownDescription: "The TMDB ID of the media.",
				Required:            true,
			},
			"seasons": schema.ListAttribute{
				MarkdownDescription: "List of season numbers to request (TV only).",
				Optional:            true,
				ElementType:         types.Int64Type,
			},
			"is_4k": schema.BoolAttribute{
				MarkdownDescription: "Whether to request in 4K.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"server_id": schema.Int64Attribute{
				MarkdownDescription: "Override the server ID for the request.",
				Optional:            true,
			},
			"profile_id": schema.Int64Attribute{
				MarkdownDescription: "Override the quality profile ID for the request.",
				Optional:            true,
			},
			"root_folder": schema.StringAttribute{
				MarkdownDescription: "Override the root folder for the request.",
				Optional:            true,
			},
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the user making the request (defaults to current user).",
				Optional:            true,
			},
			"status": schema.Int64Attribute{
				MarkdownDescription: "The status of the request (1: Pending, 2: Approved, 3: Declined).",
				Computed:            true,
			},
		},
	}
}

func (r *RequestResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RequestResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RequestModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]any{
		"mediaType": data.MediaType.ValueString(),
		"mediaId":   data.MediaID.ValueInt64(),
		"is4k":      data.Is4K.ValueBool(),
	}

	if !data.Seasons.IsNull() && !data.Seasons.IsUnknown() {
		var seasons []int64
		resp.Diagnostics.Append(data.Seasons.ElementsAs(ctx, &seasons, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		payload["seasons"] = seasons
	}

	if !data.ServerID.IsNull() && !data.ServerID.IsUnknown() {
		payload["serverId"] = data.ServerID.ValueInt64()
	}
	if !data.ProfileID.IsNull() && !data.ProfileID.IsUnknown() {
		payload["profileId"] = data.ProfileID.ValueInt64()
	}
	if !data.RootFolder.IsNull() && !data.RootFolder.IsUnknown() {
		payload["rootFolder"] = data.RootFolder.ValueString()
	}
	if !data.UserID.IsNull() && !data.UserID.IsUnknown() {
		payload["userId"] = data.UserID.ValueInt64()
	}

	body, _ := json.Marshal(payload)
	res, err := r.client.Request(ctx, "POST", "/api/v1/request", string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}
	if !HandleAPIResponse(ctx, res, &resp.Diagnostics, "Create") {
		return
	}

	extractedID, ok := ExtractIDFromJSON(res.Body)
	if !ok {
		resp.Diagnostics.AddError("Create Failed", "Could not extract request ID from response")
		return
	}

	data.ID = types.StringValue(extractedID)

	diags := r.readRequest(ctx, extractedID, &data)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RequestResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RequestModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := r.readRequest(ctx, data.ID.ValueString(), &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RequestResource) readRequest(ctx context.Context, requestID string, data *RequestModel) diag.Diagnostics {
	var diags diag.Diagnostics

	res, err := r.client.Request(ctx, "GET", "/api/v1/request/"+requestID, "", nil)
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

	if status, ok := m["status"].(float64); ok {
		data.Status = types.Int64Value(int64(status))
	}

	if media, ok := m["media"].(map[string]any); ok {
		if mediaType, ok := media["mediaType"].(string); ok {
			data.MediaType = types.StringValue(mediaType)
		}
		if tmdbId, ok := media["tmdbId"].(float64); ok {
			data.MediaID = types.Int64Value(int64(tmdbId))
		}
	}

	if is4k, ok := m["is4k"].(bool); ok {
		data.Is4K = types.BoolValue(is4k)
	}

	if requestedBy, ok := m["requestedBy"].(map[string]any); ok {
		if userId, ok := requestedBy["id"].(float64); ok {
			data.UserID = types.Int64Value(int64(userId))
		}
	}

	return diags
}

func (r *RequestResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RequestModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]any{
		"mediaType": data.MediaType.ValueString(),
		"is4k":      data.Is4K.ValueBool(),
	}

	if !data.Seasons.IsNull() && !data.Seasons.IsUnknown() {
		var seasons []int64
		resp.Diagnostics.Append(data.Seasons.ElementsAs(ctx, &seasons, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		payload["seasons"] = seasons
	}

	body, _ := json.Marshal(payload)
	res, err := r.client.Request(ctx, "PUT", "/api/v1/request/"+data.ID.ValueString(), string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}
	if !HandleAPIResponse(ctx, res, &resp.Diagnostics, "Update") {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RequestResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RequestModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.Request(ctx, "DELETE", "/api/v1/request/"+data.ID.ValueString(), "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Delete Failed", err.Error())
		return
	}
	if res.StatusCode != 404 && !HandleAPIResponse(ctx, res, &resp.Diagnostics, "Delete") {
		return
	}
}

func (r *RequestResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
