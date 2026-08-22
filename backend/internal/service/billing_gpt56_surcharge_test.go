//go:build unit

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// gpt-5.6-luna 计费加倍：实际扣费按官方目录价 2x；Sol 保持官方原价。
// 展示路径（模型广场官方参考价）读取 PricingService 的 LiteLLM 原始目录，
// 不经过 BillingService.GetModelPricing，因此仍显示官方原价。
// TestMain 默认关闭该开关，这里显式开启验证加倍行为。
func TestGPT56SolBillingSurcharge(t *testing.T) {
	SetGPT56SolBillingSurchargeEnabled(true)
	t.Cleanup(func() { SetGPT56SolBillingSurchargeEnabled(false) })
	svc := newTestBillingService()

	// fallback 官方价：input $5/M、output $30/M、cache write $6.25/M、cache read $0.5/M。
	pricing, err := svc.GetModelPricing("gpt-5.6-sol")
	require.NoError(t, err)
	require.InEpsilon(t, 5e-6, pricing.InputPricePerToken, 1e-12, "sol input stays official")
	require.InEpsilon(t, 30e-6, pricing.OutputPricePerToken, 1e-12, "sol output stays official")
	require.InEpsilon(t, 6.25e-6, pricing.CacheCreationPricePerToken, 1e-12, "sol cache write stays official")
	require.InEpsilon(t, 0.5e-6, pricing.CacheReadPricePerToken, 1e-12, "sol cache read stays official")
	require.InEpsilon(t, 10e-6, pricing.InputPricePerTokenPriority, 1e-12, "sol priority input stays official")
	require.InEpsilon(t, 60e-6, pricing.OutputPricePerTokenPriority, 1e-12, "sol priority output stays official")

	// gpt-5.6 裸名是 Sol 的官方别名，同样加倍。
	bare, err := svc.GetModelPricing("gpt-5.6")
	require.NoError(t, err)
	require.InEpsilon(t, 5e-6, bare.InputPricePerToken, 1e-12, "bare gpt-5.6 stays official")

	// Terra 不加倍；Luna 静默加倍。
	terra, err := svc.GetModelPricing("gpt-5.6-terra")
	require.NoError(t, err)
	require.InEpsilon(t, 2e-6, terra.InputPricePerToken, 1e-12, "terra stays official")
	luna, err := svc.GetModelPricing("gpt-5.6-luna")
	require.NoError(t, err)
	require.InEpsilon(t, 2*0.2e-6, luna.InputPricePerToken, 1e-12, "luna input 2x")
	require.InEpsilon(t, 2*1.2e-6, luna.OutputPricePerToken, 1e-12, "luna output 2x")
	require.InEpsilon(t, 2*0.25e-6, luna.CacheCreationPricePerToken, 1e-12, "luna cache write 2x")
	require.InEpsilon(t, 2*0.02e-6, luna.CacheReadPricePerToken, 1e-12, "luna cache read 2x")
	require.InEpsilon(t, 2*0.4e-6, luna.InputPricePerTokenPriority, 1e-12, "luna priority input 2x")
	require.InEpsilon(t, 2*2.4e-6, luna.OutputPricePerTokenPriority, 1e-12, "luna priority output 2x")

	// 共享 fallback 条目不能被污染：重复取价仍是官方原价。
	again, err := svc.GetModelPricing("gpt-5.6-sol")
	require.NoError(t, err)
	require.InEpsilon(t, 5e-6, again.InputPricePerToken, 1e-12, "repeat sol lookup stays official")
}

// 「其他」运行时开关不再影响 Sol；开关开启或关闭均为官方原价。
func TestGPT56SolBillingSurchargeToggle(t *testing.T) {
	t.Cleanup(func() { SetGPT56SolBillingSurchargeEnabled(false) })
	svc := newTestBillingService()

	SetGPT56SolBillingSurchargeEnabled(false)
	off, err := svc.GetModelPricing("gpt-5.6-sol")
	require.NoError(t, err)
	require.InEpsilon(t, 5e-6, off.InputPricePerToken, 1e-12, "disabled = official price")
	require.InEpsilon(t, 30e-6, off.OutputPricePerToken, 1e-12, "disabled = official price")

	SetGPT56SolBillingSurchargeEnabled(true)
	on, err := svc.GetModelPricing("gpt-5.6-sol")
	require.NoError(t, err)
	require.InEpsilon(t, 5e-6, on.InputPricePerToken, 1e-12, "enabled still official")
}
