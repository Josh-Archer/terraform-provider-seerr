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

var _ datasource.DataSource = &JellyfinUsersDataSource{}

type JellyfinUsersDataSource struct {
	client *APIClient
}

type JellyfinUsersDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Users types.List   `tfsdk:"users"`
	Total types.Int64  `tfsdk:"total"`
}

type JellyfinUserModel struct {
	ID       types.String `tfsdk:"id"`
	Username types.String `tfsdk:"username"`
	Email    types.String `tfsdk:"email"`
	UserType types.String `tfsdk:"user_type"`
}

func NewJellyfinUsersDataSource() datasource.DataSource {
	return &JellyfinUsersDataSource{}
}

func (d *JellyfinUsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jellyfin_users"
}

func (d *JellyfinUsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch the list of registered users on the connected Jellyfin server for selection and import.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier.",
				Computed:            true,
			},
			"total": schema.Int64Attribute{
				MarkdownDescription: "Total number of users found on the Jellyfin server.",
				Computed:            true,
			},
			"users": schema.ListNestedAttribute{
				MarkdownDescription: "List of users registered on Jellyfin.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Jellyfin user ID.",
							Computed:            true,
						},
						"username": schema.StringAttribute{
							MarkdownDescription: "Jellyfin username.",
							Computed:            true,
						},
						"email": schema.StringAttribute{
							MarkdownDescription: "Jellyfin user email address.",
							Computed:            true,
						},
						"user_type": schema.StringAttribute{
							MarkdownDescription: "User type on Jellyfin.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *JellyfinUsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

type rawJellyfinUserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	UserType string `json:"userType"`
}

func (d *JellyfinUsersDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured API Client", "API client was not provided during configuration.")
		return
	}

	res, err := d.client.Request(ctx, "GET", "/api/v1/settings/jellyfin/users", "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !HandleAPIResponse(ctx, res, &resp.Diagnostics, "Read") {
		return
	}

	var rawUsers []rawJellyfinUserResponse
	if err := json.Unmarshal(res.Body, &rawUsers); err != nil {
		resp.Diagnostics.AddError("Error Parsing Jellyfin Users Response", fmt.Sprintf("Failed to parse response: %s", err))
		return
	}

	userElements := make([]JellyfinUserModel, 0, len(rawUsers))
	for _, u := range rawUsers {
		userElements = append(userElements, JellyfinUserModel{
			ID:       types.StringValue(u.ID),
			Username: types.StringValue(u.Username),
			Email:    types.StringValue(u.Email),
			UserType: types.StringValue(u.UserType),
		})
	}

	userList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":        types.StringType,
			"username":  types.StringType,
			"email":     types.StringType,
			"user_type": types.StringType,
		},
	}, userElements)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := JellyfinUsersDataSourceModel{
		ID:    types.StringValue("jellyfin_users"),
		Users: userList,
		Total: types.Int64Value(int64(len(rawUsers))),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
