package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

const PlazaPopularityDays = 7
const PlazaPopularityLimit = 10

// PlazaModelUsage contains aggregate counts only, never user/request identifiers.
type PlazaModelUsage struct {
	GroupID  int64
	Platform string
	Model    string
	Requests int64
}

type plazaUsageReader interface {
	ListModelPlazaUsage(context.Context, time.Time, time.Time) ([]PlazaModelUsage, error)
}

type plazaPopularityCache struct {
	mu        sync.Mutex
	expiresAt time.Time
	usage     []PlazaModelUsage
}

func (s *ModelPlazaService) popularPlazaGroups(ctx context.Context, groups []PlazaGroup) ([]PlazaGroup, error) {
	if s.usageRepo == nil {
		return groups, nil
	}
	reader, ok := s.usageRepo.(plazaUsageReader)
	if !ok {
		return nil, fmt.Errorf("model popularity reader unavailable")
	}
	cache := &s.popularityCache
	cache.mu.Lock()
	defer cache.mu.Unlock()
	now := time.Now()
	if !now.Before(cache.expiresAt) {
		queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		usage, err := reader.ListModelPlazaUsage(queryCtx, now.Add(-PlazaPopularityDays*24*time.Hour), now)
		cancel()
		if err != nil {
			return nil, fmt.Errorf("read popular models: %w", err)
		}
		cache.usage = usage
		cache.expiresAt = now.Add(5 * time.Minute)
	}
	return selectPopularPlazaGroups(groups, cache.usage, PlazaPopularityLimit), nil
}

func plazaPopularityKey(platform, model string) string {
	return platform + ":" + strings.ToLower(strings.TrimSpace(model))
}

// Count only current public routes. A private/deleted group must not affect the
// public ranking, including when a group's visibility changes within the cache TTL.
func selectPopularPlazaGroups(groups []PlazaGroup, usage []PlazaModelUsage, limit int) []PlazaGroup {
	configured := make(map[int64]map[string]bool)
	for _, g := range groups {
		if g.IsExclusive {
			continue
		}
		configured[g.ID] = make(map[string]bool)
		for _, m := range g.Models {
			configured[g.ID][plazaPopularityKey(m.Platform, m.Name)] = true
		}
	}
	counts := make(map[string]int64)
	for _, row := range usage {
		key := plazaPopularityKey(row.Platform, row.Model)
		if row.Requests > 0 && configured[row.GroupID][key] {
			counts[key] += row.Requests
		}
	}
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if counts[keys[i]] != counts[keys[j]] {
			return counts[keys[i]] > counts[keys[j]]
		}
		return keys[i] < keys[j]
	})
	if limit < 0 {
		limit = 0
	}
	if len(keys) > limit {
		keys = keys[:limit]
	}
	ranks := make(map[string]int, len(keys))
	for i, key := range keys {
		ranks[key] = i + 1
	}
	out := make([]PlazaGroup, 0, len(groups))
	for _, g := range groups {
		if g.IsExclusive {
			continue
		}
		models := make([]PlazaModel, 0)
		for _, m := range g.Models {
			if rank := ranks[plazaPopularityKey(m.Platform, m.Name)]; rank > 0 {
				m.PopularityRank = rank
				models = append(models, m)
			}
		}
		if len(models) == 0 {
			continue
		}
		sort.Slice(models, func(i, j int) bool { return models[i].PopularityRank < models[j].PopularityRank })
		g.Models = models
		out = append(out, g)
	}
	return out
}
