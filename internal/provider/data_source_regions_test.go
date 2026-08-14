package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestRegionsDataSource_Schema(t *testing.T) {
	ds := NewRegionsDataSource()
	var resp datasource.SchemaResponse
	ds.Schema(nil, datasource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema diagnostics error: %v", resp.Diagnostics)
	}

	if _, ok := resp.Schema.Attributes["regions"]; !ok {
		t.Error("Expected 'regions' attribute in schema")
	}
	if _, ok := resp.Schema.Attributes["total"]; !ok {
		t.Error("Expected 'total' attribute in schema")
	}
}

func TestRegionsDataSource_Read(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/regions" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]map[string]interface{}{
				{"iso_3166_1": "US", "english_name": "United States", "name": "United States"},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	client := NewClient(u, "test-key", "test-agent", true, defaultRequestTimeout)

	ds := &RegionsDataSource{client: client}

	var metaResp datasource.MetadataResponse
	ds.Metadata(nil, datasource.MetadataRequest{ProviderTypeName: "seerr"}, &metaResp)
	if metaResp.TypeName != "seerr_regions" {
		t.Errorf("Expected type name seerr_regions, got %s", metaResp.TypeName)
	}
}
