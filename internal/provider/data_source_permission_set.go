package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &permissionSetDataSource{}

var PermissionsMap = map[string]int64{
	"admin":                 2,
	"manage_settings":       4,
	"manage_users":          8,
	"manage_requests":       16,
	"request":               32,
	"vote":                  64,
	"auto_approve":          128,
	"auto_approve_movie":    256,
	"auto_approve_tv":       512,
	"request_4k":            1024,
	"request_4k_movie":      2048,
	"request_4k_tv":         4096,
	"request_advanced":      8192,
	"request_view":          16384,
	"auto_approve_4k":       32768,
	"auto_approve_4k_movie": 65536,
	"auto_approve_4k_tv":    131072,
	"request_movie":         262144,
	"request_tv":            524288,
	"manage_issues":         1048576,
	"view_issues":           2097152,
	"create_issues":         4194304,
	"auto_request":          8388608,
	"auto_request_movie":    16777216,
	"auto_request_tv":       33554432,
	"recent_view":           67108864,
	"watchlist_view":        134217728,
	"manage_blocklist":      268435456,
	"view_blocklist":        1073741824,
}

func NewPermissionSetDataSource() datasource.DataSource {
	return &permissionSetDataSource{}
}

type permissionSetDataSource struct{}

type permissionSetDataSourceModel struct {
	Permissions         types.Int64 `tfsdk:"permissions"`
	Admin               types.Bool  `tfsdk:"admin"`
	ManageSettings      types.Bool  `tfsdk:"manage_settings"`
	ManageUsers         types.Bool  `tfsdk:"manage_users"`
	ManageRequests      types.Bool  `tfsdk:"manage_requests"`
	Request             types.Bool  `tfsdk:"request"`
	Vote                types.Bool  `tfsdk:"vote"`
	AutoApprove         types.Bool  `tfsdk:"auto_approve"`
	AutoApproveMovie    types.Bool  `tfsdk:"auto_approve_movie"`
	AutoApproveTv       types.Bool  `tfsdk:"auto_approve_tv"`
	Request4K           types.Bool  `tfsdk:"request_4k"`
	Request4KMovie      types.Bool  `tfsdk:"request_4k_movie"`
	Request4KTv         types.Bool  `tfsdk:"request_4k_tv"`
	RequestAdvanced     types.Bool  `tfsdk:"request_advanced"`
	RequestView         types.Bool  `tfsdk:"request_view"`
	AutoApprove4K       types.Bool  `tfsdk:"auto_approve_4k"`
	AutoApprove4KMovie  types.Bool  `tfsdk:"auto_approve_4k_movie"`
	AutoApprove4KTv     types.Bool  `tfsdk:"auto_approve_4k_tv"`
	RequestMovie        types.Bool  `tfsdk:"request_movie"`
	RequestTv           types.Bool  `tfsdk:"request_tv"`
	ManageIssues        types.Bool  `tfsdk:"manage_issues"`
	ViewIssues          types.Bool  `tfsdk:"view_issues"`
	CreateIssues        types.Bool  `tfsdk:"create_issues"`
	AutoRequest         types.Bool  `tfsdk:"auto_request"`
	AutoRequestMovie    types.Bool  `tfsdk:"auto_request_movie"`
	AutoRequestTv       types.Bool  `tfsdk:"auto_request_tv"`
	RecentView          types.Bool  `tfsdk:"recent_view"`
	WatchlistView       types.Bool  `tfsdk:"watchlist_view"`
	ManageBlocklist     types.Bool  `tfsdk:"manage_blocklist"`
	ViewBlocklist       types.Bool  `tfsdk:"view_blocklist"`
}

func (d *permissionSetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permission_set"
}

func (d *permissionSetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := map[string]schema.Attribute{
		"permissions": schema.Int64Attribute{
			Computed:    true,
			Description: "The combined permission bitmask integer. Use this value with `seerr_user_permissions`.",
		},
	}

	for key := range PermissionsMap {
		attributes[key] = schema.BoolAttribute{
			Optional:    true,
			Description: "Permission: " + key,
		}
	}

	resp.Schema = schema.Schema{
		Description: "Compose a set of Seerr permissions into a single bitmask integer.\n\n" +
			"Common permission sets:\n" +
			"- **User**: `request`, `request_movie`, `request_tv` (Value: 786464)\n" +
			"- **Power User**: `request`, `request_movie`, `request_tv`, `request_4k`, `auto_approve` (Value: 787616)\n" +
			"- **Admin**: `admin` (Value: 2)",
		Attributes: attributes,
	}
}

func (d *permissionSetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data permissionSetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var total int64 = 0

	addIfTrue := func(b types.Bool, val int64) {
		if !b.IsNull() && !b.IsUnknown() && b.ValueBool() {
			total += val
		}
	}

	addIfTrue(data.Admin, PermissionsMap["admin"])
	addIfTrue(data.ManageSettings, PermissionsMap["manage_settings"])
	addIfTrue(data.ManageUsers, PermissionsMap["manage_users"])
	addIfTrue(data.ManageRequests, PermissionsMap["manage_requests"])
	addIfTrue(data.Request, PermissionsMap["request"])
	addIfTrue(data.Vote, PermissionsMap["vote"])
	addIfTrue(data.AutoApprove, PermissionsMap["auto_approve"])
	addIfTrue(data.AutoApproveMovie, PermissionsMap["auto_approve_movie"])
	addIfTrue(data.AutoApproveTv, PermissionsMap["auto_approve_tv"])
	addIfTrue(data.Request4K, PermissionsMap["request_4k"])
	addIfTrue(data.Request4KMovie, PermissionsMap["request_4k_movie"])
	addIfTrue(data.Request4KTv, PermissionsMap["request_4k_tv"])
	addIfTrue(data.RequestAdvanced, PermissionsMap["request_advanced"])
	addIfTrue(data.RequestView, PermissionsMap["request_view"])
	addIfTrue(data.AutoApprove4K, PermissionsMap["auto_approve_4k"])
	addIfTrue(data.AutoApprove4KMovie, PermissionsMap["auto_approve_4k_movie"])
	addIfTrue(data.AutoApprove4KTv, PermissionsMap["auto_approve_4k_tv"])
	addIfTrue(data.RequestMovie, PermissionsMap["request_movie"])
	addIfTrue(data.RequestTv, PermissionsMap["request_tv"])
	addIfTrue(data.ManageIssues, PermissionsMap["manage_issues"])
	addIfTrue(data.ViewIssues, PermissionsMap["view_issues"])
	addIfTrue(data.CreateIssues, PermissionsMap["create_issues"])
	addIfTrue(data.AutoRequest, PermissionsMap["auto_request"])
	addIfTrue(data.AutoRequestMovie, PermissionsMap["auto_request_movie"])
	addIfTrue(data.AutoRequestTv, PermissionsMap["auto_request_tv"])
	addIfTrue(data.RecentView, PermissionsMap["recent_view"])
	addIfTrue(data.WatchlistView, PermissionsMap["watchlist_view"])
	addIfTrue(data.ManageBlocklist, PermissionsMap["manage_blocklist"])
	addIfTrue(data.ViewBlocklist, PermissionsMap["view_blocklist"])

	data.Permissions = types.Int64Value(total)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
