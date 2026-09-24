//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListPlazaGroups_DirectGroupPricingWithoutChannels(t *testing.T) {
	g := Group{ID: 10, Name: "Direct", Platform: "zhipu", RateMultiplier: 1,
		ModelPricing: []ChannelModelPricing{{Platform: "zhipu", Models: []string{"glm-direct"}, InputPrice: testPtrFloat64(0.8e-6)}},
	}
	out, err := newPlazaService(nil, []Group{g}, nil).ListGroups(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Len(t, out[0].Models, 1)
	require.Equal(t, "glm-direct", out[0].Models[0].Name)
	require.Equal(t, 0.8e-6, *out[0].Models[0].Pricing.InputPrice)
}

func TestListPlazaGroups_DirectAllowlistAndRouting(t *testing.T) {
	g := Group{ID: 10, Platform: "deepseek", ModelAllowlist: GroupModelAllowlist{Enabled: true, Models: []string{"deepseek-direct", "deepseek-*"}},
		ModelRoutingEnabled: true, ModelRouting: map[string][]int64{"deepseek-routed": {1}, "deepseek-empty": {}, "other": {2}},
	}
	out, err := newPlazaService(nil, []Group{g}, nil).ListGroups(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Len(t, out[0].Models, 2)
	require.Equal(t, "deepseek-direct", out[0].Models[0].Name)
	require.Equal(t, "deepseek-routed", out[0].Models[1].Name)
	require.Nil(t, out[0].Models[0].Pricing)
}

func TestListPlazaGroups_DoesNotPublishDisabledDefaults(t *testing.T) {
	g := Group{ID: 10, Platform: "openai", ModelAllowlist: GroupModelAllowlist{Models: []string{"gpt-ui-default"}}, ModelRouting: map[string][]int64{"gpt-disabled-route": {1}}}
	out, err := newPlazaService(nil, []Group{g}, nil).ListGroups(context.Background())
	require.NoError(t, err)
	require.Empty(t, out)
}

func TestListPlazaGroups_GroupPricePrecedenceAndAllowlist(t *testing.T) {
	g := Group{ID: 10, Platform: "openai", ModelAllowlist: GroupModelAllowlist{Enabled: true, Models: []string{"gpt-allowed"}},
		ModelPricing: []ChannelModelPricing{{Platform: "openai", Models: []string{"gpt-allowed"}, InputPrice: testPtrFloat64(0)}, {Platform: "anthropic", Models: []string{"claude-wrong-platform"}}},
	}
	ch := plazaPricedChannel(1, "channel", []int64{10}, "openai", "gpt-allowed", "gpt-blocked")
	out, err := newPlazaService([]Channel{ch}, []Group{g}, nil).ListGroups(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Len(t, out[0].Models, 1)
	require.Equal(t, 0.0, *out[0].Models[0].Pricing.InputPrice)
	require.Equal(t, 3e-6, *ch.ModelPricing[0].InputPrice)
}

func TestListPlazaGroups_CompositeRequiresExplicitPlatform(t *testing.T) {
	g := Group{ID: 10, Platform: PlatformComposite, ModelAllowlist: GroupModelAllowlist{Enabled: true, Models: []string{"*"}},
		ModelPricing: []ChannelModelPricing{{Platform: "openai", Models: []string{"gpt-direct", "GPT-direct"}}, {Platform: PlatformComposite, Models: []string{"invalid"}}},
	}
	out, err := newPlazaService(nil, []Group{g}, nil).ListGroups(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Len(t, out[0].Models, 1)
	require.Equal(t, "openai", out[0].Models[0].Platform)
}
