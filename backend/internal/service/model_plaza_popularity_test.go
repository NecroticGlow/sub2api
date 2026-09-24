//go:build unit

package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPlazaPopularity_TopTenPublicModelsAndAllCurrentRoutes(t *testing.T) {
	groups := []PlazaGroup{{ID: 1}, {ID: 2}, {ID: 3, IsExclusive: true}}
	var usage []PlazaModelUsage
	for i := 0; i < 12; i++ {
		name := fmt.Sprintf("gpt-%02d", i)
		groups[0].Models = append(groups[0].Models, PlazaModel{Name: name, Platform: PlatformOpenAI})
		usage = append(usage, PlazaModelUsage{GroupID: 1, Platform: PlatformOpenAI, Model: name, Requests: int64(100 - i)})
	}
	groups[1].Models = []PlazaModel{{Name: "GPT-01", Platform: PlatformOpenAI}, {Name: "unused", Platform: PlatformOpenAI}}
	groups[2].Models = []PlazaModel{{Name: "gpt-11", Platform: PlatformOpenAI}}
	usage = append(usage,
		PlazaModelUsage{GroupID: 2, Platform: PlatformOpenAI, Model: "gpt-01", Requests: 100},
		PlazaModelUsage{GroupID: 3, Platform: PlatformOpenAI, Model: "gpt-11", Requests: 99999},
		PlazaModelUsage{GroupID: 1, Platform: PlatformOpenAI, Model: "removed-model", Requests: 99999},
		PlazaModelUsage{GroupID: 99, Platform: PlatformOpenAI, Model: "gpt-11", Requests: 99999},
	)
	out := selectPopularPlazaGroups(groups, usage, 10)
	require.Len(t, out, 2)
	require.Len(t, out[0].Models, 10)
	require.Equal(t, "gpt-01", out[0].Models[0].Name)
	require.Equal(t, 1, out[0].Models[0].PopularityRank)
	require.Len(t, out[1].Models, 1)
	require.Equal(t, 1, out[1].Models[0].PopularityRank)
	require.NotContains(t, plazaModelsByName(out[0].Models), "gpt-11")
	require.Zero(t, groups[0].Models[0].PopularityRank)
}

func TestPlazaPopularity_NoHistoryAndStableTies(t *testing.T) {
	groups := []PlazaGroup{{ID: 1, Models: []PlazaModel{
		{Name: "b", Platform: PlatformOpenAI}, {Name: "a", Platform: PlatformOpenAI},
		{Name: "a", Platform: PlatformAnthropic},
	}}}
	require.Empty(t, selectPopularPlazaGroups(groups, nil, 10))
	usage := []PlazaModelUsage{
		{GroupID: 1, Platform: PlatformOpenAI, Model: "b", Requests: 5},
		{GroupID: 1, Platform: PlatformOpenAI, Model: " A ", Requests: 5},
	}
	out := selectPopularPlazaGroups(groups, usage, 1)
	require.Len(t, out[0].Models, 1)
	require.Equal(t, "a", out[0].Models[0].Name)
	require.Equal(t, PlatformOpenAI, out[0].Models[0].Platform)
}

type plazaUsageStub struct {
	UsageLogRepository
	calls      int
	err        error
	start, end time.Time
}

func (r *plazaUsageStub) ListModelPlazaUsage(_ context.Context, start, end time.Time) ([]PlazaModelUsage, error) {
	r.calls++
	r.start, r.end = start, end
	return []PlazaModelUsage{{GroupID: 1, Platform: PlatformOpenAI, Model: "gpt-test", Requests: 10}}, r.err
}

func TestPlazaPopularity_SevenDayWindowCacheAndErrors(t *testing.T) {
	repo := &plazaUsageStub{}
	svc := newPlazaService(nil, nil, nil)
	svc.usageRepo = repo
	groups := []PlazaGroup{{ID: 1, Models: []PlazaModel{{Name: "gpt-test", Platform: PlatformOpenAI}}}}
	for i := 0; i < 2; i++ {
		out, err := svc.popularPlazaGroups(context.Background(), groups)
		require.NoError(t, err)
		require.Len(t, out, 1)
	}
	require.Equal(t, 1, repo.calls)
	require.Equal(t, 7*24*time.Hour, repo.end.Sub(repo.start))
	// Visibility changes take effect even while the usage snapshot is cached.
	groups[0].IsExclusive = true
	out, err := svc.popularPlazaGroups(context.Background(), groups)
	require.NoError(t, err)
	require.Empty(t, out)
	svc.popularityCache.expiresAt = time.Time{}
	repo.err = errors.New("database unavailable")
	out, err = svc.popularPlazaGroups(context.Background(), groups)
	require.ErrorIs(t, err, repo.err)
	require.Nil(t, out) // never silently replace the ranking with a full catalog
}
