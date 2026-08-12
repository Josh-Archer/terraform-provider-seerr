package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestWatchlistResource_Schema(t *testing.T) {
	r := NewWatchlistResource()
	var resp resource.SchemaResponse
	r.Schema(nil, resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema diagnostics error: %v", resp.Diagnostics)
	}

	attrs := []string{"id", "tmdb_id", "media_type", "title", "overview"}
	for _, attr := range attrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Expected '%s' attribute in schema", attr)
		}
	}
}

func TestWatchlistResource_Metadata(t *testing.T) {
	r := NewWatchlistResource()
	var metaResp resource.MetadataResponse
	r.Metadata(nil, resource.MetadataRequest{ProviderTypeName: "seerr"}, &metaResp)
	if metaResp.TypeName != "seerr_watchlist" {
		t.Errorf("Expected type name seerr_watchlist, got %s", metaResp.TypeName)
	}
}
