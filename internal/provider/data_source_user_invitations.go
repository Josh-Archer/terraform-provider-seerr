package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &UserInvitationsDataSource{}

type UserInvitationsDataSource struct {
	client *APIClient
}

type UserInvitationSummaryModel struct {
	ID    types.String `tfsdk:"id"`
	Email types.String `tfsdk:"email"`
}

type UserInvitationsDataSourceModel struct {
	ID          types.String                 `tfsdk:"id"`
	Invitations []UserInvitationSummaryModel `tfsdk:"invitations"`
}

func NewUserInvitationsDataSource() datasource.DataSource {
	return &UserInvitationsDataSource{}
}

func (d *UserInvitationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_invitations"
}

func (d *UserInvitationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Deprecated and unsupported. " + invitationUnsupportedMessage + ". Reading this data source fails without making API requests. Retained for compatibility with existing configurations.",
		DeprecationMessage:  invitationUnsupportedMessage,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder ID for the data source.",
				Computed:            true,
			},
			"invitations": schema.ListNestedAttribute{
				MarkdownDescription: "List of invitations.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The ID of the invitation.",
							Computed:            true,
						},
						"email": schema.StringAttribute{
							MarkdownDescription: "The email address of the invitation.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *UserInvitationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Configure Type", fmt.Sprintf("Expected *APIClient, got %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *UserInvitationsDataSource) Read(_ context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	resp.Diagnostics.AddError("Invitations Unsupported", invitationUnsupportedMessage)
}
