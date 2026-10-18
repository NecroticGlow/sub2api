package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/stretchr/testify/require"
)

type claudeResetTestRepo struct {
	AccountRepository
	account    *Account
	failUpdate bool
}

func (r *claudeResetTestRepo) GetByID(context.Context, int64) (*Account, error) {
	return r.account, nil
}
func (r *claudeResetTestRepo) UpdateExtra(_ context.Context, _ int64, updates map[string]any) error {
	if r.failUpdate {
		return errors.New("synthetic database failure")
	}
	for k, v := range updates {
		r.account.Extra[k] = v
	}
	return nil
}

type claudeResetTestFetcher struct {
	program    *ClaudeResetProgram
	calls      int
	requestIDs []string
}

func (f *claudeResetTestFetcher) FetchUsage(context.Context, string, string) (*ClaudeUsageResponse, error) {
	return &ClaudeUsageResponse{CedarEmber: f.program}, nil
}
func (f *claudeResetTestFetcher) FetchUsageWithOptions(context.Context, *ClaudeUsageFetchOptions) (*ClaudeUsageResponse, error) {
	return &ClaudeUsageResponse{CedarEmber: f.program}, nil
}
func (f *claudeResetTestFetcher) ResetQuota(_ context.Context, _ *ClaudeUsageFetchOptions, _ string, _ string, requestID string) (*ClaudeResetResult, error) {
	f.calls++
	f.requestIDs = append(f.requestIDs, requestID)
	return &ClaudeResetResult{Result: "reset", Cleared: []string{"five_hour", "seven_day"}}, nil
}
func newClaudeResetTestService() (*AccountUsageService, *claudeResetTestFetcher, *claudeResetTestRepo) {
	repo := &claudeResetTestRepo{account: &Account{ID: 42, Platform: PlatformAnthropic, Type: AccountTypeOAuth,
		Credentials: map[string]any{"access_token": "synthetic", "org_uuid": "00000000-0000-4000-8000-000000000042"}, Extra: map[string]any{}}}
	fetcher := &claudeResetTestFetcher{program: &ClaudeResetProgram{Eligible: true, NextGrantID: "test-grant", Grants: []ClaudeResetGrant{{ID: "test-grant", ResetsLeft: 1, UsableNow: true, EndsAt: time.Now().Add(time.Hour).Format(time.RFC3339)}}}}
	s := &AccountUsageService{accountRepo: repo, usageFetcher: fetcher, cache: NewUsageCache(), claudeResetIdempotency: NewIdempotencyCoordinator(newInMemoryIdempotencyRepo(), DefaultIdempotencyConfig())}
	return s, fetcher, repo
}
func TestClaudeResetQueryPersistsSharedSnapshot(t *testing.T) {
	s, f, r := newClaudeResetTestService()
	status, err := s.QueryClaudeResetCredits(context.Background(), 42, false)
	require.NoError(t, err)
	require.Equal(t, 1, status.AvailableCount)
	require.Empty(t, r.account.Extra)
	_, err = s.QueryClaudeResetCredits(context.Background(), 42, true)
	require.NoError(t, err)
	require.Equal(t, 1, claudeResetSnapshot(r.account.Extra).AvailableCount)
	f.program = nil
	status, err = s.QueryClaudeResetCredits(context.Background(), 42, true)
	require.NoError(t, err)
	require.Nil(t, status)
	require.Nil(t, claudeResetSnapshot(r.account.Extra))
}
func TestClaudeResetRejectsUnsafeConsumption(t *testing.T) {
	for _, mode := range []string{"unconfirmed", "wrong-platform", "apikey", "no-org", "expired", "paused", "not-usable", "cooldown", "wrong-next", "stale-count", "unknown", "ineligible"} {
		t.Run(mode, func(t *testing.T) {
			s, f, r := newClaudeResetTestService()
			input := ClaudeResetInput{Confirmed: true, GrantID: "test-grant", ResetsLeft: 1}
			switch mode {
			case "unconfirmed":
				input.Confirmed = false
			case "wrong-platform":
				r.account.Platform = PlatformOpenAI
			case "apikey":
				r.account.Type = AccountTypeAPIKey
			case "no-org":
				delete(r.account.Credentials, "org_uuid")
			case "expired":
				f.program.Grants[0].EndsAt = time.Now().Add(-time.Hour).Format(time.RFC3339)
			case "paused":
				f.program.Grants[0].Paused = true
			case "not-usable":
				f.program.Grants[0].UsableNow = false
			case "cooldown":
				f.program.CooldownUntil = time.Now().Add(time.Hour).Format(time.RFC3339)
			case "wrong-next":
				f.program.NextGrantID = "another"
			case "stale-count":
				input.ResetsLeft = 2
			case "unknown":
				f.program = nil
			case "ineligible":
				f.program.Eligible = false
			}
			_, err := s.ResetClaudeQuota(context.Background(), 42, input)
			require.Error(t, err)
			require.Zero(t, f.calls)
		})
	}
}
func TestClaudeResetDurableReplayDoesNotConsumeTwice(t *testing.T) {
	s, f, r := newClaudeResetTestService()
	input := ClaudeResetInput{Confirmed: true, GrantID: "test-grant", ResetsLeft: 1}
	first, err := s.ResetClaudeQuota(context.Background(), 42, input)
	require.NoError(t, err)
	f.program = nil // upstream grant is gone; replay still returns recorded success
	second, err := s.ResetClaudeQuota(context.Background(), 42, input)
	require.NoError(t, err)
	require.Equal(t, 1, f.calls)
	require.Equal(t, first.AppliedAt, second.AppliedAt)
	require.Equal(t, first.AppliedAt.Format(time.RFC3339Nano), r.account.Extra["claude_quota_reset_at"])
	require.Len(t, f.requestIDs[0], 36)
	require.Equal(t, byte('5'), f.requestIDs[0][14])
}
func TestClaudeResetConsumptionSurvivesSnapshotFailure(t *testing.T) {
	s, f, r := newClaudeResetTestService()
	r.failUpdate = true
	result, err := s.ResetClaudeQuota(context.Background(), 42, ClaudeResetInput{Confirmed: true, GrantID: "test-grant", ResetsLeft: 1})
	require.NoError(t, err)
	require.Equal(t, "reset", result.Result)
	require.True(t, result.RefreshWarning)
	require.Equal(t, 1, f.calls)
}
func TestClaudeResetPartialResetDoesNotClearBlockedWindows(t *testing.T) {
	s, _, _ := newClaudeResetTestService()
	u := &ClaudeUsageResponse{}
	u.FiveHour.ResetsAt = time.Now().Add(time.Hour).Format(time.RFC3339)
	u.SevenDay.ResetsAt = time.Now().Add(time.Hour).Format(time.RFC3339)
	s.cache.apiCache.Store(int64(42), &apiUsageCache{response: u, timestamp: time.Now()})
	require.True(t, s.ClaudeQuotaCanResume(42))
	u.SevenDayOpus.Utilization = 100
	require.False(t, s.ClaudeQuotaCanResume(42))
}
func TestClaudeResetSnapshotExpiryAndSanitization(t *testing.T) {
	now := time.Now()
	var usage ClaudeUsageResponse
	require.NoError(t, json.Unmarshal([]byte(`{"cedar_ember":{"eligible":true,"event_props":{"email":"never-expose"},"grants":[{"id":"old","resets_left":1,"ends_at":"2020-01-01T00:00:00Z"},{"id":"new","resets_left":1}]}}`), &usage))
	status := buildClaudeResetStatus(usage.CedarEmber, now)
	require.Equal(t, 1, status.AvailableCount)
	b, err := json.Marshal(status)
	require.NoError(t, err)
	require.NotContains(t, string(b), "never-expose")
	status.FetchedAt = now.Add(-25 * time.Hour)
	require.Nil(t, claudeResetSnapshot(map[string]any{claudeResetSnapshotKey: status}))
}

type claudeWeeklyStatsTestRepo struct {
	UsageLogRepository
	start time.Time
	calls int
}

func (r *claudeWeeklyStatsTestRepo) GetAccountWindowStats(_ context.Context, _ int64, start time.Time) (*usagestats.AccountStats, error) {
	r.start = start
	r.calls++
	return &usagestats.AccountStats{Cost: 12, StandardCost: 24, UserCost: 3}, nil
}
func TestClaudeWeeklyStatsUsesWeeklyBoundaryAndManualReset(t *testing.T) {
	s, _, r := newClaudeResetTestService()
	repo := &claudeWeeklyStatsTestRepo{}
	s.usageLogRepo = repo
	end := time.Now().Add(48 * time.Hour).Truncate(time.Second)
	usage := &UsageInfo{SevenDay: &UsageProgress{Utilization: 40, ResetsAt: &end}}
	s.addClaudeWeeklyStats(context.Background(), r.account, usage)
	require.Equal(t, end.Add(-7*24*time.Hour), repo.start)
	require.Equal(t, 12.0, usage.SevenDay.WindowStats.Cost)
	s.addClaudeWeeklyStats(context.Background(), r.account, usage)
	require.Equal(t, 1, repo.calls)
	reset := time.Now().Add(-time.Minute).UTC()
	r.account.Extra["claude_quota_reset_at"] = reset.Format(time.RFC3339Nano)
	s.addClaudeWeeklyStats(context.Background(), r.account, usage)
	require.True(t, reset.Equal(repo.start))
	require.Equal(t, 2, repo.calls)
}
