package repository

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestClaudeResetHTTPContract(t *testing.T) {
	var query, ua, auth, path string
	var payload map[string]string
	posts := 0
	srv := newLocalTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		ua = r.UserAgent()
		auth = r.Header.Get("Authorization")
		path = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"cedar_ember":{"eligible":true,"grants":[{"id":"grant","resets_left":1}]}}`))
			return
		}
		posts++
		_ = json.NewDecoder(r.Body).Decode(&payload)
		_, _ = w.Write([]byte(`{"result":"reset","cleared":["five_hour","seven_day"]}`))
	}))
	defer srv.Close()
	f := &claudeUsageService{usageURL: srv.URL, allowPrivateHosts: true}
	usage, err := f.FetchUsage(context.Background(), "synthetic", "")
	require.NoError(t, err)
	require.Contains(t, query, "cedar_ember=1")
	require.Contains(t, query, "skip_spend=1")
	require.Contains(t, ua, "claude-cli/")
	require.Equal(t, 1, usage.CedarEmber.Grants[0].ResetsLeft)
	require.Equal(t, 0, posts)
	result, err := f.ResetQuota(context.Background(), &service.ClaudeUsageFetchOptions{AccessToken: "synthetic"}, "org", "grant", "request-id")
	require.NoError(t, err)
	require.Equal(t, "reset", result.Result)
	require.Equal(t, 1, posts)
	require.Equal(t, "/api/organizations/org/reset_rate_limits", path)
	require.Equal(t, "Bearer synthetic", auth)
	require.Equal(t, map[string]string{"program": "cedar_ember", "grant_id": "grant", "request_id": "request-id"}, payload)
}
func TestClaudeResetHTTPDoesNotRetryOrExposeRawErrors(t *testing.T) {
	for _, body := range []string{`not-json`, `{"result":"unknown"}`, `{"error":"secret-upstream-body"}`} {
		t.Run(body, func(t *testing.T) {
			calls := 0
			srv := newLocalTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; _, _ = w.Write([]byte(body)) }))
			defer srv.Close()
			f := &claudeUsageService{usageURL: srv.URL, allowPrivateHosts: true}
			_, err := f.ResetQuota(context.Background(), &service.ClaudeUsageFetchOptions{AccessToken: "synthetic"}, "org", "grant", "id")
			require.Error(t, err)
			require.NotContains(t, err.Error(), "secret-upstream-body")
			require.Equal(t, 1, calls)
		})
	}
}
