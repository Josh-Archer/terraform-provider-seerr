package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// newWholeRowSeerr serves a fake Seerr where every write handler loads the
// whole user row, pauses, merges its own fields, then saves the whole row
// back - the same lost-update shape as the real TypeORM handlers.
func newWholeRowSeerr(t *testing.T, fake *fakeSeerrUserRow) *APIClient {
	t.Helper()
	write := func(w http.ResponseWriter, r *http.Request, delay time.Duration, allowPermissions bool) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		row := fake.snapshot()
		time.Sleep(delay)
		for k, v := range body {
			if k == "id" || (k == "permissions" && !allowPermissions) {
				continue
			}
			row[k] = v
		}
		fake.save(row)
		_ = json.NewEncoder(w).Encode(row)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && (r.URL.Path == "/api/v1/user/1" ||
			r.URL.Path == "/api/v1/user/1/settings/main" ||
			r.URL.Path == "/api/v1/user/1/settings/notifications"):
			_ = json.NewEncoder(w).Encode(fake.snapshot())
		case r.Method == http.MethodPut && r.URL.Path == "/api/v1/user/1":
			write(w, r, 50*time.Millisecond, true)
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/user/1/settings/main":
			write(w, r, 150*time.Millisecond, false)
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/user/1/settings/notifications":
			write(w, r, 150*time.Millisecond, false)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	baseURL, _ := url.Parse(srv.URL)
	return NewClient(baseURL, "tok", "agent", false, defaultRequestTimeout, 0, 0)
}

func quotaModelFor(limit int64) *UserQuotaModel {
	return &UserQuotaModel{
		UserID:          types.Int64Value(1),
		MovieQuotaLimit: types.Int64Value(limit),
		MovieQuotaDays:  types.Int64Null(),
		TvQuotaLimit:    types.Int64Null(),
		TvQuotaDays:     types.Int64Null(),
	}
}

func userModelFor(username, locale string) UserModel {
	return UserModel{
		ID:                    types.StringValue("1"),
		Email:                 types.StringValue("u@example.com"),
		Username:              types.StringValue(username),
		PlexID:                types.StringNull(),
		PlexUsername:          types.StringNull(),
		PlexAuthToken:         types.StringNull(),
		JellyfinUsername:      types.StringNull(),
		JellyfinPassword:      types.StringNull(),
		UserType:              types.Int64Null(),
		Permissions:           types.Int64Null(),
		Locale:                types.StringValue(locale),
		DiscoverRegion:        types.StringNull(),
		StreamingRegion:       types.StringNull(),
		OriginalLanguage:      types.StringNull(),
		MovieQuotaLimit:       types.Int64Null(),
		MovieQuotaDays:        types.Int64Null(),
		TvQuotaLimit:          types.Int64Null(),
		TvQuotaDays:           types.Int64Null(),
		GlobalMovieQuotaDays:  types.Int64Null(),
		GlobalMovieQuotaLimit: types.Int64Null(),
		GlobalTvQuotaLimit:    types.Int64Null(),
		GlobalTvQuotaDays:     types.Int64Null(),
		WatchlistSyncMovies:   types.BoolNull(),
		WatchlistSyncTv:       types.BoolNull(),
	}
}

// seerr_user Update, seerr_user_permissions and seerr_user_quota for the same
// user applied in parallel must not revert each other's whole-row saves.
func TestUserResourceSerializesWithPermissionsAndQuota(t *testing.T) {
	fake := &fakeSeerrUserRow{row: map[string]any{
		"id": float64(1), "username": "old", "email": "u@example.com", "permissions": float64(2), "locale": "en",
	}}
	client := newWholeRowSeerr(t, fake)
	ctx := context.Background()

	users := &UserResource{client: client}
	var userSchema resource.SchemaResponse
	users.Schema(ctx, resource.SchemaRequest{}, &userSchema)
	userPlan := tfsdk.Plan{Schema: userSchema.Schema}
	if d := userPlan.Set(ctx, userModelFor("newname", "fr")); d.HasError() {
		t.Fatalf("user plan: %v", d)
	}
	userState := tfsdk.State{Schema: userSchema.Schema}
	if d := userState.Set(ctx, userModelFor("old", "en")); d.HasError() {
		t.Fatalf("user state: %v", d)
	}

	perms := &UserPermissionsResource{client: client}
	var permSchema resource.SchemaResponse
	perms.Schema(ctx, resource.SchemaRequest{}, &permSchema)
	permPlan := tfsdk.Plan{Schema: permSchema.Schema}
	if d := permPlan.Set(ctx, &UserPermissionsModel{
		ID: types.StringUnknown(), UserID: types.Int64Value(1), Permissions: types.Int64Value(26),
		ResponseJSON: types.StringUnknown(), StatusCode: types.Int64Unknown(),
	}); d.HasError() {
		t.Fatalf("perm plan: %v", d)
	}
	quota := &UserQuotaResource{client: client}

	userResp := resource.UpdateResponse{State: tfsdk.State{Schema: userSchema.Schema}}
	permResp := resource.CreateResponse{State: tfsdk.State{Schema: permSchema.Schema}}
	var quotaErr error
	var wg sync.WaitGroup
	wg.Add(3)
	// seerr_user_permissions writes first, then seerr_user_quota loads the row
	// and holds its copy for ~150ms. Unlocked, the seerr_user base PUT loads
	// the row before the quota save lands and saves after it, dropping the
	// quota limit.
	go func() {
		defer wg.Done()
		perms.Create(ctx, resource.CreateRequest{Plan: permPlan}, &permResp)
	}()
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
		quotaErr = quota.applyQuota(ctx, quotaModelFor(5))
	}()
	go func() {
		defer wg.Done()
		time.Sleep(180 * time.Millisecond)
		users.Update(ctx, resource.UpdateRequest{Plan: userPlan, State: userState}, &userResp)
	}()
	wg.Wait()

	if quotaErr != nil {
		t.Fatalf("applyQuota: %v", quotaErr)
	}
	if userResp.Diagnostics.HasError() {
		t.Fatalf("user update: %v", userResp.Diagnostics)
	}
	if permResp.Diagnostics.HasError() {
		t.Fatalf("permissions create: %v", permResp.Diagnostics)
	}
	row := fake.snapshot()
	if row["permissions"] != float64(26) || row["movieQuotaLimit"] != float64(5) ||
		row["username"] != "newname" || row["locale"] != "fr" {
		t.Fatalf("a concurrent same-user write was reverted; final row = %v", row)
	}
	var got UserModel
	if d := userResp.State.Get(ctx, &got); d.HasError() {
		t.Fatalf("user state: %v", d)
	}
	if got.Locale.ValueString() != "fr" || got.Username.ValueString() != "newname" {
		t.Fatalf("user read-back = %s/%s, want newname/fr", got.Username.ValueString(), got.Locale.ValueString())
	}
}

// seerr_user_notification_settings and seerr_user_quota for the same user
// applied in parallel must not revert each other's whole-row saves.
func TestUserNotificationSettingsSerializesWithQuota(t *testing.T) {
	fake := &fakeSeerrUserRow{row: map[string]any{"id": float64(1), "permissions": float64(2), "discordId": ""}}
	client := newWholeRowSeerr(t, fake)
	ctx := context.Background()

	notif := &UserNotificationSettingsResource{client: client}
	var schema resource.SchemaResponse
	notif.Schema(ctx, resource.SchemaRequest{}, &schema)
	plan := tfsdk.Plan{Schema: schema.Schema}
	if d := plan.Set(ctx, &UserNotificationSettingsResourceModel{
		ID: types.StringUnknown(), UserID: types.Int64Value(1),
		EmailEnabled: types.BoolNull(), PGPKey: types.StringNull(), DiscordEnabled: types.BoolNull(),
		DiscordID: types.StringValue("discord-123"), PushbulletAccessToken: types.StringNull(),
		PushoverApplicationToken: types.StringNull(), PushoverUserKey: types.StringNull(), PushoverSound: types.StringNull(),
		TelegramEnabled: types.BoolNull(), TelegramBotUsername: types.StringNull(), TelegramChatID: types.StringNull(),
		TelegramMessageThreadID: types.StringNull(), TelegramSendSilently: types.BoolNull(), WebpushEnabled: types.BoolNull(),
		ResponseJSON: types.StringUnknown(), StatusCode: types.Int64Unknown(),
	}); d.HasError() {
		t.Fatalf("plan: %v", d)
	}
	quota := &UserQuotaResource{client: client}

	resp := resource.CreateResponse{State: tfsdk.State{Schema: schema.Schema}}
	var quotaErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		quotaErr = quota.applyQuota(ctx, quotaModelFor(7))
	}()
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
		notif.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)
	}()
	wg.Wait()

	if quotaErr != nil {
		t.Fatalf("applyQuota: %v", quotaErr)
	}
	if resp.Diagnostics.HasError() {
		t.Fatalf("create: %v", resp.Diagnostics)
	}
	row := fake.snapshot()
	if row["discordId"] != "discord-123" || row["movieQuotaLimit"] != float64(7) {
		t.Fatalf("a concurrent same-user write was reverted; final row = %v", row)
	}
}
