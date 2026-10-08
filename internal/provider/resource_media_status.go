package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	stringvalidator "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &MediaStatusResource{}
var _ resource.ResourceWithImportState = &MediaStatusResource{}

type MediaStatusResource struct{ client *APIClient }

type MediaStatusModel struct {
	ID       types.String `tfsdk:"id"`
	MediaID  types.String `tfsdk:"media_id"`
	Status   types.String `tfsdk:"status"`
	Is4K     types.Bool   `tfsdk:"is4k"`
	Triggers types.Map    `tfsdk:"triggers"`
}

var mediaIDPattern = regexp.MustCompile(`^[1-9][0-9]*$`)

func NewMediaStatusResource() resource.Resource { return &MediaStatusResource{} }

func (r *MediaStatusResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_media_status"
}

func (r *MediaStatusResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Set a Seerr media item's availability status through `POST /api/v1/media/{mediaId}/{status}`. Destroying this action resource does not change Seerr media status.",
		Attributes: map[string]schema.Attribute{
			"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"media_id": schema.StringAttribute{MarkdownDescription: "The positive Seerr internal media ID. Changing it replaces this action.", Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, Validators: []validator.String{stringvalidator.RegexMatches(mediaIDPattern, "must be a positive Seerr media ID")}},
			"status":   schema.StringAttribute{MarkdownDescription: "The status to set. `partial` is only supported for TV series. The current Seerr route does not implement the OpenAPI `deleted` enum, so it is excluded.", Required: true, Validators: []validator.String{stringvalidator.OneOf("available", "partial", "processing", "pending", "unknown")}},
			"is4k":     schema.BoolAttribute{MarkdownDescription: "When true, updates the 4K status field. Otherwise updates the regular status field.", Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"triggers": schema.MapAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "Changing a value re-applies the requested status."},
		},
	}
}

func (r *MediaStatusResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MediaStatusResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MediaStatusModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Set Media Status Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MediaStatusResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MediaStatusModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Set Media Status Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MediaStatusResource) apply(ctx context.Context, data *MediaStatusModel) error {
	body, err := json.Marshal(map[string]any{"is4k": data.Is4K.ValueBool()})
	if err != nil {
		return err
	}
	apiPath := fmt.Sprintf("/api/v1/media/%s/%s", data.MediaID.ValueString(), data.Status.ValueString())
	res, err := r.client.Request(ctx, "POST", apiPath, string(body), nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}
	data.ID = data.MediaID
	return nil
}

func (r *MediaStatusResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MediaStatusModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	res, err := r.client.Request(ctx, "GET", "/api/v1/media/"+data.MediaID.ValueString(), "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Media Status Failed", err.Error())
		return
	}
	if res.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Media Status Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}
	var media map[string]any
	if err := json.Unmarshal(res.Body, &media); err != nil {
		resp.Diagnostics.AddError("Read Media Status Failed", "failed to parse API response: "+err.Error())
		return
	}
	key := "status"
	if data.Is4K.ValueBool() {
		key = "status4k"
	}
	status, ok := int64ValueFromAny(media[key])
	if !ok {
		resp.Diagnostics.AddError("Read Media Status Failed", fmt.Sprintf("API response did not contain numeric %s", key))
		return
	}
	value, ok := mediaStatusName(status)
	if !ok {
		resp.Diagnostics.AddError("Read Media Status Failed", fmt.Sprintf("API returned unsupported media status %d", status))
		return
	}
	data.Status = types.StringValue(value)
	data.ID = data.MediaID
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func mediaStatusName(status int64) (string, bool) {
	switch status {
	case 1:
		return "unknown", true
	case 2:
		return "pending", true
	case 3:
		return "processing", true
	case 4:
		return "partial", true
	case 5:
		return "available", true
	default:
		return "", false
	}
}

func (r *MediaStatusResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// This resource models a status action. Deleting it must not delete the media item.
}

func (r *MediaStatusResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	mediaID, is4k, hasSuffix := strings.Cut(req.ID, ":")
	if !mediaIDPattern.MatchString(mediaID) || (hasSuffix && is4k != "4k") {
		resp.Diagnostics.AddError("Invalid Import ID", "Use a Seerr media ID, optionally followed by :4k.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), mediaID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("media_id"), mediaID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("is4k"), hasSuffix)...)
}
