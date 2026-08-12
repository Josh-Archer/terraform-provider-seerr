package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestJellyfinUsersDataSource_Schema(t *testing.T) {
	ds := NewJellyfinUsersDataSource()
	var resp datasource.SchemaResponse
	ds.Schema(nil, datasource.SchemaRequest{}, &resp)

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

func TestJellyfinUsersDataSource_Read(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/settings/jellyfin/users" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]map[string]interface{}{
				{
					"id":       "jf1",
					"username": "jfadmin",
					"email":    "admin@jellyfin.local",
					"userType": "administrator",
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	client := NewClient(u, "test-key", "test-agent", true, defaultRequestTimeout)

	ds := &JellyfinUsersDataSource{client: client}

	var metaResp datasource.MetadataResponse
	ds.Metadata(nil, datasource.MetadataRequest{ProviderTypeName: "seerr"}, &metaResp)
	if metaResp.TypeName != "seerr_jellyfin_users" {
		t.Errorf("Expected type name seerr_jellyfin_users, got %s", metaResp.TypeName)
	}
}
