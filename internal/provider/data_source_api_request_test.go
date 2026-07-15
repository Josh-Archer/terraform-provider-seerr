package provider

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func apiRequestTestConfig(t *testing.T, d *APIRequestDataSource, model APIRequestDataSourceModel) (tfsdk.Config, tfsdk.State) {
	t.Helper()
	var schemaResp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("API request schema diagnostics: %v", schemaResp.Diagnostics)
	}
	configuredState := tfsdk.State{Schema: schemaResp.Schema}
	if diags := configuredState.Set(context.Background(), &model); diags.HasError() {
		t.Fatalf("set API request config: %v", diags)
	}
	config := tfsdk.Config{Raw: configuredState.Raw, Schema: schemaResp.Schema}
	return config, tfsdk.State{Schema: schemaResp.Schema}
}

func TestAPIRequestDataSourceRead(t *testing.T) {
	var gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		gotMethod = req.Method
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		gotBody = string(body)
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"queued":true}`))
	}))
	defer srv.Close()
	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	d := &APIRequestDataSource{client: NewClient(baseURL, "key", "test-agent", false, defaultRequestTimeout)}
	model := APIRequestDataSourceModel{
		Path:             types.StringValue("/api/v1/test"),
		Method:           types.StringValue(http.MethodPost),
		Headers:          types.MapNull(types.StringType),
		RequestBodyJSON:  types.StringValue(`{"request":true}`),
		ResponseBodyJSON: types.StringUnknown(),
		StatusCode:       types.Int64Unknown(),
	}
	config, state := apiRequestTestConfig(t, d, model)
	resp := datasource.ReadResponse{State: state}
	d.Read(context.Background(), datasource.ReadRequest{Config: config}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", resp.Diagnostics)
	}
	if gotMethod != http.MethodPost || gotBody != `{"request":true}` {
		t.Fatalf("request = %s %s", gotMethod, gotBody)
	}
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("get API request state: %v", diags)
	}
	if model.StatusCode.ValueInt64() != http.StatusAccepted || !strings.Contains(model.ResponseBodyJSON.ValueString(), "queued") {
		t.Fatalf("unexpected response state: %#v", model)
	}
}

func TestAPIRequestDataSourceReportsInvalidPath(t *testing.T) {
	baseURL, err := url.Parse("http://127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	d := &APIRequestDataSource{client: NewClient(baseURL, "key", "test-agent", false, defaultRequestTimeout)}
	model := APIRequestDataSourceModel{
		Path:             types.StringValue(""),
		Method:           types.StringNull(),
		Headers:          types.MapNull(types.StringType),
		RequestBodyJSON:  types.StringNull(),
		ResponseBodyJSON: types.StringUnknown(),
		StatusCode:       types.Int64Unknown(),
	}
	config, state := apiRequestTestConfig(t, d, model)
	resp := datasource.ReadResponse{State: state}
	d.Read(context.Background(), datasource.ReadRequest{Config: config}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected empty path to produce a request diagnostic")
	}
}
