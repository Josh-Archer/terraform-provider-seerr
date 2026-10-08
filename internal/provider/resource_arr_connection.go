package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	stringvalidator "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ARRConnectionTestResource{}

type ARRConnectionTestResource struct{ client *APIClient }

type ARRConnectionTestModel struct {
	ID           types.String `tfsdk:"id"`
	Service      types.String `tfsdk:"service"`
	Hostname     types.String `tfsdk:"hostname"`
	Port         types.Int64  `tfsdk:"port"`
	APIKey       types.String `tfsdk:"api_key"`
	UseSSL       types.Bool   `tfsdk:"use_ssl"`
	BaseURL      types.String `tfsdk:"base_url"`
	Triggers     types.Map    `tfsdk:"triggers"`
	ResponseJSON types.String `tfsdk:"response_json"`
}

func NewARRConnectionTestResource() resource.Resource { return &ARRConnectionTestResource{} }

func (r *ARRConnectionTestResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_arr_connection_test"
}

func (r *ARRConnectionTestResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Test a Radarr or Sonarr connection using `/api/v1/settings/{service}/test`. The API key is sensitive in Terraform output, but Terraform still stores it in state; protect state access and backups. Destroying this action resource does not change Seerr settings.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"service":       schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, Validators: []validator.String{stringvalidator.OneOf("radarr", "sonarr")}},
			"hostname":      schema.StringAttribute{Required: true},
			"port":          schema.Int64Attribute{Required: true, Validators: []validator.Int64{int64validator.Between(1, 65535)}},
			"api_key":       schema.StringAttribute{Required: true, Sensitive: true},
			"use_ssl":       schema.BoolAttribute{Required: true},
			"base_url":      schema.StringAttribute{Optional: true},
			"triggers":      schema.MapAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "Changing a value re-runs the connection test."},
			"response_json": schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "JSON response from the Seerr test endpoint, including connection profiles, root folders, tags, and version-specific fields."},
		},
	}
}

func (r *ARRConnectionTestResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ARRConnectionTestResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ARRConnectionTestModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.test(ctx, &data); err != nil {
		resp.Diagnostics.AddError("ARR Connection Test Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ARRConnectionTestResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ARRConnectionTestModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.test(ctx, &data); err != nil {
		resp.Diagnostics.AddError("ARR Connection Test Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ARRConnectionTestResource) test(ctx context.Context, data *ARRConnectionTestModel) error {
	payload := map[string]any{
		"hostname": data.Hostname.ValueString(),
		"port":     data.Port.ValueInt64(),
		"apiKey":   data.APIKey.ValueString(),
		"useSsl":   data.UseSSL.ValueBool(),
	}
	if !data.BaseURL.IsNull() && !data.BaseURL.IsUnknown() {
		payload["baseUrl"] = data.BaseURL.ValueString()
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	apiPath := fmt.Sprintf("/api/v1/settings/%s/test", data.Service.ValueString())
	res, err := r.client.Request(ctx, "POST", apiPath, string(body), nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}
	// The action remains the same Terraform object when connection inputs change.
	data.ID = data.Service
	data.ResponseJSON = types.StringValue(string(res.Body))
	return nil
}

func (r *ARRConnectionTestResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ARRConnectionTestModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Connection tests are transient actions. Preserve the last response until
	// one of the configured inputs changes and invokes Update.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ARRConnectionTestResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// This action only probes a service and does not own any remote resource.
}
