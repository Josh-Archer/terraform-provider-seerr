package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &OverrideRuleResource{}
var _ resource.ResourceWithImportState = &OverrideRuleResource{}

type OverrideRuleResource struct {
	client *APIClient
}

type OverrideRuleModel struct {
	ID              types.String `tfsdk:"id"`
	Users           types.String `tfsdk:"users"`
	Genre           types.String `tfsdk:"genre"`
	Language        types.String `tfsdk:"language"`
	Keywords        types.String `tfsdk:"keywords"`
	ProfileID       types.Int64  `tfsdk:"profile_id"`
	RootFolder      types.String `tfsdk:"root_folder"`
	Tags            types.String `tfsdk:"tags"`
	RadarrServiceID types.Int64  `tfsdk:"radarr_service_id"`
	SonarrServiceID types.Int64  `tfsdk:"sonarr_service_id"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}

func NewOverrideRuleResource() resource.Resource { return &OverrideRuleResource{} }

func (r *OverrideRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_override_rule"
}

func overrideRuleResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"users":             rschema.StringAttribute{Optional: true, Computed: true},
		"genre":             rschema.StringAttribute{Optional: true, Computed: true},
		"language":          rschema.StringAttribute{Optional: true, Computed: true},
		"keywords":          rschema.StringAttribute{Optional: true, Computed: true},
		"profile_id":        rschema.Int64Attribute{Optional: true, Computed: true},
		"root_folder":       rschema.StringAttribute{Optional: true, Computed: true},
		"tags":              rschema.StringAttribute{Optional: true, Computed: true},
		"radarr_service_id": rschema.Int64Attribute{Optional: true, Computed: true},
		"sonarr_service_id": rschema.Int64Attribute{Optional: true, Computed: true},
		"created_at":        rschema.StringAttribute{Computed: true},
		"updated_at":        rschema.StringAttribute{Computed: true},
	}
}

func overrideRuleDataSourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id":                schema.StringAttribute{Required: true},
		"users":             schema.StringAttribute{Computed: true},
		"genre":             schema.StringAttribute{Computed: true},
		"language":          schema.StringAttribute{Computed: true},
		"keywords":          schema.StringAttribute{Computed: true},
		"profile_id":        schema.Int64Attribute{Computed: true},
		"root_folder":       schema.StringAttribute{Computed: true},
		"tags":              schema.StringAttribute{Computed: true},
		"radarr_service_id": schema.Int64Attribute{Computed: true},
		"sonarr_service_id": schema.Int64Attribute{Computed: true},
		"created_at":        schema.StringAttribute{Computed: true},
		"updated_at":        schema.StringAttribute{Computed: true},
	}
}

func (r *OverrideRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: "Manage Seerr override rules via `/api/v1/overrideRule`.",
		Attributes:          overrideRuleResourceSchema(),
	}
}

func (r *OverrideRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OverrideRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OverrideRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(buildOverrideRulePayload(&data))
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}

	res, err := r.client.Request(ctx, "POST", "/api/v1/overrideRule", string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	if err := applyOverrideRuleBody(&data, res.Body); err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OverrideRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OverrideRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, err := r.fetchOverrideRule(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data = *found
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OverrideRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OverrideRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(buildOverrideRulePayload(&data))
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}

	res, err := r.client.Request(ctx, "PUT", "/api/v1/overrideRule/"+data.ID.ValueString(), string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Update Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	if err := applyOverrideRuleBody(&data, res.Body); err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OverrideRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OverrideRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.Request(ctx, "DELETE", "/api/v1/overrideRule/"+data.ID.ValueString(), "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Delete Failed", err.Error())
		return
	}
	if res.StatusCode != 404 && !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Delete Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
	}
}

func (r *OverrideRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildOverrideRulePayload(data *OverrideRuleModel) map[string]any {
	payload := map[string]any{}
	setOptionalString(payload, "users", data.Users)
	setOptionalString(payload, "genre", data.Genre)
	setOptionalString(payload, "language", data.Language)
	setOptionalString(payload, "keywords", data.Keywords)
	setOptionalInt64(payload, "profileId", data.ProfileID)
	setOptionalString(payload, "rootFolder", data.RootFolder)
	setOptionalString(payload, "tags", data.Tags)
	setOptionalInt64(payload, "radarrServiceId", data.RadarrServiceID)
	setOptionalInt64(payload, "sonarrServiceId", data.SonarrServiceID)
	return payload
}

func applyOverrideRuleBody(data *OverrideRuleModel, body []byte) error {
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return err
	}
	return applyOverrideRuleMap(data, decoded)
}

func applyOverrideRuleMap(data *OverrideRuleModel, decoded map[string]any) error {
	data.Users = types.StringNull()
	data.Genre = types.StringNull()
	data.Language = types.StringNull()
	data.Keywords = types.StringNull()
	data.ProfileID = types.Int64Null()
	data.RootFolder = types.StringNull()
	data.Tags = types.StringNull()
	data.RadarrServiceID = types.Int64Null()
	data.SonarrServiceID = types.Int64Null()
	data.CreatedAt = types.StringNull()
	data.UpdatedAt = types.StringNull()

	switch v := decoded["id"].(type) {
	case float64:
		data.ID = types.StringValue(strconv.FormatInt(int64(v), 10))
	case string:
		data.ID = types.StringValue(v)
	default:
		return fmt.Errorf("override rule id missing from response")
	}

	if v, ok := stringValueFromAny(decoded["users"]); ok {
		data.Users = types.StringValue(v)
	}
	if v, ok := stringValueFromAny(decoded["genre"]); ok {
		data.Genre = types.StringValue(v)
	}
	if v, ok := stringValueFromAny(decoded["language"]); ok {
		data.Language = types.StringValue(v)
	}
	if v, ok := stringValueFromAny(decoded["keywords"]); ok {
		data.Keywords = types.StringValue(v)
	}
	if v, ok := int64ValueFromAny(decoded["profileId"]); ok {
		data.ProfileID = types.Int64Value(v)
	}
	if v, ok := stringValueFromAny(decoded["rootFolder"]); ok {
		data.RootFolder = types.StringValue(v)
	}
	if v, ok := stringValueFromAny(decoded["tags"]); ok {
		data.Tags = types.StringValue(v)
	}
	if v, ok := int64ValueFromAny(decoded["radarrServiceId"]); ok {
		data.RadarrServiceID = types.Int64Value(v)
	}
	if v, ok := int64ValueFromAny(decoded["sonarrServiceId"]); ok {
		data.SonarrServiceID = types.Int64Value(v)
	}
	if v, ok := stringValueFromAny(decoded["createdAt"]); ok {
		data.CreatedAt = types.StringValue(v)
	}
	if v, ok := stringValueFromAny(decoded["updatedAt"]); ok {
		data.UpdatedAt = types.StringValue(v)
	}

	return nil
}

func (r *OverrideRuleResource) fetchOverrideRule(ctx context.Context, id string) (*OverrideRuleModel, error) {
	res, err := r.client.Request(ctx, "GET", "/api/v1/overrideRule", "", nil)
	if err != nil {
		return nil, err
	}
	if !StatusIsOK(res.StatusCode) {
		return nil, fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}

	var decoded []map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err != nil {
		return nil, err
	}

	for _, item := range decoded {
		var currentID string
		switch v := item["id"].(type) {
		case float64:
			currentID = strconv.FormatInt(int64(v), 10)
		case string:
			currentID = v
		}
		if currentID != id {
			continue
		}

		var model OverrideRuleModel
		if err := applyOverrideRuleMap(&model, item); err != nil {
			return nil, err
		}
		return &model, nil
	}

	return nil, nil
}
