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

func TestGenresDataSource_Schema(t *testing.T) {
	ds := NewGenresDataSource()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema diagnostics error: %v", resp.Diagnostics)
	}

	if _, ok := resp.Schema.Attributes["genres"]; !ok {
		t.Error("Expected 'genres' attribute in schema")
	}
	if _, ok := resp.Schema.Attributes["total"]; !ok {
		t.Error("Expected 'total' attribute in schema")
	}
}

func TestGenresDataSource_Read(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/genres/movie" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]map[string]interface{}{
				{"id": 28, "name": "Action"},
				{"id": 35, "name": "Comedy"},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	client := NewClient(u, "test-key", "test-agent", true, defaultRequestTimeout, 0, 0)

	ds := &GenresDataSource{client: client}

	var metaResp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "seerr"}, &metaResp)
	if metaResp.TypeName != "seerr_genres" {
		t.Errorf("Expected type name seerr_genres, got %s", metaResp.TypeName)
	}
}
