package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestLanguagesDataSource_Schema(t *testing.T) {
	ds := NewLanguagesDataSource()
	var resp datasource.SchemaResponse
	ds.Schema(nil, datasource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema diagnostics error: %v", resp.Diagnostics)
	}

	if _, ok := resp.Schema.Attributes["languages"]; !ok {
		t.Error("Expected 'languages' attribute in schema")
	}
	if _, ok := resp.Schema.Attributes["total"]; !ok {
		t.Error("Expected 'total' attribute in schema")
	}
}

func TestLanguagesDataSource_Read(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/languages" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]map[string]interface{}{
				{"iso_639_1": "en", "english_name": "English", "name": "English"},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	client := NewClient(u, "test-key", "test-agent", true, defaultRequestTimeout)

	ds := &LanguagesDataSource{client: client}

	var metaResp datasource.MetadataResponse
	ds.Metadata(nil, datasource.MetadataRequest{ProviderTypeName: "seerr"}, &metaResp)
	if metaResp.TypeName != "seerr_languages" {
		t.Errorf("Expected type name seerr_languages, got %s", metaResp.TypeName)
	}
}
