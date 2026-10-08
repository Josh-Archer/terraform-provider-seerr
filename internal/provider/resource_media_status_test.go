package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestMediaStatusApplySends4KStatusAndReturnsAPIError(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	statusCode := http.StatusOK
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		gotPath = req.URL.Path
		require.Equal(t, http.MethodPost, req.Method)
		require.NoError(t, json.NewDecoder(req.Body).Decode(&gotBody))
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(`{"id":7}`))
	}))
	defer srv.Close()
	baseURL, err := url.Parse(srv.URL)
	require.NoError(t, err)
	r := &MediaStatusResource{client: NewClient(baseURL, "test-key", "test-agent", false, defaultRequestTimeout, 0, 0)}
	data := MediaStatusModel{MediaID: types.StringValue("7"), Status: types.StringValue("available"), Is4K: types.BoolValue(true)}
	require.NoError(t, r.apply(context.Background(), &data))
	require.Equal(t, "/api/v1/media/7/available", gotPath)
	require.Equal(t, true, gotBody["is4k"])
	require.Equal(t, "7", data.ID.ValueString())

	statusCode = http.StatusBadRequest
	err = r.apply(context.Background(), &data)
	require.ErrorContains(t, err, "status 400")
}

func TestMediaStatusReadPreservesReadOnlyStatuses(t *testing.T) {
	tests := []struct {
		name   string
		status int
		is4k   bool
		want   string
	}{
		{name: "regular blocklisted", status: 6, want: "blocklisted"},
		{name: "4k deleted", status: 7, is4k: true, want: "deleted"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(fmt.Sprintf(`{"id":7,"status":%d,"status4k":%d}`, tt.status, tt.status)))
			}))
			defer srv.Close()
			baseURL, err := url.Parse(srv.URL)
			require.NoError(t, err)
			r := &MediaStatusResource{client: NewClient(baseURL, "test-key", "test-agent", false, defaultRequestTimeout, 0, 0)}
			var schemaResp resource.SchemaResponse
			r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
			state := tfsdk.State{Schema: schemaResp.Schema}
			setDiags := state.Set(context.Background(), &MediaStatusModel{ID: types.StringValue("7"), MediaID: types.StringValue("7"), Status: types.StringValue("available"), Is4K: types.BoolValue(tt.is4k), Triggers: types.MapNull(types.StringType)})
			require.False(t, setDiags.HasError(), "%v", setDiags)
			resp := resource.ReadResponse{State: state}
			r.Read(context.Background(), resource.ReadRequest{State: state}, &resp)
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			var got MediaStatusModel
			require.False(t, resp.State.Get(context.Background(), &got).HasError())
			require.Equal(t, tt.want, got.Status.ValueString())
		})
	}
}

func TestMediaStatusNameMapsSeerrStatuses(t *testing.T) {
	for number, name := range map[int64]string{1: "unknown", 2: "pending", 3: "processing", 4: "partial", 5: "available", 6: "blocklisted", 7: "deleted"} {
		got, ok := mediaStatusName(number)
		require.True(t, ok)
		require.Equal(t, name, got)
	}
	_, ok := mediaStatusName(8)
	require.False(t, ok)
}
