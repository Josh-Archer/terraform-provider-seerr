package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &UserQuotaDataSource{}

// UserQuotaDataSource reads per-user quota settings and global quota defaults.
type UserQuotaDataSource struct {
	client *APIClient
}

// UserQuotaDataSourceModel mirrors UserQuotaModel without the id field (data sources don't need it).
type UserQuotaDataSourceModel struct {
	UserID                types.Int64 `tfsdk:"user_id"`
	MovieQuotaLimit       types.Int64 `tfsdk:"movie_quota_limit"`
	MovieQuotaDays        types.Int64 `tfsdk:"movie_quota_days"`
	TvQuotaLimit          types.Int64 `tfsdk:"tv_quota_limit"`
	TvQuotaDays           types.Int64 `tfsdk:"tv_quota_days"`
	GlobalMovieQuotaLimit types.Int64 `tfsdk:"global_movie_quota_limit"`
	GlobalMovieQuotaDays  types.Int64 `tfsdk:"global_movie_quota_days"`
	GlobalTvQuotaLimit    types.Int64 `tfsdk:"global_tv_quota_limit"`
	GlobalTvQuotaDays     types.Int64 `tfsdk:"global_tv_quota_days"`
}

func NewUserQuotaDataSource() datasource.DataSource { return &UserQuotaDataSource{} }

func (d *UserQuotaDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_quota"
}

func (d *UserQuotaDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read per-user movie and TV request quotas from Seerr via `/api/v1/user/{userId}/settings/main`.\n\n" +
			"A quota value of `0` means the user inherits the global instance default. " +
			"Use `global_movie_quota_limit` / `global_tv_quota_limit` to inspect those defaults.",
		Attributes: map[string]schema.Attribute{
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the user to look up.",
				Required:            true,
			},
			"movie_quota_limit": schema.Int64Attribute{
				MarkdownDescription: "Per-user movie request quota limit (`0` = inherit global).",
				Computed:            true,
			},
			"movie_quota_days": schema.Int64Attribute{
				MarkdownDescription: "Per-user movie quota rolling window in days (`0` = inherit global).",
				Computed:            true,
			},
			"tv_quota_limit": schema.Int64Attribute{
				MarkdownDescription: "Per-user TV request quota limit (`0` = inherit global).",
				Computed:            true,
			},
			"tv_quota_days": schema.Int64Attribute{
				MarkdownDescription: "Per-user TV quota rolling window in days (`0` = inherit global).",
				Computed:            true,
			},
			"global_movie_quota_limit": schema.Int64Attribute{
				MarkdownDescription: "Instance-wide default movie quota limit.",
				Computed:            true,
			},
			"global_movie_quota_days": schema.Int64Attribute{
				MarkdownDescription: "Instance-wide default movie quota period in days.",
				Computed:            true,
			},
			"global_tv_quota_limit": schema.Int64Attribute{
				MarkdownDescription: "Instance-wide default TV quota limit.",
				Computed:            true,
			},
			"global_tv_quota_days": schema.Int64Attribute{
				MarkdownDescription: "Instance-wide default TV quota period in days.",
				Computed:            true,
			},
		},
	}
}

func (d *UserQuotaDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserQuotaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserQuotaDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiPath := userQuotaPath(data.UserID.ValueInt64())
	res, err := d.client.Request(ctx, "GET", apiPath, "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	var decoded map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err != nil {
		resp.Diagnostics.AddError("Read Failed", "failed to parse response: "+err.Error())
		return
	}

	if v, ok := int64ValueFromAny(decoded["movieQuotaLimit"]); ok {
		data.MovieQuotaLimit = types.Int64Value(v)
	} else {
		data.MovieQuotaLimit = types.Int64Value(0)
	}
	if v, ok := int64ValueFromAny(decoded["movieQuotaDays"]); ok {
		data.MovieQuotaDays = types.Int64Value(v)
	} else {
		data.MovieQuotaDays = types.Int64Value(0)
	}
	if v, ok := int64ValueFromAny(decoded["tvQuotaLimit"]); ok {
		data.TvQuotaLimit = types.Int64Value(v)
	} else {
		data.TvQuotaLimit = types.Int64Value(0)
	}
	if v, ok := int64ValueFromAny(decoded["tvQuotaDays"]); ok {
		data.TvQuotaDays = types.Int64Value(v)
	} else {
		data.TvQuotaDays = types.Int64Value(0)
	}
	if v, ok := int64ValueFromAny(decoded["globalMovieQuotaLimit"]); ok {
		data.GlobalMovieQuotaLimit = types.Int64Value(v)
	} else {
		data.GlobalMovieQuotaLimit = types.Int64Null()
	}
	if v, ok := int64ValueFromAny(decoded["globalMovieQuotaDays"]); ok {
		data.GlobalMovieQuotaDays = types.Int64Value(v)
	} else {
		data.GlobalMovieQuotaDays = types.Int64Null()
	}
	if v, ok := int64ValueFromAny(decoded["globalTvQuotaLimit"]); ok {
		data.GlobalTvQuotaLimit = types.Int64Value(v)
	} else {
		data.GlobalTvQuotaLimit = types.Int64Null()
	}
	if v, ok := int64ValueFromAny(decoded["globalTvQuotaDays"]); ok {
		data.GlobalTvQuotaDays = types.Int64Value(v)
	} else {
		data.GlobalTvQuotaDays = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
