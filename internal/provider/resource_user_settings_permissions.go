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

var _ resource.Resource = &UserSettingsPermissionsResource{}
var _ resource.ResourceWithImportState = &UserSettingsPermissionsResource{}

type UserSettingsPermissionsResource struct {
	client *APIClient
}

type UserSettingsPermissionsModel struct {
	ID                 types.String `tfsdk:"id"`
	UserID             types.Int64  `tfsdk:"user_id"`
	AutoApproveMovies  types.Bool   `tfsdk:"auto_approve_movies"`
	AutoApproveTV      types.Bool   `tfsdk:"auto_approve_tv"`
	AutoApprove4KMovies types.Bool   `tfsdk:"auto_approve_4k_movies"`
	AutoApprove4KTV     types.Bool   `tfsdk:"auto_approve_4k_tv"`
}

func NewUserSettingsPermissionsResource() resource.Resource {
	return &UserSettingsPermissionsResource{}
}

func (r *UserSettingsPermissionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_settings_permissions"
}

func (r *UserSettingsPermissionsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage granular Seerr user settings permissions via /api/v1/user/{userId}/settings/permissions.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the user whose settings permissions to manage.",
				Required:            true,
			},
			"auto_approve_movies": schema.BoolAttribute{
				MarkdownDescription: "Whether the user's movie requests are automatically approved.",
				Optional:            true,
				Computed:            true,
			},
			"auto_approve_tv": schema.BoolAttribute{
				MarkdownDescription: "Whether the user's TV series requests are automatically approved.",
				Optional:            true,
				Computed:            true,
			},
			"auto_approve_4k_movies": schema.BoolAttribute{
				MarkdownDescription: "Whether the user's 4K movie requests are automatically approved.",
				Optional:            true,
				Computed:            true,
			},
			"auto_approve_4k_tv": schema.BoolAttribute{
				MarkdownDescription: "Whether the user's 4K TV series requests are automatically approved.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *UserSettingsPermissionsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserSettingsPermissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserSettingsPermissionsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.updatePermissions(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%d", data.UserID.ValueInt64()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserSettingsPermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserSettingsPermissionsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userID := data.UserID.ValueInt64()
	res, err := r.client.Request(ctx, "GET", fmt.Sprintf("/api/v1/user/%d/settings/permissions", userID), "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if res.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	var decoded map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err != nil {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("failed to decode response: %s", err))
		return
	}

	if v, ok := decoded["autoApproveMovies"].(bool); ok {
		data.AutoApproveMovies = types.BoolValue(v)
	}
	if v, ok := decoded["autoApproveTv"].(bool); ok {
		data.AutoApproveTV = types.BoolValue(v)
	}
	if v, ok := decoded["autoApprove4kMovies"].(bool); ok {
		data.AutoApprove4KMovies = types.BoolValue(v)
	}
	if v, ok := decoded["autoApprove4kTv"].(bool); ok {
		data.AutoApprove4KTV = types.BoolValue(v)
	}

	data.ID = types.StringValue(fmt.Sprintf("%d", userID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserSettingsPermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserSettingsPermissionsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.updatePermissions(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%d", data.UserID.ValueInt64()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserSettingsPermissionsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No DELETE route for user settings permissions.
}

func (r *UserSettingsPermissionsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := requireInt64ID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), id)...)
}

func (r *UserSettingsPermissionsResource) updatePermissions(ctx context.Context, data *UserSettingsPermissionsModel) error {
	payload := map[string]any{}
	if !data.AutoApproveMovies.IsNull() && !data.AutoApproveMovies.IsUnknown() {
		payload["autoApproveMovies"] = data.AutoApproveMovies.ValueBool()
	}
	if !data.AutoApproveTV.IsNull() && !data.AutoApproveTV.IsUnknown() {
		payload["autoApproveTv"] = data.AutoApproveTV.ValueBool()
	}
	if !data.AutoApprove4KMovies.IsNull() && !data.AutoApprove4KMovies.IsUnknown() {
		payload["autoApprove4kMovies"] = data.AutoApprove4KMovies.ValueBool()
	}
	if !data.AutoApprove4KTV.IsNull() && !data.AutoApprove4KTV.IsUnknown() {
		payload["autoApprove4kTv"] = data.AutoApprove4KTV.ValueBool()
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	res, err := r.client.Request(ctx, "POST", fmt.Sprintf("/api/v1/user/%d/settings/permissions", data.UserID.ValueInt64()), string(body), nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}

	return nil
}
