package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &RequestResource{}
var _ resource.ResourceWithImportState = &RequestResource{}

type RequestResource struct {
	client *APIClient
}

type RequestModel struct {
	ID                        types.String `tfsdk:"id"`
	MediaType                 types.String `tfsdk:"media_type"`
	MediaID                   types.Int64  `tfsdk:"media_id"`
	SeerrMediaID              types.Int64  `tfsdk:"seerr_media_id"`
	Seasons                   types.List   `tfsdk:"seasons"`
	Is4K                      types.Bool   `tfsdk:"is_4k"`
	ServerID                  types.Int64  `tfsdk:"server_id"`
	ProfileID                 types.Int64  `tfsdk:"profile_id"`
	RootFolder                types.String `tfsdk:"root_folder"`
	UserID                    types.Int64  `tfsdk:"user_id"`
	Status                    types.Int64  `tfsdk:"status"`
	StatusWaitTimeoutSeconds  types.Int64  `tfsdk:"status_wait_timeout_seconds"`
	StatusWaitIntervalSeconds types.Int64  `tfsdk:"status_wait_interval_seconds"`
}

func NewRequestResource() resource.Resource {
	return &RequestResource{}
}

func (r *RequestResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_request"
}

func (r *RequestResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Seerr media requests.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"media_type": schema.StringAttribute{
				MarkdownDescription: "The type of media (movie or tv).",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("movie", "tv"),
				},
			},
			"media_id": schema.Int64Attribute{
				MarkdownDescription: "The TMDB ID of the media.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"seerr_media_id": schema.Int64Attribute{
				MarkdownDescription: "The Seerr-internal media ID created or associated with the request.",
				Computed:            true,
			},
			"seasons": schema.ListAttribute{
				MarkdownDescription: "List of season numbers to request (TV only).",
				Optional:            true,
				ElementType:         types.Int64Type,
			},
			"is_4k": schema.BoolAttribute{
				MarkdownDescription: "Whether to request in 4K.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"server_id": schema.Int64Attribute{
				MarkdownDescription: "Override the server ID for the request.",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"profile_id": schema.Int64Attribute{
				MarkdownDescription: "Override the quality profile ID for the request.",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"root_folder": schema.StringAttribute{
				MarkdownDescription: "Override the root folder for the request.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the user making the request (defaults to current user).",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"status": schema.Int64Attribute{
				MarkdownDescription: "The desired status of the request (1: Pending, 2: Approved, 3: Declined).",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int64{
					int64validator.OneOf(1, 2, 3),
				},
			},
			"status_wait_timeout_seconds": schema.Int64Attribute{
				MarkdownDescription: "Maximum seconds to wait after applying desired status for the observed request status to match. Defaults to 0 (waiting disabled).",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"status_wait_interval_seconds": schema.Int64Attribute{
				MarkdownDescription: "Seconds between status polls when status_wait_timeout_seconds is greater than 0. Defaults to 2; minimum 1.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(2),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
		},
	}
}

func (r *RequestResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RequestResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RequestModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	desiredStatus := data.Status

	payload := map[string]any{
		"mediaType": data.MediaType.ValueString(),
		"mediaId":   data.MediaID.ValueInt64(),
		"is4k":      data.Is4K.ValueBool(),
	}

	if !data.Seasons.IsNull() && !data.Seasons.IsUnknown() {
		var seasons []int64
		resp.Diagnostics.Append(data.Seasons.ElementsAs(ctx, &seasons, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		payload["seasons"] = seasons
	}

	if !data.ServerID.IsNull() && !data.ServerID.IsUnknown() {
		payload["serverId"] = data.ServerID.ValueInt64()
	}
	if !data.ProfileID.IsNull() && !data.ProfileID.IsUnknown() {
		payload["profileId"] = data.ProfileID.ValueInt64()
	}
	if !data.RootFolder.IsNull() && !data.RootFolder.IsUnknown() {
		payload["rootFolder"] = data.RootFolder.ValueString()
	}
	if !data.UserID.IsNull() && !data.UserID.IsUnknown() {
		payload["userId"] = data.UserID.ValueInt64()
	}

	body, _ := json.Marshal(payload)
	res, err := r.client.Request(ctx, "POST", "/api/v1/request", string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}
	if !HandleAPIResponse(ctx, res, &resp.Diagnostics, "Create") {
		return
	}

	extractedID, ok := ExtractIDFromJSON(res.Body)
	if !ok {
		resp.Diagnostics.AddError("Create Failed", "Could not extract request ID from response")
		return
	}

	data.ID = types.StringValue(extractedID)

	diags := r.readRequest(ctx, extractedID, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyRequestStatus(ctx, extractedID, desiredStatus); err != nil {
		resp.Diagnostics.AddError("Update Status Failed", err.Error())
		return
	}
	if err := r.waitForDesiredRequestStatus(ctx, extractedID, desiredStatus, data); err != nil {
		resp.Diagnostics.AddError("Wait For Request Status Failed", err.Error())
		return
	}
	diags = r.readRequest(ctx, extractedID, &data)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RequestResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RequestModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := r.readRequest(ctx, data.ID.ValueString(), &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.ID.IsNull() || data.ID.IsUnknown() || data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RequestResource) readRequest(ctx context.Context, requestID string, data *RequestModel) diag.Diagnostics {
	var diags diag.Diagnostics

	res, err := r.client.Request(ctx, "GET", "/api/v1/request/"+requestID, "", nil)
	if err != nil {
		diags.AddError("Read Failed", err.Error())
		return diags
	}
	if res.StatusCode == 404 {
		data.ID = types.StringNull()
		return diags
	}
	if !HandleAPIResponse(ctx, res, &diags, "Read") {
		return diags
	}

	var m map[string]any
	if err := json.Unmarshal(res.Body, &m); err != nil {
		diags.AddError("Read Failed", err.Error())
		return diags
	}

	if status, ok := m["status"].(float64); ok {
		data.Status = types.Int64Value(int64(status))
	}

	if media, ok := m["media"].(map[string]any); ok {
		if mediaID, ok := media["id"].(float64); ok {
			data.SeerrMediaID = types.Int64Value(int64(mediaID))
		}
		if mediaType, ok := media["mediaType"].(string); ok {
			data.MediaType = types.StringValue(mediaType)
		}
		if tmdbId, ok := media["tmdbId"].(float64); ok {
			data.MediaID = types.Int64Value(int64(tmdbId))
		}
	}

	if is4k, ok := m["is4k"].(bool); ok {
		data.Is4K = types.BoolValue(is4k)
	}

	if requestedBy, ok := m["requestedBy"].(map[string]any); ok {
		if userId, ok := requestedBy["id"].(float64); ok {
			data.UserID = types.Int64Value(int64(userId))
		}
	}

	return diags
}

func (r *RequestResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RequestModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]any{
		"mediaType": data.MediaType.ValueString(),
		"is4k":      data.Is4K.ValueBool(),
	}

	if !data.Seasons.IsNull() && !data.Seasons.IsUnknown() {
		var seasons []int64
		resp.Diagnostics.Append(data.Seasons.ElementsAs(ctx, &seasons, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		payload["seasons"] = seasons
	}

	body, _ := json.Marshal(payload)
	res, err := r.client.Request(ctx, "PUT", "/api/v1/request/"+data.ID.ValueString(), string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}
	if !HandleAPIResponse(ctx, res, &resp.Diagnostics, "Update") {
		return
	}

	if err := r.applyRequestStatus(ctx, data.ID.ValueString(), data.Status); err != nil {
		resp.Diagnostics.AddError("Update Status Failed", err.Error())
		return
	}
	if err := r.waitForDesiredRequestStatus(ctx, data.ID.ValueString(), data.Status, data); err != nil {
		resp.Diagnostics.AddError("Wait For Request Status Failed", err.Error())
		return
	}

	diags := r.readRequest(ctx, data.ID.ValueString(), &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RequestResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RequestModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.Request(ctx, "DELETE", "/api/v1/request/"+data.ID.ValueString(), "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Delete Failed", err.Error())
		return
	}
	if res.StatusCode != 404 && !HandleAPIResponse(ctx, res, &resp.Diagnostics, "Delete") {
		return
	}
}

func (r *RequestResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *RequestResource) applyRequestStatus(ctx context.Context, requestID string, status types.Int64) error {
	if status.IsNull() || status.IsUnknown() {
		return nil
	}
	statusPath, ok := requestStatusPath(status.ValueInt64())
	if !ok {
		return fmt.Errorf("unsupported request status %d; valid values are 1 (pending), 2 (approved), and 3 (declined)", status.ValueInt64())
	}
	res, err := r.client.Request(ctx, "POST", fmt.Sprintf("/api/v1/request/%s/%s", requestID, statusPath), "", nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}
	return nil
}

func (r *RequestResource) waitForDesiredRequestStatus(ctx context.Context, requestID string, status types.Int64, data RequestModel) error {
	if status.IsNull() || status.IsUnknown() {
		return nil
	}
	timeoutSeconds := int64(0)
	if !data.StatusWaitTimeoutSeconds.IsNull() && !data.StatusWaitTimeoutSeconds.IsUnknown() {
		timeoutSeconds = data.StatusWaitTimeoutSeconds.ValueInt64()
	}
	if timeoutSeconds <= 0 {
		return nil
	}

	intervalSeconds := int64(2)
	if !data.StatusWaitIntervalSeconds.IsNull() && !data.StatusWaitIntervalSeconds.IsUnknown() {
		intervalSeconds = data.StatusWaitIntervalSeconds.ValueInt64()
	}
	if intervalSeconds < 1 {
		intervalSeconds = 1
	}

	return waitForRequestStatus(
		ctx,
		r.client,
		requestID,
		status.ValueInt64(),
		time.Duration(timeoutSeconds)*time.Second,
		time.Duration(intervalSeconds)*time.Second,
		nil,
	)
}

// waitForRequestStatus polls GET /api/v1/request/{id} until the observed status
// matches wantStatus, the timeout elapses, or ctx is cancelled.
// sleep may be nil to use a context-aware default sleep (injectable for tests).
//
// The wait timeout is applied as a context deadline so individual HTTP GETs cannot
// outlive status_wait_timeout_seconds when the HTTP client timeout is larger.
func waitForRequestStatus(ctx context.Context, client *APIClient, requestID string, wantStatus int64, timeout, interval time.Duration, sleep func(context.Context, time.Duration) error) error {
	if sleep == nil {
		sleep = sleepContext
	}
	if interval <= 0 {
		interval = 2 * time.Second
	}
	if timeout <= 0 {
		return nil
	}

	deadline := time.Now().Add(timeout)
	waitCtx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	var lastStatus int64 = -1

	for {
		if err := waitCtx.Err(); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return fmt.Errorf("timed out after %s waiting for request %s to reach status %d (last observed %d)", timeout, requestID, wantStatus, lastStatus)
			}
			return fmt.Errorf("waiting for request %s status %d: %w", requestID, wantStatus, err)
		}

		res, err := client.Request(waitCtx, "GET", "/api/v1/request/"+requestID, "", nil)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || waitCtx.Err() != nil && errors.Is(waitCtx.Err(), context.DeadlineExceeded) {
				return fmt.Errorf("timed out after %s waiting for request %s to reach status %d (last observed %d)", timeout, requestID, wantStatus, lastStatus)
			}
			if errors.Is(err, context.Canceled) || waitCtx.Err() != nil && errors.Is(waitCtx.Err(), context.Canceled) {
				return fmt.Errorf("waiting for request %s status %d: %w", requestID, wantStatus, context.Canceled)
			}
			return fmt.Errorf("waiting for request %s status: %w", requestID, err)
		}
		if !StatusIsOK(res.StatusCode) {
			return fmt.Errorf("waiting for request %s status: status %d: %s", requestID, res.StatusCode, string(res.Body))
		}

		var m map[string]any
		if err := json.Unmarshal(res.Body, &m); err != nil {
			return fmt.Errorf("waiting for request %s status: parse response: %w", requestID, err)
		}
		statusVal, ok := m["status"].(float64)
		if !ok {
			return fmt.Errorf("waiting for request %s status: response missing numeric status field", requestID)
		}
		lastStatus = int64(statusVal)
		if lastStatus == wantStatus {
			return nil
		}

		remaining := time.Until(deadline)
		if remaining <= 0 {
			return fmt.Errorf("timed out after %s waiting for request %s to reach status %d (last observed %d)", timeout, requestID, wantStatus, lastStatus)
		}

		sleepFor := interval
		if remaining < sleepFor {
			sleepFor = remaining
		}
		if err := sleep(waitCtx, sleepFor); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return fmt.Errorf("timed out after %s waiting for request %s to reach status %d (last observed %d)", timeout, requestID, wantStatus, lastStatus)
			}
			if errors.Is(err, context.Canceled) {
				return fmt.Errorf("waiting for request %s status %d: %w", requestID, wantStatus, err)
			}
			return err
		}
	}
}

func sleepContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func requestStatusPath(status int64) (string, bool) {
	switch status {
	case 1:
		return "pending", true
	case 2:
		return "approve", true
	case 3:
		return "decline", true
	default:
		return "", false
	}
}
