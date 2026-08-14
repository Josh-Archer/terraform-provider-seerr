package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestWatchlistDataSource_Schema(t *testing.T) {
	ds := NewWatchlistDataSource()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema diagnostics error: %v", resp.Diagnostics)
	}

	attrs := []string{"id", "user_id", "total", "watchlist"}
	for _, attr := range attrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Expected '%s' attribute in schema", attr)
		}
	}
}

func TestWatchlistDataSource_Metadata(t *testing.T) {
	ds := NewWatchlistDataSource()
	var metaResp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "seerr"}, &metaResp)
	if metaResp.TypeName != "seerr_watchlist" {
		t.Errorf("Expected type name seerr_watchlist, got %s", metaResp.TypeName)
	}
}
