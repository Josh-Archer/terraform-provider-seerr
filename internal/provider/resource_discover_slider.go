package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &DiscoverSliderResource{}

type DiscoverSliderResource struct {
	client *APIClient
}

type DiscoverSliderItemModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Type      types.Int64  `tfsdk:"type"`
	IsBuiltIn types.Bool   `tfsdk:"is_built_in"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	Title     types.String `tfsdk:"title"`
	Data      types.String `tfsdk:"data"`
}

type DiscoverSliderModel struct {
	ID      types.String              `tfsdk:"id"`
	Sliders []DiscoverSliderItemModel `tfsdk:"sliders"`
}

func NewDiscoverSliderResource() resource.Resource { return &DiscoverSliderResource{} }

func (r *DiscoverSliderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_discover_slider"
}

func (r *DiscoverSliderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage discovery sliders in Seerr via `/api/v1/settings/discover`. This is a singleton resource that manages the entire list of sliders to ensure consistent ordering.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"sliders": schema.ListNestedBlock{
				MarkdownDescription: "The list of discovery sliders. The order of this list determines the display order in Seerr.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "The internal ID of the slider.",
						},
						"type": schema.Int64Attribute{
							Required:            true,
							MarkdownDescription: "The type of the slider (e.g., 1 for TV, 2 for Movie).",
						},
						"is_built_in": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether this is a built-in slider.",
						},
						"enabled": schema.BoolAttribute{
							Required:            true,
							MarkdownDescription: "Whether the slider is enabled.",
						},
						"title": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "The title of the slider (only for custom sliders).",
						},
						"data": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "The data/query for the slider (only for custom sliders).",
						},
					},
				},
			},
		},
	}
}

func (r *DiscoverSliderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DiscoverSliderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DiscoverSliderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.updateSliders(ctx, data.Sliders); err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}

	data.ID = types.StringValue("settings")
	resp.Diagnostics.Append(r.readSliders(ctx, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
func (r *DiscoverSliderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DiscoverSliderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue("settings")
	resp.Diagnostics.Append(r.readSliders(ctx, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DiscoverSliderResource) readSliders(ctx context.Context, data *DiscoverSliderModel) diag.Diagnostics {
	var diags diag.Diagnostics

	res, err := r.client.Request(ctx, "GET", "/api/v1/settings/discover", "", nil)
	if err != nil {
		diags.AddError("Read Failed", err.Error())
		return diags
	}

	if !StatusIsOK(res.StatusCode) {
		diags.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return diags
	}

	var apiSliders []map[string]any
	if err := json.Unmarshal(res.Body, &apiSliders); err != nil {
		diags.AddError("Read Failed", fmt.Sprintf("failed to decode response: %s", err))
		return diags
	}

	data.Sliders = make([]DiscoverSliderItemModel, 0, len(apiSliders))
	for _, s := range apiSliders {
		item := DiscoverSliderItemModel{
			ID:      r.toInt64(s["id"]),
			Type:    r.toInt64(s["type"]),
			Enabled: r.toBool(s["enabled"]),
		}
		if v, ok := s["isBuiltIn"].(bool); ok {
			item.IsBuiltIn = types.BoolValue(v)
		}
		if v, ok := s["title"].(string); ok {
			item.Title = types.StringValue(v)
		} else {
			item.Title = types.StringNull()
		}
		if v, ok := s["data"].(string); ok {
			item.Data = types.StringValue(v)
		} else {
			item.Data = types.StringNull()
		}
		data.Sliders = append(data.Sliders, item)
	}

	return diags
}

func (r *DiscoverSliderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DiscoverSliderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.updateSliders(ctx, data.Sliders); err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}

	data.ID = types.StringValue("settings")
	resp.Diagnostics.Append(r.readSliders(ctx, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DiscoverSliderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Deleting the singleton doesn't make sense to delete sliders from Seerr,
	// but we could "reset" them if we wanted. For now, no-op.
}

func (r *DiscoverSliderResource) updateSliders(ctx context.Context, sliders []DiscoverSliderItemModel) error {
	payload := make([]map[string]any, 0, len(sliders))
	for _, s := range sliders {
		item := map[string]any{
			"type":    s.Type.ValueInt64(),
			"enabled": s.Enabled.ValueBool(),
		}
		if !s.ID.IsNull() && !s.ID.IsUnknown() {
			item["id"] = s.ID.ValueInt64()
		}
		if !s.Title.IsNull() && !s.Title.IsUnknown() {
			item["title"] = s.Title.ValueString()
		} else {
			item["title"] = ""
		}
		if !s.Data.IsNull() && !s.Data.IsUnknown() {
			item["data"] = s.Data.ValueString()
		} else {
			item["data"] = ""
		}
		payload = append(payload, item)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	res, err := r.client.Request(ctx, "POST", "/api/v1/settings/discover", string(body), nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}

	return nil
}

func (r *DiscoverSliderResource) toInt64(v any) types.Int64 {
	switch val := v.(type) {
	case float64:
		return types.Int64Value(int64(val))
	case int64:
		return types.Int64Value(val)
	case int:
		return types.Int64Value(int64(val))
	}
	return types.Int64Null()
}

func (r *DiscoverSliderResource) toBool(v any) types.Bool {
	if b, ok := v.(bool); ok {
		return types.BoolValue(b)
	}
	return types.BoolNull()
}
