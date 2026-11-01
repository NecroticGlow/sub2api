package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func deepSeekOpenCodeFixture(id, groupID int64, platform string) Account {
	return Account{ID: id, Platform: platform, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Concurrency: 1,
		GroupIDs: []int64{groupID}, Credentials: map[string]any{"api_key": "synthetic-not-a-token", "model_mapping": map[string]any{"deepseek-v4.1-flash": "deepseek-v4.1-flash", "beam-free": "beam-free"}}}
}

func TestDeepSeekOpenCodePlatformCompatibilityIsOneWay(t *testing.T) {
	for _, tc := range []struct {
		request, account string
		allowed          bool
	}{
		{PlatformDeepseek, PlatformOpenCodeGo, true}, {PlatformDeepseek, PlatformDeepseek, true},
		{PlatformOpenCodeGo, PlatformDeepseek, false}, {PlatformOpenAI, PlatformOpenCodeGo, false},
		{PlatformGrok, PlatformOpenCodeGo, false}, {PlatformDeepseek, PlatformOpenAI, false},
	} {
		require.Equal(t, tc.allowed, openAIAccountPlatformMatches(&Account{Platform: tc.account}, tc.request), tc)
	}
	require.False(t, openAIAccountPlatformMatches(nil, PlatformDeepseek))
}

type deepSeekOpenCodeSnapshot struct {
	SchedulerCache
	byPlatform map[string][]*Account
	reads      []SchedulerBucket
}

func (c *deepSeekOpenCodeSnapshot) GetSnapshot(_ context.Context, bucket SchedulerBucket) ([]*Account, bool, error) {
	c.reads = append(c.reads, bucket)
	accounts, hit := c.byPlatform[bucket.Platform]
	return accounts, hit, nil
}

func TestDeepSeekOpenCodeMergesCanonicalSnapshotsAndFailsClosed(t *testing.T) {
	groupID := int64(83)
	native := deepSeekOpenCodeFixture(1, groupID, PlatformDeepseek)
	openCode := deepSeekOpenCodeFixture(2, groupID, PlatformOpenCodeGo)
	cache := &deepSeekOpenCodeSnapshot{byPlatform: map[string][]*Account{PlatformDeepseek: {&native}, PlatformOpenCodeGo: {&openCode}}}
	snapshots := NewSchedulerSnapshotService(cache, nil, nil, nil, snapshotReaderConfig())
	defer snapshots.Stop()
	svc := &OpenAIGatewayService{schedulerSnapshot: snapshots}
	accounts, err := svc.listSchedulableAccounts(context.Background(), &groupID, PlatformDeepseek)
	require.NoError(t, err)
	require.Equal(t, []int64{1, 2}, []int64{accounts[0].ID, accounts[1].ID})
	for _, bucket := range cache.reads {
		require.Equal(t, groupID, bucket.GroupID)
	}
	delete(cache.byPlatform, PlatformOpenCodeGo)
	accounts, err = svc.listSchedulableAccounts(context.Background(), &groupID, PlatformDeepseek)
	require.ErrorIs(t, err, ErrSchedulerCacheNotReady)
	require.Nil(t, accounts, "never route around a missing provider snapshot")
}

func TestDeepSeekOpenCodeSelectionKeepsBindingsModelsAndPausedAccounts(t *testing.T) {
	groupID := int64(83)
	openCode := deepSeekOpenCodeFixture(2, groupID, PlatformOpenCodeGo)
	foreign := deepSeekOpenCodeFixture(3, 84, PlatformOpenCodeGo)
	repo := schedulerGroupAwareOpenAIAccountRepo{schedulerTestOpenAIAccountRepo{accounts: []Account{openCode, foreign}}}
	svc := &OpenAIGatewayService{accountRepo: repo, cache: &schedulerTestGatewayCache{}, concurrencyService: NewConcurrencyService(schedulerTestConcurrencyCache{})}
	ctx := context.Background()
	accounts, err := svc.listSchedulableAccounts(ctx, &groupID, PlatformDeepseek)
	require.NoError(t, err)
	require.Len(t, accounts, 1)
	require.Equal(t, openCode.ID, accounts[0].ID)
	selected, err := svc.selectAccountForModelWithExclusions(ctx, &groupID, PlatformDeepseek, "", "deepseek-v4.1-flash", nil, false, 0, "", false)
	require.NoError(t, err)
	require.Equal(t, PlatformOpenCodeGo, selected.Platform)
	scheduler := &defaultOpenAIAccountScheduler{service: svc, stats: newOpenAIAccountRuntimeStats()}
	req := OpenAIAccountScheduleRequest{GroupID: &groupID, Platform: PlatformDeepseek, RequestedModel: "deepseek-v4.1-flash", RequiredTransport: OpenAIUpstreamTransportAny}
	selection, _, _, _, err := scheduler.selectByLoadBalance(ctx, req)
	require.NoError(t, err)
	require.Equal(t, openCode.ID, selection.Account.ID)
	if selection.ReleaseFunc != nil {
		selection.ReleaseFunc()
	}
	req.SessionHash, req.StickyAccountID = "synthetic-session", openCode.ID
	selection, yielded, err := scheduler.selectBySessionHash(ctx, req)
	require.NoError(t, err)
	require.False(t, yielded)
	require.NotNil(t, selection)
	require.Equal(t, openCode.ID, selection.Account.ID)
	if selection.ReleaseFunc != nil {
		selection.ReleaseFunc()
	}
	paused := openCode
	paused.Schedulable = false
	require.False(t, isOpenAICompatibleAccountEligibleForRequest(ctx, &paused, &groupID, PlatformDeepseek, req.RequestedModel, false, ""))
	require.False(t, isOpenAICompatibleAccountEligibleForRequest(ctx, &openCode, &groupID, PlatformDeepseek, "unsupported-model", false, ""))
	limited := openCode
	limited.AccountGroups = []AccountGroup{{AccountID: limited.ID, GroupID: groupID, AllowedModels: []string{"beam-free"}}}
	require.False(t, isOpenAICompatibleAccountEligibleForRequest(ctx, &limited, &groupID, PlatformDeepseek, req.RequestedModel, false, ""))
	require.True(t, isOpenAICompatibleAccountEligibleForRequest(ctx, &limited, &groupID, PlatformDeepseek, "beam-free", false, ""))
	otherPool, err := svc.listSchedulableAccounts(ctx, &groupID, PlatformOpenAI)
	require.NoError(t, err)
	require.Empty(t, otherPool)
}

func TestDeepSeekOpenCodeModelsIncludeBoundProviderAndRespectPermissions(t *testing.T) {
	groupID := int64(83)
	account := deepSeekOpenCodeFixture(2, groupID, PlatformOpenCodeGo)
	account.AccountGroups = []AccountGroup{{AccountID: account.ID, GroupID: groupID, AllowedModels: []string{"deepseek-v4.1-flash"}}}
	repo := schedulerTestOpenAIAccountRepo{accounts: []Account{account}}
	svc := &GatewayService{accountRepo: repo}
	require.Equal(t, []string{"deepseek-v4.1-flash"}, svc.GetAvailableModels(context.Background(), &groupID, PlatformDeepseek))
	models := plazaAccountModels(&Group{ID: groupID, Platform: PlatformDeepseek}, []*Account{&account}, true)
	require.Len(t, models, 1)
	require.Equal(t, "deepseek-v4.1-flash", models[0].Name)
	require.Equal(t, PlatformDeepseek, models[0].Platform)
	require.Empty(t, plazaAccountModels(&Group{ID: groupID, Platform: PlatformGrok}, []*Account{&account}, false))
}
