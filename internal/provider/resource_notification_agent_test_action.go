package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &NotificationAgentTestResource{}
var _ resource.ResourceWithImportState = &NotificationAgentTestResource{}

type NotificationAgentTestResource struct {
	client *APIClient
}

type NotificationAgentTestModel struct {
	ID       types.String `tfsdk:"id"`
	Agent    types.String `tfsdk:"agent"`
	Triggers types.Map    `tfsdk:"triggers"`
}

func NewNotificationAgentTestResource() resource.Resource {
	return &NotificationAgentTestResource{}
}

func (r *NotificationAgentTestResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_agent_test"
}

func (r *NotificationAgentTestResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Trigger a test notification delivery via `/api/v1/settings/notifications/{agent}/test`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"agent": schema.StringAttribute{
				MarkdownDescription: "The notification agent type to test (e.g. `discord`, `email`, `gotify`, `lunasea`, `ntfy`, `pushbullet`, `pushover`, `slack`, `telegram`, `webhook`, `webpush`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"triggers": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Arbitrary map of values (such as timestamps or hashes) that, when changed, will re-trigger the test notification delivery.",
				Optional:            true,
			},
		},
	}
}

func (r *NotificationAgentTestResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NotificationAgentTestResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NotificationAgentTestModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.testNotificationAgent(ctx, data.Agent.ValueString()); err != nil {
		resp.Diagnostics.AddError("Notification Agent Test Failed", err.Error())
		return
	}

	data.ID = data.Agent
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NotificationAgentTestResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationAgentTestModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NotificationAgentTestResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NotificationAgentTestModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.testNotificationAgent(ctx, data.Agent.ValueString()); err != nil {
		resp.Diagnostics.AddError("Notification Agent Test Failed", err.Error())
		return
	}

	data.ID = data.Agent
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NotificationAgentTestResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Destroys operational action state; no remote side-effects on delete.
}

func (r *NotificationAgentTestResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("agent"), req.ID)...)
}

func (r *NotificationAgentTestResource) testNotificationAgent(ctx context.Context, agent string) error {
	apiPath := fmt.Sprintf("/api/v1/settings/notifications/%s/test", agent)
	res, err := r.client.Request(ctx, "POST", apiPath, "", nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}
	return nil
}
