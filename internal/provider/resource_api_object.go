package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &APIObjectResource{}
var _ resource.ResourceWithImportState = &APIObjectResource{}

type APIObjectResource struct {
	client *APIClient
}

type APIObjectModel struct {
	ID               types.String `tfsdk:"id"`
	Path             types.String `tfsdk:"path"`
	Headers          types.Map    `tfsdk:"headers"`
	RequestBodyJSON  types.String `tfsdk:"request_body_json"`
	DeleteBodyJSON   types.String `tfsdk:"delete_body_json"`
	ReadMethod       types.String `tfsdk:"read_method"`
	CreateMethod     types.String `tfsdk:"create_method"`
	UpdateMethod     types.String `tfsdk:"update_method"`
	DeleteMethod     types.String `tfsdk:"delete_method"`
	SkipDelete       types.Bool   `tfsdk:"skip_delete"`
	SuppressNotFound types.Bool   `tfsdk:"suppress_not_found"`
	ResponseBodyJSON types.String `tfsdk:"response_body_json"`
	StatusCode       types.Int64  `tfsdk:"status_code"`
}

func NewAPIObjectResource() resource.Resource {
	return &APIObjectResource{}
}

func (r *APIObjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_object"
}

func (r *APIObjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Universal Seerr API object resource. This can manage any Seerr endpoint by path and method configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Terraform resource identifier.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "Endpoint path or absolute URL, for example `/api/v1/settings/main`.",
				Required:            true,
			},
			"headers": schema.MapAttribute{
				MarkdownDescription: "Extra request headers.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"request_body_json": schema.StringAttribute{
				MarkdownDescription: "Optional request body JSON for create/update operations.",
				Optional:            true,
			},
			"delete_body_json": schema.StringAttribute{
				MarkdownDescription: "Optional request body JSON for delete operations.",
				Optional:            true,
			},
			"read_method": schema.StringAttribute{
				MarkdownDescription: "HTTP method used for read operations.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("GET"),
			},
			"create_method": schema.StringAttribute{
				MarkdownDescription: "HTTP method used for create operations.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("PUT"),
			},
			"update_method": schema.StringAttribute{
				MarkdownDescription: "HTTP method used for update operations.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("PUT"),
			},
			"delete_method": schema.StringAttribute{
				MarkdownDescription: "HTTP method used for delete operations.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("DELETE"),
			},
			"skip_delete": schema.BoolAttribute{
				MarkdownDescription: "If true, no HTTP call is made during delete. Useful for settings endpoints with no delete route.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"suppress_not_found": schema.BoolAttribute{
				MarkdownDescription: "If true, a 404 during read removes the resource from state.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"response_body_json": schema.StringAttribute{
				MarkdownDescription: "Raw JSON response body from the latest operation.",
				Computed:            true,
			},
			"status_code": schema.Int64Attribute{
				MarkdownDescription: "HTTP status code from the latest operation.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *APIObjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Configure Type", fmt.Sprintf("Expected *APIClient, got %T", req.ProviderData))
		return
	}

	r.client = client
}

func (r *APIObjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data APIObjectModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	headers, diags := mapFromTypesMap(ctx, data.Headers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Request(ctx, data.CreateMethod.ValueString(), data.Path.ValueString(), data.RequestBodyJSON.ValueString(), headers)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}
	if !StatusIsOK(httpResp.StatusCode) {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("Seerr returned status %d: %s", httpResp.StatusCode, string(httpResp.Body)))
		return
	}

	data.StatusCode = types.Int64Value(int64(httpResp.StatusCode))
	data.ResponseBodyJSON = types.StringValue(string(httpResp.Body))
	if id, ok := ExtractIDFromJSON(httpResp.Body); ok {
		data.ID = types.StringValue(id)
	} else {
		data.ID = types.StringValue(strings.TrimSpace(data.Path.ValueString()))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APIObjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data APIObjectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	headers, diags := mapFromTypesMap(ctx, data.Headers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Request(ctx, data.ReadMethod.ValueString(), data.Path.ValueString(), "", headers)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	if httpResp.StatusCode == 404 && data.SuppressNotFound.ValueBool() {
		resp.State.RemoveResource(ctx)
		return
	}
	if !StatusIsOK(httpResp.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("Seerr returned status %d: %s", httpResp.StatusCode, string(httpResp.Body)))
		return
	}

	data.StatusCode = types.Int64Value(int64(httpResp.StatusCode))
	data.ResponseBodyJSON = types.StringValue(string(httpResp.Body))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APIObjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data APIObjectModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	headers, diags := mapFromTypesMap(ctx, data.Headers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Request(ctx, data.UpdateMethod.ValueString(), data.Path.ValueString(), data.RequestBodyJSON.ValueString(), headers)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}
	if !StatusIsOK(httpResp.StatusCode) {
		resp.Diagnostics.AddError("Update Failed", fmt.Sprintf("Seerr returned status %d: %s", httpResp.StatusCode, string(httpResp.Body)))
		return
	}

	data.StatusCode = types.Int64Value(int64(httpResp.StatusCode))
	data.ResponseBodyJSON = types.StringValue(string(httpResp.Body))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APIObjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data APIObjectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.SkipDelete.ValueBool() {
		return
	}

	headers, diags := mapFromTypesMap(ctx, data.Headers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Request(ctx, data.DeleteMethod.ValueString(), data.Path.ValueString(), data.DeleteBodyJSON.ValueString(), headers)
	if err != nil {
		resp.Diagnostics.AddError("Delete Failed", err.Error())
		return
	}
	if httpResp.StatusCode == 404 {
		return
	}
	if !StatusIsOK(httpResp.StatusCode) {
		resp.Diagnostics.AddError("Delete Failed", fmt.Sprintf("Seerr returned status %d: %s", httpResp.StatusCode, string(httpResp.Body)))
	}
}

func (r *APIObjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("path"), req.ID)...)
}
