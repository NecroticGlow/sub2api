//go:build unit

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildAccountForCreateKeepsInitialQualityAndRetry(t *testing.T) {
	plan := &ScheduledTestPlan{ModelID: "gpt-6-astra", CronExpression: "*/30 * * * *"}
	for _, retry := range []int{DefaultRateLimit429RetryCount, 2} {
		input := &CreateAccountInput{
			Name: "synthetic", Platform: PlatformOpenAI, Type: AccountTypeOAuth,
			Concurrency: 3, InitialQualityPlan: plan,
		}
		if retry != DefaultRateLimit429RetryCount {
			input.RateLimit429RetryCount = &retry
		}
		account, err := buildAccountForCreate(input, nil)
		require.NoError(t, err)
		require.Same(t, plan, account.InitialQualityPlan)
		require.NotNil(t, account.RateLimit429RetryCount)
		require.Equal(t, retry, *account.RateLimit429RetryCount)
	}
}
