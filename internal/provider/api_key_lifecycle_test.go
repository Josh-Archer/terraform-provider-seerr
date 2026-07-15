package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func apiKeyResourceSchema(t *testing.T, r *APIKeyResource) resource.SchemaResponse {
	t.Helper()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("API key resource schema diagnostics: %v", resp.Diagnostics)
	}
	return resp
}

func apiKeyPlan(t *testing.T, schemaResp resource.SchemaResponse) tfsdk.Plan {
	t.Helper()
	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	model := APIKeyModel{ApiKey: types.StringUnknown(), StatusCode: types.Int64Unknown()}
	if diags := plan.Set(context.Background(), &model); diags.HasError() {
		t.Fatalf("set API key plan: %v", diags)
	}
	return plan
}

func TestAPIKeyResourceCreateAndRead(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case req.Method == http.MethodPost && req.URL.Path == "/api/v1/settings/main/regenerate":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"apiKey":"regenerated-key"}`))
		case req.Method == http.MethodGet && req.URL.Path == "/api/v1/settings/main":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"apiKey":"current-key"}`))
		default:
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		}
	}))
	defer srv.Close()
	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	r := &APIKeyResource{client: NewClient(baseURL, "old-key", "test-agent", false, defaultRequestTimeout)}
	schemaResp := apiKeyResourceSchema(t, r)

	createResp := resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Plan: apiKeyPlan(t, schemaResp)}, &createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("create diagnostics: %v", createResp.Diagnostics)
	}
	var model APIKeyModel
	if diags := createResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("get create state: %v", diags)
	}
	if model.ApiKey.ValueString() != "regenerated-key" || model.StatusCode.ValueInt64() != http.StatusCreated {
		t.Fatalf("unexpected create state: %#v", model)
	}

	readResp := resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, &readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", readResp.Diagnostics)
	}
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("get read state: %v", diags)
	}
	if model.ApiKey.ValueString() != "current-key" || model.StatusCode.ValueInt64() != http.StatusOK {
		t.Fatalf("unexpected read state: %#v", model)
	}
}

func TestAPIKeyResourceResponseErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{name: "http error", statusCode: http.StatusInternalServerError, body: `{"message":"boom"}`},
		{name: "invalid json", statusCode: http.StatusOK, body: `{`},
		{name: "missing api key", statusCode: http.StatusOK, body: `{"applicationUrl":"https://example.test"}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(test.statusCode)
				_, _ = w.Write([]byte(test.body))
			}))
			defer srv.Close()
			baseURL, err := url.Parse(srv.URL)
			if err != nil {
				t.Fatal(err)
			}
			r := &APIKeyResource{client: NewClient(baseURL, "key", "test-agent", false, defaultRequestTimeout)}
			schemaResp := apiKeyResourceSchema(t, r)

			createResp := resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
			r.Create(context.Background(), resource.CreateRequest{Plan: apiKeyPlan(t, schemaResp)}, &createResp)
			if !createResp.Diagnostics.HasError() {
				t.Fatal("expected invalid regenerate response to produce a diagnostic")
			}

			state := tfsdk.State{Schema: schemaResp.Schema}
			model := APIKeyModel{ApiKey: types.StringValue("old-key"), StatusCode: types.Int64Value(http.StatusOK)}
			if diags := state.Set(context.Background(), &model); diags.HasError() {
				t.Fatalf("set API key state: %v", diags)
			}
			readResp := resource.ReadResponse{State: state}
			r.Read(context.Background(), resource.ReadRequest{State: state}, &readResp)
			if !readResp.Diagnostics.HasError() {
				t.Fatal("expected invalid settings response to produce a diagnostic")
			}
		})
	}
}

func TestAPIKeyDataSourceReadAndHTTPError(t *testing.T) {
	for _, statusCode := range []int{http.StatusOK, http.StatusForbidden} {
		t.Run(http.StatusText(statusCode), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.Method != http.MethodGet || req.URL.Path != "/api/v1/settings/main" {
					t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
				}
				w.WriteHeader(statusCode)
				_, _ = w.Write([]byte(`{"apiKey":"data-source-key"}`))
			}))
			defer srv.Close()
			baseURL, err := url.Parse(srv.URL)
			if err != nil {
				t.Fatal(err)
			}
			d := &APIKeyDataSource{client: NewClient(baseURL, "key", "test-agent", false, defaultRequestTimeout)}
			var schemaResp datasource.SchemaResponse
			d.Schema(context.Background(), datasource.SchemaRequest{}, &schemaResp)
			configured := tfsdk.State{Schema: schemaResp.Schema}
			model := APIKeyDataSourceModel{ApiKey: types.StringUnknown(), StatusCode: types.Int64Unknown()}
			if diags := configured.Set(context.Background(), &model); diags.HasError() {
				t.Fatalf("set data source config: %v", diags)
			}
			config := tfsdk.Config{Raw: configured.Raw, Schema: schemaResp.Schema}
			resp := datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
			d.Read(context.Background(), datasource.ReadRequest{Config: config}, &resp)

			if statusCode != http.StatusOK {
				if !resp.Diagnostics.HasError() || !strings.Contains(resp.Diagnostics.Errors()[0].Detail(), "403") {
					t.Fatalf("expected 403 diagnostic, got %v", resp.Diagnostics)
				}
				return
			}
			if resp.Diagnostics.HasError() {
				t.Fatalf("read diagnostics: %v", resp.Diagnostics)
			}
			if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
				t.Fatalf("get data source state: %v", diags)
			}
			if model.ApiKey.ValueString() != "data-source-key" || model.StatusCode.ValueInt64() != http.StatusOK {
				t.Fatalf("unexpected data source state: %#v", model)
			}
		})
	}
}
