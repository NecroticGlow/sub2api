//go:build unit

package service

import (
	"os"
	"testing"
)

// TestMain 在 unit 测试中默认关闭「其他」（gpt-5.6-sol 计费加倍）运行时开关：
// 存量计价测试断言的是官方目录价语义；加价行为由 billing_gpt56_surcharge_test.go
// 显式开启并单独验证。
func TestMain(m *testing.M) {
	SetGPT56SolBillingSurchargeEnabled(false)
	os.Exit(m.Run())
}
