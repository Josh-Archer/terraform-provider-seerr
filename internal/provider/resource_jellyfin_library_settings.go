package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &JellyfinLibrarySettingsResource{}
var _ resource.ResourceWithImportState = &JellyfinLibrarySettingsResource{}

type JellyfinLibrarySettingsResource struct {
	client *APIClient
}

type JellyfinLibraryModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

type JellyfinLibrarySettingsModel struct {
	ID               types.String `tfsdk:"id"`
	SyncOnRead       types.Bool   `tfsdk:"sync_on_read"`
	EnabledLibraries types.Set    `tfsdk:"enabled_libraries"`
	Libraries        types.List   `tfsdk:"libraries"`
}

func NewJellyfinLibrarySettingsResource() resource.Resource {
	return &JellyfinLibrarySettingsResource{}
}

func (r *JellyfinLibrarySettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jellyfin_library_settings"
}

func (r *JellyfinLibrarySettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Jellyfin library enable/disable settings via `/api/v1/settings/jellyfin/library`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"sync_on_read": schema.BoolAttribute{
				MarkdownDescription: "Whether to pass `sync=true` when reading libraries from Jellyfin.",
				Optional:            true,
				Computed:            true,
			},
			"enabled_libraries": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Set of Jellyfin library IDs to enable in Seerr. Libraries omitted will be disabled.",
				Required:            true,
			},
			"libraries": schema.ListNestedAttribute{
				MarkdownDescription: "List of available Jellyfin libraries discovered on the server.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"enabled": schema.BoolAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (r *JellyfinLibrarySettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *JellyfinLibrarySettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data JellyfinLibrarySettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.SyncOnRead.IsUnknown() || data.SyncOnRead.IsNull() {
		data.SyncOnRead = types.BoolValue(false)
	}

	if err := r.updateJellyfinLibraries(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}

	data.ID = types.StringValue("jellyfin_library_settings")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JellyfinLibrarySettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data JellyfinLibrarySettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readJellyfinLibraries(ctx, &data); err != nil {
		if IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	data.ID = types.StringValue("jellyfin_library_settings")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JellyfinLibrarySettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data JellyfinLibrarySettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.SyncOnRead.IsUnknown() || data.SyncOnRead.IsNull() {
		data.SyncOnRead = types.BoolValue(false)
	}

	if err := r.updateJellyfinLibraries(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}

	data.ID = types.StringValue("jellyfin_library_settings")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JellyfinLibrarySettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Removes resource from state.
}

func (r *JellyfinLibrarySettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "jellyfin_library_settings")...)
}

func (r *JellyfinLibrarySettingsResource) updateJellyfinLibraries(ctx context.Context, data *JellyfinLibrarySettingsModel) error {
	var enabledList []string
	if !data.EnabledLibraries.IsNull() && !data.EnabledLibraries.IsUnknown() {
		diags := data.EnabledLibraries.ElementsAs(ctx, &enabledList, false)
		if diags.HasError() {
			return fmt.Errorf("failed to extract enabled_libraries")
		}
	}

	apiPath := "/api/v1/settings/jellyfin/library"
	if len(enabledList) > 0 {
		enableQuery := ""
		for i, id := range enabledList {
			if i > 0 {
				enableQuery += ","
			}
			enableQuery += id
		}
		apiPath = fmt.Sprintf("/api/v1/settings/jellyfin/library?enable=%s", enableQuery)
	}
	res, err := r.client.Request(ctx, "GET", apiPath, "", nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}

	return r.parseJellyfinLibraryResponse(ctx, res.Body, data)
}

func (r *JellyfinLibrarySettingsResource) readJellyfinLibraries(ctx context.Context, data *JellyfinLibrarySettingsModel) error {
	apiPath := "/api/v1/settings/jellyfin/library"
	if data.SyncOnRead.ValueBool() {
		apiPath += "?sync=true"
	}

	res, err := r.client.Request(ctx, "GET", apiPath, "", nil)
	if err != nil {
		return err
	}
	if res.StatusCode == 404 {
		return ErrNotFound
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}

	return r.parseJellyfinLibraryResponse(ctx, res.Body, data)
}

func (r *JellyfinLibrarySettingsResource) parseJellyfinLibraryResponse(ctx context.Context, body []byte, data *JellyfinLibrarySettingsModel) error {
	var rawLibs []struct {
		ID      any    `json:"id"`
		Name    string `json:"name"`
		Enabled bool   `json:"enabled"`
	}

	if err := json.Unmarshal(body, &rawLibs); err != nil {
		return fmt.Errorf("failed to decode jellyfin library response: %w", err)
	}

	libModels := make([]JellyfinLibraryModel, 0, len(rawLibs))
	enabledIDs := make([]string, 0)

	for _, l := range rawLibs {
		idStr := fmt.Sprintf("%v", l.ID)
		libModels = append(libModels, JellyfinLibraryModel{
			ID:      types.StringValue(idStr),
			Name:    types.StringValue(l.Name),
			Enabled: types.BoolValue(l.Enabled),
		})
		if l.Enabled {
			enabledIDs = append(enabledIDs, idStr)
		}
	}

	elemType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":      types.StringType,
			"name":    types.StringType,
			"enabled": types.BoolType,
		},
	}

	listVal, diags := types.ListValueFrom(ctx, elemType, libModels)
	if diags.HasError() {
		return fmt.Errorf("failed to build libraries list value")
	}
	data.Libraries = listVal

	setVal, diags := types.SetValueFrom(ctx, types.StringType, enabledIDs)
	if diags.HasError() {
		return fmt.Errorf("failed to build enabled_libraries set value")
	}
	data.EnabledLibraries = setVal

	return nil
}
