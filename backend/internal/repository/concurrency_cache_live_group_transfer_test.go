package repository

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestLiveLeaseTransferPreservesUserGroupConcurrency(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	cache := NewConcurrencyCache(client, 15, 900).(*concurrencyCache)
	ctx := context.Background()
	acquired, err := cache.AcquireAccountSlot(ctx, 10, 1, "account")
	require.NoError(t, err)
	require.True(t, acquired)
	acquired, err = cache.AcquireUserGroupSlot(ctx, 20, 40, 2, 1, "user-group")
	require.NoError(t, err)
	require.True(t, acquired)
	require.NoError(t, client.ZAdd(ctx, apiKeySlotKey(30), redis.Z{Score: float64(serverTime(t, client)), Member: "key"}).Err())
	request := service.LiveLeaseTransferRequest{AccountID: 10, AccountMax: 1, AccountRequestID: "account", UserID: 20, UserMax: 2, UserRequestID: "user-group", APIKeyID: 30, APIKeyMax: 1, KeyRequestID: "key", LeaseID: "joint", GroupID: 40, GroupMax: 1, GroupRequestID: "user-group"}
	acquired, err = cache.AcquireLiveLeaseTransferring(ctx, request)
	require.NoError(t, err)
	require.True(t, acquired)
	require.ErrorIs(t, client.ZScore(ctx, userGroupSlotKey(20, 40), "user-group").Err(), redis.Nil)
	require.NoError(t, client.ZScore(ctx, liveUserGroupSlotKey(20, 40), "joint").Err())
	acquired, err = cache.AcquireUserGroupSlot(ctx, 20, 40, 2, 1, "blocked")
	require.NoError(t, err)
	require.False(t, acquired)
	require.NoError(t, cache.ReleaseLiveLeaseAccount(ctx, 10, "joint"))
	acquired, err = cache.AcquireAccountSlot(ctx, 11, 1, "retry-account")
	require.NoError(t, err)
	require.True(t, acquired)
	request.AccountID, request.AccountRequestID, request.ReplacedAccountID = 11, "retry-account", 10
	acquired, err = cache.MigrateLiveLeaseAccount(ctx, request)
	require.NoError(t, err)
	require.True(t, acquired)
	refreshed, err := cache.RefreshLiveLeaseForGroup(ctx, 11, 20, 40, 30, "joint")
	require.NoError(t, err)
	require.True(t, refreshed)
	require.NoError(t, cache.ReleaseLiveLeaseAccount(ctx, 11, "joint"))
	require.NoError(t, cache.ReleaseLiveLeaseForGroup(ctx, 0, 20, 40, 30, "joint"))
	require.ErrorIs(t, client.ZScore(ctx, liveUserGroupSlotKey(20, 40), "joint").Err(), redis.Nil)
	acquired, err = cache.AcquireUserGroupSlot(ctx, 20, 40, 2, 1, "after-release")
	require.NoError(t, err)
	require.True(t, acquired)
}

func serverTime(t *testing.T, client *redis.Client) int64 {
	t.Helper()
	now, err := client.Time(context.Background()).Result()
	require.NoError(t, err)
	return now.Unix()
}
