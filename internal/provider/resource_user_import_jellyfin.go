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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &UserImportJellyfinResource{}
var _ resource.ResourceWithImportState = &UserImportJellyfinResource{}

type UserImportJellyfinResource struct {
	client *APIClient
}

type ImportedJellyfinUserModel struct {
	ID             types.String `tfsdk:"id"`
	Username       types.String `tfsdk:"username"`
	Email          types.String `tfsdk:"email"`
	JellyfinUserID types.String `tfsdk:"jellyfin_user_id"`
}

type UserImportJellyfinModel struct {
	ID              types.String `tfsdk:"id"`
	JellyfinUserIDs types.Set    `tfsdk:"jellyfin_user_ids"`
	ImportedUsers   types.List   `tfsdk:"imported_users"`
	ImportedCount   types.Int64  `tfsdk:"imported_count"`
	Triggers        types.Map    `tfsdk:"triggers"`
}

func NewUserImportJellyfinResource() resource.Resource {
	return &UserImportJellyfinResource{}
}

func (r *UserImportJellyfinResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_import_jellyfin"
}

func (r *UserImportJellyfinResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Import users from a connected Jellyfin media server into Seerr via `/api/v1/user/import-from-jellyfin`.\n\n" +
			"This resource facilitates the **bootstrap sequence**: Media Server Setup → `seerr_user_import_jellyfin` → `seerr_permission_set` / `seerr_user_quota`.\n\n" +
			"If `jellyfin_user_ids` is specified, only matching Jellyfin users are imported. If omitted or empty, all available Jellyfin users on the server are imported.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Resource identifier (defaults to `jellyfin_import`).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"jellyfin_user_ids": schema.SetAttribute{
				MarkdownDescription: "Optional set of specific Jellyfin user IDs to import. If omitted, all available Jellyfin users are imported.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"imported_users": schema.ListNestedAttribute{
				MarkdownDescription: "List of users imported from Jellyfin.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The numeric Seerr user ID.",
							Computed:            true,
						},
						"username": schema.StringAttribute{
							MarkdownDescription: "The user's display name.",
							Computed:            true,
						},
						"email": schema.StringAttribute{
							MarkdownDescription: "The user's email address.",
							Computed:            true,
						},
						"jellyfin_user_id": schema.StringAttribute{
							MarkdownDescription: "The user's Jellyfin user ID.",
							Computed:            true,
						},
					},
				},
			},
			"imported_count": schema.Int64Attribute{
				MarkdownDescription: "Number of users imported.",
				Computed:            true,
			},
			"triggers": schema.MapAttribute{
				MarkdownDescription: "Arbitrary map of trigger values (e.g. `version = \"1.0\"`) to force a re-import action when values change.",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *UserImportJellyfinResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserImportJellyfinResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserImportJellyfinModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.performImport(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Jellyfin User Import Failed", err.Error())
		return
	}

	data.ID = types.StringValue("jellyfin_import")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserImportJellyfinResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserImportJellyfinModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch all users to refresh imported Jellyfin users list
	results, err := fetchAllPaginatedResults(ctx, r.client, "/api/v1/user", defaultPaginationPageSize)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	var imported []ImportedJellyfinUserModel
	for _, u := range results {
		jellyfinID, _ := u["jellyfinUserId"].(string)
		if jellyfinID == "" {
			continue
		}

		user := ImportedJellyfinUserModel{
			JellyfinUserID: types.StringValue(jellyfinID),
		}

		switch v := u["id"].(type) {
		case float64:
			user.ID = types.StringValue(fmt.Sprintf("%.0f", v))
		case string:
			user.ID = types.StringValue(v)
		}

		if e, ok := u["email"].(string); ok {
			user.Email = types.StringValue(e)
		}
		if un, ok := u["username"].(string); ok {
			user.Username = types.StringValue(un)
		}

		imported = append(imported, user)
	}

	listVal, diags := buildImportedJellyfinUserListValue(ctx, imported)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ImportedUsers = listVal
	data.ImportedCount = types.Int64Value(int64(len(imported)))
	if data.ID.IsNull() || data.ID.ValueString() == "" {
		data.ID = types.StringValue("jellyfin_import")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserImportJellyfinResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserImportJellyfinModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.performImport(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Jellyfin User Import Update Failed", err.Error())
		return
	}

	data.ID = types.StringValue("jellyfin_import")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserImportJellyfinResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// One-shot import action: deletion removes from Terraform state only.
}

func (r *UserImportJellyfinResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *UserImportJellyfinResource) performImport(ctx context.Context, data *UserImportJellyfinModel) error {
	payload := map[string]any{}
	if !data.JellyfinUserIDs.IsNull() && !data.JellyfinUserIDs.IsUnknown() {
		var jellyfinUserIDs []string
		diags := data.JellyfinUserIDs.ElementsAs(ctx, &jellyfinUserIDs, false)
		if diags.HasError() {
			return fmt.Errorf("failed to parse jellyfin_user_ids")
		}
		if len(jellyfinUserIDs) > 0 {
			payload["jellyfinUserIds"] = jellyfinUserIDs
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal import request: %w", err)
	}

	res, err := r.client.Request(ctx, "POST", "/api/v1/user/import-from-jellyfin", string(body), nil)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}

	var importedRaw []map[string]any
	if err := json.Unmarshal(res.Body, &importedRaw); err != nil {
		return fmt.Errorf("failed to parse API response: %w", err)
	}

	var imported []ImportedJellyfinUserModel
	for _, u := range importedRaw {
		user := ImportedJellyfinUserModel{}

		switch v := u["id"].(type) {
		case float64:
			user.ID = types.StringValue(fmt.Sprintf("%.0f", v))
		case string:
			user.ID = types.StringValue(v)
		}

		if e, ok := u["email"].(string); ok {
			user.Email = types.StringValue(e)
		}
		if un, ok := u["username"].(string); ok {
			user.Username = types.StringValue(un)
		}
		if j, ok := u["jellyfinUserId"].(string); ok {
			user.JellyfinUserID = types.StringValue(j)
		}

		imported = append(imported, user)
	}

	listVal, diags := buildImportedJellyfinUserListValue(ctx, imported)
	if diags.HasError() {
		return fmt.Errorf("failed to construct imported_users attribute")
	}

	data.ImportedUsers = listVal
	data.ImportedCount = types.Int64Value(int64(len(imported)))
	return nil
}

func importedJellyfinUserAttrType() map[string]attr.Type {
	return map[string]attr.Type{
		"id":               types.StringType,
		"username":         types.StringType,
		"email":            types.StringType,
		"jellyfin_user_id": types.StringType,
	}
}

func buildImportedJellyfinUserListValue(_ context.Context, imported []ImportedJellyfinUserModel) (types.List, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: importedJellyfinUserAttrType()}
	elems := make([]attr.Value, 0, len(imported))
	for _, item := range imported {
		obj, diags := types.ObjectValue(importedJellyfinUserAttrType(), map[string]attr.Value{
			"id":               item.ID,
			"username":         item.Username,
			"email":            item.Email,
			"jellyfin_user_id": item.JellyfinUserID,
		})
		if diags.HasError() {
			return types.ListNull(elemType), diags
		}
		elems = append(elems, obj)
	}
	return types.ListValue(elemType, elems)
}
