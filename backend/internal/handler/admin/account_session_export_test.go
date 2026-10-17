package admin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/require"
)

type sessionExportAdminStub struct{ service.AdminService }

func (s *sessionExportAdminStub) GetAccount(_ context.Context, id int64) (*service.Account, error) {
	return &service.Account{ID: id, Platform: service.PlatformOpenAI, Type: service.AccountTypeOAuth, Credentials: map[string]any{"chatgpt_user_id": "user-test"}}, nil
}
func TestExportOpenAISessionValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &AccountHandler{adminService: &sessionExportAdminStub{}}
	r := gin.New()
	r.POST("/accounts/:id/export-session", h.ExportOpenAISession)
	for _, tc := range []struct {
		id, body string
		status   int
		message  string
	}{
		{"bad", `{}`, 400, "Invalid account ID"},
		{"0", `{}`, 400, "Invalid account ID"},
		{"1", `broken`, 400, "Invalid Session export request"},
		{"1", `{}`, 400, "Cookie is required"},
		{"1", `{"cookie":"` + strings.Repeat("x", 33<<10) + `"}`, 400, "Invalid Session export request"},
	} {
		w := httptest.NewRecorder()
		request := httptest.NewRequest("POST", "/accounts/"+tc.id+"/export-session", strings.NewReader(tc.body))
		request.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, request)
		require.Equal(t, tc.status, w.Code)
		require.Contains(t, w.Body.String(), tc.message)
		require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
	}
}

type sessionExportHandlerTransport func(*http.Request) (*http.Response, error)

func (f sessionExportHandlerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
func TestExportOpenAISessionReturnsPayloadWithoutEnvelopeInData(t *testing.T) {
	body := `{"user":{"id":"user-test"},"accessToken":"offline-web-token","expires":"2099-01-01T00:00:00Z","upstreamField":true}`
	oauth := service.NewOpenAIOAuthService(nil, nil)
	oauth.SetPrivacyClientFactory(func(string) (*req.Client, error) {
		client := req.C()
		client.GetClient().Transport = sessionExportHandlerTransport(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
		})
		return client, nil
	})
	h := &AccountHandler{adminService: &sessionExportAdminStub{}, openaiOAuthService: oauth}
	r := gin.New()
	r.POST("/accounts/:id/export-session", h.ExportOpenAISession)
	w := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/accounts/1/export-session", strings.NewReader(`{"cookie":"__Secure-next-auth.session-token=offline-cookie"}`))
	request.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, request)
	require.Equal(t, 200, w.Code)
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &envelope))
	require.JSONEq(t, body, string(envelope.Data))
	require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
	require.NotContains(t, w.Body.String(), "offline-cookie")
}

type sessionCookieHandlerRepo struct {
	ciphertext string
	deleted    bool
}

func (r *sessionCookieHandlerRepo) Get(context.Context, int64) (string, error) {
	return r.ciphertext, nil
}
func (r *sessionCookieHandlerRepo) Save(_ context.Context, _ int64, ciphertext string) error {
	r.ciphertext = ciphertext
	return nil
}
func (r *sessionCookieHandlerRepo) Delete(context.Context, int64) error {
	r.deleted = true
	r.ciphertext = ""
	return nil
}
func TestOpenAISessionCookieHandlersNeverReturnSecrets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &sessionCookieHandlerRepo{ciphertext: "private-ciphertext"}
	oauth := service.NewOpenAIOAuthService(nil, nil)
	oauth.SetSessionCookieStorage(repo, nil, false)
	h := &AccountHandler{adminService: &sessionExportAdminStub{}, openaiOAuthService: oauth}
	router := gin.New()
	router.GET("/accounts/:id/session-cookie", h.GetOpenAISessionCookie)
	router.DELETE("/accounts/:id/session-cookie", h.DeleteOpenAISessionCookie)
	for _, method := range []string{http.MethodGet, http.MethodDelete} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(method, "/accounts/7/session-cookie", nil))
		require.Equal(t, http.StatusOK, recorder.Code)
		require.NotContains(t, recorder.Body.String(), "private-ciphertext")
		require.Equal(t, "no-store", recorder.Header().Get("Cache-Control"))
		if method == http.MethodGet {
			require.Contains(t, recorder.Body.String(), `"configured":true`)
			require.Contains(t, recorder.Body.String(), `"storage_enabled":false`)
		}
	}
	require.True(t, repo.deleted)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/accounts/0/session-cookie", nil))
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}
