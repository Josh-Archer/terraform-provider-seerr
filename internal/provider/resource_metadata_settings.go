package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &MetadataSettingsResource{}
var _ resource.ResourceWithImportState = &MetadataSettingsResource{}

type MetadataSettingsResource struct {
	client *APIClient
}

type MetadataSettingsModel struct {
	ID       types.String `tfsdk:"id"`
	TV       types.String `tfsdk:"tv"`
	Anime    types.String `tfsdk:"anime"`
	Response types.String `tfsdk:"response_json"`
}

func NewMetadataSettingsResource() resource.Resource { return &MetadataSettingsResource{} }

func (r *MetadataSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metadata_settings"
}

func metadataProviderValidators() []validator.String {
	return []validator.String{stringvalidator.OneOf("tmdb", "tvdb")}
}

func metadataSettingsResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manage Seerr metadata provider settings via `/api/v1/settings/metadata`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tv": schema.StringAttribute{
				MarkdownDescription: "Metadata provider for TV metadata. Valid values are `tmdb` and `tvdb`.",
				Optional:            true,
				Computed:            true,
				Validators:          metadataProviderValidators(),
			},
			"anime": schema.StringAttribute{
				MarkdownDescription: "Metadata provider for anime metadata. Valid values are `tmdb` and `tvdb`.",
				Optional:            true,
				Computed:            true,
				Validators:          metadataProviderValidators(),
			},
			"response_json": schema.StringAttribute{
				MarkdownDescription: "Raw JSON response body from the latest read.",
				Computed:            true,
			},
		},
	}
}

func (r *MetadataSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = metadataSettingsResourceSchema()
}

func (r *MetadataSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MetadataSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MetadataSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.applyMetadataSettings(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MetadataSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MetadataSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.readMetadataSettings(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MetadataSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MetadataSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.applyMetadataSettings(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MetadataSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *MetadataSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *MetadataSettingsResource) applyMetadataSettings(ctx context.Context, data *MetadataSettingsModel) error {
	payload := map[string]any{}
	if !data.TV.IsNull() && !data.TV.IsUnknown() {
		payload["tv"] = data.TV.ValueString()
	}
	if !data.Anime.IsNull() && !data.Anime.IsUnknown() {
		payload["anime"] = data.Anime.ValueString()
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	res, err := r.client.Request(ctx, "PUT", "/api/v1/settings/metadata", string(body), nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}
	return r.readMetadataSettings(ctx, data)
}

func (r *MetadataSettingsResource) readMetadataSettings(ctx context.Context, data *MetadataSettingsModel) error {
	res, err := r.client.Request(ctx, "GET", "/api/v1/settings/metadata", "", nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}
	var decoded map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err != nil {
		return err
	}
	applyMetadataSettingsMap(data, decoded)
	data.ID = types.StringValue("metadata")
	data.Response = types.StringValue(string(res.Body))
	return nil
}

func applyMetadataSettingsMap(data *MetadataSettingsModel, decoded map[string]any) {
	data.TV = types.StringNull()
	data.Anime = types.StringNull()
	if v, ok := decoded["tv"].(string); ok {
		data.TV = types.StringValue(v)
	}
	if v, ok := decoded["anime"].(string); ok {
		data.Anime = types.StringValue(v)
	}
}
