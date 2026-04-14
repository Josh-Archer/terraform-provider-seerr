package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	stringvalidator "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &BlocklistResource{}
var _ resource.ResourceWithImportState = &BlocklistResource{}

type BlocklistResource struct {
	client *APIClient
}

type BlocklistModel struct {
	ID              types.String `tfsdk:"id"`
	TMDBID          types.Int64  `tfsdk:"tmdb_id"`
	MediaType       types.String `tfsdk:"media_type"`
	Title           types.String `tfsdk:"title"`
	UserID          types.Int64  `tfsdk:"user_id"`
	BlocklistedTags types.String `tfsdk:"blocklisted_tags"`
	CreatedAt       types.String `tfsdk:"created_at"`
}

func NewBlocklistResource() resource.Resource { return &BlocklistResource{} }

func (r *BlocklistResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blocklist"
}

func blocklistResourceAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"tmdb_id": rschema.Int64Attribute{Required: true},
		"media_type": rschema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				stringvalidator.OneOf("movie", "tv"),
			},
		},
		"title": rschema.StringAttribute{Optional: true, Computed: true},
		"user_id": rschema.Int64Attribute{
			MarkdownDescription: "User ID recorded as the actor who manually blocklisted this media.",
			Required:            true,
		},
		"blocklisted_tags": rschema.StringAttribute{Optional: true, Computed: true},
		"created_at":       rschema.StringAttribute{Computed: true},
	}
}

func blocklistDataSourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id":               schema.StringAttribute{Computed: true},
		"tmdb_id":          schema.Int64Attribute{Required: true},
		"media_type":       schema.StringAttribute{Required: true},
		"title":            schema.StringAttribute{Computed: true},
		"user_id":          schema.Int64Attribute{Computed: true},
		"blocklisted_tags": schema.StringAttribute{Computed: true},
		"created_at":       schema.StringAttribute{Computed: true},
	}
}

func (r *BlocklistResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: "Manage manual Seerr blocklist entries via `/api/v1/blocklist`.",
		Attributes:          blocklistResourceAttributes(),
	}
}

func (r *BlocklistResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BlocklistResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BlocklistModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(buildBlocklistPayload(&data))
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}

	res, err := r.client.Request(ctx, "POST", "/api/v1/blocklist", string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}
	if res.StatusCode != 201 && !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	data.ID = blocklistID(data.MediaType.ValueString(), data.TMDBID.ValueInt64())
	if err := r.refreshBlocklist(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BlocklistResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BlocklistModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.refreshBlocklist(ctx, &data); err != nil {
		if strings.Contains(err.Error(), "status 404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BlocklistResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BlocklistModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Seerr exposes blocklist as create/delete semantics, so "update" re-reads state only.
	if err := r.refreshBlocklist(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BlocklistResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BlocklistModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/api/v1/blocklist/%d?mediaType=%s", data.TMDBID.ValueInt64(), url.QueryEscape(data.MediaType.ValueString()))
	res, err := r.client.Request(ctx, "DELETE", endpoint, "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Delete Failed", err.Error())
		return
	}
	if res.StatusCode != 404 && res.StatusCode != 204 && !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Delete Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
	}
}

func (r *BlocklistResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import id", "Use import format `<media_type>:<tmdb_id>`, for example `movie:438631`.")
		return
	}
	tmdbID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import id", fmt.Sprintf("invalid tmdb id %q: %s", parts[1], err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("media_type"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tmdb_id"), tmdbID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func buildBlocklistPayload(data *BlocklistModel) map[string]any {
	payload := map[string]any{
		"tmdbId":    data.TMDBID.ValueInt64(),
		"mediaType": data.MediaType.ValueString(),
		"user":      data.UserID.ValueInt64(),
	}
	setOptionalString(payload, "title", data.Title)
	setOptionalString(payload, "blocklistedTags", data.BlocklistedTags)
	return payload
}

func applyBlocklistMap(data *BlocklistModel, decoded map[string]any) error {
	data.Title = types.StringNull()
	data.UserID = types.Int64Null()
	data.BlocklistedTags = types.StringNull()
	data.CreatedAt = types.StringNull()

	tmdbID, ok := int64ValueFromAny(decoded["tmdbId"])
	if !ok {
		return fmt.Errorf("tmdbId missing from blocklist response")
	}
	data.TMDBID = types.Int64Value(tmdbID)

	mediaType, ok := stringValueFromAny(decoded["mediaType"])
	if !ok || mediaType == "" {
		return fmt.Errorf("mediaType missing from blocklist response")
	}
	data.MediaType = types.StringValue(mediaType)
	data.ID = blocklistID(mediaType, tmdbID)

	if v, ok := stringValueFromAny(decoded["title"]); ok {
		data.Title = types.StringValue(v)
	}
	if userMap, ok := decoded["user"].(map[string]any); ok {
		if v, ok := int64ValueFromAny(userMap["id"]); ok {
			data.UserID = types.Int64Value(v)
		}
	}
	if v, ok := int64ValueFromAny(decoded["userId"]); ok {
		data.UserID = types.Int64Value(v)
	}
	if v, ok := stringValueFromAny(decoded["blocklistedTags"]); ok {
		data.BlocklistedTags = types.StringValue(v)
	}
	if v, ok := stringValueFromAny(decoded["createdAt"]); ok {
		data.CreatedAt = types.StringValue(v)
	}
	return nil
}

func (r *BlocklistResource) refreshBlocklist(ctx context.Context, data *BlocklistModel) error {
	endpoint := fmt.Sprintf("/api/v1/blocklist/%d?mediaType=%s", data.TMDBID.ValueInt64(), url.QueryEscape(data.MediaType.ValueString()))
	res, err := r.client.Request(ctx, "GET", endpoint, "", nil)
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

	return applyBlocklistMap(data, decoded)
}

func blocklistID(mediaType string, tmdbID int64) types.String {
	return types.StringValue(mediaType + ":" + strconv.FormatInt(tmdbID, 10))
}
