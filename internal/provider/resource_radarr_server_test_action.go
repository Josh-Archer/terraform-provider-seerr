package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &RadarrServerTestResource{}
var _ resource.ResourceWithImportState = &RadarrServerTestResource{}

type RadarrServerTestResource struct {
	client *APIClient
}

type RadarrServerTestModel struct {
	ID            types.String `tfsdk:"id"`
	Hostname      types.String `tfsdk:"hostname"`
	Port          types.Int64  `tfsdk:"port"`
	APIKey        types.String `tfsdk:"api_key"`
	UseSSL        types.Bool   `tfsdk:"use_ssl"`
	BaseURL       types.String `tfsdk:"base_url"`
	Triggers      types.Map    `tfsdk:"triggers"`
	ProfilesCount types.Int64  `tfsdk:"profiles_count"`
}

func NewRadarrServerTestResource() resource.Resource {
	return &RadarrServerTestResource{}
}

func (r *RadarrServerTestResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_radarr_server_test"
}

func (r *RadarrServerTestResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Test connection to a Radarr server via `/api/v1/settings/radarr/test`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Resource identifier (formatted as `hostname:port` or `hostname:port/base_url`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "The hostname or IP address of the Radarr server.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "The port number of the Radarr server.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The API key for authenticating with the Radarr server.",
				Required:            true,
				Sensitive:           true,
			},
			"use_ssl": schema.BoolAttribute{
				MarkdownDescription: "Whether to connect to the Radarr server using SSL/TLS (HTTPS).",
				Required:            true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Optional URL path prefix for the Radarr server (e.g. `/radarr`).",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"triggers": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Arbitrary map of values (such as timestamps or hashes) that, when changed, will re-trigger the connection test.",
				Optional:            true,
			},
			"profiles_count": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The number of quality profiles returned by the Radarr server upon successful connection test.",
			},
		},
	}
}

func (r *RadarrServerTestResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RadarrServerTestResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RadarrServerTestModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError("Provider Error", "APIClient not configured")
		return
	}

	if err := r.testRadarrServer(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Radarr Server Test Failed", err.Error())
		return
	}

	baseURL := ""
	if !data.BaseURL.IsNull() && !data.BaseURL.IsUnknown() {
		baseURL = data.BaseURL.ValueString()
	}
	data.ID = types.StringValue(serverTestID(data.Hostname.ValueString(), data.Port.ValueInt64(), baseURL))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RadarrServerTestResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RadarrServerTestModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RadarrServerTestResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RadarrServerTestModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError("Provider Error", "APIClient not configured")
		return
	}

	if err := r.testRadarrServer(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Radarr Server Test Failed", err.Error())
		return
	}

	baseURL := ""
	if !data.BaseURL.IsNull() && !data.BaseURL.IsUnknown() {
		baseURL = data.BaseURL.ValueString()
	}
	data.ID = types.StringValue(serverTestID(data.Hostname.ValueString(), data.Port.ValueInt64(), baseURL))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RadarrServerTestResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Destroys operational action state; no remote side-effects on delete.
}

func (r *RadarrServerTestResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *RadarrServerTestResource) testRadarrServer(ctx context.Context, data *RadarrServerTestModel) error {
	testBody := map[string]any{
		"hostname": data.Hostname.ValueString(),
		"port":     data.Port.ValueInt64(),
		"apiKey":   data.APIKey.ValueString(),
		"useSsl":   data.UseSSL.ValueBool(),
	}
	if !data.BaseURL.IsNull() && !data.BaseURL.IsUnknown() && strings.TrimSpace(data.BaseURL.ValueString()) != "" {
		testBody["baseUrl"] = strings.TrimSpace(data.BaseURL.ValueString())
	}

	bodyBytes, err := json.Marshal(testBody)
	if err != nil {
		return fmt.Errorf("failed to encode request payload: %w", err)
	}

	res, err := r.client.Request(ctx, "POST", "/api/v1/settings/radarr/test", string(bodyBytes), nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}

	var testResult struct {
		Profiles []struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"profiles"`
	}
	if err := json.Unmarshal(res.Body, &testResult); err != nil {
		return fmt.Errorf("failed to parse test response: %w", err)
	}

	data.ProfilesCount = types.Int64Value(int64(len(testResult.Profiles)))
	return nil
}

func serverTestID(hostname string, port int64, baseURL string) string {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL != "" {
		return fmt.Sprintf("%s:%d%s", hostname, port, baseURL)
	}
	return fmt.Sprintf("%s:%d", hostname, port)
}
