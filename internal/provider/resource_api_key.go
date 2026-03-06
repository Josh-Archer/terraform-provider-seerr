package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &APIKeyResource{}

type APIKeyResource struct {
	client *APIClient
}

type APIKeyModel struct {
	ApiKey     types.String `tfsdk:"api_key"`
	StatusCode types.Int64  `tfsdk:"status_code"`
}

func NewAPIKeyResource() resource.Resource { return &APIKeyResource{} }

func (r *APIKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *APIKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage the Seerr API key. Creating this resource will regenerate the API key.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The current Seerr API key.",
				Computed:            true,
				Sensitive:           true,
			},
			"status_code": schema.Int64Attribute{
				MarkdownDescription: "HTTP status code.",
				Computed:            true,
			},
		},
	}
}

func (r *APIKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *APIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data APIKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.Request(ctx, "POST", "/api/v1/settings/main/regenerate", "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Regenerate Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Regenerate Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	var decoded map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err != nil {
		resp.Diagnostics.AddError("Regenerate Failed", fmt.Sprintf("failed to decode response: %s", err))
		return
	}

	apiKey, ok := decoded["apiKey"].(string)
	if !ok {
		resp.Diagnostics.AddError("Regenerate Failed", "apiKey not found in response")
		return
	}

	data.ApiKey = types.StringValue(apiKey)
	data.StatusCode = types.Int64Value(int64(res.StatusCode))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data APIKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.Request(ctx, "GET", "/api/v1/settings/main", "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	var decoded map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err != nil {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("failed to decode response: %s", err))
		return
	}

	apiKey, ok := decoded["apiKey"].(string)
	if !ok {
		resp.Diagnostics.AddError("Read Failed", "apiKey not found in response")
		return
	}

	data.ApiKey = types.StringValue(apiKey)
	data.StatusCode = types.Int64Value(int64(res.StatusCode))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APIKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Update will regenerate the key again.
	r.Create(ctx, resource.CreateRequest{Plan: req.Plan}, &resource.CreateResponse{State: resp.State, Diagnostics: resp.Diagnostics})
}

func (r *APIKeyResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Delete does nothing.
}
