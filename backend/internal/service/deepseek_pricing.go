package service

import (
	"strings"
	"time"
)

// DeepSeek V4 uses Beijing-time peak windows.  The official price page defines
// peak hours as 09:00-12:00 and 14:00-18:00 (UTC+8); off-peak is half price.
const (
	deepSeekPeakMultiplier   = 2.0
	deepSeekBeijingUTCOffset = 8 * 60 * 60
)

var deepSeekBeijingLocation = time.FixedZone("Asia/Shanghai", deepSeekBeijingUTCOffset)

// isDeepSeekV4Model recognizes the native and ClinePass slugs that represent
// DeepSeek V4 Flash/Pro. Legacy chat/reasoner aliases are retained for billing
// compatibility and map to V4 Flash in the fallback pricing table.
func isDeepSeekV4Model(model string) bool {
	m := strings.ToLower(strings.TrimSpace(model))
	return strings.Contains(m, "deepseek-v4-flash") ||
		strings.Contains(m, "deepseek-v4-flash-vision-exp") ||
		strings.Contains(m, "deepseek-v4-pro") ||
		strings.Contains(m, "deepseek-chat") ||
		strings.Contains(m, "deepseek-reasoner")
}

// IsDeepSeekPeakTime reports whether a timestamp falls in an official
// DeepSeek V4 peak window. The timestamp is converted to Beijing time even
// when the server's configured timezone is different.
func IsDeepSeekPeakTime(at time.Time) bool {
	if at.IsZero() {
		at = time.Now()
	}
	local := at.In(deepSeekBeijingLocation)
	minutes := local.Hour()*60 + local.Minute()
	return (minutes >= 9*60 && minutes < 12*60) ||
		(minutes >= 14*60 && minutes < 18*60)
}

// DeepSeekPeakMultiplier returns the official 2x peak multiplier or 1x
// off-peak multiplier. It is deliberately separate from user group peak-rate
// configuration so DeepSeek's upstream tariff is applied automatically.
func DeepSeekPeakMultiplier(at time.Time) float64 {
	if IsDeepSeekPeakTime(at) {
		return deepSeekPeakMultiplier
	}
	return 1
}

func applyDeepSeekPeakMultiplier(model string, base float64, at time.Time) float64 {
	if !isDeepSeekV4Model(model) {
		return base
	}
	return base * DeepSeekPeakMultiplier(at)
}
