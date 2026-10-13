package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const invitationUnsupportedMessage = "Seerr does not expose an invitation API; use seerr_user to create local users"

var _ resource.Resource = &UserInvitationResource{}
var _ resource.ResourceWithImportState = &UserInvitationResource{}

type UserInvitationResource struct {
	client *APIClient
}

type UserInvitationModel struct {
	ID    types.String `tfsdk:"id"`
	Email types.String `tfsdk:"email"`
}

func NewUserInvitationResource() resource.Resource {
	return &UserInvitationResource{}
}

func (r *UserInvitationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_invitation"
}

func (r *UserInvitationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Deprecated and unsupported. " + invitationUnsupportedMessage + ". Create, update and import fail; refresh drops any existing state and delete succeeds, both with a warning. No API requests are made.",
		DeprecationMessage:  invitationUnsupportedMessage,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the invitation.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The email address to invite.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *UserInvitationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserInvitationResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Invitations Unsupported", invitationUnsupportedMessage)
}

func (r *UserInvitationResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	// Nothing can exist server-side; drop stale state so plans and removal keep working.
	resp.State.RemoveResource(ctx)
	resp.Diagnostics.AddWarning("Invitations Unsupported", invitationUnsupportedMessage+"; removing this resource from state")
}

func (r *UserInvitationResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Invitations Unsupported", invitationUnsupportedMessage)
}

func (r *UserInvitationResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	// No API call: the framework removes the resource from state.
	resp.Diagnostics.AddWarning("Invitations Unsupported", invitationUnsupportedMessage)
}

func (r *UserInvitationResource) ImportState(_ context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.AddError("Invitations Unsupported", invitationUnsupportedMessage)
}
