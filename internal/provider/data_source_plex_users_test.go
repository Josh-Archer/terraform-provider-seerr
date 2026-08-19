package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestPlexUsersDataSource_Schema(t *testing.T) {
	ds := NewPlexUsersDataSource()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema diagnostics error: %v", resp.Diagnostics)
	}

	if _, ok := resp.Schema.Attributes["users"]; !ok {
		t.Error("Expected 'users' attribute in schema")
	}
	if _, ok := resp.Schema.Attributes["total"]; !ok {
		t.Error("Expected 'total' attribute in schema")
	}
}

func TestPlexUsersDataSource_Read(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/settings/plex/users" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]map[string]interface{}{
				{
					"id":       "p1",
					"title":    "Plex Admin",
					"username": "plexadmin",
					"email":    "admin@plex.local",
					"thumb":    "https://plex.tv/thumb.jpg",
					"userType": "admin",
				},
				{
					"id":       "p2",
					"title":    "Plex Guest",
					"username": "plexguest",
					"email":    "guest@plex.local",
					"thumb":    "",
					"userType": "user",
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	client := NewClient(u, "test-key", "test-agent", true, defaultRequestTimeout, 0, 0)

	ds := &PlexUsersDataSource{client: client}

	var metaResp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "seerr"}, &metaResp)
	if metaResp.TypeName != "seerr_plex_users" {
		t.Errorf("Expected type name seerr_plex_users, got %s", metaResp.TypeName)
	}
}
