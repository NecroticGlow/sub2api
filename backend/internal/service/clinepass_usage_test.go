package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func clinePassTestAccount(baseURL string) *Account {
	return &Account{ID: 1135, Platform: PlatformDeepseek, Type: AccountTypeAPIKey,
		Credentials: map[string]any{"base_url": baseURL, "api_key": "test-cline-key", "account_mode": AccountModePayG}}
}

func TestClinePassAccountEligibility(t *testing.T) {
	for _, tc := range []struct {
		base string
		want bool
	}{
		{"https://api.cline.bot", true},
		{"https://api.cline.bot/api/v1", true},
		{"https://API.CLINE.BOT:443/api/v1", true},
		{"http://api.cline.bot", false},
		{"https://api.cline.bot:8443", false},
		{"https://api.cline.bot.attacker.example", false},
		{"https://attacker.example/api.cline.bot", false},
		{"https://api.cline.bot@attacker.example", false},
		{"https://someone@api.cline.bot", false},
		{"https://api.deepseek.com", false},
		{"https://ollama.com", false},
		{"", false},
	} {
		t.Run(tc.base, func(t *testing.T) {
			account := clinePassTestAccount(tc.base)
			require.Equal(t, tc.want, account.IsClinePassAccount())
			if tc.want {
				require.Equal(t, PlatformDeepseek, account.GetCodingPlanProvider())
				require.NoError(t, validateCodingPlanAccount(account))
			}
		})
	}
	account := clinePassTestAccount("https://api.cline.bot")
	account.Platform = PlatformOpenAI
	require.False(t, account.IsClinePassAccount())
	account.Platform, account.Type = PlatformDeepseek, AccountTypeOAuth
	require.False(t, account.IsClinePassAccount())
	var empty *Account
	require.False(t, empty.IsClinePassAccount())
}

const clinePassUsageFixture = `{"success":true,"data":{"limits":[
  {"type":"monthly","percentUsed":50,"resetsAt":"2026-10-20T04:42:27.567187626Z"},
  {"type":"five_hour","percentUsed":2,"resetsAt":"2026-09-30T15:43:33.563387847Z"},
  {"type":"weekly","percentUsed":51,"resetsAt":"2026-10-04T05:09:08.565240917Z"}
]}}`

func TestClinePassUsageAccountSwitch(t *testing.T) {
	account := clinePassTestAccount("https://api.cline.bot/v1")
	require.True(t, account.IsClinePassUsageEnabled(), "legacy accounts stay enabled")
	for _, enabled := range []any{true, "true"} {
		account.Credentials["clinepass_usage_enabled"] = enabled
		require.True(t, account.IsClinePassUsageEnabled())
	}
	for _, disabled := range []any{false, "false", nil, "invalid", 0, map[string]any{}} {
		account.Credentials["clinepass_usage_enabled"] = disabled
		require.True(t, account.IsClinePassAccount(), "disabling does not change account identity")
		require.False(t, account.IsClinePassUsageEnabled())
		require.Empty(t, account.GetCodingPlanProvider())
		requireReason(t, validateCodingPlanAccount(account), "CN_QUOTA_DISABLED")
	}
	account.Credentials["clinepass_usage_enabled"] = true
	account.Credentials["base_url"] = "https://relay.example/v1"
	require.False(t, account.IsClinePassUsageEnabled(), "a switch cannot bypass the official-host safety guard")
	var empty *Account
	require.False(t, empty.IsClinePassUsageEnabled())
}

func TestClinePassUsageDisabledDoesNotQueryOrChangeSnapshot(t *testing.T) {
	account := clinePassTestAccount("https://api.cline.bot/v1")
	account.Credentials["clinepass_usage_enabled"] = false
	account.Extra = map[string]any{"deepseek_5h_used_percent": 42.0}
	repo := &cnBalanceProbeRepo{account: account}
	upstream := &httpUpstreamRecorder{}
	quota := NewCNProviderQuotaService(repo, nil, upstream, nil)
	_, err := quota.QueryUsage(context.Background(), account.ID)
	requireReason(t, err, "CN_QUOTA_DISABLED")
	_, err = quota.QueryUsageForAccount(context.Background(), account)
	requireReason(t, err, "CN_QUOTA_DISABLED")
	_, err = NewCNProviderBalanceService(repo, nil, upstream, nil).QueryBalance(context.Background(), account.ID)
	requireReason(t, err, "CN_BALANCE_CODING_PLAN")
	require.Empty(t, upstream.requests)
	require.Empty(t, repo.extraWrites)
	require.Equal(t, 42.0, account.Extra["deepseek_5h_used_percent"])
}

func TestParseClinePassUsageTiers(t *testing.T) {
	tiers := parseClinePassUsageTiers([]byte(clinePassUsageFixture))
	require.Equal(t, []CNQuotaTier{
		{Window: "5h", UsedPercent: 2, ResetAt: "2026-09-30T15:43:33Z"},
		{Window: "weekly", UsedPercent: 51, ResetAt: "2026-10-04T05:09:08Z"},
		{Window: "monthly", UsedPercent: 50, ResetAt: "2026-10-20T04:42:27Z"},
	}, tiers)
	for _, body := range []string{`{}`, `{bad`, `{"data":null}`, `{"data":{"limits":[]}}`,
		`{"data":{"limits":[{"type":"five_hour"},{"type":"weekly","percentUsed":"NaN"},{"type":"monthly","percentUsed":-1}]}}`} {
		require.Empty(t, parseClinePassUsageTiers([]byte(body)))
	}
	tiers = parseClinePassUsageTiers([]byte(`{"data":{"limits":[
		{"type":"five_hour","percentUsed":0},
		{"type":"weekly","percentUsed":"103.5"},
		{"type":"weekly","percentUsed":99},
		{"type":"unknown","percentUsed":12}
	]}}`))
	require.Equal(t, []CNQuotaTier{{Window: "5h", UsedPercent: 0}, {Window: "weekly", UsedPercent: 103.5}}, tiers)
}

func TestClinePassUsageQueryUsesFixedEndpointAndBearerAndPersists(t *testing.T) {
	account := clinePassTestAccount("https://api.cline.bot/v1")
	account.Credentials["header_override_enabled"] = true
	account.Credentials["header_overrides"] = map[string]any{"Authorization": "other-auth"}
	repo := &cnBalanceProbeRepo{account: account}
	upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(clinePassUsageFixture))}}
	svc := NewCNProviderQuotaService(repo, nil, upstream, cnProbeAllowlistConfig("api.cline.bot"))
	result, err := svc.QueryUsage(context.Background(), account.ID)
	require.NoError(t, err)
	require.True(t, result.Success)
	require.True(t, result.Persisted)
	require.Equal(t, PlatformDeepseek, result.Provider)
	require.Equal(t, "clinepass", result.Source)
	require.Equal(t, "ClinePass", result.PlanLevel)
	require.Equal(t, http.MethodGet, upstream.lastReq.Method)
	require.Equal(t, clinePassUsageURL, upstream.lastReq.URL.String())
	require.Equal(t, "Bearer test-cline-key", upstream.lastReq.Header.Get("Authorization"))
	authHeaderCount := 0
	for name, values := range upstream.lastReq.Header {
		if strings.EqualFold(name, "Authorization") {
			authHeaderCount += len(values)
		}
	}
	require.Equal(t, 1, authHeaderCount)
	require.Len(t, repo.extraWrites, 1)
	require.Equal(t, float64(2), repo.extraWrites[0]["deepseek_5h_used_percent"])
	require.Equal(t, float64(51), repo.extraWrites[0]["deepseek_weekly_used_percent"])
	require.Equal(t, float64(50), repo.extraWrites[0]["deepseek_monthly_used_percent"])
	_, err = NewCNProviderBalanceService(repo, nil, upstream, nil).QueryBalance(context.Background(), account.ID)
	requireReason(t, err, "CN_BALANCE_CODING_PLAN")
	// Invalid balance requests do not make a second network call.
	require.Len(t, upstream.requests, 1)
}

func TestClinePassUsageFailureDoesNotOverwriteSnapshot(t *testing.T) {
	for _, tc := range []struct {
		status int
		body   string
	}{
		{401, `{"error":"expired token"}`},
		{403, `{"error":"forbidden"}`},
		{429, `{"error":"try later"}`},
		{500, `{"error":"upstream failed"}`},
		{200, `{"success":false,"data":{"limits":[{"type":"five_hour","percentUsed":0}]}}`},
		{200, `{"success":true,"data":null}`},
		{200, `{not-json`},
	} {
		t.Run(tc.body, func(t *testing.T) {
			repo := &cnBalanceProbeRepo{account: clinePassTestAccount("https://api.cline.bot")}
			upstream := &cnBalanceResponseUpstream{statusCode: tc.status, body: tc.body}
			result, err := NewCNProviderQuotaService(repo, nil, upstream, nil).QueryUsage(context.Background(), repo.account.ID)
			require.NoError(t, err)
			require.False(t, result.Success)
			require.False(t, result.Persisted)
			require.NotEmpty(t, result.Error)
			require.Empty(t, repo.extraWrites)
		})
	}
}

func TestClinePassUsageRejectedBeforeNetworkWhenUnconfigured(t *testing.T) {
	account := clinePassTestAccount("https://api.cline.bot")
	repo := &cnBalanceProbeRepo{account: account}
	upstream := &recordingHTTPUpstream{}
	svc := NewCNProviderQuotaService(repo, nil, upstream, cnProbeAllowlistConfig("api.deepseek.com"))
	_, err := svc.QueryUsage(context.Background(), account.ID)
	requireReason(t, err, "CN_QUOTA_URL_REJECTED")
	require.Zero(t, upstream.calls)
	account.Credentials["api_key"] = ""
	_, err = svc.QueryUsage(context.Background(), account.ID)
	requireReason(t, err, "CN_QUOTA_NO_APIKEY")
	require.Zero(t, upstream.calls)
}
