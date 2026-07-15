package provider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func apiObjectTestSchema(t *testing.T, r *APIObjectResource) resource.SchemaResponse {
	t.Helper()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("API object schema diagnostics: %v", resp.Diagnostics)
	}
	return resp
}

func apiObjectTestModel(path string) APIObjectModel {
	return APIObjectModel{
		ID:               types.StringUnknown(),
		Path:             types.StringValue(path),
		Headers:          types.MapNull(types.StringType),
		RequestBodyJSON:  types.StringValue(`{"phase":"create"}`),
		DeleteBodyJSON:   types.StringValue(`{"cleanup":true}`),
		ReadMethod:       types.StringValue(http.MethodGet),
		CreateMethod:     types.StringValue(http.MethodPost),
		UpdateMethod:     types.StringValue(http.MethodPatch),
		DeleteMethod:     types.StringValue(http.MethodDelete),
		SkipDelete:       types.BoolValue(false),
		SuppressNotFound: types.BoolValue(true),
		ResponseBodyJSON: types.StringUnknown(),
		StatusCode:       types.Int64Unknown(),
	}
}

func apiObjectTestPlan(t *testing.T, schemaResp resource.SchemaResponse, model APIObjectModel) tfsdk.Plan {
	t.Helper()
	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	if diags := plan.Set(context.Background(), &model); diags.HasError() {
		t.Fatalf("set API object plan: %v", diags)
	}
	return plan
}

func apiObjectTestState(t *testing.T, schemaResp resource.SchemaResponse, model APIObjectModel) tfsdk.State {
	t.Helper()
	state := tfsdk.State{Schema: schemaResp.Schema}
	if diags := state.Set(context.Background(), &model); diags.HasError() {
		t.Fatalf("set API object state: %v", diags)
	}
	return state
}

func TestAPIObjectLifecycle(t *testing.T) {
	var requests []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		requests = append(requests, fmt.Sprintf("%s %s %s", req.Method, req.URL.Path, body))
		w.Header().Set("Content-Type", "application/json")
		switch req.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":42,"phase":"created"}`))
		case http.MethodPatch:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":42,"phase":"updated"}`))
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":42,"phase":"read"}`))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected method %s", req.Method)
		}
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	r := &APIObjectResource{client: NewClient(baseURL, "key", "test-agent", false, defaultRequestTimeout)}
	schemaResp := apiObjectTestSchema(t, r)
	model := apiObjectTestModel("/api/v1/example")

	createResp := resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Plan: apiObjectTestPlan(t, schemaResp, model)}, &createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("create diagnostics: %v", createResp.Diagnostics)
	}
	if diags := createResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("get create state: %v", diags)
	}
	if got := model.ID.ValueString(); got != "42" {
		t.Fatalf("create ID = %q, want 42", got)
	}
	if got := model.StatusCode.ValueInt64(); got != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", got, http.StatusCreated)
	}

	model.RequestBodyJSON = types.StringValue(`{"phase":"update"}`)
	updateResp := resource.UpdateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Update(context.Background(), resource.UpdateRequest{Plan: apiObjectTestPlan(t, schemaResp, model)}, &updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("update diagnostics: %v", updateResp.Diagnostics)
	}

	readResp := resource.ReadResponse{State: updateResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: updateResp.State}, &readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", readResp.Diagnostics)
	}
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("get read state: %v", diags)
	}
	if got := model.ResponseBodyJSON.ValueString(); !strings.Contains(got, `"phase":"read"`) {
		t.Fatalf("read response = %q", got)
	}

	var deleteResp resource.DeleteResponse
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp.State}, &deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("delete diagnostics: %v", deleteResp.Diagnostics)
	}

	wantRequests := []string{
		`POST /api/v1/example {"phase":"create"}`,
		`PATCH /api/v1/example {"phase":"update"}`,
		`GET /api/v1/example `,
		`DELETE /api/v1/example {"cleanup":true}`,
	}
	if got := strings.Join(requests, "\n"); got != strings.Join(wantRequests, "\n") {
		t.Fatalf("requests:\n%s\nwant:\n%s", got, strings.Join(wantRequests, "\n"))
	}
}

func TestAPIObjectLifecycleHTTPFailures(t *testing.T) {
	tests := []struct {
		name string
		run  func(context.Context, *APIObjectResource, resource.SchemaResponse) bool
	}{
		{
			name: "create",
			run: func(ctx context.Context, r *APIObjectResource, schemaResp resource.SchemaResponse) bool {
				var resp resource.CreateResponse
				resp.State = tfsdk.State{Schema: schemaResp.Schema}
				r.Create(ctx, resource.CreateRequest{Plan: apiObjectTestPlan(t, schemaResp, apiObjectTestModel("/failure"))}, &resp)
				return resp.Diagnostics.HasError()
			},
		},
		{
			name: "read",
			run: func(ctx context.Context, r *APIObjectResource, schemaResp resource.SchemaResponse) bool {
				model := apiObjectTestModel("/failure")
				model.ID = types.StringValue("failure")
				model.SuppressNotFound = types.BoolValue(false)
				state := apiObjectTestState(t, schemaResp, model)
				resp := resource.ReadResponse{State: state}
				r.Read(ctx, resource.ReadRequest{State: state}, &resp)
				return resp.Diagnostics.HasError()
			},
		},
		{
			name: "update",
			run: func(ctx context.Context, r *APIObjectResource, schemaResp resource.SchemaResponse) bool {
				var resp resource.UpdateResponse
				resp.State = tfsdk.State{Schema: schemaResp.Schema}
				r.Update(ctx, resource.UpdateRequest{Plan: apiObjectTestPlan(t, schemaResp, apiObjectTestModel("/failure"))}, &resp)
				return resp.Diagnostics.HasError()
			},
		},
		{
			name: "delete",
			run: func(ctx context.Context, r *APIObjectResource, schemaResp resource.SchemaResponse) bool {
				model := apiObjectTestModel("/failure")
				model.ID = types.StringValue("failure")
				state := apiObjectTestState(t, schemaResp, model)
				var resp resource.DeleteResponse
				r.Delete(ctx, resource.DeleteRequest{State: state}, &resp)
				return resp.Diagnostics.HasError()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"message":"boom"}`))
			}))
			defer srv.Close()
			baseURL, err := url.Parse(srv.URL)
			if err != nil {
				t.Fatal(err)
			}
			r := &APIObjectResource{client: NewClient(baseURL, "key", "test-agent", false, defaultRequestTimeout)}
			if !test.run(context.Background(), r, apiObjectTestSchema(t, r)) {
				t.Fatal("expected HTTP 500 to produce an error diagnostic")
			}
		})
	}
}

func TestAPIObjectReadSuppressesNotFoundAndDeleteIsIdempotent(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	r := &APIObjectResource{client: NewClient(baseURL, "key", "test-agent", false, defaultRequestTimeout)}
	schemaResp := apiObjectTestSchema(t, r)
	model := apiObjectTestModel("/missing")
	model.ID = types.StringValue("missing")
	state := apiObjectTestState(t, schemaResp, model)

	readResp := resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, &readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", readResp.Diagnostics)
	}
	if !readResp.State.Raw.IsNull() {
		t.Fatal("expected suppressed 404 to remove resource state")
	}

	var deleteResp resource.DeleteResponse
	r.Delete(context.Background(), resource.DeleteRequest{State: state}, &deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("delete diagnostics: %v", deleteResp.Diagnostics)
	}
	if requests != 2 {
		t.Fatalf("requests = %d, want 2", requests)
	}

	model.SkipDelete = types.BoolValue(true)
	state = apiObjectTestState(t, schemaResp, model)
	r.Delete(context.Background(), resource.DeleteRequest{State: state}, &deleteResp)
	if requests != 2 {
		t.Fatalf("skip_delete issued an HTTP request; requests = %d", requests)
	}
}
