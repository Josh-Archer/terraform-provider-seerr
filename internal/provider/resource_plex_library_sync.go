package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &PlexLibrarySyncResource{}
var _ resource.ResourceWithImportState = &PlexLibrarySyncResource{}

type PlexLibrarySyncResource struct {
	client *APIClient
}

type PlexLibrarySyncModel struct {
	ID       types.String  `tfsdk:"id"`
	Running  types.Bool    `tfsdk:"running"`
	Progress types.Float64 `tfsdk:"progress"`
	Total    types.Float64 `tfsdk:"total"`
	Triggers types.Map     `tfsdk:"triggers"`
}

func NewPlexLibrarySyncResource() resource.Resource {
	return &PlexLibrarySyncResource{}
}

func (r *PlexLibrarySyncResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plex_library_sync"
}

func (r *PlexLibrarySyncResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Trigger and monitor full Plex library scans via `/api/v1/settings/plex/sync`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"running": schema.BoolAttribute{
				MarkdownDescription: "Whether a Plex library scan is currently running.",
				Computed:            true,
			},
			"progress": schema.Float64Attribute{
				MarkdownDescription: "Current scan progress percentage or count.",
				Computed:            true,
			},
			"total": schema.Float64Attribute{
				MarkdownDescription: "Total items to scan.",
				Computed:            true,
			},
			"triggers": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Optional arbitrary map of values that will trigger a new Plex scan when changed.",
				Optional:            true,
			},
		},
	}
}

func (r *PlexLibrarySyncResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PlexLibrarySyncResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PlexLibrarySyncModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.triggerScan(ctx, true); err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}

	if err := r.readStatus(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}

	data.ID = types.StringValue("plex_library_sync")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PlexLibrarySyncResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PlexLibrarySyncModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readStatus(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	data.ID = types.StringValue("plex_library_sync")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PlexLibrarySyncResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PlexLibrarySyncModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.triggerScan(ctx, true); err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}

	if err := r.readStatus(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}

	data.ID = types.StringValue("plex_library_sync")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PlexLibrarySyncResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Send cancel request if running
	_ = r.triggerScan(ctx, false)
}

func (r *PlexLibrarySyncResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *PlexLibrarySyncResource) triggerScan(ctx context.Context, start bool) error {
	bodyMap := map[string]bool{"start": start, "cancel": !start}
	bodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		return err
	}

	res, err := r.client.Request(ctx, "POST", "/api/v1/settings/plex/sync", string(bodyBytes), nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}

	return nil
}

func (r *PlexLibrarySyncResource) readStatus(ctx context.Context, data *PlexLibrarySyncModel) error {
	res, err := r.client.Request(ctx, "GET", "/api/v1/settings/plex/sync", "", nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}

	var status struct {
		Running  bool    `json:"running"`
		Progress float64 `json:"progress"`
		Total    float64 `json:"total"`
	}

	if err := json.Unmarshal(res.Body, &status); err != nil {
		return fmt.Errorf("failed to parse sync status: %w", err)
	}

	data.Running = types.BoolValue(status.Running)
	data.Progress = types.Float64Value(status.Progress)
	data.Total = types.Float64Value(status.Total)
	return nil
}
