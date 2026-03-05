package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func TestProviderMetadata(t *testing.T) {
	p := &SeerrProvider{version: "test"}
	resp := &provider.MetadataResponse{}
	p.Metadata(context.Background(), provider.MetadataRequest{}, resp)

	if resp.TypeName != "seerr" {
		t.Fatalf("expected type name seerr, got %s", resp.TypeName)
	}
	if resp.Version != "test" {
		t.Fatalf("expected version test, got %s", resp.Version)
	}
}
