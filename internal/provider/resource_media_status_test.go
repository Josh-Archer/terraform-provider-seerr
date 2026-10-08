package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

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

func TestMediaStatusNameMapsWritableSeerrStatuses(t *testing.T) {
	for number, name := range map[int64]string{1: "unknown", 2: "pending", 3: "processing", 4: "partial", 5: "available"} {
		got, ok := mediaStatusName(number)
		require.True(t, ok)
		require.Equal(t, name, got)
	}
	_, ok := mediaStatusName(6) // blocklisted is not writable through the status route
	require.False(t, ok)
}
