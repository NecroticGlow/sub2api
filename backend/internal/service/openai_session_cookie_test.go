package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/require"
)

type sessionCookieTestRepo struct {
	data   map[int64]string
	writes int
	err    error
}

func (r *sessionCookieTestRepo) Get(_ context.Context, id int64) (string, error) {
	return r.data[id], r.err
}
func (r *sessionCookieTestRepo) Save(_ context.Context, id int64, value string) error {
	if r.err != nil {
		return r.err
	}
	r.data[id] = value
	r.writes++
	return nil
}
func (r *sessionCookieTestRepo) Delete(_ context.Context, id int64) error {
	if r.err != nil {
		return r.err
	}
	delete(r.data, id)
	return nil
}

// A deterministic test double; production uses the existing AES-GCM encryptor.
type sessionCookieTestEncryptor struct{ err error }

func (e sessionCookieTestEncryptor) Encrypt(s string) (string, error) {
	return base64.StdEncoding.EncodeToString([]byte(s)), e.err
}
func (e sessionCookieTestEncryptor) Decrypt(s string) (string, error) {
	if e.err != nil {
		return "", e.err
	}
	b, err := base64.StdEncoding.DecodeString(s)
	return string(b), err
}

func sessionCookieTestService(t *testing.T) (*OpenAIOAuthService, *sessionCookieTestRepo, *Account, *int) {
	t.Helper()
	repo := &sessionCookieTestRepo{data: map[int64]string{}}
	account := sessionExportAccount()
	account.ID = 7
	calls := new(int)
	svc := &OpenAIOAuthService{privacyClientFactory: func(string) (*req.Client, error) {
		client := req.C()
		client.GetClient().Transport = sessionExportTransport(func(r *http.Request) (*http.Response, error) {
			*calls++
			require.Equal(t, "__Secure-next-auth.session-token.0=first; __Secure-next-auth.session-token.1=second", r.Header.Get("Cookie"))
			body := `{"user":{"id":"user-test","email":"test@example.test"},"accessToken":"offline-token","expires":"2099-01-01T00:00:00Z"}`
			return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
		})
		return client, nil
	}}
	svc.SetSessionCookieStorage(repo, sessionCookieTestEncryptor{}, true)
	return svc, repo, account, calls
}

const testSavedSessionCookie = "__Secure-next-auth.session-token.0=first; __Secure-next-auth.session-token.1=second"

func TestSavedChatGPTSessionCookieLifecycle(t *testing.T) {
	ctx := context.Background()
	svc, repo, account, calls := sessionCookieTestService(t)
	state, err := svc.ChatGPTSessionCookieState(ctx, account)
	require.NoError(t, err)
	require.False(t, state.Configured)
	require.True(t, state.StorageEnabled)
	body, err := svc.ExportSavedChatGPTSession(ctx, account, testSavedSessionCookie, true)
	require.NoError(t, err)
	require.Contains(t, string(body), "offline-token")
	require.Equal(t, 1, repo.writes)
	require.NotContains(t, repo.data[7], "session-token")
	state, err = svc.ChatGPTSessionCookieState(ctx, account)
	require.NoError(t, err)
	require.True(t, state.Configured)
	stateJSON, err := json.Marshal(state)
	require.NoError(t, err)
	require.JSONEq(t, `{"configured":true,"storage_enabled":true}`, string(stateJSON))
	// Re-create the service to ensure this uses persistent storage, not process memory.
	restarted, _, _, _ := sessionCookieTestService(t)
	restarted.SetSessionCookieStorage(repo, sessionCookieTestEncryptor{}, true)
	_, err = restarted.ExportSavedChatGPTSession(ctx, account, "", false)
	require.NoError(t, err)
	require.Equal(t, 1, repo.writes)
	_, err = svc.ExportSavedChatGPTSession(ctx, account, testSavedSessionCookie, false)
	require.NoError(t, err)
	require.Equal(t, 1, repo.writes)
	require.Equal(t, 2, *calls)
	// Deletion works even if the encryption key is no longer available.
	svc.SetSessionCookieStorage(repo, nil, false)
	require.NoError(t, svc.DeleteChatGPTSessionCookie(ctx, account))
	require.Empty(t, repo.data)
	_, err = svc.ExportSavedChatGPTSession(ctx, account, "", false)
	require.ErrorContains(t, err, "Cookie is required")
}

func TestSavedChatGPTSessionCookieRejectsInvalidReplacement(t *testing.T) {
	svc, repo, account, _ := sessionCookieTestService(t)
	_, err := svc.ExportSavedChatGPTSession(context.Background(), account, testSavedSessionCookie, true)
	require.NoError(t, err)
	original := repo.data[7]
	account.Credentials["chatgpt_user_id"] = "other-user"
	_, err = svc.ExportSavedChatGPTSession(context.Background(), account, testSavedSessionCookie, true)
	require.ErrorContains(t, err, "different user")
	require.Equal(t, original, repo.data[7])
	require.Equal(t, 1, repo.writes)
	_, err = svc.ExportSavedChatGPTSession(context.Background(), account, "", false)
	require.ErrorContains(t, err, "different user")
}

func TestSavedChatGPTSessionCookieStorageFailures(t *testing.T) {
	for _, scenario := range []string{"no fixed key", "corrupt ciphertext", "different account", "database read", "decrypt error", "encrypt error", "database write", "no new cookie"} {
		t.Run(scenario, func(t *testing.T) {
			svc, repo, account, calls := sessionCookieTestService(t)
			cookie, remember := "", false
			switch scenario {
			case "no fixed key":
				svc.sessionCookieStorageEnabled = false
				cookie = testSavedSessionCookie
				remember = true
			case "corrupt ciphertext":
				repo.data[7] = "invalid-ciphertext"
			case "different account":
				repo.data[7], _ = (sessionCookieTestEncryptor{}).Encrypt(`{"account_id":8,"cookie":"secret"}`)
			case "database read":
				repo.err = errors.New("sensitive-diagnostic")
			case "decrypt error":
				repo.data[7] = "present"
				svc.sessionCookieEncryptor = sessionCookieTestEncryptor{err: errors.New("sensitive-diagnostic")}
			case "encrypt error":
				cookie = testSavedSessionCookie
				remember = true
				svc.sessionCookieEncryptor = sessionCookieTestEncryptor{err: errors.New("sensitive-diagnostic")}
			case "database write":
				cookie = testSavedSessionCookie
				remember = true
				repo.err = errors.New("sensitive-diagnostic")
			case "no new cookie":
				remember = true
			}
			body, err := svc.ExportSavedChatGPTSession(context.Background(), account, cookie, remember)
			require.Error(t, err)
			require.Empty(t, body)
			require.NotContains(t, err.Error(), "sensitive-diagnostic")
			require.Zero(t, repo.writes)
			if scenario != "encrypt error" && scenario != "database write" {
				require.Zero(t, *calls)
			}
		})
	}
}

func TestSavedChatGPTSessionCookieStateAndDeleteErrorsAreSanitized(t *testing.T) {
	svc, repo, account, _ := sessionCookieTestService(t)
	repo.err = errors.New("sensitive-diagnostic")
	_, err := svc.ChatGPTSessionCookieState(context.Background(), account)
	require.Error(t, err)
	require.NotContains(t, err.Error(), "sensitive-diagnostic")
	err = svc.DeleteChatGPTSessionCookie(context.Background(), account)
	require.Error(t, err)
	require.NotContains(t, err.Error(), "sensitive-diagnostic")
	account.Type = AccountTypeAPIKey
	_, err = svc.ChatGPTSessionCookieState(context.Background(), account)
	require.ErrorContains(t, err, "OpenAI OAuth")
	err = svc.DeleteChatGPTSessionCookie(context.Background(), account)
	require.ErrorContains(t, err, "OpenAI OAuth")
}
