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

var _ resource.Resource = &BackupSettingsResource{}
var _ resource.ResourceWithImportState = &BackupSettingsResource{}

type BackupSettingsResource struct {
	client *APIClient
}

type BackupSettingsModel struct {
	ID           types.String `tfsdk:"id"`
	Schedule     types.String `tfsdk:"schedule"`
	Retention    types.Int64  `tfsdk:"retention"`
	StoragePath  types.String `tfsdk:"storage_path"`
	ResponseJSON types.String `tfsdk:"response_json"`
	StatusCode   types.Int64  `tfsdk:"status_code"`
}

func NewBackupSettingsResource() resource.Resource { return &BackupSettingsResource{} }

func (r *BackupSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backup_settings"
}

func (r *BackupSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Seerr backup settings via /api/v1/settings/backups.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"schedule": schema.StringAttribute{
				MarkdownDescription: "The backup schedule in cron format.",
				Optional:            true,
				Computed:            true,
			},
			"retention": schema.Int64Attribute{
				MarkdownDescription: "The number of backups to retain.",
				Optional:            true,
				Computed:            true,
			},
			"storage_path": schema.StringAttribute{
				MarkdownDescription: "The path where backups are stored.",
				Optional:            true,
				Computed:            true,
			},
			"response_json": schema.StringAttribute{
				MarkdownDescription: "Raw JSON response body.",
				Computed:            true,
			},
			"status_code": schema.Int64Attribute{
				MarkdownDescription: "HTTP status code.",
				Computed:            true,
			},
		},
	}
}

func (r *BackupSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BackupSettingsResource) buildPayload(data *BackupSettingsModel) map[string]any {
	payload := make(map[string]any)
	if !data.Schedule.IsNull() && !data.Schedule.IsUnknown() {
		payload["schedule"] = data.Schedule.ValueString()
	}
	if !data.Retention.IsNull() && !data.Retention.IsUnknown() {
		payload["retention"] = data.Retention.ValueInt64()
	}
	if !data.StoragePath.IsNull() && !data.StoragePath.IsUnknown() {
		payload["storage_path"] = data.StoragePath.ValueString()
	}

	return payload
}

func (r *BackupSettingsResource) applyDecodedSettings(data *BackupSettingsModel, decoded map[string]any) {
	if v, ok := decoded["schedule"].(string); ok {
		data.Schedule = types.StringValue(v)
	}
	if v, ok := decoded["retention"].(float64); ok {
		data.Retention = types.Int64Value(int64(v))
	}
	if v, ok := decoded["storage_path"].(string); ok {
		data.StoragePath = types.StringValue(v)
	}
}

func (r *BackupSettingsResource) refreshState(ctx context.Context, data *BackupSettingsModel) error {
	res, err := r.client.Request(ctx, "GET", "/api/v1/settings/backups", "", nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}

	var decoded map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err != nil {
		return fmt.Errorf("failed to decode response: %s", err)
	}

	data.StatusCode = types.Int64Value(int64(res.StatusCode))
	data.ResponseJSON = types.StringValue(string(res.Body))
	r.applyDecodedSettings(data, decoded)
	return nil
}

func (r *BackupSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BackupSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.buildPayload(&data))
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("failed to marshal payload: %s", err))
		return
	}

	res, err := r.client.Request(ctx, "POST", "/api/v1/settings/backups", string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	data.ID = types.StringValue("backups")
	data.StatusCode = types.Int64Value(int64(res.StatusCode))
	data.ResponseJSON = types.StringValue(string(res.Body))
	var decoded map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err != nil {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("failed to decode response: %s", err))
		return
	}
	r.applyDecodedSettings(&data, decoded)
	if err := r.refreshState(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BackupSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BackupSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.refreshState(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	data.ID = types.StringValue("backups")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BackupSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BackupSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.buildPayload(&data))
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", fmt.Sprintf("failed to marshal payload: %s", err))
		return
	}

	res, err := r.client.Request(ctx, "POST", "/api/v1/settings/backups", string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Update Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	data.ID = types.StringValue("backups")
	data.StatusCode = types.Int64Value(int64(res.StatusCode))
	data.ResponseJSON = types.StringValue(string(res.Body))

	var decoded map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err != nil {
		resp.Diagnostics.AddError("Update Failed", fmt.Sprintf("failed to decode response: %s", err))
		return
	}
	r.applyDecodedSettings(&data, decoded)
	if err := r.refreshState(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BackupSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// There is no concept of "deleting" backup settings in Seerr; it is a singleton.
}

func (r *BackupSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
