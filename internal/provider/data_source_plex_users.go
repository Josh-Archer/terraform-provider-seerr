package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &PlexUsersDataSource{}

type PlexUsersDataSource struct {
	client *APIClient
}

type PlexUsersDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Users types.List   `tfsdk:"users"`
	Total types.Int64  `tfsdk:"total"`
}

type PlexUserModel struct {
	ID       types.String `tfsdk:"id"`
	Title    types.String `tfsdk:"title"`
	Username types.String `tfsdk:"username"`
	Email    types.String `tfsdk:"email"`
	Thumb    types.String `tfsdk:"thumb"`
	UserType types.String `tfsdk:"user_type"`
}

func NewPlexUsersDataSource() datasource.DataSource {
	return &PlexUsersDataSource{}
}

func (d *PlexUsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plex_users"
}

func (d *PlexUsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch the list of registered users on the connected Plex server for selection and import.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier.",
				Computed:            true,
			},
			"total": schema.Int64Attribute{
				MarkdownDescription: "Total number of users found on the Plex server.",
				Computed:            true,
			},
			"users": schema.ListNestedAttribute{
				MarkdownDescription: "List of users registered on Plex.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Plex user ID.",
							Computed:            true,
						},
						"title": schema.StringAttribute{
							MarkdownDescription: "User title / display name.",
							Computed:            true,
						},
						"username": schema.StringAttribute{
							MarkdownDescription: "Plex username.",
							Computed:            true,
						},
						"email": schema.StringAttribute{
							MarkdownDescription: "Plex user email address.",
							Computed:            true,
						},
						"thumb": schema.StringAttribute{
							MarkdownDescription: "User avatar / thumbnail URL.",
							Computed:            true,
						},
						"user_type": schema.StringAttribute{
							MarkdownDescription: "User type on Plex.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *PlexUsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *APIClient, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

type rawPlexUserResponse struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Thumb    string `json:"thumb"`
	UserType string `json:"userType"`
}

func (d *PlexUsersDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured API Client", "API client was not provided during configuration.")
		return
	}

	res, err := d.client.Request(ctx, "GET", "/api/v1/settings/plex/users", "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !HandleAPIResponse(ctx, res, &resp.Diagnostics, "Read") {
		return
	}

	var rawUsers []rawPlexUserResponse
	if err := json.Unmarshal(res.Body, &rawUsers); err != nil {
		resp.Diagnostics.AddError("Error Parsing Plex Users Response", fmt.Sprintf("Failed to parse response: %s", err))
		return
	}

	userElements := make([]PlexUserModel, 0, len(rawUsers))
	for _, u := range rawUsers {
		userElements = append(userElements, PlexUserModel{
			ID:       types.StringValue(u.ID),
			Title:    types.StringValue(u.Title),
			Username: types.StringValue(u.Username),
			Email:    types.StringValue(u.Email),
			Thumb:    types.StringValue(u.Thumb),
			UserType: types.StringValue(u.UserType),
		})
	}

	userList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":        types.StringType,
			"title":     types.StringType,
			"username":  types.StringType,
			"email":     types.StringType,
			"thumb":     types.StringType,
			"user_type": types.StringType,
		},
	}, userElements)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := PlexUsersDataSourceModel{
		ID:    types.StringValue("plex_users"),
		Users: userList,
		Total: types.Int64Value(int64(len(rawUsers))),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
