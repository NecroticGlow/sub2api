//go:build unit

package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAPIKeyGroupFallback_ImagesRetainImageCooldownAndAlias(t *testing.T) {
	for _, fallback := range []bool{false, true} {
		t.Run(map[bool]string{false: "primary", true: "fallback"}[fallback], func(t *testing.T) {
			resetOpenAIAdvancedSchedulerSettingCacheForTest()
			defer resetOpenAIAdvancedSchedulerSettingCacheForTest()
			primaryID, fallbackID := int64(53001), int64(53002)
			accountGroup := primaryID
			if fallback {
				accountGroup = fallbackID
			}
			cooled := imageRoutingAccount(53011, map[string]any{"image-alias": "flux-example"})
			cooled.GroupIDs, cooled.Priority = []int64{accountGroup}, 1
			setAccountModelRateLimitSnapshot(&cooled, openAIImageGenerationRateLimitKey,
				time.Now().Add(time.Hour), openAIImageCapabilityLossReason, time.Now())
			available := imageRoutingAccount(53012, map[string]any{"image-alias": "flux-example"})
			available.GroupIDs, available.Priority = []int64{accountGroup}, 2
			svc := newAPIKeyFallbackOpenAIService([]Account{cooled, available})
			selection, _, err := svc.SelectAccountWithSchedulerForImages(
				apiKeyFallbackTestContext(primaryID, fallbackID), &primaryID, "", "image-alias", nil,
				OpenAIImagesCapabilityNative)
			require.NoError(t, err)
			require.NotNil(t, selection)
			if selection.ReleaseFunc != nil {
				defer selection.ReleaseFunc()
			}
			require.Equal(t, available.ID, selection.Account.ID)
			require.Equal(t, fallback, selection.UsedAPIKeyFallback)
			if fallback {
				require.Equal(t, fallbackID, selection.RoutedGroupID)
			}
		})
	}
}
