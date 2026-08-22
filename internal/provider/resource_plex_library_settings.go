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

var _ resource.Resource = &PlexLibrarySettingsResource{}
var _ resource.ResourceWithImportState = &PlexLibrarySettingsResource{}

type PlexLibrarySettingsResource struct {
	client *APIClient
}

type PlexLibraryModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

type PlexLibrarySettingsModel struct {
	ID               types.String `tfsdk:"id"`
	SyncOnRead       types.Bool   `tfsdk:"sync_on_read"`
	EnabledLibraries types.Set    `tfsdk:"enabled_libraries"`
	Libraries        types.List   `tfsdk:"libraries"`
}

func NewPlexLibrarySettingsResource() resource.Resource {
	return &PlexLibrarySettingsResource{}
}

func (r *PlexLibrarySettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plex_library_settings"
}

func (r *PlexLibrarySettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Plex library enable/disable settings via `/api/v1/settings/plex/library`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"sync_on_read": schema.BoolAttribute{
				MarkdownDescription: "Whether to pass `sync=true` when reading libraries from Plex.",
				Optional:            true,
				Computed:            true,
			},
			"enabled_libraries": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Set of Plex library IDs to enable in Seerr. Libraries omitted will be disabled.",
				Required:            true,
			},
			"libraries": schema.ListNestedAttribute{
				MarkdownDescription: "List of available Plex libraries discovered on the server.",
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

func (r *PlexLibrarySettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PlexLibrarySettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PlexLibrarySettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.SyncOnRead.IsUnknown() || data.SyncOnRead.IsNull() {
		data.SyncOnRead = types.BoolValue(false)
	}

	if err := r.updatePlexLibraries(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}

	data.ID = types.StringValue("plex_library_settings")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PlexLibrarySettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PlexLibrarySettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readPlexLibraries(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	data.ID = types.StringValue("plex_library_settings")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PlexLibrarySettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PlexLibrarySettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.SyncOnRead.IsUnknown() || data.SyncOnRead.IsNull() {
		data.SyncOnRead = types.BoolValue(false)
	}

	if err := r.updatePlexLibraries(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}

	data.ID = types.StringValue("plex_library_settings")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PlexLibrarySettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Removes resource from state.
}

func (r *PlexLibrarySettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "plex_library_settings")...)
}

func (r *PlexLibrarySettingsResource) updatePlexLibraries(ctx context.Context, data *PlexLibrarySettingsModel) error {
	var enabledList []string
	if !data.EnabledLibraries.IsNull() && !data.EnabledLibraries.IsUnknown() {
		diags := data.EnabledLibraries.ElementsAs(ctx, &enabledList, false)
		if diags.HasError() {
			return fmt.Errorf("failed to extract enabled_libraries")
		}
	}

	apiPath := "/api/v1/settings/plex/library"
	if len(enabledList) > 0 {
		enableQuery := ""
		for i, id := range enabledList {
			if i > 0 {
				enableQuery += ","
			}
			enableQuery += id
		}
		apiPath = fmt.Sprintf("/api/v1/settings/plex/library?enable=%s", enableQuery)
	}
	res, err := r.client.Request(ctx, "GET", apiPath, "", nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}

	return r.parsePlexLibraryResponse(ctx, res.Body, data)
}

func (r *PlexLibrarySettingsResource) readPlexLibraries(ctx context.Context, data *PlexLibrarySettingsModel) error {
	apiPath := "/api/v1/settings/plex/library"
	if data.SyncOnRead.ValueBool() {
		apiPath += "?sync=true"
	}

	res, err := r.client.Request(ctx, "GET", apiPath, "", nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}

	return r.parsePlexLibraryResponse(ctx, res.Body, data)
}

func (r *PlexLibrarySettingsResource) parsePlexLibraryResponse(ctx context.Context, body []byte, data *PlexLibrarySettingsModel) error {
	var rawLibs []struct {
		ID      any    `json:"id"`
		Name    string `json:"name"`
		Enabled bool   `json:"enabled"`
	}

	if err := json.Unmarshal(body, &rawLibs); err != nil {
		return fmt.Errorf("failed to decode plex library response: %w", err)
	}

	libModels := make([]PlexLibraryModel, 0, len(rawLibs))
	enabledIDs := make([]string, 0)

	for _, l := range rawLibs {
		idStr := fmt.Sprintf("%v", l.ID)
		libModels = append(libModels, PlexLibraryModel{
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
