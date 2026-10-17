package admin

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestFindCodexNamedEmailMatchRefreshScope(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*codexImportAccount)
		want bool
	}{
		{"rotated RT retains identity", func(*codexImportAccount) {}, true},
		{"access-only stays separate", func(v *codexImportAccount) { v.RefreshToken = "" }, false},
		{"different client stays separate", func(v *codexImportAccount) { v.ClientID = "another-oauth-client" }, false},
		{"different workspace even in same org", func(v *codexImportAccount) { v.AccountID = "workspace-b" }, false},
		{"different user stays separate", func(v *codexImportAccount) { v.UserID = "user-b" }, false},
		{"different organization stays separate", func(v *codexImportAccount) { v.Organization = "org-b" }, false},
		{"agent identity stays separate", func(v *codexImportAccount) { v.IsAgentIdentity = true }, false},
		{"name must match exactly", func(v *codexImportAccount) { v.Name = "OTHER" }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			accounts := []service.Account{{ID: 48, Name: "offline-account", Platform: service.PlatformOpenAI, Type: service.AccountTypeOAuth,
				Credentials: map[string]any{"email": "offline@example.invalid", "client_id": openai.ClientID,
					"chatgpt_account_id": "workspace-a", "chatgpt_user_id": "user-a", "organization_id": "org-a", "refresh_token": "old"}}}
			item := &codexImportAccount{Name: "offline-account", Email: "offline@example.invalid", RefreshToken: "rotated",
				AccountID: "workspace-a", UserID: "user-a", Organization: "org-a"}
			tc.edit(item)
			matched := findCodexNamedRefreshMatch(accounts, item.Name, item)
			if tc.want {
				require.NotNil(t, matched)
				require.Equal(t, int64(48), matched.ID)
			} else {
				require.Nil(t, matched)
			}
		})
	}
}
