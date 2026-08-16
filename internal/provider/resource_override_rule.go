package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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
	ID               types.String `tfsdk:"id"`
	Users            types.String `tfsdk:"users"`
	Genre            types.String `tfsdk:"genre"`
	Genres           types.List   `tfsdk:"genres"`
	Language         types.String `tfsdk:"language"`
	Languages        types.List   `tfsdk:"languages"`
	OriginalLanguage types.String `tfsdk:"original_language"`
	Keywords         types.String `tfsdk:"keywords"`
	ProfileID        types.Int64  `tfsdk:"profile_id"`
	RootFolder       types.String `tfsdk:"root_folder"`
	Tags             types.String `tfsdk:"tags"`
	TagIDs           types.List   `tfsdk:"tag_ids"`
	Roles            types.String `tfsdk:"roles"`
	UserRoles        types.List   `tfsdk:"user_roles"`
	RadarrServiceID  types.Int64  `tfsdk:"radarr_service_id"`
	SonarrServiceID  types.Int64  `tfsdk:"sonarr_service_id"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
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
		"genres":            rschema.ListAttribute{Optional: true, Computed: true, ElementType: types.Int64Type},
		"language":          rschema.StringAttribute{Optional: true, Computed: true},
		"languages":         rschema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
		"original_language": rschema.StringAttribute{Optional: true, Computed: true},
		"keywords":          rschema.StringAttribute{Optional: true, Computed: true},
		"profile_id":        rschema.Int64Attribute{Optional: true, Computed: true},
		"root_folder":       rschema.StringAttribute{Optional: true, Computed: true},
		"tags":              rschema.StringAttribute{Optional: true, Computed: true},
		"tag_ids":           rschema.ListAttribute{Optional: true, Computed: true, ElementType: types.Int64Type},
		"roles":             rschema.StringAttribute{Optional: true, Computed: true},
		"user_roles":        rschema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
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
		"genres":            schema.ListAttribute{Computed: true, ElementType: types.Int64Type},
		"language":          schema.StringAttribute{Computed: true},
		"languages":         schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"original_language": schema.StringAttribute{Computed: true},
		"keywords":          schema.StringAttribute{Computed: true},
		"profile_id":        schema.Int64Attribute{Computed: true},
		"root_folder":       schema.StringAttribute{Computed: true},
		"tags":              schema.StringAttribute{Computed: true},
		"tag_ids":           schema.ListAttribute{Computed: true, ElementType: types.Int64Type},
		"roles":             schema.StringAttribute{Computed: true},
		"user_roles":        schema.ListAttribute{Computed: true, ElementType: types.StringType},
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

	// If genres list is set, sync with genre string
	if !data.Genres.IsNull() && !data.Genres.IsUnknown() {
		var gList []int64
		_ = data.Genres.ElementsAs(context.Background(), &gList, false)
		if len(gList) > 0 {
			var strList []string
			for _, id := range gList {
				strList = append(strList, strconv.FormatInt(id, 10))
			}
			if data.Genre.IsNull() || data.Genre.IsUnknown() {
				payload["genre"] = strings.Join(strList, ",")
			}
			payload["genres"] = gList
		}
	}

	// If tag_ids list is set, sync with tags string
	if !data.TagIDs.IsNull() && !data.TagIDs.IsUnknown() {
		var tList []int64
		_ = data.TagIDs.ElementsAs(context.Background(), &tList, false)
		if len(tList) > 0 {
			var strList []string
			for _, id := range tList {
				strList = append(strList, strconv.FormatInt(id, 10))
			}
			if data.Tags.IsNull() || data.Tags.IsUnknown() {
				payload["tags"] = strings.Join(strList, ",")
			}
			payload["tagIds"] = tList
		}
	}

	// If user_roles list or roles string is set
	if !data.UserRoles.IsNull() && !data.UserRoles.IsUnknown() {
		var rList []string
		_ = data.UserRoles.ElementsAs(context.Background(), &rList, false)
		if len(rList) > 0 {
			payload["userRoles"] = strings.Join(rList, ",")
			if data.Roles.IsNull() || data.Roles.IsUnknown() {
				payload["roles"] = strings.Join(rList, ",")
			}
		}
	} else if !data.Roles.IsNull() && !data.Roles.IsUnknown() {
		setOptionalString(payload, "userRoles", data.Roles)
		setOptionalString(payload, "roles", data.Roles)
	}

	// If languages list is set
	if !data.Languages.IsNull() && !data.Languages.IsUnknown() {
		var lList []string
		_ = data.Languages.ElementsAs(context.Background(), &lList, false)
		if len(lList) > 0 {
			if data.Language.IsNull() || data.Language.IsUnknown() {
				payload["language"] = strings.Join(lList, ",")
			}
			payload["languages"] = lList
		}
	}

	// If original_language is set
	if !data.OriginalLanguage.IsNull() && !data.OriginalLanguage.IsUnknown() {
		setOptionalString(payload, "originalLanguage", data.OriginalLanguage)
		setOptionalString(payload, "original_language", data.OriginalLanguage)
	}

	return payload
}

func applyOverrideRuleBody(data *OverrideRuleModel, body []byte) error {
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return err
	}
	return applyOverrideRuleMap(data, decoded)
}

func parseIntListFromStringOrSlice(v any) types.List {
	if v == nil {
		return types.ListNull(types.Int64Type)
	}
	var ints []int64
	switch raw := v.(type) {
	case string:
		if strings.TrimSpace(raw) == "" {
			return types.ListNull(types.Int64Type)
		}
		parts := strings.Split(raw, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if n, err := strconv.ParseInt(p, 10, 64); err == nil {
				ints = append(ints, n)
			}
		}
	case []any:
		for _, item := range raw {
			if n, ok := int64ValueFromAny(item); ok {
				ints = append(ints, n)
			}
		}
	case []int64:
		ints = raw
	}

	if len(ints) == 0 {
		return types.ListNull(types.Int64Type)
	}
	elems := make([]types.Int64, len(ints))
	for i, val := range ints {
		elems[i] = types.Int64Value(val)
	}
	res, _ := types.ListValueFrom(context.Background(), types.Int64Type, elems)
	return res
}

func parseStringListFromStringOrSlice(v any) types.List {
	if v == nil {
		return types.ListNull(types.StringType)
	}
	var stringsList []string
	switch raw := v.(type) {
	case string:
		if strings.TrimSpace(raw) == "" {
			return types.ListNull(types.StringType)
		}
		parts := strings.Split(raw, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				stringsList = append(stringsList, p)
			}
		}
	case []any:
		for _, item := range raw {
			if s, ok := stringValueFromAny(item); ok && s != "" {
				stringsList = append(stringsList, s)
			}
		}
	case []string:
		stringsList = raw
	}

	if len(stringsList) == 0 {
		return types.ListNull(types.StringType)
	}
	elems := make([]types.String, len(stringsList))
	for i, val := range stringsList {
		elems[i] = types.StringValue(val)
	}
	res, _ := types.ListValueFrom(context.Background(), types.StringType, elems)
	return res
}

func applyOverrideRuleMap(data *OverrideRuleModel, decoded map[string]any) error {
	data.Users = types.StringNull()
	data.Genre = types.StringNull()
	data.Genres = types.ListNull(types.Int64Type)
	data.Language = types.StringNull()
	data.Languages = types.ListNull(types.StringType)
	data.OriginalLanguage = types.StringNull()
	data.Keywords = types.StringNull()
	data.ProfileID = types.Int64Null()
	data.RootFolder = types.StringNull()
	data.Tags = types.StringNull()
	data.TagIDs = types.ListNull(types.Int64Type)
	data.Roles = types.StringNull()
	data.UserRoles = types.ListNull(types.StringType)
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
		data.Genres = parseIntListFromStringOrSlice(v)
	} else if gRaw := decoded["genres"]; gRaw != nil {
		data.Genres = parseIntListFromStringOrSlice(gRaw)
		if !data.Genres.IsNull() {
			var gList []int64
			_ = data.Genres.ElementsAs(context.Background(), &gList, false)
			var strList []string
			for _, id := range gList {
				strList = append(strList, strconv.FormatInt(id, 10))
			}
			data.Genre = types.StringValue(strings.Join(strList, ","))
		}
	}
	if v, ok := stringValueFromAny(decoded["language"]); ok {
		data.Language = types.StringValue(v)
		data.Languages = parseStringListFromStringOrSlice(v)
	} else if lRaw := decoded["languages"]; lRaw != nil {
		data.Languages = parseStringListFromStringOrSlice(lRaw)
		if !data.Languages.IsNull() {
			var lList []string
			_ = data.Languages.ElementsAs(context.Background(), &lList, false)
			data.Language = types.StringValue(strings.Join(lList, ","))
		}
	}
	if v, ok := stringValueFromAny(decoded["originalLanguage"]); ok {
		data.OriginalLanguage = types.StringValue(v)
	} else if v, ok := stringValueFromAny(decoded["original_language"]); ok {
		data.OriginalLanguage = types.StringValue(v)
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
		data.TagIDs = parseIntListFromStringOrSlice(v)
	} else if tRaw := decoded["tagIds"]; tRaw != nil {
		data.TagIDs = parseIntListFromStringOrSlice(tRaw)
		if !data.TagIDs.IsNull() {
			var tList []int64
			_ = data.TagIDs.ElementsAs(context.Background(), &tList, false)
			var strList []string
			for _, id := range tList {
				strList = append(strList, strconv.FormatInt(id, 10))
			}
			data.Tags = types.StringValue(strings.Join(strList, ","))
		}
	}
	if v, ok := stringValueFromAny(decoded["userRoles"]); ok {
		data.Roles = types.StringValue(v)
		data.UserRoles = parseStringListFromStringOrSlice(v)
	} else if v, ok := stringValueFromAny(decoded["roles"]); ok {
		data.Roles = types.StringValue(v)
		data.UserRoles = parseStringListFromStringOrSlice(v)
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
