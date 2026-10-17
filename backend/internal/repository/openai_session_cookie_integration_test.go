//go:build integration

package repository

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/require"
)

type sessionCookieIntegrationTransport func(*http.Request) (*http.Response, error)

func (f sessionCookieIntegrationTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestOpenAISessionCookieEncryptedPersistence(t *testing.T) {
	ctx := context.Background()
	account := mustCreateAccount(t, integrationEntClient, &service.Account{Name: "session-cookie-integration", Platform: service.PlatformOpenAI, Type: service.AccountTypeOAuth, Credentials: map[string]any{"chatgpt_user_id": "offline-user"}})
	store := NewOpenAISessionCookieRepository(integrationDB)
	t.Cleanup(func() { _ = store.Delete(ctx, account.ID) })
	cfg := &config.Config{Totp: config.TotpConfig{EncryptionKey: strings.Repeat("42", 32)}}
	calls := 0
	makeService := func() *service.OpenAIOAuthService {
		encryptor, err := NewAESEncryptor(cfg)
		require.NoError(t, err)
		svc := service.NewOpenAIOAuthService(nil, nil)
		svc.SetSessionCookieStorage(store, encryptor, true)
		svc.SetPrivacyClientFactory(func(string) (*req.Client, error) {
			client := req.C()
			client.GetClient().Transport = sessionCookieIntegrationTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				require.Equal(t, "__Secure-next-auth.session-token=offline-cookie", r.Header.Get("Cookie"))
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"user":{"id":"offline-user"},"accessToken":"offline-token","expires":"2099-01-01T00:00:00Z"}`))}, nil
			})
			return client, nil
		})
		return svc
	}
	svc := makeService()
	_, err := svc.ExportSavedChatGPTSession(ctx, account, "__Secure-next-auth.session-token=offline-cookie", true)
	require.NoError(t, err)
	ciphertext, err := store.Get(ctx, account.ID)
	require.NoError(t, err)
	require.NotEmpty(t, ciphertext)
	require.NotContains(t, ciphertext, "offline-cookie")
	restarted := makeService()
	body, err := restarted.ExportSavedChatGPTSession(ctx, account, "", false)
	require.NoError(t, err)
	require.True(t, json.Valid(body))
	require.Equal(t, 2, calls)
	state, err := restarted.ChatGPTSessionCookieState(ctx, account)
	require.NoError(t, err)
	require.True(t, state.Configured)
	require.NoError(t, restarted.DeleteChatGPTSessionCookie(ctx, account))
	ciphertext, err = store.Get(ctx, account.ID)
	require.NoError(t, err)
	require.Empty(t, ciphertext)
	_, err = svc.ExportSavedChatGPTSession(ctx, account, "__Secure-next-auth.session-token=offline-cookie", true)
	require.NoError(t, err)
	accounts := NewAccountRepository(integrationEntClient, integrationDB, nil)
	require.NoError(t, accounts.Delete(ctx, account.ID))
	ciphertext, err = store.Get(ctx, account.ID)
	require.NoError(t, err)
	require.Empty(t, ciphertext)
	require.Error(t, store.Save(ctx, account.ID, "encrypted-fixture"))
}
