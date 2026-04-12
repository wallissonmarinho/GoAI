package gemini

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckKeys_ok(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1beta/models", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"models":[{"name":"models/gemini-2.0-flash"}]}`))
	}))
	defer srv.Close()

	out := CheckKeys(context.Background(), srv.Client(), srv.URL, "test", []string{"k1"})
	require.Len(t, out, 1)
	require.True(t, out[0].OK)
	require.Equal(t, "ok", out[0].Status)
	require.Equal(t, 200, out[0].HTTPStatus)
}

func TestCheckKeys_quotaJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"error":{"code":429,"message":"Resource exhausted","status":"RESOURCE_EXHAUSTED"}}`))
	}))
	defer srv.Close()

	out := CheckKeys(context.Background(), srv.Client(), srv.URL, "test", []string{"k1"})
	require.Len(t, out, 1)
	require.False(t, out[0].OK)
	require.Equal(t, "quota_or_rate", out[0].Status)
}

func TestCheckKeys_429(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`too fast`))
	}))
	defer srv.Close()

	out := CheckKeys(context.Background(), srv.Client(), srv.URL, "test", []string{"k1"})
	require.Equal(t, "quota_or_rate", out[0].Status)
	require.Contains(t, out[0].Detail, "too fast")
}

func TestCheckKeys_emptyKeys(t *testing.T) {
	require.Nil(t, CheckKeys(context.Background(), http.DefaultClient, "http://x", "ua", nil))
}

func TestClassifyAPIError(t *testing.T) {
	require.Equal(t, "quota_or_rate", classifyAPIError("Quota exceeded", ""))
	require.Equal(t, "auth_error", classifyAPIError("API key invalid", ""))
	require.Equal(t, "api_error", classifyAPIError("something else", ""))
}
