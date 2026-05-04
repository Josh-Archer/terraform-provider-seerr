package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &RequestRetryResource{}
var _ resource.ResourceWithImportState = &RequestRetryResource{}

type RequestRetryResource struct {
	client *APIClient
}

type RequestRetryModel struct {
	ID        types.String `tfsdk:"id"`
	RequestID types.Int64  `tfsdk:"request_id"`
	Trigger   types.String `tfsdk:"trigger"`
	Status    types.Int64  `tfsdk:"status"`
}

func NewRequestRetryResource() resource.Resource { return &RequestRetryResource{} }

func (r *RequestRetryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_request_retry"
}

func (r *RequestRetryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retry a failed Seerr media request via `/api/v1/request/{requestId}/retry`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"request_id": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"trigger": schema.StringAttribute{
				MarkdownDescription: "Optional operator-controlled value. Changing it retries the request again.",
				Optional:            true,
			},
			"status": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (r *RequestRetryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RequestRetryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RequestRetryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.retryRequest(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Retry Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RequestRetryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RequestRetryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.readRequestRetry(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if data.ID.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RequestRetryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RequestRetryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.retryRequest(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Retry Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RequestRetryResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *RequestRetryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	requestID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("expected numeric request ID, got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("request_id"), requestID)...)
}

func (r *RequestRetryResource) retryRequest(ctx context.Context, data *RequestRetryModel) error {
	endpoint := fmt.Sprintf("/api/v1/request/%d/retry", data.RequestID.ValueInt64())
	res, err := r.client.Request(ctx, "POST", endpoint, "", nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}
	return applyRequestRetryResponse(data, res.Body)
}

func (r *RequestRetryResource) readRequestRetry(ctx context.Context, data *RequestRetryModel) error {
	endpoint := fmt.Sprintf("/api/v1/request/%d", data.RequestID.ValueInt64())
	res, err := r.client.Request(ctx, "GET", endpoint, "", nil)
	if err != nil {
		return err
	}
	if res.StatusCode == 404 {
		data.ID = types.StringNull()
		return nil
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}
	return applyRequestRetryResponse(data, res.Body)
}

func applyRequestRetryResponse(data *RequestRetryModel, body []byte) error {
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return err
	}
	if id, ok := int64ValueFromAny(decoded["id"]); ok {
		data.ID = types.StringValue(fmt.Sprintf("%d", id))
		data.RequestID = types.Int64Value(id)
	} else if !data.RequestID.IsNull() && !data.RequestID.IsUnknown() {
		data.ID = types.StringValue(fmt.Sprintf("%d", data.RequestID.ValueInt64()))
	}
	if status, ok := int64ValueFromAny(decoded["status"]); ok {
		data.Status = types.Int64Value(status)
	}
	return nil
}
