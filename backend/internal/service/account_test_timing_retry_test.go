//go:build unit

package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/requesttiming"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/stretchr/testify/require"
)

type accountTimingRetryStub struct{ calls int }

func (s *accountTimingRetryStub) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	s.calls++
	status := http.StatusTooManyRequests
	if s.calls > 1 {
		status = http.StatusOK
	}
	return &http.Response{StatusCode: status, Request: req, Header: http.Header{"Retry-After": {"0"}}, Body: io.NopCloser(strings.NewReader("synthetic"))}, nil
}

func (s *accountTimingRetryStub) DoWithTLS(req *http.Request, proxy string, account int64, concurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return s.Do(req, proxy, account, concurrency)
}

func TestAccountConnectionTimingPreservesCustom429Retry(t *testing.T) {
	for _, mode := range []string{"http", "tls", "oauth"} {
		t.Run(mode, func(t *testing.T) {
			collector := requesttiming.New(time.Now(), 0)
			req, err := http.NewRequestWithContext(requesttiming.With(context.Background(), collector), http.MethodPost, "https://example.invalid/test", bytes.NewBufferString(`{"model":"synthetic"}`))
			require.NoError(t, err)
			budget := 1
			account := &Account{ID: 42, Concurrency: 1, RateLimit429RetryCount: &budget}
			upstream := &accountTimingRetryStub{}
			var resp *http.Response
			switch mode {
			case "http":
				resp, err = doAccountHTTPUpstream(timedAccountTestUpstream(upstream), req, "", account)
			case "tls":
				resp, err = doAccountHTTPUpstreamWithTLS(timedAccountTestUpstream(upstream), req, "", account, nil)
			case "oauth":
				svc := &AccountTestService{httpUpstream: upstream}
				resp, err = svc.doOpenAIAccountTestUpstream(req, "", account, false, nil)
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			require.Equal(t, 2, upstream.calls)
			require.NoError(t, resp.Body.Close())
			var snapshot requesttiming.Snapshot
			collector.WhenFinished(func(value requesttiming.Snapshot) { snapshot = value })
			collector.Finish(http.StatusOK, false)
			require.Len(t, snapshot.Attempts, 2)
			require.Equal(t, http.StatusTooManyRequests, snapshot.Attempts[0].Status)
			require.Equal(t, http.StatusOK, snapshot.Attempts[1].Status)
		})
	}
}
