package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	providerConfig = `
provider "seerr" {
  url      = "http://localhost:5055"
  api_key  = "dummy_api_key_for_testing"
}
`
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"seerr": providerserver.NewProtocol6WithError(New("test")()),
}

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
