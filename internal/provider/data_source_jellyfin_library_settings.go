package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &JellyfinLibrarySettingsDataSource{}

type JellyfinLibrarySettingsDataSource struct {
	client *APIClient
}

func NewJellyfinLibrarySettingsDataSource() datasource.DataSource {
	return &JellyfinLibrarySettingsDataSource{}
}

func (d *JellyfinLibrarySettingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jellyfin_library_settings"
}

func (d *JellyfinLibrarySettingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch current Jellyfin library enable/disable settings via `/api/v1/settings/jellyfin/library`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"sync_on_read": schema.BoolAttribute{
				MarkdownDescription: "Whether to pass `sync=true` when reading libraries from Jellyfin.",
				Optional:            true,
			},
			"enabled_libraries": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Set of Jellyfin library IDs enabled in Seerr.",
				Computed:            true,
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

func (d *JellyfinLibrarySettingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Configure Type", fmt.Sprintf("Expected *APIClient, got %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *JellyfinLibrarySettingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data JellyfinLibrarySettingsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiPath := "/api/v1/settings/jellyfin/library"
	if !data.SyncOnRead.IsNull() && data.SyncOnRead.ValueBool() {
		apiPath += "?sync=true"
	}

	res, err := d.client.Request(ctx, "GET", apiPath, "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	var rawLibs []struct {
		ID      any    `json:"id"`
		Name    string `json:"name"`
		Enabled bool   `json:"enabled"`
	}

	if err := json.Unmarshal(res.Body, &rawLibs); err != nil {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("failed to decode response: %s", err.Error()))
		return
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
		resp.Diagnostics.Append(diags...)
		return
	}
	data.Libraries = listVal

	setVal, diags := types.SetValueFrom(ctx, types.StringType, enabledIDs)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.EnabledLibraries = setVal
	data.ID = types.StringValue("jellyfin_library_settings")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
