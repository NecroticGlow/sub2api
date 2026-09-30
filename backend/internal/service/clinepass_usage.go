package service

import (
	"math"
	"net/url"
	"strings"

	"github.com/tidwall/gjson"
)

const clinePassUsageURL = "https://api.cline.bot/api/v1/users/me/plan/usage-limits"

// IsClinePassAccount recognizes only DeepSeek API-key accounts using the real
// Cline API host. Never send an unrelated relay's key to Cline's quota endpoint.
func (a *Account) IsClinePassAccount() bool {
	if a == nil || a.Platform != PlatformDeepseek || a.Type != AccountTypeAPIKey {
		return false
	}
	u, err := url.Parse(strings.TrimSpace(a.GetCredential("base_url")))
	return err == nil && strings.EqualFold(u.Scheme, "https") && u.User == nil &&
		strings.EqualFold(u.Hostname(), "api.cline.bot") && (u.Port() == "" || u.Port() == "443")
}

// parseClinePassUsageTiers follows GET /users/me/plan/usage-limits:
// data.limits[].{type: five_hour|weekly|monthly, percentUsed, resetsAt}.
// Unknown or malformed windows are skipped rather than shown as zero usage.
func parseClinePassUsageTiers(body []byte) []CNQuotaTier {
	if !gjson.ValidBytes(body) {
		return nil
	}
	limits := gjson.GetBytes(body, "data.limits")
	if !limits.IsArray() {
		return nil
	}
	windows := make(map[string]CNQuotaTier, 3)
	limits.ForEach(func(_, entry gjson.Result) bool {
		window := strings.ToLower(strings.TrimSpace(entry.Get("type").String()))
		if window == "five_hour" {
			window = "5h"
		}
		if window != "5h" && window != "weekly" && window != "monthly" {
			return true
		}
		if _, exists := windows[window]; exists {
			return true
		}
		percent, ok := cnParseF64(entry.Get("percentUsed").Value())
		if !ok || math.IsNaN(percent) || math.IsInf(percent, 0) || percent < 0 {
			return true
		}
		windows[window] = CNQuotaTier{
			Window: window, UsedPercent: percent,
			ResetAt: cnNormalizeResetTime(entry.Get("resetsAt").Value()),
		}
		return true
	})
	tiers := make([]CNQuotaTier, 0, len(windows))
	for _, window := range []string{"5h", "weekly", "monthly"} {
		if tier, exists := windows[window]; exists {
			tiers = append(tiers, tier)
		}
	}
	return tiers
}
