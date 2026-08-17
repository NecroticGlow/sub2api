//go:build unit

package handler

import (
	"os"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

// TestMain 在 unit 测试中默认关闭「其他」（gpt-5.6-sol 计费加倍）运行时开关：
// handler 层的计费断言（如 WS 每轮用量）以官方目录价为基准；加价行为在
// service 包的 billing_gpt56_surcharge_test.go 中单独验证。
func TestMain(m *testing.M) {
	service.SetGPT56SolBillingSurchargeEnabled(false)
	os.Exit(m.Run())
}
