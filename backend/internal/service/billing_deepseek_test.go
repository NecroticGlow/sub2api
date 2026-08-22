//go:build unit

package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// DeepSeek 计费必须使用官方定价（https://api-docs.deepseek.com/quick_start/pricing），
// 即使流量经由第三方上游（如 ClinePass 的 cline-pass/deepseek-v4-* slug）转发。
// ClinePass 参考价（V4 Pro $1.74/$3.48）是第三方口径，不得用于计费。
func TestDeepSeekPricingUsesOfficialRates(t *testing.T) {
	svc := newTestBillingService()

	officialV4Pro := struct{ in, out float64 }{4.5 / 7e6, 13.5 / 7e6}    // ¥4.5 / ¥13.5 per MTok, off-peak
	officialV4Flash := struct{ in, out float64 }{1.5 / 7e6, 4.5 / 7e6}   // ¥1.5 / ¥4.5 per MTok, off-peak

	cases := []struct {
		model string
		in    float64
		out   float64
	}{
		{"deepseek-v4-pro", officialV4Pro.in, officialV4Pro.out},
		{"deepseek-v4-flash", officialV4Flash.in, officialV4Flash.out},
		// ClinePass 上游 slug：按官方 V4 价计费，而非 ClinePass 参考价。
		{"cline-pass/deepseek-v4-pro", officialV4Pro.in, officialV4Pro.out},
		{"cline-pass/deepseek-v4-flash", officialV4Flash.in, officialV4Flash.out},
	}

	for _, tc := range cases {
		t.Run(tc.model, func(t *testing.T) {
			pricing, err := svc.GetModelPricing(tc.model)
			require.NoError(t, err)
			require.NotNil(t, pricing)
			require.InEpsilon(t, tc.in, pricing.InputPricePerToken, 1e-12, "input price")
			require.InEpsilon(t, tc.out, pricing.OutputPricePerToken, 1e-12, "output price")
		})
	}
}

func TestDeepSeekPricingOverridesStaleDynamicCatalog(t *testing.T) {
	stale := &PricingService{pricingData: map[string]*LiteLLMModelPricing{
		"deepseek-v4-pro": {
			InputCostPerToken:       4.35e-7,
			OutputCostPerToken:      8.7e-7,
			CacheReadInputTokenCost: 3.625e-9,
		},
	}}
	svc := NewBillingService(nil, stale)

	pricing, err := svc.GetModelPricing("deepseek-v4-pro")
	require.NoError(t, err)
	require.InEpsilon(t, 4.5/7e6, pricing.InputPricePerToken, 1e-12)
	require.InEpsilon(t, 13.5/7e6, pricing.OutputPricePerToken, 1e-12)
	require.InEpsilon(t, 0.15/7e6, pricing.CacheReadPricePerToken, 1e-12)
}

func TestDeepSeekPeakMultiplierUsesBeijingWindows(t *testing.T) {
	utc := time.FixedZone("UTC", 0)
	for _, tc := range []struct {
		name string
		at   time.Time
		want float64
	}{
		{"peak-morning", time.Date(2026, 8, 17, 1, 0, 0, 0, utc), 2},   // 09:00 Beijing
		{"peak-afternoon", time.Date(2026, 8, 17, 6, 0, 0, 0, utc), 2}, // 14:00 Beijing
		{"off-peak", time.Date(2026, 8, 17, 5, 59, 0, 0, utc), 1},      // 13:59 Beijing
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, DeepSeekPeakMultiplier(tc.at))
		})
	}
}

func TestDeepSeekSilentDoublePreservesPeakMultiplier(t *testing.T) {
	SetGPT56SolBillingSurchargeEnabled(true)
	t.Cleanup(func() { SetGPT56SolBillingSurchargeEnabled(false) })
	svc := newTestBillingService()

	pricing, err := svc.GetModelPricing("cline-pass/deepseek-v4-flash")
	require.NoError(t, err)
	require.InEpsilon(t, 2*(1.5/7e6), pricing.InputPricePerToken, 1e-12, "hidden base multiplier")
	require.InEpsilon(t, 2*(4.5/7e6), pricing.OutputPricePerToken, 1e-12, "hidden base multiplier")
	require.InEpsilon(t, 2*(0.05/7e6), pricing.CacheReadPricePerToken, 1e-12, "hidden cache multiplier")

	utc := time.FixedZone("UTC", 0)
	offPeak := time.Date(2026, 8, 17, 5, 59, 0, 0, utc) // 13:59 Beijing
	peak := time.Date(2026, 8, 17, 6, 0, 0, 0, utc)     // 14:00 Beijing
	require.Equal(t, 1.0, applyDeepSeekPeakMultiplier("deepseek-v4-flash", 1, offPeak))
	require.Equal(t, 2.0, applyDeepSeekPeakMultiplier("deepseek-v4-flash", 1, peak))
	require.InEpsilon(t, 2*(1.5/7e6), pricing.InputPricePerToken*DeepSeekPeakMultiplier(offPeak), 1e-12)
	require.InEpsilon(t, 4*(1.5/7e6), pricing.InputPricePerToken*DeepSeekPeakMultiplier(peak), 1e-12)
}

// ClinePass 全系模型（含非 DeepSeek 厂商）都必须能解析出各厂商官方口径的兜底价，
// 防止 ClinePass 分组的任一模型按 $0 计费或被 fail-closed 拒绝。
func TestClinePassModelsAllHaveFallbackPricing(t *testing.T) {
	svc := newTestBillingService()

	models := []string{
		"cline-pass/deepseek-v4-pro",
		"cline-pass/deepseek-v4-flash",
		"cline-pass/glm-5.2",
		"cline-pass/kimi-k3",
		"cline-pass/kimi-k2.7-code",
		"cline-pass/kimi-k2.6",
		"cline-pass/mimo-v2.5",
		"cline-pass/mimo-v2.5-pro",
		"cline-pass/minimax-m3",
		"cline-pass/qwen3.8-max",
		"cline-pass/qwen3.7-max",
		"cline-pass/qwen3.7-plus",
	}

	for _, model := range models {
		t.Run(model, func(t *testing.T) {
			pricing, err := svc.GetModelPricing(model)
			require.NoError(t, err)
			require.NotNil(t, pricing)
			require.Greater(t, pricing.InputPricePerToken, 0.0, "input price must be positive")
			require.Greater(t, pricing.OutputPricePerToken, 0.0, "output price must be positive")
		})
	}
}
