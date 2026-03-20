package provider

import (
	"context"
	"encoding/json"
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
		MarkdownDescription: "Get information about all existing Seerr user invitations.",
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

func (d *UserInvitationsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserInvitationsDataSourceModel

	// Fetch invitations
	res, err := d.client.Request(ctx, "GET", "/api/v1/user/invites", "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	var results []map[string]any
	if err := json.Unmarshal(res.Body, &results); err != nil {
		resp.Diagnostics.AddError("Read Failed", "Failed to parse API response: "+err.Error())
		return
	}

	for _, inv := range results {
		invitation := UserInvitationSummaryModel{}

		idRaw := inv["id"]
		switch v := idRaw.(type) {
		case float64:
			invitation.ID = types.StringValue(fmt.Sprintf("%.0f", v))
		case string:
			invitation.ID = types.StringValue(v)
		}

		if e, ok := inv["email"].(string); ok {
			invitation.Email = types.StringValue(e)
		}

		data.Invitations = append(data.Invitations, invitation)
	}

	data.ID = types.StringValue("all_invitations")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
