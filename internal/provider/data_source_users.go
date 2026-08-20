package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &UsersDataSource{}

type UsersDataSource struct {
	client *APIClient
}

type UserSummaryModel struct {
	ID          types.String `tfsdk:"id"`
	Email       types.String `tfsdk:"email"`
	Username    types.String `tfsdk:"username"`
	DisplayName types.String `tfsdk:"display_name"`
	UserType    types.Int64  `tfsdk:"user_type"`
	Permissions types.Int64  `tfsdk:"permissions"`
}

type UsersDataSourceModel struct {
	ID                   types.String       `tfsdk:"id"`
	FilterUserType       types.Int64        `tfsdk:"filter_user_type"`
	FilterPermissionsHas types.Int64        `tfsdk:"filter_permissions_has"`
	Users                []UserSummaryModel `tfsdk:"users"`
}

func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

func (d *UsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *UsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get information about all existing Seerr users.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder ID for the data source.",
				Computed:            true,
			},
			"filter_user_type": schema.Int64Attribute{
				MarkdownDescription: "Filter by user type (1=Plex, 2=Local, 3=Jellyfin, 4=Emby).",
				Optional:            true,
			},
			"filter_permissions_has": schema.Int64Attribute{
				MarkdownDescription: "Filter to users who have ALL bits in this bitmask set.",
				Optional:            true,
			},
			"users": schema.ListNestedAttribute{
				MarkdownDescription: "List of users.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The ID of the user.",
							Computed:            true,
						},
						"email": schema.StringAttribute{
							MarkdownDescription: "The email address of the user.",
							Computed:            true,
						},
						"username": schema.StringAttribute{
							MarkdownDescription: "The username of the user.",
							Computed:            true,
						},
						"display_name": schema.StringAttribute{
							MarkdownDescription: "The display name of the user.",
							Computed:            true,
						},
						"user_type": schema.Int64Attribute{
							MarkdownDescription: "The type of user (1=Plex, 2=Local, 3=Jellyfin, 4=Emby).",
							Computed:            true,
						},
						"permissions": schema.Int64Attribute{
							MarkdownDescription: "Permissions bitmask for the user.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *UsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UsersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	allUsers, err := d.fetchUsers(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	var filteredUsers []UserSummaryModel
	for _, u := range allUsers {
		if !data.FilterUserType.IsNull() && !data.FilterUserType.IsUnknown() {
			if u.UserType.ValueInt64() != data.FilterUserType.ValueInt64() {
				continue
			}
		}
		if !data.FilterPermissionsHas.IsNull() && !data.FilterPermissionsHas.IsUnknown() {
			filterPerm := data.FilterPermissionsHas.ValueInt64()
			if (u.Permissions.ValueInt64() & filterPerm) != filterPerm {
				continue
			}
		}
		filteredUsers = append(filteredUsers, u)
	}

	if filteredUsers == nil {
		filteredUsers = []UserSummaryModel{}
	}
	data.Users = filteredUsers

	idStr := "all_users"
	if !data.FilterUserType.IsNull() && !data.FilterUserType.IsUnknown() {
		idStr += fmt.Sprintf("_type_%d", data.FilterUserType.ValueInt64())
	}
	if !data.FilterPermissionsHas.IsNull() && !data.FilterPermissionsHas.IsUnknown() {
		idStr += fmt.Sprintf("_perms_%d", data.FilterPermissionsHas.ValueInt64())
	}
	data.ID = types.StringValue(idStr)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// fetchUsers retrieves every user page from the Seerr API.
func (d *UsersDataSource) fetchUsers(ctx context.Context) ([]UserSummaryModel, error) {
	results, err := fetchAllPaginatedResults(ctx, d.client, "/api/v1/user", defaultPaginationPageSize)
	if err != nil {
		return nil, err
	}

	users := make([]UserSummaryModel, 0, len(results))
	for _, u := range results {
		user := UserSummaryModel{}

		idRaw := u["id"]
		switch v := idRaw.(type) {
		case float64:
			user.ID = types.StringValue(fmt.Sprintf("%.0f", v))
		case string:
			user.ID = types.StringValue(v)
		}

		if e, ok := u["email"].(string); ok {
			user.Email = types.StringValue(e)
		}
		if un, ok := u["username"].(string); ok {
			user.Username = types.StringValue(un)
		}
		if dn, ok := u["displayName"].(string); ok {
			user.DisplayName = types.StringValue(dn)
		}
		if ut, ok := u["userType"].(float64); ok {
			user.UserType = types.Int64Value(int64(ut))
		}
		if p, ok := u["permissions"].(float64); ok {
			user.Permissions = types.Int64Value(int64(p))
		} else if pStr, ok := u["permissions"].(string); ok {
			if pInt, err := strconv.ParseInt(pStr, 10, 64); err == nil {
				user.Permissions = types.Int64Value(pInt)
			}
		}

		users = append(users, user)
	}
	return users, nil
}
