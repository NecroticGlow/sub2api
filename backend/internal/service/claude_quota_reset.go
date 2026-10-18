package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"slices"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const claudeResetSnapshotKey = "claude_reset_credit_snapshot"

// Claude's banked resets are separate from paid extra-usage credits. Unknown
// (null program) is not an empty inventory. Never expose the event_props block.
type ClaudeResetGrant struct {
	ID         string   `json:"id"`
	ResetsLeft int      `json:"resets_left"`
	StartsAt   string   `json:"starts_at,omitempty"`
	EndsAt     string   `json:"ends_at,omitempty"`
	Clears     []string `json:"clears,omitempty"`
	Paused     bool     `json:"paused"`
	UsableNow  bool     `json:"usable_now"`
}

type ClaudeResetProgram struct {
	Eligible         bool               `json:"eligible"`
	IneligibleReason string             `json:"ineligible_reason,omitempty"`
	NextGrantID      string             `json:"next_grant_id,omitempty"`
	CooldownUntil    string             `json:"cooldown_until,omitempty"`
	Grants           []ClaudeResetGrant `json:"grants"`
}

type ClaudeResetStatus struct {
	*ClaudeResetProgram
	AvailableCount int       `json:"available_count"`
	FetchedAt      time.Time `json:"fetched_at"`
}

type ClaudeResetInput struct {
	Confirmed  bool   `json:"confirmed"`
	GrantID    string `json:"grant_id"`
	ResetsLeft int    `json:"resets_left"`
}

type ClaudeResetResult struct {
	AppliedAt      *time.Time `json:"applied_at,omitempty"`
	RefreshWarning bool       `json:"refresh_warning,omitempty"`
	Result         string     `json:"result"`
	Cleared        []string   `json:"cleared,omitempty"`
	CooldownUntil  string     `json:"cooldown_until,omitempty"`
}

// Separate optional port keeps other usage fetchers (and test doubles) read-only.
type ClaudeQuotaResetter interface {
	ResetQuota(context.Context, *ClaudeUsageFetchOptions, string, string, string) (*ClaudeResetResult, error)
}

var claudeResetIdentifier = regexp.MustCompile(`^[A-Za-z0-9_-]{1,160}$`)

func claudeGrantCurrent(g ClaudeResetGrant, now time.Time) bool {
	if g.ResetsLeft <= 0 {
		return false
	}
	if g.StartsAt != "" {
		t, err := time.Parse(time.RFC3339Nano, g.StartsAt)
		if err != nil || now.Before(t) {
			return false
		}
	}
	if g.EndsAt != "" {
		t, err := time.Parse(time.RFC3339Nano, g.EndsAt)
		if err != nil || !now.Before(t) {
			return false
		}
	}
	return true
}

func buildClaudeResetStatus(p *ClaudeResetProgram, now time.Time) *ClaudeResetStatus {
	if p == nil {
		return nil
	}
	status := &ClaudeResetStatus{ClaudeResetProgram: p, FetchedAt: now}
	for _, g := range p.Grants {
		if claudeGrantCurrent(g, now) {
			status.AvailableCount += g.ResetsLeft
		}
	}
	return status
}

func claudeResetSnapshot(extra map[string]any) *ClaudeResetStatus {
	raw, ok := extra[claudeResetSnapshotKey]
	if !ok || raw == nil {
		return nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var status ClaudeResetStatus
	if json.Unmarshal(b, &status) != nil || status.ClaudeResetProgram == nil || status.FetchedAt.IsZero() {
		return nil
	}
	// Do not let a stale or expired grant enable consumption after a page reload.
	if time.Since(status.FetchedAt) > 24*time.Hour {
		return nil
	}
	fresh := buildClaudeResetStatus(status.ClaudeResetProgram, time.Now())
	fresh.FetchedAt = status.FetchedAt
	return fresh
}

func (s *AccountUsageService) claudeResetAccount(ctx context.Context, id int64) (*Account, error) {
	a, err := s.accountRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if a == nil || a.Platform != PlatformAnthropic || a.Type != AccountTypeOAuth || a.IsSyntheticUITest() {
		return nil, infraerrors.BadRequest("CLAUDE_RESET_ACCOUNT_UNSUPPORTED", "reset credits require a Claude OAuth account")
	}
	return a, nil
}

func (s *AccountUsageService) QueryClaudeResetCredits(ctx context.Context, id int64, persist bool) (*ClaudeResetStatus, error) {
	a, err := s.claudeResetAccount(ctx, id)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	usage, err := s.fetchOAuthUsageRaw(ctx, a)
	if err != nil {
		return nil, infraerrors.New(502, "CLAUDE_QUOTA_QUERY_FAILED", "Claude usage query failed; check authorization and proxy availability")
	}
	status := buildClaudeResetStatus(usage.CedarEmber, time.Now())
	if persist {
		if err := s.accountRepo.UpdateExtra(ctx, id, map[string]any{claudeResetSnapshotKey: status}); err != nil {
			return nil, err
		}
		if s.cache != nil {
			s.cache.apiCache.Store(id, &apiUsageCache{response: usage, timestamp: time.Now()})
		}
	}
	return status, nil
}

func (s *AccountUsageService) ResetClaudeQuota(ctx context.Context, id int64, input ClaudeResetInput) (*ClaudeResetResult, error) {
	if !input.Confirmed || !claudeResetIdentifier.MatchString(input.GrantID) || input.ResetsLeft < 1 || input.ResetsLeft > 1000 {
		return nil, infraerrors.BadRequest("CLAUDE_RESET_CONFIRM_REQUIRED", "confirm the queried reset grant and remaining count before consuming it")
	}
	a, err := s.claudeResetAccount(ctx, id)
	if err != nil {
		return nil, err
	}
	org := a.GetCredential("org_uuid")
	if !claudeResetIdentifier.MatchString(org) {
		return nil, infraerrors.BadRequest("CLAUDE_RESET_ORG_REQUIRED", "Claude organization ID is required")
	}
	resetter, ok := s.usageFetcher.(ClaudeQuotaResetter)
	if !ok || s.claudeResetIdempotency == nil {
		return nil, infraerrors.ServiceUnavailable("CLAUDE_RESET_UNAVAILABLE", "Claude quota reset is not configured")
	}
	// The SAME grant/count always has the SAME upstream request UUID, across
	// admins, duplicate local accounts, retries, lost responses and restarts.
	sum := sha256.Sum256([]byte(fmt.Sprintf("claude-reset:%s:%s:%d", org, input.GrantID, input.ResetsLeft)))
	key := fmt.Sprintf("%x", sum)
	sum[6] = (sum[6] & 0x0f) | 0x50
	sum[8] = (sum[8] & 0x3f) | 0x80
	requestID := fmt.Sprintf("%x-%x-%x-%x-%x", sum[0:4], sum[4:6], sum[6:8], sum[8:10], sum[10:16])
	execution, err := s.claudeResetIdempotency.Execute(ctx, IdempotencyExecuteOptions{
		Scope: "claude_manual_reset", ActorScope: key, Method: http.MethodPost, Route: "/admin/claude/reset-quota",
		IdempotencyKey: key, Payload: map[string]any{"grant_id": input.GrantID, "resets_left": input.ResetsLeft},
		RequireKey: true, TTL: 90 * 24 * time.Hour, ExecutionTimeout: 50 * time.Second,
	}, func(execCtx context.Context) (any, error) {
		// Re-load the account and re-query upstream INSIDE the durable claim.
		current, err := s.claudeResetAccount(execCtx, id)
		if err != nil {
			return nil, err
		}
		if current.GetCredential("org_uuid") != org {
			return nil, infraerrors.Conflict("CLAUDE_RESET_ACCOUNT_CHANGED", "account identity changed; query again")
		}
		usage, err := s.fetchOAuthUsageRaw(execCtx, current)
		if err != nil {
			return nil, infraerrors.New(502, "CLAUDE_QUOTA_QUERY_FAILED", "Claude reset verification query failed")
		}
		p := usage.CedarEmber
		if p == nil || !p.Eligible || p.NextGrantID != input.GrantID {
			return nil, infraerrors.Conflict("CLAUDE_RESET_NOT_AVAILABLE", "reset grant is unknown, ineligible or no longer next; query again")
		}
		if p.CooldownUntil != "" {
			until, err := time.Parse(time.RFC3339Nano, p.CooldownUntil)
			if err != nil || until.After(time.Now()) {
				return nil, infraerrors.Conflict("CLAUDE_RESET_COOLDOWN", "reset is in cooldown")
			}
		}
		var selected *ClaudeResetGrant
		for i := range p.Grants {
			if p.Grants[i].ID == input.GrantID {
				selected = &p.Grants[i]
				break
			}
		}
		if selected == nil || selected.ResetsLeft != input.ResetsLeft || !claudeGrantCurrent(*selected, time.Now()) || selected.Paused || !selected.UsableNow {
			return nil, infraerrors.Conflict("CLAUDE_RESET_STALE", "reset grant is expired, paused, unusable or changed; query again")
		}
		opts, err := s.claudeQuotaOptions(execCtx, current)
		if err != nil {
			return nil, err
		}
		result, err := resetter.ResetQuota(execCtx, opts, org, input.GrantID, requestID)
		if err == nil && result != nil && (result.Result == "reset" || result.Result == "already_used") {
			now := time.Now().UTC()
			result.AppliedAt = &now
		}
		return result, err
	})
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(execution.Data)
	if err != nil {
		return nil, err
	}
	var result ClaudeResetResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	if s.cache != nil {
		s.cache.apiCache.Delete(id)
		s.cache.windowStatsCache.Delete(id)
	}
	if result.AppliedAt != nil {
		postCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		updates := map[string]any{claudeResetSnapshotKey: nil}
		if slices.Contains(result.Cleared, "seven_day") {
			updates["claude_quota_reset_at"] = result.AppliedAt.Format(time.RFC3339Nano)
		}
		if err := s.accountRepo.UpdateExtra(postCtx, id, updates); err != nil {
			slog.Warn("claude_reset_snapshot_invalidation_failed", "account_id", id)
			result.RefreshWarning = true
		}
	}
	return &result, nil
}

// A grant can clear only some windows. Never clear all local cooldowns based
// solely on a successful redemption while a fresh upstream window is still full.
func (s *AccountUsageService) ClaudeQuotaCanResume(id int64) bool {
	if s.cache == nil {
		return false
	}
	value, ok := s.cache.apiCache.Load(id)
	if !ok {
		return false
	}
	item, ok := value.(*apiUsageCache)
	if !ok || item.response == nil || time.Since(item.timestamp) > 30*time.Second {
		return false
	}
	u := item.response
	return u.FiveHour.ResetsAt != "" && u.SevenDay.ResetsAt != "" &&
		u.FiveHour.Utilization < 100 && u.SevenDay.Utilization < 100 && u.SevenDaySonnet.Utilization < 100 && u.SevenDayOverageIncluded.Utilization < 100 &&
		u.SevenDayOpus.Utilization < 100 && u.SevenDayOAuthApps.Utilization < 100 && u.SevenDayCowork.Utilization < 100 && u.SevenDayOmelette.Utilization < 100
}

// Match the existing GPT estimate: account-window cost / official utilization.
// Claude's manual reset does NOT move the weekly reset date, so start costs at
// the last successful manual reset when it is later than the weekly boundary.
type claudeWeeklyStatsCache struct {
	stats     *WindowStats
	timestamp time.Time
	start     time.Time
}

func (s *AccountUsageService) addClaudeWeeklyStats(ctx context.Context, a *Account, usage *UsageInfo) {
	if a.Platform != PlatformAnthropic || a.Type != AccountTypeOAuth || s.usageLogRepo == nil || s.cache == nil || usage.SevenDay == nil || usage.SevenDay.ResetsAt == nil {
		return
	}
	now := time.Now()
	end := *usage.SevenDay.ResetsAt
	if !end.After(now) || end.After(now.Add(8*24*time.Hour)) {
		return
	}
	start := end.Add(-7 * 24 * time.Hour)
	if value, ok := a.Extra["claude_quota_reset_at"].(string); ok {
		if reset, err := time.Parse(time.RFC3339Nano, value); err == nil && reset.After(start) && !reset.After(now) {
			start = reset
		}
	}
	key := fmt.Sprintf("claude-week:%d", a.ID)
	if cached, ok := s.cache.windowStatsCache.Load(key); ok {
		if item, ok := cached.(*claudeWeeklyStatsCache); ok && item.start.Equal(start) && time.Since(item.timestamp) < windowStatsCacheTTL {
			usage.SevenDay.WindowStats = item.stats
			return
		}
	}
	stats, err := s.usageLogRepo.GetAccountWindowStats(ctx, a.ID, start)
	if err != nil {
		return
	}
	usage.SevenDay.WindowStats = windowStatsFromAccountStats(stats)
	s.cache.windowStatsCache.Store(key, &claudeWeeklyStatsCache{stats: usage.SevenDay.WindowStats, timestamp: now, start: start})
}

func (s *AccountUsageService) claudeQuotaOptions(ctx context.Context, a *Account) (*ClaudeUsageFetchOptions, error) {
	token := a.GetCredential("access_token")
	if s.claudeTokenProvider != nil {
		var err error
		token, err = s.claudeTokenProvider.GetAccessToken(ctx, a)
		if err != nil {
			return nil, infraerrors.New(502, "CLAUDE_QUOTA_AUTH_FAILED", "Claude authorization could not be refreshed")
		}
	}
	if token == "" {
		return nil, infraerrors.BadRequest("CLAUDE_QUOTA_TOKEN_REQUIRED", "Claude access token is missing")
	}
	opts := &ClaudeUsageFetchOptions{AccessToken: token, AccountID: a.ID}
	if a.Proxy != nil && a.ProxyID != nil {
		opts.ProxyURL = a.Proxy.URL()
	}
	if s.tlsFPProfileService != nil {
		opts.TLSProfile = s.tlsFPProfileService.ResolveTLSProfile(a)
	}
	return opts, nil
}
