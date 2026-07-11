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
	Permissions types.Int64  `tfsdk:"permissions"`
}

type UsersDataSourceModel struct {
	ID    types.String       `tfsdk:"id"`
	Users []UserSummaryModel `tfsdk:"users"`
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

func (d *UsersDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UsersDataSourceModel

	users, err := d.fetchUsers(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	data.Users = users
	data.ID = types.StringValue("all_users")

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
