package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestAPIRequestDataSourceResponseIsSensitive(t *testing.T) {
	dataSource := NewAPIRequestDataSource()
	var response datasource.SchemaResponse
	dataSource.Schema(context.Background(), datasource.SchemaRequest{}, &response)
	if response.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", response.Diagnostics)
	}

	attribute, ok := response.Schema.Attributes["response_body_json"].(schema.StringAttribute)
	if !ok {
		t.Fatal("response_body_json is not a string attribute")
	}
	if !attribute.Sensitive {
		t.Fatal("response_body_json must be sensitive because arbitrary API responses can contain credentials")
	}
}
