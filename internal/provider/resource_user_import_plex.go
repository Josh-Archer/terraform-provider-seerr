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

var _ resource.Resource = &UserImportPlexResource{}
var _ resource.ResourceWithImportState = &UserImportPlexResource{}

type UserImportPlexResource struct {
	client *APIClient
}

type ImportedPlexUserModel struct {
	ID       types.String `tfsdk:"id"`
	Username types.String `tfsdk:"username"`
	Email    types.String `tfsdk:"email"`
	PlexID   types.String `tfsdk:"plex_id"`
}

type UserImportPlexModel struct {
	ID            types.String `tfsdk:"id"`
	PlexIDs       types.Set    `tfsdk:"plex_ids"`
	ImportedUsers types.List   `tfsdk:"imported_users"`
	ImportedCount types.Int64  `tfsdk:"imported_count"`
	Triggers      types.Map    `tfsdk:"triggers"`
}

func NewUserImportPlexResource() resource.Resource {
	return &UserImportPlexResource{}
}

func (r *UserImportPlexResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_import_plex"
}

func (r *UserImportPlexResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Import users from a connected Plex media server into Seerr via `/api/v1/user/import-from-plex`.\n\n" +
			"This resource facilitates the **bootstrap sequence**: Media Server Setup → `seerr_user_import_plex` → `seerr_permission_set` / `seerr_user_quota`.\n\n" +
			"If `plex_ids` is specified, only the matching Plex users are imported. If omitted or empty, all available Plex users on the server are imported.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Resource identifier (defaults to `plex_import`).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"plex_ids": schema.SetAttribute{
				MarkdownDescription: "Optional set of specific Plex user IDs to import. If omitted, all available Plex users are imported.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"imported_users": schema.ListNestedAttribute{
				MarkdownDescription: "List of users imported from Plex.",
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
						"plex_id": schema.StringAttribute{
							MarkdownDescription: "The user's Plex ID.",
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

func (r *UserImportPlexResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserImportPlexResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserImportPlexModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.performImport(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Plex User Import Failed", err.Error())
		return
	}

	data.ID = types.StringValue("plex_import")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserImportPlexResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserImportPlexModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch all users to refresh imported Plex users list
	results, err := fetchAllPaginatedResults(ctx, r.client, "/api/v1/user", defaultPaginationPageSize)
	if err != nil {
		if IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	var imported []ImportedPlexUserModel
	for _, u := range results {
		plexID, _ := u["plexId"].(string)
		if plexID == "" {
			// Skip non-Plex users
			continue
		}

		user := ImportedPlexUserModel{
			PlexID: types.StringValue(plexID),
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

	listVal, diags := buildImportedPlexUserListValue(ctx, imported)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ImportedUsers = listVal
	data.ImportedCount = types.Int64Value(int64(len(imported)))
	if data.ID.IsNull() || data.ID.ValueString() == "" {
		data.ID = types.StringValue("plex_import")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserImportPlexResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserImportPlexModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.performImport(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Plex User Import Update Failed", err.Error())
		return
	}

	data.ID = types.StringValue("plex_import")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserImportPlexResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// One-shot import action: deletion removes from Terraform state only.
}

func (r *UserImportPlexResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *UserImportPlexResource) performImport(ctx context.Context, data *UserImportPlexModel) error {
	payload := map[string]any{}
	if !data.PlexIDs.IsNull() && !data.PlexIDs.IsUnknown() {
		var plexIDs []string
		diags := data.PlexIDs.ElementsAs(ctx, &plexIDs, false)
		if diags.HasError() {
			return fmt.Errorf("failed to parse plex_ids")
		}
		if len(plexIDs) > 0 {
			payload["plexIds"] = plexIDs
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal import request: %w", err)
	}

	res, err := r.client.Request(ctx, "POST", "/api/v1/user/import-from-plex", string(body), nil)
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

	var imported []ImportedPlexUserModel
	for _, u := range importedRaw {
		user := ImportedPlexUserModel{}

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
		if p, ok := u["plexId"].(string); ok {
			user.PlexID = types.StringValue(p)
		}

		imported = append(imported, user)
	}

	listVal, diags := buildImportedPlexUserListValue(ctx, imported)
	if diags.HasError() {
		return fmt.Errorf("failed to construct imported_users attribute")
	}

	data.ImportedUsers = listVal
	data.ImportedCount = types.Int64Value(int64(len(imported)))
	return nil
}

func importedPlexUserAttrType() map[string]attr.Type {
	return map[string]attr.Type{
		"id":       types.StringType,
		"username": types.StringType,
		"email":    types.StringType,
		"plex_id":  types.StringType,
	}
}

func buildImportedPlexUserListValue(_ context.Context, imported []ImportedPlexUserModel) (types.List, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: importedPlexUserAttrType()}
	elems := make([]attr.Value, 0, len(imported))
	for _, item := range imported {
		obj, diags := types.ObjectValue(importedPlexUserAttrType(), map[string]attr.Value{
			"id":       item.ID,
			"username": item.Username,
			"email":    item.Email,
			"plex_id":  item.PlexID,
		})
		if diags.HasError() {
			return types.ListNull(elemType), diags
		}
		elems = append(elems, obj)
	}
	return types.ListValue(elemType, elems)
}
