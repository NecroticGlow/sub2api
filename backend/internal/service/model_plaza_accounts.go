package service

import (
	"context"
	"strings"
)

// Production uses a narrow projection: the public directory never needs account
// tokens, API keys, endpoints, or transient scheduler/usage state.
type plazaAccountReader interface {
	ListModelPlazaAccounts(context.Context) ([]Account, error)
}

func (s *ModelPlazaService) plazaAccountsByGroup(ctx context.Context) (map[int64][]*Account, error) {
	byGroup := make(map[int64][]*Account)
	if s.accountRepo == nil {
		return byGroup, nil
	}
	var accounts []Account
	var err error
	if reader, ok := s.accountRepo.(plazaAccountReader); ok {
		accounts, err = reader.ListModelPlazaAccounts(ctx)
	} else {
		accounts, err = s.accountRepo.ListActive(ctx)
	}
	if err != nil {
		return nil, err
	}
	for i := range accounts {
		for _, gid := range accounts[i].GroupIDs {
			byGroup[gid] = append(byGroup[gid], &accounts[i])
		}
	}
	return byGroup, nil
}

func plazaDefaultModelIDs(platform string) []string {
	switch platform {
	case PlatformAnthropic, PlatformOpenAI, PlatformGemini, PlatformAntigravity, PlatformGrok, PlatformOpenCodeGo:
		return defaultModelsListCandidateIDs(platform)
	default:
		// CN providers require their own configuration; the generic admin helper
		// otherwise falls back to Claude, which would publish unrelated models.
		return nil
	}
}

func plazaAccountModels(g *Group, accounts []*Account, noChannelModels bool) []PlazaModel {
	var out []PlazaModel
	matchedAccount := false
	add := func(name, platform string) {
		if strings.TrimSpace(name) != "" && !strings.ContainsAny(name, "*?") {
			out = append(out, PlazaModel{Name: name, Platform: platform})
		}
	}
	for _, account := range accounts {
		platform := account.Platform
		bridge := platform == PlatformAntigravity && (g.Platform == PlatformGemini || g.Platform == PlatformAnthropic)
		if g.Platform == PlatformComposite {
			if !isConcreteRequestPlatform(platform) {
				continue
			}
		} else if platform != g.Platform && !bridge {
			continue
		}
		matchedAccount = true
		mapping := account.GetModelMapping()
		addAccountModel := func(name string) {
			if bridge {
				prefix := "gemini"
				if g.Platform == PlatformAnthropic {
					prefix = "claude"
				}
				if !strings.HasPrefix(strings.ToLower(name), prefix) {
					return
				}
				add(name, g.Platform)
				return
			}
			add(name, platform)
		}
		for name := range mapping {
			addAccountModel(name)
		}
		// Expand configured wildcards against the maintained provider catalog.
		// Unmapped/passthrough accounts expose the same defaults as /v1/models.
		for _, name := range plazaDefaultModelIDs(platform) {
			if len(mapping) == 0 || account.IsOpenAIPassthroughEnabled() || account.IsModelSupported(name) {
				addAccountModel(name)
			}
		}
	}
	if !matchedAccount && noChannelModels && len(g.ModelPricing) == 0 && !g.ModelAllowlist.Enabled && !g.ModelRoutingEnabled {
		for _, name := range plazaDefaultModelIDs(g.Platform) {
			add(name, g.Platform)
		}
	}
	return out
}
