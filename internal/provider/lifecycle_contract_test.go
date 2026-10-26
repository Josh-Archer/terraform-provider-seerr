package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Lifecycle contract tests exercise import/refresh, null versus empty, external
// deletion, retries, and secret handling through httptest. They sit beside the
// existing resource tests and do not replace them. Nothing here talks to a live
// Seerr, and nothing sets TF_ACC.

const lifecycleProviderKey = "lifecycle-provider-key"

type lifecycleServer struct {
	t         *testing.T
	srv       *httptest.Server
	mu        sync.Mutex
	hits      int
	badAPIKey bool
}

func newLifecycleServer(t *testing.T, handler func(http.ResponseWriter, *http.Request, int)) *lifecycleServer {
	t.Helper()
	ls := &lifecycleServer{t: t}
	ls.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ls.mu.Lock()
		ls.hits++
		n := ls.hits
		if r.Header.Get("X-Api-Key") != lifecycleProviderKey {
			ls.badAPIKey = true
		}
		ls.mu.Unlock()
		handler(w, r, n)
	}))
	t.Cleanup(ls.srv.Close)
	return ls
}

func (ls *lifecycleServer) client(retries int) *APIClient {
	ls.t.Helper()
	parsed, err := url.Parse(ls.srv.URL)
	if err != nil {
		ls.t.Fatal(err)
	}
	return NewClient(parsed, lifecycleProviderKey, "lifecycle-contract", false, defaultRequestTimeout, retries, time.Millisecond)
}

func (ls *lifecycleServer) Hits() int {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	return ls.hits
}

func (ls *lifecycleServer) requireProviderKey(t *testing.T) {
	t.Helper()
	ls.mu.Lock()
	defer ls.mu.Unlock()
	if ls.hits == 0 {
		t.Fatal("expected at least one upstream request")
	}
	if ls.badAPIKey {
		t.Fatal("upstream request did not send the provider API key")
	}
}

func diagnosticsText(diags diag.Diagnostics) string {
	var b strings.Builder
	for _, d := range diags {
		fmt.Fprintf(&b, "%s: %s\n", d.Summary(), d.Detail())
	}
	return b.String()
}

func requireNoErrors(t *testing.T, diags diag.Diagnostics) {
	t.Helper()
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics:\n%s", diagnosticsText(diags))
	}
}

func requireNoSecretLeak(t *testing.T, diags diag.Diagnostics, secrets ...string) {
	t.Helper()
	text := diagnosticsText(diags)
	for _, secret := range secrets {
		if secret != "" && strings.Contains(text, secret) {
			t.Fatalf("diagnostics leaked secret %q:\n%s", secret, text)
		}
	}
}

func lifecycleTypeName(t *testing.T, res resource.Resource) string {
	t.Helper()
	var resp resource.MetadataResponse
	res.Metadata(t.Context(), resource.MetadataRequest{ProviderTypeName: "seerr"}, &resp)
	if resp.TypeName == "" {
		t.Fatalf("%T has an empty type name", res)
	}
	return resp.TypeName
}

func lifecycleSchema(t *testing.T, res resource.Resource) schema.Schema {
	t.Helper()
	var resp resource.SchemaResponse
	res.Schema(t.Context(), resource.SchemaRequest{}, &resp)
	requireNoErrors(t, resp.Diagnostics)
	return resp.Schema
}

func lifecycleConfigure(t *testing.T, res resource.Resource, client *APIClient) {
	t.Helper()
	configurable, ok := res.(resource.ResourceWithConfigure)
	if !ok {
		t.Fatalf("%T does not implement resource.ResourceWithConfigure", res)
	}
	var resp resource.ConfigureResponse
	configurable.Configure(t.Context(), resource.ConfigureRequest{ProviderData: client}, &resp)
	requireNoErrors(t, resp.Diagnostics)
}

func lifecycleState(t *testing.T, res resource.Resource, model any) tfsdk.State {
	t.Helper()
	state := lifecycleTypedState(t, res)
	if model != nil {
		requireNoErrors(t, state.Set(t.Context(), model))
	}
	return state
}

// lifecycleTypedState is an object whose every attribute is a typed null.
// ImportStatePassthroughID only writes the id. Starting from an untyped raw
// value makes the following Read write attributes that have no element type.
func lifecycleTypedState(t *testing.T, res resource.Resource) tfsdk.State {
	t.Helper()
	state := tfsdk.State{Schema: lifecycleSchema(t, res)}
	state.Raw = nullSchemaValue(state.Schema.Type().TerraformType(t.Context()))
	return state
}

func nullSchemaValue(typ tftypes.Type) tftypes.Value {
	obj, ok := typ.(tftypes.Object)
	if !ok {
		return tftypes.NewValue(typ, nil)
	}
	attrs := make(map[string]tftypes.Value, len(obj.AttributeTypes))
	for name, attrType := range obj.AttributeTypes {
		attrs[name] = tftypes.NewValue(attrType, nil)
	}
	return tftypes.NewValue(obj, attrs)
}

func lifecycleRead(t *testing.T, res resource.Resource, state tfsdk.State) resource.ReadResponse {
	t.Helper()
	resp := resource.ReadResponse{State: state}
	res.Read(t.Context(), resource.ReadRequest{State: state}, &resp)
	return resp
}

func lifecycleImport(t *testing.T, res resource.Resource, id string) resource.ImportStateResponse {
	t.Helper()
	importer, ok := res.(resource.ResourceWithImportState)
	if !ok {
		t.Fatalf("%T does not implement resource.ResourceWithImportState", res)
	}
	resp := resource.ImportStateResponse{State: lifecycleTypedState(t, res)}
	importer.ImportState(t.Context(), resource.ImportStateRequest{ID: id}, &resp)
	return resp
}

func writeJSON(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}

func okHandler(t *testing.T, path, body string) func(http.ResponseWriter, *http.Request, int) {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request, _ int) {
		if r.URL.Path != path {
			t.Errorf("path = %s, want %s", r.URL.Path, path)
			http.NotFound(w, r)
			return
		}
		writeJSON(w, body)
	}
}

func configureRead(t *testing.T, res resource.Resource, model any, path, body string) resource.ReadResponse {
	t.Helper()
	srv := newLifecycleServer(t, okHandler(t, path, body))
	lifecycleConfigure(t, res, srv.client(0))
	resp := lifecycleRead(t, res, lifecycleState(t, res, model))
	requireNoErrors(t, resp.Diagnostics)
	srv.requireProviderKey(t)
	return resp
}

func importThenRead(t *testing.T, res resource.Resource, importID, path, body string, mutate func(*testing.T, *tfsdk.State)) resource.ReadResponse {
	t.Helper()
	srv := newLifecycleServer(t, okHandler(t, path, body))
	lifecycleConfigure(t, res, srv.client(0))
	imported := lifecycleImport(t, res, importID)
	requireNoErrors(t, imported.Diagnostics)
	if mutate != nil {
		mutate(t, &imported.State)
	}
	resp := lifecycleRead(t, res, imported.State)
	requireNoErrors(t, resp.Diagnostics)
	srv.requireProviderKey(t)
	requireNoSecretLeak(t, resp.Diagnostics, lifecycleProviderKey)
	return resp
}

func decodeState[T any](t *testing.T, state tfsdk.State) T {
	t.Helper()
	var out T
	requireNoErrors(t, state.Get(t.Context(), &out))
	return out
}

func assertString(t *testing.T, got types.String, wantNull bool, want string) {
	t.Helper()
	if wantNull {
		if !got.IsNull() {
			t.Fatalf("string = %q, want null", got.ValueString())
		}
		return
	}
	if got.IsNull() {
		t.Fatalf("string is null, want %q", want)
	}
	if got.ValueString() != want {
		t.Fatalf("string = %q, want %q", got.ValueString(), want)
	}
}

func assertRemoved(t *testing.T, resp resource.ReadResponse) {
	t.Helper()
	requireNoErrors(t, resp.Diagnostics)
	if !resp.State.Raw.IsNull() {
		t.Fatalf("expected removed state, raw = %s", resp.State.Raw)
	}
}

func assertKept(t *testing.T, resp resource.ReadResponse) {
	t.Helper()
	if resp.State.Raw.IsNull() {
		t.Fatal("expected state to be kept")
	}
}

func collectSensitiveAttribute(name string, attr schema.Attribute, out *[]string) {
	if attr.IsSensitive() {
		*out = append(*out, name)
	}
	switch nested := attr.(type) {
	case schema.SingleNestedAttribute:
		for childName, child := range nested.Attributes {
			collectSensitiveAttribute(name+"."+childName, child, out)
		}
	case schema.ListNestedAttribute:
		for childName, child := range nested.NestedObject.Attributes {
			collectSensitiveAttribute(name+"."+childName, child, out)
		}
	case schema.SetNestedAttribute:
		for childName, child := range nested.NestedObject.Attributes {
			collectSensitiveAttribute(name+"."+childName, child, out)
		}
	case schema.MapNestedAttribute:
		for childName, child := range nested.NestedObject.Attributes {
			collectSensitiveAttribute(name+"."+childName, child, out)
		}
	}
}

func collectSensitiveBlock(name string, block schema.Block, out *[]string) {
	switch nested := block.(type) {
	case schema.SingleNestedBlock:
		for childName, child := range nested.Attributes {
			collectSensitiveAttribute(name+"."+childName, child, out)
		}
		for childName, child := range nested.Blocks {
			collectSensitiveBlock(name+"."+childName, child, out)
		}
	case schema.ListNestedBlock:
		for childName, child := range nested.NestedObject.Attributes {
			collectSensitiveAttribute(name+"."+childName, child, out)
		}
		for childName, child := range nested.NestedObject.Blocks {
			collectSensitiveBlock(name+"."+childName, child, out)
		}
	case schema.SetNestedBlock:
		for childName, child := range nested.NestedObject.Attributes {
			collectSensitiveAttribute(name+"."+childName, child, out)
		}
		for childName, child := range nested.NestedObject.Blocks {
			collectSensitiveBlock(name+"."+childName, child, out)
		}
	}
}

func lifecycleSensitivePaths(t *testing.T, res resource.Resource) []string {
	t.Helper()
	sch := lifecycleSchema(t, res)
	var paths []string
	for name, attr := range sch.Attributes {
		collectSensitiveAttribute(name, attr, &paths)
	}
	for name, block := range sch.Blocks {
		collectSensitiveBlock(name, block, &paths)
	}
	sort.Strings(paths)
	return paths
}

func issuePrior(id string) *IssueModel {
	return &IssueModel{
		ID:         types.StringValue(id),
		CreatedBy:  types.ObjectNull(issueUserAttrTypes),
		ModifiedBy: types.ObjectNull(issueUserAttrTypes),
	}
}

func requestPrior(id string) *RequestModel {
	return &RequestModel{
		ID:          types.StringValue(id),
		Seasons:     types.ListNull(types.Int64Type),
		RequestedBy: types.ObjectNull(requestUserAttrTypes),
		ModifiedBy:  types.ObjectNull(requestUserAttrTypes),
	}
}

func sonarrPrior(apiKey types.String) *SonarrServerModel {
	const id int64 = 4
	return &SonarrServerModel{
		ID:        types.StringValue(fmt.Sprintf("%d", id)),
		ServerID:  types.Int64Value(id),
		APIKey:    apiKey,
		Tags:      types.ListNull(types.Int64Type),
		AnimeTags: types.ListNull(types.Int64Type),
	}
}

func radarrPrior(apiKey types.String) *RadarrServerModel {
	const id int64 = 3
	return &RadarrServerModel{
		ID:       types.StringValue(fmt.Sprintf("%d", id)),
		ServerID: types.Int64Value(id),
		APIKey:   apiKey,
		Tags:     types.ListNull(types.Int64Type),
	}
}

func apiObjectPrior(suppress bool) *APIObjectModel {
	return &APIObjectModel{
		ID:               types.StringValue("/api/v1/settings/main"),
		Path:             types.StringValue("/api/v1/settings/main"),
		Headers:          types.MapNull(types.StringType),
		ReadMethod:       types.StringValue(http.MethodGet),
		SuppressNotFound: types.BoolValue(suppress),
	}
}

func TestLifecycleContractResourcesSupportImport(t *testing.T) {
	for _, newRes := range RegisteredResources {
		res := newRes()
		name := lifecycleTypeName(t, res)
		if _, ok := res.(resource.ResourceWithImportState); !ok {
			t.Errorf("%s does not implement resource.ResourceWithImportState", name)
		}
	}
}

func TestLifecycleContractSensitiveSchema(t *testing.T) {
	want := map[string][]string{
		"seerr_api_key":            {"api_key"},
		"seerr_emby_settings":      {"api_key"},
		"seerr_jellyfin_settings":  {"api_key"},
		"seerr_tautulli_settings":  {"api_key"},
		"seerr_network_settings":   {"proxy.password"},
		"seerr_radarr_server":      {"api_key"},
		"seerr_radarr_server_test": {"api_key"},
		"seerr_sonarr_server":      {"api_key"},
		"seerr_sonarr_server_test": {"api_key"},
		"seerr_user": {
			"jellyfin_password",
			"notification_settings.pushbullet_access_token",
			"notification_settings.pushover_application_token",
			"notification_settings.pushover_user_key",
			"plex_auth_token",
		},
		"seerr_user_notification_settings": {
			"pushbullet_access_token",
			"pushover_application_token",
			"pushover_user_key",
		},
		"seerr_notification_email": {
			"email.auth_pass",
			"email.pgp_password",
			"email.pgp_private_key",
		},
		"seerr_notification_discord":    {"discord.webhook_url"},
		"seerr_notification_gotify":     {"gotify.token"},
		"seerr_notification_lunasea":    {"lunasea.webhook_url"},
		"seerr_notification_ntfy":       {"ntfy.password", "ntfy.token"},
		"seerr_notification_pushbullet": {"pushbullet.access_token"},
		"seerr_notification_pushover":   {"pushover.access_token", "pushover.user_token"},
		"seerr_notification_telegram":   {"telegram.bot_api"},
		"seerr_notification_slack":      {"slack.webhook_url"},
		"seerr_notification_webhook":    {"webhook.auth_header", "webhook.webhook_url"},
	}
	for name, paths := range want {
		sort.Strings(paths)
		want[name] = paths
	}

	seen := map[string]bool{}
	for _, newRes := range RegisteredResources {
		res := newRes()
		name := lifecycleTypeName(t, res)
		seen[name] = true
		got := lifecycleSensitivePaths(t, res)
		expected := want[name]
		if !slices.Equal(got, expected) {
			t.Errorf("%s sensitive paths = %#v, want %#v", name, got, expected)
		}
	}
	for name := range want {
		if !seen[name] {
			t.Errorf("sensitive schema lists unregistered resource %s", name)
		}
	}
}

func TestLifecycleContractExternalDeletion(t *testing.T) {
	notFound := func(w http.ResponseWriter, _ *http.Request, _ int) {
		http.Error(w, "gone", http.StatusNotFound)
	}
	cases := []struct {
		name        string
		new         func() resource.Resource
		model       any
		wantRemoved bool
		wantError   bool
	}{
		{name: "api_key", new: NewAPIKeyResource, model: &APIKeyModel{ID: types.StringValue("api_key"), ApiKey: types.StringValue("stored-key")}, wantRemoved: true},
		{name: "backup_settings", new: NewBackupSettingsResource, model: &BackupSettingsModel{ID: types.StringValue("backups")}, wantRemoved: true},
		{name: "metadata_settings", new: NewMetadataSettingsResource, model: &MetadataSettingsModel{ID: types.StringValue("metadata")}, wantRemoved: true},
		{name: "main_settings", new: NewMainSettingsResource, model: &MainSettingsModel{ID: types.StringValue("main")}, wantRemoved: true},
		{name: "plex_settings", new: NewPlexSettingsResource, model: &PlexSettingsModel{ID: types.StringValue("plex")}, wantRemoved: true},
		{name: "jellyfin_settings", new: NewJellyfinSettingsResource, model: &JellyfinSettingsModel{ID: types.StringValue("jellyfin"), APIKey: types.StringValue("jf-secret")}, wantRemoved: true},
		{name: "emby_settings", new: NewEmbySettingsResource, model: &EmbySettingsModel{ID: types.StringValue("emby"), APIKey: types.StringValue("emby-secret")}, wantRemoved: true},
		{name: "tautulli_settings", new: NewTautulliSettingsResource, model: &TautulliSettingsModel{ID: types.StringValue("tautulli"), APIKey: types.StringValue("tautulli-secret")}, wantRemoved: true},
		{name: "network_settings", new: NewNetworkSettingsResource, model: &NetworkSettingsModel{ID: types.StringValue("network")}, wantRemoved: true},
		{name: "sonarr_server", new: NewSonarrServerResource, model: sonarrPrior(types.StringValue("sonarr-secret")), wantRemoved: true},
		{name: "radarr_server", new: NewRadarrServerResource, model: radarrPrior(types.StringValue("radarr-secret")), wantRemoved: true},
		{name: "job_schedule", new: NewJobScheduleResource, model: &JobScheduleModel{ID: types.StringValue("plex-scan"), JobID: types.StringValue("plex-scan")}, wantRemoved: true},
		{name: "issue", new: NewIssueResource, model: issuePrior("11"), wantRemoved: true},
		{name: "request", new: NewRequestResource, model: requestPrior("42"), wantRemoved: true},
		{name: "blocklist", new: NewBlocklistResource, model: &BlocklistModel{ID: types.StringValue("movie:99"), TMDBID: types.Int64Value(99), MediaType: types.StringValue("movie")}, wantRemoved: true},
		{name: "user_notification_settings", new: NewUserNotificationSettingsResource, model: &UserNotificationSettingsResourceModel{ID: types.StringValue("9"), UserID: types.Int64Value(9), PushbulletAccessToken: types.StringValue("pb-secret")}, wantRemoved: true},
		{name: "api_object_suppress", new: NewAPIObjectResource, model: apiObjectPrior(true), wantRemoved: true},
		{name: "api_object_surface_error", new: NewAPIObjectResource, model: apiObjectPrior(false), wantError: true},
		{name: "plex_library_sync", new: NewPlexLibrarySyncResource, model: &PlexLibrarySyncModel{ID: types.StringValue("plex_library_sync"), Triggers: types.MapNull(types.StringType)}, wantError: true},
		{name: "jellyfin_library_sync", new: NewJellyfinLibrarySyncResource, model: &JellyfinLibrarySyncModel{ID: types.StringValue("jellyfin_library_sync"), Triggers: types.MapNull(types.StringType)}, wantError: true},
		{name: "emby_library_sync", new: NewEmbyLibrarySyncResource, model: &EmbyLibrarySyncModel{ID: types.StringValue("emby_library_sync"), Triggers: types.MapNull(types.StringType)}, wantError: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := newLifecycleServer(t, notFound)
			res := tc.new()
			lifecycleConfigure(t, res, srv.client(0))
			resp := lifecycleRead(t, res, lifecycleState(t, res, tc.model))
			if resp.Diagnostics.HasError() != tc.wantError {
				t.Fatalf("error = %v, want %v\n%s", resp.Diagnostics.HasError(), tc.wantError, diagnosticsText(resp.Diagnostics))
			}
			requireNoSecretLeak(t, resp.Diagnostics, lifecycleProviderKey, "stored-key", "jf-secret", "emby-secret", "tautulli-secret", "sonarr-secret", "radarr-secret", "pb-secret")
			if tc.wantRemoved {
				assertRemoved(t, resp)
			} else {
				assertKept(t, resp)
			}
			srv.requireProviderKey(t)
		})
	}
}

func TestLifecycleContractMissingArrServerRemovesState(t *testing.T) {
	body := `[{"id":99,"name":"other","tags":[]}]`
	for _, tc := range []struct {
		name  string
		new   func() resource.Resource
		model any
		path  string
	}{
		{name: "sonarr", new: NewSonarrServerResource, model: sonarrPrior(types.StringValue("sonarr-secret")), path: "/api/v1/settings/sonarr"},
		{name: "radarr", new: NewRadarrServerResource, model: radarrPrior(types.StringValue("radarr-secret")), path: "/api/v1/settings/radarr"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.new()
			resp := configureRead(t, res, tc.model, tc.path, body)
			assertRemoved(t, resp)
		})
	}

	t.Run("job_schedule", func(t *testing.T) {
		res := NewJobScheduleResource()
		resp := configureRead(t, res, &JobScheduleModel{
			ID:       types.StringValue("plex-scan"),
			JobID:    types.StringValue("plex-scan"),
			Schedule: types.StringValue("0 1 * * *"),
		}, "/api/v1/settings/jobs", `[{"id":"other-job","schedule":"0 0 * * *"}]`)
		assertRemoved(t, resp)
	})
}

func TestLifecycleContractImportAndRefresh(t *testing.T) {
	t.Run("api_key", func(t *testing.T) {
		res := NewAPIKeyResource()
		resp := importThenRead(t, res, "api_key", "/api/v1/settings/main", `{"apiKey":"server-real-api-key"}`, nil)
		got := decodeState[APIKeyModel](t, resp.State)
		assertString(t, got.ID, false, "api_key")
		assertString(t, got.ApiKey, false, "server-real-api-key")
		if got.ApiKey.ValueString() == lifecycleProviderKey {
			t.Fatal("resource stored the provider credential")
		}
		requireNoSecretLeak(t, resp.Diagnostics, "server-real-api-key")
	})

	t.Run("backup_settings", func(t *testing.T) {
		res := NewBackupSettingsResource()
		resp := importThenRead(t, res, "ignored", "/api/v1/settings/backups", `{"schedule":"0 2 * * *","retention":30,"storage_path":"/var/backups"}`, nil)
		got := decodeState[BackupSettingsModel](t, resp.State)
		assertString(t, got.ID, false, "backups")
		assertString(t, got.Schedule, false, "0 2 * * *")
		if got.Retention.ValueInt64() != 30 {
			t.Fatalf("retention = %d", got.Retention.ValueInt64())
		}
	})

	t.Run("metadata_settings", func(t *testing.T) {
		res := NewMetadataSettingsResource()
		resp := importThenRead(t, res, "ignored", "/api/v1/settings/metadatas", `{"tv":"tvdb","anime":"tmdb"}`, nil)
		got := decodeState[MetadataSettingsModel](t, resp.State)
		assertString(t, got.ID, false, "metadata")
		assertString(t, got.TV, false, "tvdb")
		assertString(t, got.Anime, false, "tmdb")
	})

	t.Run("main_settings", func(t *testing.T) {
		res := NewMainSettingsResource()
		resp := importThenRead(t, res, "ignored", "/api/v1/settings/main", `{"applicationTitle":"Seerr","applicationUrl":"https://seerr.example.com"}`, nil)
		got := decodeState[MainSettingsModel](t, resp.State)
		assertString(t, got.ID, false, "main")
		assertString(t, got.AppTitle, false, "Seerr")
		assertString(t, got.ApplicationURL, false, "https://seerr.example.com")
	})

	t.Run("plex_settings", func(t *testing.T) {
		res := NewPlexSettingsResource()
		resp := importThenRead(t, res, "ignored", "/api/v1/settings/plex", `{"name":"Plex","ip":"127.0.0.1","port":32400,"useSsl":false}`, nil)
		got := decodeState[PlexSettingsModel](t, resp.State)
		assertString(t, got.ID, false, "plex")
		assertString(t, got.Name, false, "Plex")
		assertString(t, got.IP, false, "127.0.0.1")
	})

	t.Run("job_schedule", func(t *testing.T) {
		res := NewJobScheduleResource()
		resp := importThenRead(t, res, "plex-scan", "/api/v1/settings/jobs", `[{"id":"plex-scan","schedule":"0 0 * * *"}]`, nil)
		got := decodeState[JobScheduleModel](t, resp.State)
		assertString(t, got.ID, false, "plex-scan")
		assertString(t, got.JobID, false, "plex-scan")
		assertString(t, got.Schedule, false, "0 0 * * *")
		assertString(t, got.PreviousSchedule, false, "0 0 * * *")
	})

	t.Run("sonarr_server", func(t *testing.T) {
		res := NewSonarrServerResource()
		resp := importThenRead(t, res, "4", "/api/v1/settings/sonarr", `[{"id":4,"name":"main-sonarr","hostname":"sonarr.local","port":8989,"apiKey":"upstream-secret","tags":[],"animeTags":null}]`, nil)
		got := decodeState[SonarrServerModel](t, resp.State)
		assertString(t, got.ID, false, "4")
		if got.ServerID.ValueInt64() != 4 {
			t.Fatalf("server_id = %d", got.ServerID.ValueInt64())
		}
		assertString(t, got.Name, false, "main-sonarr")
		assertString(t, got.APIKey, true, "")
		if got.Tags.IsNull() || len(got.Tags.Elements()) != 0 {
			t.Fatalf("tags = %#v, want empty list", got.Tags)
		}
		if !got.AnimeTags.IsNull() {
			t.Fatal("anime_tags should stay null")
		}
		requireNoSecretLeak(t, resp.Diagnostics, "upstream-secret")
	})

	t.Run("radarr_server", func(t *testing.T) {
		res := NewRadarrServerResource()
		resp := importThenRead(t, res, "3", "/api/v1/settings/radarr", `[{"id":3,"name":"main-radarr","hostname":"radarr.local","port":7878,"apiKey":"upstream-secret","tags":[]}]`, nil)
		got := decodeState[RadarrServerModel](t, resp.State)
		assertString(t, got.Name, false, "main-radarr")
		assertString(t, got.APIKey, true, "")
		if got.Tags.IsNull() || len(got.Tags.Elements()) != 0 {
			t.Fatalf("tags = %#v, want empty list", got.Tags)
		}
	})

	t.Run("issue", func(t *testing.T) {
		res := NewIssueResource()
		body := `{"id":11,"issueType":4,"status":1,"createdAt":"2026-01-01T00:00:00.000Z","comments":[],"media":{"id":550},"createdBy":{"id":1,"email":"user@example.com","displayName":"User"},"modifiedBy":null}`
		resp := importThenRead(t, res, "11", "/api/v1/issue/11", body, nil)
		got := decodeState[IssueModel](t, resp.State)
		assertString(t, got.ID, false, "11")
		if got.MediaID.ValueInt64() != 550 {
			t.Fatalf("media_id = %d", got.MediaID.ValueInt64())
		}
		assertString(t, got.CreatedAt, false, "2026-01-01T00:00:00.000Z")
	})

	t.Run("request", func(t *testing.T) {
		res := NewRequestResource()
		body := `{"status":1,"createdAt":"2026-01-01T00:00:00.000Z","media":{"id":7,"mediaType":"movie","tmdbId":550},"requestedBy":{"id":1,"email":"user@example.com","displayName":"User"},"modifiedBy":{"id":2,"displayName":"Admin"}}`
		resp := importThenRead(t, res, "42", "/api/v1/request/42", body, nil)
		got := decodeState[RequestModel](t, resp.State)
		assertString(t, got.ID, false, "42")
		assertString(t, got.CreatedAt, false, "2026-01-01T00:00:00.000Z")
		if got.SeerrMediaID.ValueInt64() != 7 {
			t.Fatalf("seerr_media_id = %d", got.SeerrMediaID.ValueInt64())
		}
	})

	t.Run("blocklist", func(t *testing.T) {
		res := NewBlocklistResource()
		resp := importThenRead(t, res, "movie:99", "/api/v1/blocklist/99", `{"tmdbId":99,"mediaType":"movie","title":"Example"}`, nil)
		got := decodeState[BlocklistModel](t, resp.State)
		assertString(t, got.ID, false, "movie:99")
		assertString(t, got.Title, false, "Example")
		assertString(t, got.CreatedAt, true, "")
	})

	t.Run("api_object", func(t *testing.T) {
		res := NewAPIObjectResource()
		resp := importThenRead(t, res, "/api/v1/settings/main", "/api/v1/settings/main", `{"ok":true}`, func(t *testing.T, state *tfsdk.State) {
			t.Helper()
			requireNoErrors(t, state.SetAttribute(t.Context(), path.Root("read_method"), http.MethodGet))
		})
		got := decodeState[APIObjectModel](t, resp.State)
		assertString(t, got.Path, false, "/api/v1/settings/main")
		assertString(t, got.ResponseBodyJSON, false, `{"ok":true}`)
	})

	t.Run("watchlist", func(t *testing.T) {
		res := NewWatchlistResource()
		srv := newLifecycleServer(t, func(w http.ResponseWriter, r *http.Request, _ int) {
			switch r.URL.Path {
			case "/api/v1/auth/me":
				writeJSON(w, `{"id":17}`)
			case "/api/v1/user/17/watchlist":
				writeJSON(w, `{"page":1,"totalPages":1,"totalResults":1,"results":[{"tmdbId":550,"mediaType":"movie","title":"Fight Club","overview":""}]}`)
			default:
				t.Errorf("unexpected path %s", r.URL.Path)
				http.NotFound(w, r)
			}
		})
		lifecycleConfigure(t, res, srv.client(0))
		imported := lifecycleImport(t, res, "550:movie")
		requireNoErrors(t, imported.Diagnostics)
		resp := lifecycleRead(t, res, imported.State)
		requireNoErrors(t, resp.Diagnostics)
		got := decodeState[WatchlistResourceModel](t, resp.State)
		assertString(t, got.ID, false, "550:movie")
		assertString(t, got.Title, false, "Fight Club")
		assertString(t, got.Overview, true, "")
		srv.requireProviderKey(t)
	})

	t.Run("user_notification_settings", func(t *testing.T) {
		res := NewUserNotificationSettingsResource()
		resp := importThenRead(t, res, "9", "/api/v1/user/9/settings/notifications", `{"discordEnabled":true,"discordId":"123"}`, func(t *testing.T, state *tfsdk.State) {
			t.Helper()
			requireNoErrors(t, state.SetAttribute(t.Context(), path.Root("pushbullet_access_token"), "pb-secret"))
		})
		got := decodeState[UserNotificationSettingsResourceModel](t, resp.State)
		assertString(t, got.ID, false, "9")
		if got.UserID.ValueInt64() != 9 {
			t.Fatalf("user_id = %d", got.UserID.ValueInt64())
		}
		assertString(t, got.DiscordID, false, "123")
		assertString(t, got.PushbulletAccessToken, false, "pb-secret")
		requireNoSecretLeak(t, resp.Diagnostics, "pb-secret")
	})

	t.Run("gotify", func(t *testing.T) {
		res := NewNotificationGotifyResource()
		resp := importThenRead(t, res, "gotify", "/api/v1/settings/notifications/gotify", `{"enabled":true,"embedPoster":false,"types":0,"options":{"url":"https://gotify.example","token":"gotify-from-api"}}`, nil)
		var got *NotificationAgentGotifyModel
		requireNoErrors(t, resp.State.GetAttribute(t.Context(), path.Root("gotify"), &got))
		if got == nil {
			t.Fatal("gotify block missing after import refresh")
		}
		assertString(t, got.Url, false, "https://gotify.example")
		assertString(t, got.Token, false, "gotify-from-api")
		requireNoSecretLeak(t, resp.Diagnostics, "gotify-from-api")

		rejected := lifecycleImport(t, NewNotificationGotifyResource(), "slack")
		if !rejected.Diagnostics.HasError() {
			t.Fatal("expected gotify import to reject a different agent id")
		}
	})
}

func TestLifecycleContractNullVersusEmpty(t *testing.T) {
	t.Run("backup_settings", func(t *testing.T) {
		res := NewBackupSettingsResource()
		prior := &BackupSettingsModel{ID: types.StringValue("backups"), Schedule: types.StringValue("0 0 * * *"), Retention: types.Int64Value(10), StoragePath: types.StringValue("/backups")}
		omitted := configureRead(t, res, prior, "/api/v1/settings/backups", `{"schedule":"0 1 * * *"}`)
		got := decodeState[BackupSettingsModel](t, omitted.State)
		assertString(t, got.Schedule, false, "0 1 * * *")
		if !got.Retention.IsNull() || !got.StoragePath.IsNull() {
			t.Fatalf("omitted fields should be null, retention %#v storage %#v", got.Retention, got.StoragePath)
		}

		res = NewBackupSettingsResource()
		empty := configureRead(t, res, prior, "/api/v1/settings/backups", `{"schedule":"","retention":0,"storage_path":""}`)
		got = decodeState[BackupSettingsModel](t, empty.State)
		assertString(t, got.Schedule, false, "")
		assertString(t, got.StoragePath, false, "")
		if got.Retention.IsNull() || got.Retention.ValueInt64() != 0 {
			t.Fatalf("retention = %#v, want 0", got.Retention)
		}
	})

	t.Run("metadata_settings", func(t *testing.T) {
		prior := &MetadataSettingsModel{ID: types.StringValue("metadata"), TV: types.StringValue("tmdb"), Anime: types.StringValue("tvdb")}
		omitted := configureRead(t, NewMetadataSettingsResource(), prior, "/api/v1/settings/metadatas", `{"tv":"tvdb"}`)
		got := decodeState[MetadataSettingsModel](t, omitted.State)
		assertString(t, got.TV, false, "tvdb")
		assertString(t, got.Anime, true, "")

		empty := configureRead(t, NewMetadataSettingsResource(), prior, "/api/v1/settings/metadatas", `{"tv":"","anime":null}`)
		got = decodeState[MetadataSettingsModel](t, empty.State)
		assertString(t, got.TV, false, "")
		assertString(t, got.Anime, true, "")
	})

	t.Run("main_settings", func(t *testing.T) {
		prior := &MainSettingsModel{ID: types.StringValue("main"), AppTitle: types.StringValue("Old"), Locale: types.StringValue("en")}
		omitted := configureRead(t, NewMainSettingsResource(), prior, "/api/v1/settings/main", `{"applicationTitle":"Seerr"}`)
		got := decodeState[MainSettingsModel](t, omitted.State)
		assertString(t, got.AppTitle, false, "Seerr")
		assertString(t, got.Locale, true, "")

		empty := configureRead(t, NewMainSettingsResource(), prior, "/api/v1/settings/main", `{"applicationTitle":"","locale":""}`)
		got = decodeState[MainSettingsModel](t, empty.State)
		assertString(t, got.AppTitle, false, "")
		assertString(t, got.Locale, false, "")
	})

	t.Run("plex_settings", func(t *testing.T) {
		prior := &PlexSettingsModel{ID: types.StringValue("plex"), Name: types.StringValue("Old"), IP: types.StringValue("127.0.0.1"), Port: types.Int64Value(32400)}
		omitted := configureRead(t, NewPlexSettingsResource(), prior, "/api/v1/settings/plex", `{"ip":"10.0.0.5","port":32400}`)
		got := decodeState[PlexSettingsModel](t, omitted.State)
		assertString(t, got.Name, true, "")
		assertString(t, got.IP, false, "10.0.0.5")

		empty := configureRead(t, NewPlexSettingsResource(), prior, "/api/v1/settings/plex", `{"name":"","ip":"10.0.0.5","port":32400}`)
		got = decodeState[PlexSettingsModel](t, empty.State)
		assertString(t, got.Name, false, "")
	})

	t.Run("jellyfin_url_base", func(t *testing.T) {
		prior := &JellyfinSettingsModel{ID: types.StringValue("jellyfin"), IP: types.StringValue("127.0.0.1"), Port: types.Int64Value(8096), URLBase: types.StringValue("/old"), APIKey: types.StringValue("jf-secret")}
		omitted := configureRead(t, NewJellyfinSettingsResource(), prior, "/api/v1/settings/jellyfin", `{"ip":"127.0.0.1","port":8096,"apiKey":"jf-secret"}`)
		got := decodeState[JellyfinSettingsModel](t, omitted.State)
		assertString(t, got.URLBase, true, "")

		empty := configureRead(t, NewJellyfinSettingsResource(), prior, "/api/v1/settings/jellyfin", `{"ip":"127.0.0.1","port":8096,"urlBase":"","apiKey":"jf-secret"}`)
		got = decodeState[JellyfinSettingsModel](t, empty.State)
		assertString(t, got.URLBase, false, "")
	})

	t.Run("network_proxy_hostname", func(t *testing.T) {
		prior := &NetworkSettingsModel{
			ID: types.StringValue("network"),
			Proxy: &NetworkProxyModel{
				Hostname: types.StringValue("old-proxy"),
				Password: types.StringValue("proxy-secret"),
			},
		}
		omitted := configureRead(t, NewNetworkSettingsResource(), prior, "/api/v1/settings/network", `{"proxy":{"enabled":true,"port":8080}}`)
		got := decodeState[NetworkSettingsModel](t, omitted.State)
		if got.Proxy == nil {
			t.Fatal("managed proxy block was dropped")
		}
		assertString(t, got.Proxy.Hostname, true, "")

		empty := configureRead(t, NewNetworkSettingsResource(), prior, "/api/v1/settings/network", `{"proxy":{"hostname":"","port":8080}}`)
		got = decodeState[NetworkSettingsModel](t, empty.State)
		assertString(t, got.Proxy.Hostname, false, "")
	})

	t.Run("sonarr_tags", func(t *testing.T) {
		bodyNull := `[{"id":4,"name":"sonarr","tags":null,"animeTags":[]}]`
		resp := configureRead(t, NewSonarrServerResource(), sonarrPrior(types.StringNull()), "/api/v1/settings/sonarr", bodyNull)
		got := decodeState[SonarrServerModel](t, resp.State)
		if !got.Tags.IsNull() {
			t.Fatal("json null tags should stay null")
		}
		if got.AnimeTags.IsNull() || len(got.AnimeTags.Elements()) != 0 {
			t.Fatalf("empty animeTags = %#v", got.AnimeTags)
		}

		bodyEmpty := `[{"id":4,"name":"sonarr","tags":[],"animeTags":null}]`
		resp = configureRead(t, NewSonarrServerResource(), sonarrPrior(types.StringNull()), "/api/v1/settings/sonarr", bodyEmpty)
		got = decodeState[SonarrServerModel](t, resp.State)
		if got.Tags.IsNull() || len(got.Tags.Elements()) != 0 {
			t.Fatalf("empty tags = %#v", got.Tags)
		}
		if !got.AnimeTags.IsNull() {
			t.Fatal("json null animeTags should stay null")
		}
	})

	t.Run("radarr_tags", func(t *testing.T) {
		nullResp := configureRead(t, NewRadarrServerResource(), radarrPrior(types.StringNull()), "/api/v1/settings/radarr", `[{"id":3,"name":"radarr","tags":null}]`)
		got := decodeState[RadarrServerModel](t, nullResp.State)
		if !got.Tags.IsNull() {
			t.Fatal("json null tags should stay null")
		}
		emptyResp := configureRead(t, NewRadarrServerResource(), radarrPrior(types.StringNull()), "/api/v1/settings/radarr", `[{"id":3,"name":"radarr","tags":[]}]`)
		got = decodeState[RadarrServerModel](t, emptyResp.State)
		if got.Tags.IsNull() || len(got.Tags.Elements()) != 0 {
			t.Fatalf("empty tags = %#v", got.Tags)
		}
	})

	t.Run("issue_created_at", func(t *testing.T) {
		prior := issuePrior("11")
		omitted := configureRead(t, NewIssueResource(), prior, "/api/v1/issue/11", `{"issueType":1,"status":1,"comments":[]}`)
		got := decodeState[IssueModel](t, omitted.State)
		assertString(t, got.CreatedAt, true, "")
		if got.CommentsCount.ValueInt64() != 0 {
			t.Fatalf("comments_count = %d", got.CommentsCount.ValueInt64())
		}

		empty := configureRead(t, NewIssueResource(), prior, "/api/v1/issue/11", `{"issueType":1,"status":1,"createdAt":"","comments":null}`)
		got = decodeState[IssueModel](t, empty.State)
		assertString(t, got.CreatedAt, false, "")
		if got.CommentsCount.ValueInt64() != 0 {
			t.Fatalf("null comments_count = %d, want 0", got.CommentsCount.ValueInt64())
		}
	})

	t.Run("request_created_at", func(t *testing.T) {
		prior := requestPrior("42")
		body := `{"status":1,"media":{"id":7,"mediaType":"movie","tmdbId":550},"requestedBy":{"id":1,"displayName":"User"}}`
		omitted := configureRead(t, NewRequestResource(), prior, "/api/v1/request/42", body)
		got := decodeState[RequestModel](t, omitted.State)
		assertString(t, got.CreatedAt, true, "")

		emptyBody := `{"status":1,"createdAt":"","media":{"id":7,"mediaType":"movie","tmdbId":550},"requestedBy":{"id":1,"displayName":"User"}}`
		empty := configureRead(t, NewRequestResource(), prior, "/api/v1/request/42", emptyBody)
		got = decodeState[RequestModel](t, empty.State)
		assertString(t, got.CreatedAt, false, "")
	})

	t.Run("blocklist_title", func(t *testing.T) {
		prior := &BlocklistModel{ID: types.StringValue("movie:99"), TMDBID: types.Int64Value(99), MediaType: types.StringValue("movie"), Title: types.StringValue("Old")}
		omitted := configureRead(t, NewBlocklistResource(), prior, "/api/v1/blocklist/99", `{"tmdbId":99,"mediaType":"movie"}`)
		got := decodeState[BlocklistModel](t, omitted.State)
		assertString(t, got.Title, true, "")

		empty := configureRead(t, NewBlocklistResource(), prior, "/api/v1/blocklist/99", `{"tmdbId":99,"mediaType":"movie","title":""}`)
		got = decodeState[BlocklistModel](t, empty.State)
		assertString(t, got.Title, false, "")
	})

	t.Run("api_object_body", func(t *testing.T) {
		prior := apiObjectPrior(false)
		empty := configureRead(t, NewAPIObjectResource(), prior, "/api/v1/settings/main", "")
		got := decodeState[APIObjectModel](t, empty.State)
		assertString(t, got.ResponseBodyJSON, false, "")

		object := configureRead(t, NewAPIObjectResource(), prior, "/api/v1/settings/main", `{}`)
		got = decodeState[APIObjectModel](t, object.State)
		assertString(t, got.ResponseBodyJSON, false, `{}`)
	})

	t.Run("job_schedule_blank_keeps_prior", func(t *testing.T) {
		prior := &JobScheduleModel{ID: types.StringValue("plex-scan"), JobID: types.StringValue("plex-scan"), Schedule: types.StringValue("0 1 * * *"), PreviousSchedule: types.StringValue("0 1 * * *")}
		for _, body := range []string{
			`[{"id":"plex-scan","schedule":""}]`,
			`[{"id":"plex-scan","schedule":null}]`,
			`[{"id":"plex-scan"}]`,
		} {
			resp := configureRead(t, NewJobScheduleResource(), prior, "/api/v1/settings/jobs", body)
			got := decodeState[JobScheduleModel](t, resp.State)
			assertString(t, got.Schedule, false, "0 1 * * *")
		}
	})

	t.Run("user_notification_empty_object", func(t *testing.T) {
		prior := &UserNotificationSettingsResourceModel{
			ID:                types.StringValue("9"),
			UserID:            types.Int64Value(9),
			NotificationTypes: &UserNotificationTypesModel{Email: types.Int64Value(4)},
		}
		omitted := configureRead(t, NewUserNotificationSettingsResource(), prior, "/api/v1/user/9/settings/notifications", `{}`)
		got := decodeState[UserNotificationSettingsResourceModel](t, omitted.State)
		if got.NotificationTypes == nil || got.NotificationTypes.Email.ValueInt64() != 4 {
			t.Fatalf("omitted notification_types = %#v", got.NotificationTypes)
		}

		empty := configureRead(t, NewUserNotificationSettingsResource(), prior, "/api/v1/user/9/settings/notifications", `{"notificationTypes":{}}`)
		got = decodeState[UserNotificationSettingsResourceModel](t, empty.State)
		if got.NotificationTypes == nil {
			t.Fatal("empty notificationTypes object should still produce a block")
		}
		if !got.NotificationTypes.Email.IsNull() {
			t.Fatalf("empty object email = %#v, want null", got.NotificationTypes.Email)
		}
	})
}

func TestLifecycleContractSecrets(t *testing.T) {
	t.Run("jellyfin_preserves_null_and_empty_api_key", func(t *testing.T) {
		prior := &JellyfinSettingsModel{ID: types.StringValue("jellyfin"), IP: types.StringValue("127.0.0.1"), Port: types.Int64Value(8096), APIKey: types.StringValue("jf-secret")}
		for _, body := range []string{
			`{"ip":"127.0.0.1","port":8096}`,
			`{"ip":"127.0.0.1","port":8096,"apiKey":null}`,
			`{"ip":"127.0.0.1","port":8096,"apiKey":""}`,
		} {
			resp := configureRead(t, NewJellyfinSettingsResource(), prior, "/api/v1/settings/jellyfin", body)
			got := decodeState[JellyfinSettingsModel](t, resp.State)
			assertString(t, got.APIKey, false, "jf-secret")
			requireNoSecretLeak(t, resp.Diagnostics, "jf-secret")
		}
		replaced := configureRead(t, NewJellyfinSettingsResource(), prior, "/api/v1/settings/jellyfin", `{"ip":"127.0.0.1","port":8096,"apiKey":"jf-rotated"}`)
		got := decodeState[JellyfinSettingsModel](t, replaced.State)
		assertString(t, got.APIKey, false, "jf-rotated")
	})

	t.Run("emby_preserves_empty_api_key", func(t *testing.T) {
		prior := &EmbySettingsModel{ID: types.StringValue("emby"), IP: types.StringValue("127.0.0.1"), Port: types.Int64Value(8096), APIKey: types.StringValue("emby-secret")}
		resp := configureRead(t, NewEmbySettingsResource(), prior, "/api/v1/settings/emby", `{"ip":"127.0.0.1","port":8096,"apiKey":""}`)
		got := decodeState[EmbySettingsModel](t, resp.State)
		assertString(t, got.APIKey, false, "emby-secret")
		requireNoSecretLeak(t, resp.Diagnostics, "emby-secret")
	})

	t.Run("tautulli_preserves_empty_api_key", func(t *testing.T) {
		prior := &TautulliSettingsModel{ID: types.StringValue("tautulli"), Hostname: types.StringValue("tautulli.local"), APIKey: types.StringValue("tautulli-secret")}
		resp := configureRead(t, NewTautulliSettingsResource(), prior, "/api/v1/settings/tautulli", `{"hostname":"tautulli.local","apiKey":""}`)
		got := decodeState[TautulliSettingsModel](t, resp.State)
		assertString(t, got.APIKey, false, "tautulli-secret")
		assertString(t, got.Hostname, false, "tautulli.local")
		requireNoSecretLeak(t, resp.Diagnostics, "tautulli-secret")
	})

	t.Run("network_password_null_preserved_empty_applied", func(t *testing.T) {
		prior := &NetworkSettingsModel{
			ID:    types.StringValue("network"),
			Proxy: &NetworkProxyModel{Hostname: types.StringValue("proxy.internal"), Password: types.StringValue("proxy-secret")},
		}
		omitted := configureRead(t, NewNetworkSettingsResource(), prior, "/api/v1/settings/network", `{"proxy":{"hostname":"proxy.internal","port":8080}}`)
		got := decodeState[NetworkSettingsModel](t, omitted.State)
		assertString(t, got.Proxy.Password, false, "proxy-secret")

		empty := configureRead(t, NewNetworkSettingsResource(), prior, "/api/v1/settings/network", `{"proxy":{"hostname":"proxy.internal","password":""}}`)
		got = decodeState[NetworkSettingsModel](t, empty.State)
		assertString(t, got.Proxy.Password, false, "")
		requireNoSecretLeak(t, empty.Diagnostics, "proxy-secret")
	})

	t.Run("arr_servers_keep_state_api_key", func(t *testing.T) {
		sonarr := configureRead(t, NewSonarrServerResource(), sonarrPrior(types.StringValue("sonarr-secret")), "/api/v1/settings/sonarr", `[{"id":4,"name":"sonarr","apiKey":"upstream-secret","tags":[]}]`)
		gotSonarr := decodeState[SonarrServerModel](t, sonarr.State)
		assertString(t, gotSonarr.APIKey, false, "sonarr-secret")

		radarr := configureRead(t, NewRadarrServerResource(), radarrPrior(types.StringValue("radarr-secret")), "/api/v1/settings/radarr", `[{"id":3,"name":"radarr","apiKey":"","tags":null}]`)
		gotRadarr := decodeState[RadarrServerModel](t, radarr.State)
		assertString(t, gotRadarr.APIKey, false, "radarr-secret")
		requireNoSecretLeak(t, sonarr.Diagnostics, "sonarr-secret", "upstream-secret")
		requireNoSecretLeak(t, radarr.Diagnostics, "radarr-secret")
	})

	t.Run("user_notification_tokens", func(t *testing.T) {
		prior := &UserNotificationSettingsResourceModel{
			ID:                       types.StringValue("9"),
			UserID:                   types.Int64Value(9),
			PushbulletAccessToken:    types.StringValue("pb-secret"),
			PushoverApplicationToken: types.StringValue("po-app"),
			PushoverUserKey:          types.StringValue("po-user"),
		}
		kept := configureRead(t, NewUserNotificationSettingsResource(), prior, "/api/v1/user/9/settings/notifications", `{"pushbulletAccessToken":null,"discordEnabled":false}`)
		got := decodeState[UserNotificationSettingsResourceModel](t, kept.State)
		assertString(t, got.PushbulletAccessToken, false, "pb-secret")
		assertString(t, got.PushoverUserKey, false, "po-user")

		cleared := configureRead(t, NewUserNotificationSettingsResource(), prior, "/api/v1/user/9/settings/notifications", `{"pushbulletAccessToken":"","pushoverApplicationToken":"","pushoverUserKey":""}`)
		got = decodeState[UserNotificationSettingsResourceModel](t, cleared.State)
		assertString(t, got.PushbulletAccessToken, false, "")
		assertString(t, got.PushoverApplicationToken, false, "")
		assertString(t, got.PushoverUserKey, false, "")
		requireNoSecretLeak(t, cleared.Diagnostics, "pb-secret", "po-app", "po-user")
	})

	t.Run("gotify_empty_token_preserves_state", func(t *testing.T) {
		res := NewNotificationGotifyResource()
		srv := newLifecycleServer(t, okHandler(t, "/api/v1/settings/notifications/gotify", `{"enabled":true,"embedPoster":false,"types":0,"options":{"url":"","token":""}}`))
		lifecycleConfigure(t, res, srv.client(0))
		state := lifecycleTypedState(t, res)
		ctx := t.Context()
		requireNoErrors(t, state.SetAttribute(ctx, path.Root("id"), "gotify"))
		requireNoErrors(t, state.SetAttribute(ctx, path.Root("enabled"), true))
		requireNoErrors(t, state.SetAttribute(ctx, path.Root("embed_poster"), false))
		requireNoErrors(t, state.SetAttribute(ctx, path.Root("notification_types"), types.SetNull(types.StringType)))
		requireNoErrors(t, state.SetAttribute(ctx, path.Root("gotify"), &NotificationAgentGotifyModel{
			Url:   types.StringValue(""),
			Token: types.StringValue("gotify-secret"),
		}))
		resp := lifecycleRead(t, res, state)
		requireNoErrors(t, resp.Diagnostics)
		var got *NotificationAgentGotifyModel
		requireNoErrors(t, resp.State.GetAttribute(ctx, path.Root("gotify"), &got))
		if got == nil {
			t.Fatal("gotify block missing")
		}
		assertString(t, got.Token, false, "gotify-secret")
		assertString(t, got.Url, false, "")
		requireNoSecretLeak(t, resp.Diagnostics, "gotify-secret")
		srv.requireProviderKey(t)
	})

	t.Run("upstream_error_does_not_echo_provider_key", func(t *testing.T) {
		srv := newLifecycleServer(t, func(w http.ResponseWriter, _ *http.Request, _ int) {
			http.Error(w, "boom", http.StatusInternalServerError)
		})
		res := NewAPIKeyResource()
		lifecycleConfigure(t, res, srv.client(0))
		resp := lifecycleRead(t, res, lifecycleState(t, res, &APIKeyModel{ID: types.StringValue("api_key"), ApiKey: types.StringValue("stored-key")}))
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected read error")
		}
		assertKept(t, resp)
		requireNoSecretLeak(t, resp.Diagnostics, lifecycleProviderKey, "stored-key")
	})
}

func TestLifecycleContractRetries(t *testing.T) {
	cases := []struct {
		name  string
		new   func() resource.Resource
		model any
		path  string
		body  string
	}{
		{name: "api_key", new: NewAPIKeyResource, model: &APIKeyModel{ID: types.StringValue("api_key")}, path: "/api/v1/settings/main", body: `{"apiKey":"rotated-key"}`},
		{name: "main_settings", new: NewMainSettingsResource, model: &MainSettingsModel{ID: types.StringValue("main")}, path: "/api/v1/settings/main", body: `{"applicationTitle":"Seerr"}`},
		{name: "backup_settings", new: NewBackupSettingsResource, model: &BackupSettingsModel{ID: types.StringValue("backups")}, path: "/api/v1/settings/backups", body: `{"schedule":"0 2 * * *"}`},
		{name: "jellyfin_settings", new: NewJellyfinSettingsResource, model: &JellyfinSettingsModel{ID: types.StringValue("jellyfin"), APIKey: types.StringValue("jf-secret")}, path: "/api/v1/settings/jellyfin", body: `{"ip":"127.0.0.1","port":8096,"apiKey":"jf-secret"}`},
		{name: "sonarr_server", new: NewSonarrServerResource, model: sonarrPrior(types.StringValue("sonarr-secret")), path: "/api/v1/settings/sonarr", body: `[{"id":4,"name":"sonarr","tags":[]}]`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := newLifecycleServer(t, func(w http.ResponseWriter, r *http.Request, n int) {
				if r.URL.Path != tc.path || r.Method != http.MethodGet {
					t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
					http.NotFound(w, r)
					return
				}
				if n == 1 {
					w.WriteHeader(http.StatusServiceUnavailable)
					_, _ = w.Write([]byte("unavailable"))
					return
				}
				writeJSON(w, tc.body)
			})
			res := tc.new()
			lifecycleConfigure(t, res, srv.client(1))
			resp := lifecycleRead(t, res, lifecycleState(t, res, tc.model))
			requireNoErrors(t, resp.Diagnostics)
			if srv.Hits() != 2 {
				t.Fatalf("hits = %d, want 2", srv.Hits())
			}
			requireNoSecretLeak(t, resp.Diagnostics, lifecycleProviderKey, "jf-secret", "sonarr-secret", "rotated-key")
		})

		t.Run(tc.name+"_disabled", func(t *testing.T) {
			srv := newLifecycleServer(t, func(w http.ResponseWriter, _ *http.Request, _ int) {
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte("unavailable"))
			})
			res := tc.new()
			lifecycleConfigure(t, res, srv.client(0))
			resp := lifecycleRead(t, res, lifecycleState(t, res, tc.model))
			if !resp.Diagnostics.HasError() {
				t.Fatal("expected an error when retries are disabled")
			}
			if srv.Hits() != 1 {
				t.Fatalf("hits = %d, want 1", srv.Hits())
			}
			assertKept(t, resp)
			requireNoSecretLeak(t, resp.Diagnostics, lifecycleProviderKey, "jf-secret", "sonarr-secret")
		})
	}
}

func TestLifecycleContractWatchlistDeletionAndRetry(t *testing.T) {
	prior := &WatchlistResourceModel{ID: types.StringValue("999:movie"), TMDBID: types.Int64Value(999), MediaType: types.StringValue("movie"), Title: types.StringValue("Keep me")}

	t.Run("endpoint_404_keeps_state", func(t *testing.T) {
		srv := newLifecycleServer(t, func(w http.ResponseWriter, r *http.Request, _ int) {
			if r.URL.Path == "/api/v1/auth/me" {
				writeJSON(w, `{"id":17}`)
				return
			}
			http.Error(w, "gone", http.StatusNotFound)
		})
		res := NewWatchlistResource()
		lifecycleConfigure(t, res, srv.client(0))
		resp := lifecycleRead(t, res, lifecycleState(t, res, prior))
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected an error when the watchlist endpoint 404s")
		}
		got := decodeState[WatchlistResourceModel](t, resp.State)
		assertString(t, got.ID, false, "999:movie")
		assertString(t, got.Title, false, "Keep me")
	})

	t.Run("complete_absence_removes_state", func(t *testing.T) {
		srv := newLifecycleServer(t, func(w http.ResponseWriter, r *http.Request, _ int) {
			if r.URL.Path == "/api/v1/auth/me" {
				writeJSON(w, `{"id":17}`)
				return
			}
			writeJSON(w, `{"page":1,"totalPages":0,"totalResults":0,"results":[]}`)
		})
		res := NewWatchlistResource()
		lifecycleConfigure(t, res, srv.client(0))
		resp := lifecycleRead(t, res, lifecycleState(t, res, prior))
		assertRemoved(t, resp)
	})

	t.Run("retries_watchlist_get", func(t *testing.T) {
		var listHits int
		srv := newLifecycleServer(t, func(w http.ResponseWriter, r *http.Request, _ int) {
			switch r.URL.Path {
			case "/api/v1/auth/me":
				writeJSON(w, `{"id":17}`)
			case "/api/v1/user/17/watchlist":
				listHits++
				if listHits == 1 {
					w.WriteHeader(http.StatusServiceUnavailable)
					_, _ = w.Write([]byte("unavailable"))
					return
				}
				writeJSON(w, `{"page":1,"totalPages":1,"totalResults":1,"results":[{"tmdbId":999,"mediaType":"movie","title":"Retried","overview":"kept"}]}`)
			default:
				t.Errorf("unexpected path %s", r.URL.Path)
				http.NotFound(w, r)
			}
		})
		res := NewWatchlistResource()
		lifecycleConfigure(t, res, srv.client(1))
		resp := lifecycleRead(t, res, lifecycleState(t, res, prior))
		requireNoErrors(t, resp.Diagnostics)
		if listHits != 2 {
			t.Fatalf("watchlist hits = %d, want 2", listHits)
		}
		got := decodeState[WatchlistResourceModel](t, resp.State)
		assertString(t, got.Title, false, "Retried")
		assertString(t, got.Overview, false, "kept")
	})
}

func TestLifecycleContractUserInvitationMakesNoRequest(t *testing.T) {
	srv := newLifecycleServer(t, func(w http.ResponseWriter, r *http.Request, _ int) {
		t.Errorf("invitation resource called %s %s", r.Method, r.URL.Path)
		http.Error(w, "unexpected", http.StatusInternalServerError)
	})
	res := NewUserInvitationResource()
	lifecycleConfigure(t, res, srv.client(0))

	imported := lifecycleImport(t, res, "1")
	if !imported.Diagnostics.HasError() {
		t.Fatal("expected import to fail")
	}
	if srv.Hits() != 0 {
		t.Fatalf("import hits = %d, want 0", srv.Hits())
	}

	resp := lifecycleRead(t, res, lifecycleState(t, res, &UserInvitationModel{ID: types.StringValue("1"), Email: types.StringValue("person@example.com")}))
	if resp.Diagnostics.HasError() {
		t.Fatalf("read should warn, not error: %s", diagnosticsText(resp.Diagnostics))
	}
	if resp.Diagnostics.WarningsCount() == 0 {
		t.Fatal("expected an unsupported-invitation warning")
	}
	if !resp.State.Raw.IsNull() {
		t.Fatal("expected invitation refresh to drop state")
	}
	if srv.Hits() != 0 {
		t.Fatalf("read hits = %d, want 0", srv.Hits())
	}
}
