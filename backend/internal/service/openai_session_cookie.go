package service

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// OpenAISessionCookieRepository keeps ciphertext outside account JSON exports.
type OpenAISessionCookieRepository interface {
	Get(context.Context, int64) (string, error)
	Save(context.Context, int64, string) error
	Delete(context.Context, int64) error
}

type OpenAISessionCookieState struct {
	Configured     bool `json:"configured"`
	StorageEnabled bool `json:"storage_enabled"`
}

type storedChatGPTCookie struct {
	AccountID int64  `json:"account_id"`
	Cookie    string `json:"cookie"`
}

func (s *OpenAIOAuthService) SetSessionCookieStorage(repo OpenAISessionCookieRepository, encryptor SecretEncryptor, fixedKey bool) {
	s.sessionCookieRepo = repo
	s.sessionCookieEncryptor = encryptor
	s.sessionCookieStorageEnabled = fixedKey && repo != nil && encryptor != nil
}

func validateSessionAccount(account *Account) error {
	if account == nil || account.Platform != PlatformOpenAI || account.Type != AccountTypeOAuth || account.ParentAccountID != nil {
		return infraerrors.BadRequest("OPENAI_SESSION_ACCOUNT_UNSUPPORTED", "Session export requires an OpenAI OAuth account with its own credentials")
	}
	return nil
}

func sessionStorageError() error {
	return infraerrors.ServiceUnavailable("OPENAI_SESSION_STORAGE_FAILED", "Could not access stored ChatGPT login credentials")
}

func (s *OpenAIOAuthService) ChatGPTSessionCookieState(ctx context.Context, account *Account) (*OpenAISessionCookieState, error) {
	if err := validateSessionAccount(account); err != nil {
		return nil, err
	}
	if s == nil || s.sessionCookieRepo == nil {
		return nil, sessionStorageError()
	}
	ciphertext, err := s.sessionCookieRepo.Get(ctx, account.ID)
	if err != nil {
		return nil, sessionStorageError()
	}
	return &OpenAISessionCookieState{Configured: ciphertext != "", StorageEnabled: s.sessionCookieStorageEnabled}, nil
}

func (s *OpenAIOAuthService) DeleteChatGPTSessionCookie(ctx context.Context, account *Account) error {
	if err := validateSessionAccount(account); err != nil {
		return err
	}
	if s == nil || s.sessionCookieRepo == nil {
		return sessionStorageError()
	}
	if err := s.sessionCookieRepo.Delete(ctx, account.ID); err != nil {
		return sessionStorageError()
	}
	return nil
}

// ExportSavedChatGPTSession accepts a one-off Cookie or reuses encrypted login
// credentials. A replacement is stored only after a successful identity check.
func (s *OpenAIOAuthService) ExportSavedChatGPTSession(ctx context.Context, account *Account, cookie string, remember bool) (json.RawMessage, error) {
	if err := validateSessionAccount(account); err != nil {
		return nil, err
	}
	cookie = strings.TrimSpace(cookie)
	if remember && (s == nil || !s.sessionCookieStorageEnabled) {
		return nil, infraerrors.BadRequest("OPENAI_SESSION_ENCRYPTION_REQUIRED", "Configure a fixed TOTP_ENCRYPTION_KEY before storing ChatGPT login credentials")
	}
	if remember && cookie == "" {
		return nil, infraerrors.BadRequest("OPENAI_SESSION_COOKIE_REQUIRED", "A new ChatGPT login Cookie is required to save credentials")
	}
	if cookie == "" && s != nil && s.sessionCookieRepo != nil {
		ciphertext, err := s.sessionCookieRepo.Get(ctx, account.ID)
		if err != nil {
			return nil, sessionStorageError()
		}
		if ciphertext != "" {
			if !s.sessionCookieStorageEnabled {
				return nil, infraerrors.BadRequest("OPENAI_SESSION_ENCRYPTION_REQUIRED", "Restore the original TOTP_ENCRYPTION_KEY to use stored ChatGPT login credentials")
			}
			plaintext, err := s.sessionCookieEncryptor.Decrypt(ciphertext)
			var stored storedChatGPTCookie
			if err != nil || json.Unmarshal([]byte(plaintext), &stored) != nil || stored.AccountID != account.ID || stored.Cookie == "" {
				return nil, infraerrors.New(http.StatusServiceUnavailable, "OPENAI_SESSION_DECRYPT_FAILED", "Stored ChatGPT login credentials cannot be decrypted; restore the original encryption key or replace the Cookie")
			}
			cookie = stored.Cookie
		}
	}
	body, err := s.ExportChatGPTSession(ctx, account, cookie)
	if err != nil {
		return nil, err
	}
	if remember {
		plaintext, err := json.Marshal(storedChatGPTCookie{AccountID: account.ID, Cookie: cookie})
		if err != nil {
			return nil, sessionStorageError()
		}
		ciphertext, err := s.sessionCookieEncryptor.Encrypt(string(plaintext))
		if err != nil {
			return nil, sessionStorageError()
		}
		if err := s.sessionCookieRepo.Save(ctx, account.ID, ciphertext); err != nil {
			return nil, sessionStorageError()
		}
	}
	return body, nil
}
