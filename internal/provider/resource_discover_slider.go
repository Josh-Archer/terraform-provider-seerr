package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

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
	resp.Diagnostics.Append(r.readManagedSliders(ctx, data.Sliders, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
func (r *DiscoverSliderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DiscoverSliderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue("settings")
	resp.Diagnostics.Append(r.readManagedSliders(ctx, data.Sliders, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(data.Sliders) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DiscoverSliderResource) readManagedSliders(ctx context.Context, managed []DiscoverSliderItemModel, data *DiscoverSliderModel) diag.Diagnostics {
	var diags diag.Diagnostics

	allSliders, err := r.fetchSliders(ctx)
	if err != nil {
		diags.AddError("Read Failed", err.Error())
		return diags
	}

	data.Sliders = filterManagedSliders(allSliders, managed)
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
	resp.Diagnostics.Append(r.readManagedSliders(ctx, data.Sliders, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DiscoverSliderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DiscoverSliderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.fetchSliders(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Delete Failed", err.Error())
		return
	}

	remaining := make([]DiscoverSliderItemModel, 0, len(current))
	matchedAny := false
	for _, slider := range current {
		if matchesAnyManagedSlider(slider, state.Sliders) {
			matchedAny = true
			continue
		}
		remaining = append(remaining, slider)
	}

	if !matchedAny {
		resp.State.RemoveResource(ctx)
		return
	}

	if err := r.replaceSliders(ctx, remaining); err != nil {
		resp.Diagnostics.AddError("Delete Failed", err.Error())
		return
	}
}

func (r *DiscoverSliderResource) updateSliders(ctx context.Context, sliders []DiscoverSliderItemModel) error {
	current, err := r.fetchSliders(ctx)
	if err != nil {
		return err
	}

	resolved := resolveSliderIDs(current, sliders)
	managedKeys := make([]sliderKey, 0, len(resolved))
	for _, slider := range resolved {
		managedKeys = append(managedKeys, sliderIdentity(slider))
	}

	payload := make([]DiscoverSliderItemModel, 0, len(resolved)+len(current))
	payload = append(payload, resolved...)
	for _, slider := range current {
		if slices.Contains(managedKeys, sliderIdentity(slider)) {
			continue
		}
		payload = append(payload, slider)
	}

	return r.replaceSliders(ctx, payload)
}

func (r *DiscoverSliderResource) replaceSliders(ctx context.Context, sliders []DiscoverSliderItemModel) error {
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

func (r *DiscoverSliderResource) fetchSliders(ctx context.Context) ([]DiscoverSliderItemModel, error) {
	res, err := r.client.Request(ctx, "GET", "/api/v1/settings/discover", "", nil)
	if err != nil {
		return nil, err
	}

	if !StatusIsOK(res.StatusCode) {
		return nil, fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}

	var apiSliders []map[string]any
	if err := json.Unmarshal(res.Body, &apiSliders); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	sliders := make([]DiscoverSliderItemModel, 0, len(apiSliders))
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
		sliders = append(sliders, item)
	}

	return sliders, nil
}

type sliderKey struct {
	ID    int64
	Type  int64
	Title string
	Data  string
}

func sliderIdentity(slider DiscoverSliderItemModel) sliderKey {
	key := sliderKey{
		Type:  slider.Type.ValueInt64(),
		Title: sliderStringValue(slider.Title),
		Data:  sliderStringValue(slider.Data),
	}
	if !slider.ID.IsNull() && !slider.ID.IsUnknown() {
		key.ID = slider.ID.ValueInt64()
	}
	return key
}

func sliderStringValue(v types.String) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return v.ValueString()
}

func slidersMatch(a, b DiscoverSliderItemModel) bool {
	if !a.ID.IsNull() && !a.ID.IsUnknown() && !b.ID.IsNull() && !b.ID.IsUnknown() {
		return a.ID.ValueInt64() == b.ID.ValueInt64()
	}

	return a.Type.ValueInt64() == b.Type.ValueInt64() &&
		sliderStringValue(a.Title) == sliderStringValue(b.Title) &&
		sliderStringValue(a.Data) == sliderStringValue(b.Data)
}

func resolveSliderIDs(current, desired []DiscoverSliderItemModel) []DiscoverSliderItemModel {
	resolved := make([]DiscoverSliderItemModel, 0, len(desired))
	used := make([]bool, len(current))

	for _, slider := range desired {
		resolvedSlider := slider
		for idx, existing := range current {
			if used[idx] || !slidersMatch(existing, slider) {
				continue
			}
			used[idx] = true
			resolvedSlider.ID = existing.ID
			resolvedSlider.IsBuiltIn = existing.IsBuiltIn
			break
		}
		resolved = append(resolved, resolvedSlider)
	}

	return resolved
}

func filterManagedSliders(current, managed []DiscoverSliderItemModel) []DiscoverSliderItemModel {
	filtered := make([]DiscoverSliderItemModel, 0, len(managed))
	used := make([]bool, len(current))

	for _, target := range managed {
		for idx, existing := range current {
			if used[idx] || !slidersMatch(existing, target) {
				continue
			}
			used[idx] = true
			filtered = append(filtered, existing)
			break
		}
	}

	return filtered
}

func matchesAnyManagedSlider(slider DiscoverSliderItemModel, managed []DiscoverSliderItemModel) bool {
	for _, target := range managed {
		if slidersMatch(slider, target) {
			return true
		}
	}
	return false
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
