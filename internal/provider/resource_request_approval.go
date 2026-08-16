package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &RequestApprovalResource{}
var _ resource.ResourceWithImportState = &RequestApprovalResource{}

type RequestApprovalResource struct {
	client *APIClient
}

type RequestApprovalModel struct {
	ID         types.String `tfsdk:"id"`
	RequestID  types.Int64  `tfsdk:"request_id"`
	Status     types.String `tfsdk:"status"`
	ModifiedBy types.Int64  `tfsdk:"modified_by"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

func NewRequestApprovalResource() resource.Resource {
	return &RequestApprovalResource{}
}

func (r *RequestApprovalResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_request_approval"
}

func (r *RequestApprovalResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage media request approval and decline status lifecycle via `/api/v1/request/{requestId}/{status}`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The request ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"request_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the media request to approve or decline.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The desired approval status (`approved`, `declined`, `approve`, or `decline`).",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("approved", "declined", "approve", "decline"),
				},
			},
			"modified_by": schema.Int64Attribute{
				MarkdownDescription: "The ID of the user who modified or approved/declined the request.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Date and time the request was last updated in ISO 8601 format.",
				Computed:            true,
			},
		},
	}
}

func (r *RequestApprovalResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RequestApprovalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RequestApprovalModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	requestID := data.RequestID.ValueInt64()
	statusAction := normalizeApprovalStatus(data.Status.ValueString())

	endpoint := fmt.Sprintf("/api/v1/request/%d/%s", requestID, statusAction)
	res, err := r.client.Request(ctx, "POST", endpoint, "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Create Request Approval Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Create Request Approval Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	data.ID = types.StringValue(strconv.FormatInt(requestID, 10))

	if err := r.readRequestDetails(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Read Request Approval Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RequestApprovalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RequestApprovalModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.RequestID.IsNull() || data.RequestID.IsUnknown() {
		if !data.ID.IsNull() && !data.ID.IsUnknown() && data.ID.ValueString() != "" {
			if id, err := strconv.ParseInt(data.ID.ValueString(), 10, 64); err == nil {
				data.RequestID = types.Int64Value(id)
			}
		}
	}

	if err := r.readRequestDetails(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Read Request Approval Failed", err.Error())
		return
	}

	if data.ID.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RequestApprovalResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RequestApprovalModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	requestID := data.RequestID.ValueInt64()
	statusAction := normalizeApprovalStatus(data.Status.ValueString())

	endpoint := fmt.Sprintf("/api/v1/request/%d/%s", requestID, statusAction)
	res, err := r.client.Request(ctx, "POST", endpoint, "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Update Request Approval Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Update Request Approval Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	if err := r.readRequestDetails(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Read Request Approval Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RequestApprovalResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Seerr request approvals are status state transitions; deleting removes from Terraform state.
}

func (r *RequestApprovalResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	requestID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("expected numeric request ID, got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("request_id"), requestID)...)
}

func normalizeApprovalStatus(status string) string {
	lower := strings.ToLower(strings.TrimSpace(status))
	if lower == "approved" || lower == "approve" {
		return "approve"
	}
	return "decline"
}

func (r *RequestApprovalResource) readRequestDetails(ctx context.Context, data *RequestApprovalModel) error {
	requestID := data.RequestID.ValueInt64()
	endpoint := fmt.Sprintf("/api/v1/request/%d", requestID)
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

	var m map[string]any
	if err := json.Unmarshal(res.Body, &m); err != nil {
		return fmt.Errorf("failed to parse request JSON: %w", err)
	}

	data.ID = types.StringValue(strconv.FormatInt(requestID, 10))
	data.RequestID = types.Int64Value(requestID)

	if statusVal, ok := int64ValueFromAny(m["status"]); ok {
		switch statusVal {
		case 2: // APPROVED
			if !data.Status.IsNull() && data.Status.ValueString() == "approve" {
				data.Status = types.StringValue("approve")
			} else {
				data.Status = types.StringValue("approved")
			}
		case 3: // DECLINED
			if !data.Status.IsNull() && data.Status.ValueString() == "decline" {
				data.Status = types.StringValue("decline")
			} else {
				data.Status = types.StringValue("declined")
			}
		default: // 1 = PENDING or other
			data.Status = types.StringValue("pending")
		}
	}

	if updatedAt, ok := stringValueFromAny(m["updatedAt"]); ok {
		data.UpdatedAt = types.StringValue(updatedAt)
	} else {
		data.UpdatedAt = types.StringNull()
	}

	if modifiedBy := m["modifiedBy"]; modifiedBy != nil {
		if mbMap, ok := modifiedBy.(map[string]any); ok {
			if id, ok := int64ValueFromAny(mbMap["id"]); ok {
				data.ModifiedBy = types.Int64Value(id)
			} else {
				data.ModifiedBy = types.Int64Null()
			}
		} else if id, ok := int64ValueFromAny(modifiedBy); ok {
			data.ModifiedBy = types.Int64Value(id)
		} else {
			data.ModifiedBy = types.Int64Null()
		}
	} else {
		data.ModifiedBy = types.Int64Null()
	}

	return nil
}
