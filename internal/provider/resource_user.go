package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

type UserResource struct {
	client *APIClient
}

type UserNotificationTypesModel struct {
	Discord    types.Int64 `tfsdk:"discord"`
	Email      types.Int64 `tfsdk:"email"`
	Pushbullet types.Int64 `tfsdk:"pushbullet"`
	Pushover   types.Int64 `tfsdk:"pushover"`
	Slack      types.Int64 `tfsdk:"slack"`
	Telegram   types.Int64 `tfsdk:"telegram"`
	Webhook    types.Int64 `tfsdk:"webhook"`
	Webpush    types.Int64 `tfsdk:"webpush"`
	Gotify     types.Int64 `tfsdk:"gotify"`
	Ntfy       types.Int64 `tfsdk:"ntfy"`
}

type UserNotificationSettingsModel struct {
	EmailEnabled             types.Bool                  `tfsdk:"email_enabled"`
	PGPKey                   types.String                `tfsdk:"pgp_key"`
	DiscordEnabled           types.Bool                  `tfsdk:"discord_enabled"`
	DiscordID                types.String                `tfsdk:"discord_id"`
	PushbulletAccessToken    types.String                `tfsdk:"pushbullet_access_token"`
	PushoverApplicationToken types.String                `tfsdk:"pushover_application_token"`
	PushoverUserKey          types.String                `tfsdk:"pushover_user_key"`
	PushoverSound            types.String                `tfsdk:"pushover_sound"`
	TelegramEnabled          types.Bool                  `tfsdk:"telegram_enabled"`
	TelegramBotUsername      types.String                `tfsdk:"telegram_bot_username"`
	TelegramChatID           types.String                `tfsdk:"telegram_chat_id"`
	TelegramMessageThreadID  types.String                `tfsdk:"telegram_message_thread_id"`
	TelegramSendSilently     types.Bool                  `tfsdk:"telegram_send_silently"`
	WebpushEnabled           types.Bool                  `tfsdk:"webpush_enabled"`
	NotificationTypes        *UserNotificationTypesModel `tfsdk:"notification_types"`
}

type UserModel struct {
	ID                   types.String                   `tfsdk:"id"`
	Email                types.String                   `tfsdk:"email"`
	Username             types.String                   `tfsdk:"username"`
	PlexID               types.String                   `tfsdk:"plex_id"`
	Permissions          types.Int64                    `tfsdk:"permissions"`
	Locale               types.String                   `tfsdk:"locale"`
	DiscoverRegion       types.String                   `tfsdk:"discover_region"`
	StreamingRegion      types.String                   `tfsdk:"streaming_region"`
	OriginalLanguage     types.String                   `tfsdk:"original_language"`
	WatchlistSyncMovies  types.Bool                     `tfsdk:"watchlist_sync_movies"`
	WatchlistSyncTv      types.Bool                     `tfsdk:"watchlist_sync_tv"`
	NotificationSettings *UserNotificationSettingsModel `tfsdk:"notification_settings"`
}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

func (r *UserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Seerr users and their notification settings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "User's email address. Field is ForceNew because Overseerr API doesn't support updating it.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "User's display name. Can be imported from Plex if `plex_id` is provided.",
				Required:            true,
			},
			"plex_id": schema.StringAttribute{
				MarkdownDescription: "Optional Plex ID to import a user directly from Plex.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"permissions": schema.Int64Attribute{
				MarkdownDescription: "Permissions bitmask.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"locale": schema.StringAttribute{
				MarkdownDescription: "User's preferred locale.",
				Optional:            true,
				Computed:            true,
			},
			"discover_region": schema.StringAttribute{
				MarkdownDescription: "User's preferred discovery region.",
				Optional:            true,
				Computed:            true,
			},
			"streaming_region": schema.StringAttribute{
				MarkdownDescription: "User's preferred streaming region.",
				Optional:            true,
				Computed:            true,
			},
			"original_language": schema.StringAttribute{
				MarkdownDescription: "User's preferred original language.",
				Optional:            true,
				Computed:            true,
			},
			"watchlist_sync_movies": schema.BoolAttribute{
				MarkdownDescription: "Enable watchlist sync for movies.",
				Optional:            true,
				Computed:            true,
			},
			"watchlist_sync_tv": schema.BoolAttribute{
				MarkdownDescription: "Enable watchlist sync for TV shows.",
				Optional:            true,
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"notification_settings": schema.SingleNestedBlock{
				MarkdownDescription: "User-specific notification settings.",
				Attributes: map[string]schema.Attribute{
					"email_enabled":              schema.BoolAttribute{Optional: true, Computed: true},
					"pgp_key":                    schema.StringAttribute{Optional: true},
					"discord_enabled":            schema.BoolAttribute{Optional: true, Computed: true},
					"discord_id":                 schema.StringAttribute{Optional: true},
					"pushbullet_access_token":    schema.StringAttribute{Optional: true, Sensitive: true},
					"pushover_application_token": schema.StringAttribute{Optional: true, Sensitive: true},
					"pushover_user_key":          schema.StringAttribute{Optional: true, Sensitive: true},
					"pushover_sound":             schema.StringAttribute{Optional: true},
					"telegram_enabled":           schema.BoolAttribute{Optional: true, Computed: true},
					"telegram_bot_username":      schema.StringAttribute{Optional: true},
					"telegram_chat_id":           schema.StringAttribute{Optional: true},
					"telegram_message_thread_id": schema.StringAttribute{Optional: true},
					"telegram_send_silently":     schema.BoolAttribute{Optional: true, Computed: true},
					"webpush_enabled":            schema.BoolAttribute{Computed: true},
				},
				Blocks: map[string]schema.Block{
					"notification_types": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"discord":    schema.Int64Attribute{Optional: true, Computed: true},
							"email":      schema.Int64Attribute{Optional: true, Computed: true},
							"pushbullet": schema.Int64Attribute{Optional: true, Computed: true},
							"pushover":   schema.Int64Attribute{Optional: true, Computed: true},
							"slack":      schema.Int64Attribute{Optional: true, Computed: true},
							"telegram":   schema.Int64Attribute{Optional: true, Computed: true},
							"webhook":    schema.Int64Attribute{Optional: true, Computed: true},
							"webpush":    schema.Int64Attribute{Optional: true, Computed: true},
							"gotify":     schema.Int64Attribute{Optional: true, Computed: true},
							"ntfy":       schema.Int64Attribute{Optional: true, Computed: true},
						},
					},
				},
			},
		},
	}
}

func (r *UserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var userIDStr string

	// Create or Import User
	if !data.PlexID.IsNull() && !data.PlexID.IsUnknown() && data.PlexID.ValueString() != "" {
		// Import from Plex
		importBody, _ := json.Marshal(map[string]any{
			"plexIds": []string{data.PlexID.ValueString()},
		})
		res, err := r.client.Request(ctx, "POST", "/api/v1/user/import-from-plex", string(importBody), nil)
		if err != nil {
			resp.Diagnostics.AddError("Import Plex User Failed", err.Error())
			return
		}
		if !StatusIsOK(res.StatusCode) {
			resp.Diagnostics.AddError("Import Plex User Failed", fmt.Sprintf("Status %d: %s", res.StatusCode, string(res.Body)))
			return
		}

		// Response should be an array of imported users. We need to find the ID of the new user.
		var importedUsers []map[string]any
		if err := json.Unmarshal(res.Body, &importedUsers); err != nil {
			resp.Diagnostics.AddError("Import Plex User Failed", "Failed to parse API response: "+err.Error())
			return
		}

		if len(importedUsers) == 0 {
			resp.Diagnostics.AddError("Import Plex User Failed", "API returned an empty array of imported users.")
			return
		}

		importedIDRaw := importedUsers[0]["id"]
		switch v := importedIDRaw.(type) {
		case float64:
			userIDStr = fmt.Sprintf("%.0f", v)
		case string:
			userIDStr = v
		default:
			resp.Diagnostics.AddError("Import Plex User Failed", "Could not extract user ID from import response.")
			return
		}

		// Immediately after import, update the permissions to match what Terraform expects (since import might not set them)
		updateBody, _ := json.Marshal(map[string]any{
			"permissions": data.Permissions.ValueInt64(),
		})
		_, _ = r.client.Request(ctx, "PUT", "/api/v1/user/"+userIDStr, string(updateBody), nil)

	} else {
		// Create Local User
		createBody, _ := json.Marshal(map[string]any{
			"email":       data.Email.ValueString(),
			"username":    data.Username.ValueString(),
			"permissions": data.Permissions.ValueInt64(),
		})

		res, err := r.client.Request(ctx, "POST", "/api/v1/user", string(createBody), nil)
		if err != nil {
			resp.Diagnostics.AddError("Create Failed", err.Error())
			return
		}
		if !StatusIsOK(res.StatusCode) {
			resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("Status %d: %s", res.StatusCode, string(res.Body)))
			return
		}

		extractedID, ok := ExtractIDFromJSON(res.Body)
		if !ok {
			resp.Diagnostics.AddError("Create Failed", "Could not extract user ID from response")
			return
		}
		userIDStr = extractedID
	}

	data.ID = types.StringValue(userIDStr)

	// Update main settings
	if err := r.updateMainSettings(ctx, userIDStr, &data); err != nil {
		resp.Diagnostics.AddError("Update Main Settings Failed", err.Error())
		return
	}

	// Update Notification Settings
	if data.NotificationSettings != nil {
		if err := r.updateNotificationSettings(ctx, userIDStr, data.NotificationSettings); err != nil {
			resp.Diagnostics.AddError("Update Notification Settings Failed", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userID := data.ID.ValueString()
	res, err := r.client.Request(ctx, "GET", "/api/v1/user/"+userID, "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if res.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("Status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	var userMap map[string]any
	if err := json.Unmarshal(res.Body, &userMap); err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	if email, ok := userMap["email"].(string); ok {
		data.Email = types.StringValue(strings.ToLower(email))
	}
	if username, ok := userMap["username"].(string); ok {
		data.Username = types.StringValue(strings.ToLower(username))
	}
	if p, ok := userMap["permissions"].(float64); ok {
		data.Permissions = types.Int64Value(int64(p))
	}

	// Read Main Settings
	mainRes, err := r.client.Request(ctx, "GET", fmt.Sprintf("/api/v1/user/%s/settings/main", userID), "", nil)
	if err == nil && StatusIsOK(mainRes.StatusCode) {
		var mainMap map[string]any
		if err := json.Unmarshal(mainRes.Body, &mainMap); err == nil {
			if v, ok := mainMap["locale"].(string); ok {
				data.Locale = types.StringValue(v)
			}
			if v, ok := mainMap["discoverRegion"].(string); ok {
				data.DiscoverRegion = types.StringValue(v)
			}
			if v, ok := mainMap["streamingRegion"].(string); ok {
				data.StreamingRegion = types.StringValue(v)
			}
			if v, ok := mainMap["originalLanguage"].(string); ok {
				data.OriginalLanguage = types.StringValue(v)
			}
			if v, ok := mainMap["watchlistSyncMovies"].(bool); ok {
				data.WatchlistSyncMovies = types.BoolValue(v)
			}
			if v, ok := mainMap["watchlistSyncTv"].(bool); ok {
				data.WatchlistSyncTv = types.BoolValue(v)
			}
		}
	}

	// Read Notification Settings
	notifRes, err := r.client.Request(ctx, "GET", fmt.Sprintf("/api/v1/user/%s/settings/notifications", userID), "", nil)
	if err == nil && StatusIsOK(notifRes.StatusCode) {
		var notifMap map[string]any
		if err := json.Unmarshal(notifRes.Body, &notifMap); err == nil {
			data.NotificationSettings = r.mapNotificationSettings(notifMap)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userID := data.ID.ValueString()

	// Update Base Info
	updateBody, _ := json.Marshal(map[string]any{
		"username":    data.Username.ValueString(),
		"permissions": data.Permissions.ValueInt64(),
	})

	res, err := r.client.Request(ctx, "PUT", "/api/v1/user/"+userID, string(updateBody), nil)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Update Failed", fmt.Sprintf("Status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	// Update main settings
	if err := r.updateMainSettings(ctx, userID, &data); err != nil {
		resp.Diagnostics.AddError("Update Main Settings Failed", err.Error())
		return
	}

	// Update Notification Settings
	if data.NotificationSettings != nil {
		if err := r.updateNotificationSettings(ctx, userID, data.NotificationSettings); err != nil {
			resp.Diagnostics.AddError("Update Notification Settings Failed", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.Request(ctx, "DELETE", "/api/v1/user/"+data.ID.ValueString(), "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Delete Failed", err.Error())
		return
	}
	if res.StatusCode != 404 && !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Delete Failed", fmt.Sprintf("Status %d: %s", res.StatusCode, string(res.Body)))
		return
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID

	// If ID is not an integer, try to look up user by username or email
	if _, err := strconv.Atoi(id); err != nil {
		res, err := r.client.Request(ctx, "GET", "/api/v1/user?take=1000", "", nil)
		if err != nil {
			resp.Diagnostics.AddError("Import Failed", fmt.Sprintf("failed to fetch users for lookup: %s", err))
			return
		}
		if !StatusIsOK(res.StatusCode) {
			resp.Diagnostics.AddError("Import Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
			return
		}

		var parsedResponse struct {
			Results []map[string]any `json:"results"`
		}
		if err := json.Unmarshal(res.Body, &parsedResponse); err != nil {
			resp.Diagnostics.AddError("Import Failed", "Failed to parse API response: "+err.Error())
			return
		}

		searchID := ""
		searchLower := strings.ToLower(id)
		for _, u := range parsedResponse.Results {
			if un, ok := u["username"].(string); ok && strings.ToLower(un) == searchLower {
				searchID = fmt.Sprintf("%.0f", u["id"].(float64))
				break
			}
			if e, ok := u["email"].(string); ok && strings.ToLower(e) == searchLower {
				searchID = fmt.Sprintf("%.0f", u["id"].(float64))
				break
			}
		}

		if searchID == "" {
			resp.Diagnostics.AddError("Import Failed", fmt.Sprintf("could not find user matching %q as id, username, or email", id))
			return
		}
		id = searchID
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func (r *UserResource) updateMainSettings(ctx context.Context, userID string, data *UserModel) error {
	payload := map[string]any{
		"username": data.Username.ValueString(),
		"email":    data.Email.ValueString(),
	}

	if !data.Locale.IsNull() && !data.Locale.IsUnknown() {
		payload["locale"] = data.Locale.ValueString()
	}
	if !data.DiscoverRegion.IsNull() && !data.DiscoverRegion.IsUnknown() {
		payload["discoverRegion"] = data.DiscoverRegion.ValueString()
	}
	if !data.StreamingRegion.IsNull() && !data.StreamingRegion.IsUnknown() {
		payload["streamingRegion"] = data.StreamingRegion.ValueString()
	}
	if !data.OriginalLanguage.IsNull() && !data.OriginalLanguage.IsUnknown() {
		payload["originalLanguage"] = data.OriginalLanguage.ValueString()
	}
	if !data.WatchlistSyncMovies.IsNull() && !data.WatchlistSyncMovies.IsUnknown() {
		payload["watchlistSyncMovies"] = data.WatchlistSyncMovies.ValueBool()
	}
	if !data.WatchlistSyncTv.IsNull() && !data.WatchlistSyncTv.IsUnknown() {
		payload["watchlistSyncTv"] = data.WatchlistSyncTv.ValueBool()
	}

	body, _ := json.Marshal(payload)
	path := fmt.Sprintf("/api/v1/user/%s/settings/main", userID)
	res, err := r.client.Request(ctx, "POST", path, string(body), nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("main settings status %d: %s", res.StatusCode, string(res.Body))
	}
	return nil
}

func (r *UserResource) updateNotificationSettings(ctx context.Context, userID string, settings *UserNotificationSettingsModel) error {
	payload := map[string]any{
		"emailEnabled":             settings.EmailEnabled.ValueBool(),
		"pgpKey":                   settings.PGPKey.ValueString(),
		"discordEnabled":           settings.DiscordEnabled.ValueBool(),
		"discordId":                settings.DiscordID.ValueString(),
		"pushbulletAccessToken":    settings.PushbulletAccessToken.ValueString(),
		"pushoverApplicationToken": settings.PushoverApplicationToken.ValueString(),
		"pushoverUserKey":          settings.PushoverUserKey.ValueString(),
		"pushoverSound":            settings.PushoverSound.ValueString(),
		"telegramEnabled":          settings.TelegramEnabled.ValueBool(),
		"telegramBotUsername":      settings.TelegramBotUsername.ValueString(),
		"telegramChatId":           settings.TelegramChatID.ValueString(),
		"telegramMessageThreadId":  settings.TelegramMessageThreadID.ValueString(),
		"telegramSendSilently":     settings.TelegramSendSilently.ValueBool(),
	}

	if settings.NotificationTypes != nil {
		payload["notificationTypes"] = map[string]any{
			"discord":    settings.NotificationTypes.Discord.ValueInt64(),
			"email":      settings.NotificationTypes.Email.ValueInt64(),
			"pushbullet": settings.NotificationTypes.Pushbullet.ValueInt64(),
			"pushover":   settings.NotificationTypes.Pushover.ValueInt64(),
			"slack":      settings.NotificationTypes.Slack.ValueInt64(),
			"telegram":   settings.NotificationTypes.Telegram.ValueInt64(),
			"webhook":    settings.NotificationTypes.Webhook.ValueInt64(),
			"webpush":    settings.NotificationTypes.Webpush.ValueInt64(),
			"gotify":     settings.NotificationTypes.Gotify.ValueInt64(),
			"ntfy":       settings.NotificationTypes.Ntfy.ValueInt64(),
		}
	}

	body, _ := json.Marshal(payload)
	path := fmt.Sprintf("/api/v1/user/%s/settings/notifications", userID)
	res, err := r.client.Request(ctx, "POST", path, string(body), nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("notification settings status %d: %s", res.StatusCode, string(res.Body))
	}
	return nil
}

func (r *UserResource) mapNotificationSettings(notifMap map[string]any) *UserNotificationSettingsModel {
	settings := &UserNotificationSettingsModel{}

	if v, ok := notifMap["emailEnabled"].(bool); ok {
		settings.EmailEnabled = types.BoolValue(v)
	}
	if v, ok := notifMap["pgpKey"].(string); ok {
		settings.PGPKey = types.StringValue(v)
	}
	if v, ok := notifMap["discordEnabled"].(bool); ok {
		settings.DiscordEnabled = types.BoolValue(v)
	}
	if v, ok := notifMap["discordId"].(string); ok {
		settings.DiscordID = types.StringValue(v)
	}
	if v, ok := notifMap["pushbulletAccessToken"].(string); ok {
		settings.PushbulletAccessToken = types.StringValue(v)
	}
	if v, ok := notifMap["pushoverApplicationToken"].(string); ok {
		settings.PushoverApplicationToken = types.StringValue(v)
	}
	if v, ok := notifMap["pushoverUserKey"].(string); ok {
		settings.PushoverUserKey = types.StringValue(v)
	}
	if v, ok := notifMap["pushoverSound"].(string); ok {
		settings.PushoverSound = types.StringValue(v)
	}
	if v, ok := notifMap["telegramEnabled"].(bool); ok {
		settings.TelegramEnabled = types.BoolValue(v)
	}
	if v, ok := notifMap["telegramBotUsername"].(string); ok {
		settings.TelegramBotUsername = types.StringValue(v)
	}
	if v, ok := notifMap["telegramChatId"].(string); ok {
		settings.TelegramChatID = types.StringValue(v)
	}
	if v, ok := notifMap["telegramSendSilently"].(bool); ok {
		settings.TelegramSendSilently = types.BoolValue(v)
	}
	if v, ok := notifMap["telegramMessageThreadId"].(string); ok {
		settings.TelegramMessageThreadID = types.StringValue(v)
	}

	if typesMap, ok := notifMap["notificationTypes"].(map[string]any); ok {
		settings.NotificationTypes = &UserNotificationTypesModel{
			Discord:    r.toInt64(typesMap["discord"]),
			Email:      r.toInt64(typesMap["email"]),
			Pushbullet: r.toInt64(typesMap["pushbullet"]),
			Pushover:   r.toInt64(typesMap["pushover"]),
			Slack:      r.toInt64(typesMap["slack"]),
			Telegram:   r.toInt64(typesMap["telegram"]),
			Webhook:    r.toInt64(typesMap["webhook"]),
			Webpush:    r.toInt64(typesMap["webpush"]),
			Gotify:     r.toInt64(typesMap["gotify"]),
			Ntfy:       r.toInt64(typesMap["ntfy"]),
		}
	}

	return settings
}

func (r *UserResource) toInt64(v any) types.Int64 {
	switch val := v.(type) {
	case float64:
		return types.Int64Value(int64(val))
	case int64:
		return types.Int64Value(val)
	case string:
		i, _ := strconv.ParseInt(val, 10, 64)
		return types.Int64Value(i)
	}
	return types.Int64Null()
}
