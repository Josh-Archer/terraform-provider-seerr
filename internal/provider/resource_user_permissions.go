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

var _ resource.Resource = &UserPermissionsResource{}
var _ resource.ResourceWithImportState = &UserPermissionsResource{}

type UserPermissionsResource struct {
	client *APIClient
}

type UserPermissionsModel struct {
	ID           types.String `tfsdk:"id"`
	UserID       types.Int64  `tfsdk:"user_id"`
	Permissions  types.Int64  `tfsdk:"permissions"`
	ResponseJSON types.String `tfsdk:"response_json"`
	StatusCode   types.Int64  `tfsdk:"status_code"`
}

func NewUserPermissionsResource() resource.Resource { return &UserPermissionsResource{} }

func (r *UserPermissionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_permissions"
}

func (r *UserPermissionsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Seerr user permissions via /api/v1/user/{userId}.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the user whose permissions to manage.",
				Required:            true,
			},
			"permissions": schema.Int64Attribute{
				MarkdownDescription: "Numeric permissions bitmask for the user.",
				Required:            true,
			},
			"response_json": schema.StringAttribute{
				MarkdownDescription: "Raw JSON response body from the latest operation.",
				Computed:            true,
			},
			"status_code": schema.Int64Attribute{
				MarkdownDescription: "HTTP status code from the latest operation.",
				Computed:            true,
			},
		},
	}
}

func (r *UserPermissionsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func userPermissionsPath(userID int64) string {
	return fmt.Sprintf("/api/v1/user/%d", userID)
}

func (r *UserPermissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserPermissionsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(map[string]any{"permissions": data.Permissions.ValueInt64()})
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}

	apiPath := userPermissionsPath(data.UserID.ValueInt64())
	res, err := r.client.Request(ctx, "PUT", apiPath, string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%d", data.UserID.ValueInt64()))
	data.StatusCode = types.Int64Value(int64(res.StatusCode))
	data.ResponseJSON = types.StringValue(string(res.Body))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserPermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserPermissionsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiPath := userPermissionsPath(data.UserID.ValueInt64())
	res, err := r.client.Request(ctx, "GET", apiPath, "", nil)
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
	if err := json.Unmarshal(res.Body, &decoded); err == nil {
		if raw, ok := decoded["permissions"]; ok {
			if v, ok := raw.(float64); ok {
				data.Permissions = types.Int64Value(int64(v))
			}
		}
	}

	data.StatusCode = types.Int64Value(int64(res.StatusCode))
	data.ResponseJSON = types.StringValue(string(res.Body))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserPermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserPermissionsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(map[string]any{"permissions": data.Permissions.ValueInt64()})
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}

	apiPath := userPermissionsPath(data.UserID.ValueInt64())
	res, err := r.client.Request(ctx, "PUT", apiPath, string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Update Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%d", data.UserID.ValueInt64()))
	data.StatusCode = types.Int64Value(int64(res.StatusCode))
	data.ResponseJSON = types.StringValue(string(res.Body))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserPermissionsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No DELETE route for user permissions; removing from state only.
}

func (r *UserPermissionsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := requireInt64ID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), id)...)
}
