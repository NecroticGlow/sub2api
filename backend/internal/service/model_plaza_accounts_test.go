//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type plazaAccountRepoStub struct {
	AccountRepository
	accounts []Account
	err      error
}

func (r *plazaAccountRepoStub) ListModelPlazaAccounts(context.Context) ([]Account, error) {
	return r.accounts, r.err
}

func TestListPlazaGroups_AccountMappingsAcrossPublicGroups(t *testing.T) {
	groups := []Group{
		{ID: 1, Platform: PlatformOpenAI, Name: "GPT"},
		{ID: 2, Platform: PlatformOpenAI, Name: "Claude over OpenAI"},
		{ID: 3, Platform: PlatformGemini, Name: "Gemini"},
		{ID: 4, Platform: PlatformOpenAI, Name: "Exclusive", IsExclusive: true},
	}
	svc := newPlazaService(nil, groups, nil)
	svc.accountRepo = &plazaAccountRepoStub{accounts: []Account{
		{Platform: PlatformOpenAI, GroupIDs: []int64{1}, Credentials: map[string]any{"model_mapping": map[string]any{"gpt-5.6-sol": "upstream-private-id"}}},
		{Platform: PlatformOpenAI, GroupIDs: []int64{2}, Credentials: map[string]any{"model_mapping": map[string]any{"claude-opus-4-6": "upstream"}}},
		{Platform: PlatformAntigravity, GroupIDs: []int64{3}},
		{Platform: PlatformOpenAI, GroupIDs: []int64{4}, Credentials: map[string]any{"model_mapping": map[string]any{"exclusive-model": "private"}}},
	}}
	out, err := svc.ListGroups(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 4)
	byID := map[int64]PlazaGroup{}
	for _, group := range out {
		byID[group.ID] = group
	}
	require.Equal(t, "gpt-5.6-sol", byID[1].Models[0].Name)
	require.Len(t, byID[1].Models, 1)
	require.Equal(t, "claude-opus-4-6", byID[2].Models[0].Name)
	require.Len(t, byID[2].Models, 1)
	require.NotEmpty(t, byID[3].Models)
	for _, model := range byID[3].Models {
		require.Equal(t, PlatformGemini, model.Platform)
		require.Contains(t, model.Name, "gemini")
	}
	require.True(t, byID[4].IsExclusive) // handler remains responsible for visibility
}

func TestListPlazaGroups_AccountAllowlistWildcardAndPrice(t *testing.T) {
	g := Group{ID: 1, Platform: PlatformOpenAI,
		ModelAllowlist: GroupModelAllowlist{Enabled: true, Models: []string{"gpt-5.6-*"}},
		ModelPricing:   []ChannelModelPricing{{Platform: PlatformOpenAI, Models: []string{"gpt-5.6-sol"}, InputPrice: testPtrFloat64(2e-6)}},
	}
	svc := newPlazaService(nil, []Group{g}, nil)
	svc.accountRepo = &plazaAccountRepoStub{accounts: []Account{
		{Platform: PlatformOpenAI, GroupIDs: []int64{1}, Credentials: map[string]any{"model_mapping": map[string]any{"gpt-5.6-*": "upstream", "claude-custom": "upstream"}}},
	}}
	out, err := svc.ListGroups(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 1)
	models := plazaModelsByName(out[0].Models)
	require.Contains(t, models, "gpt-5.6-sol")
	require.Contains(t, models, "gpt-5.6-luna")
	require.NotContains(t, models, "gpt-5.6-*")
	require.NotContains(t, models, "claude-custom")
	require.Equal(t, 2e-6, *models["gpt-5.6-sol"].Pricing.InputPrice)
}

func TestListPlazaGroups_DefaultsWithoutStaleAllowlist(t *testing.T) {
	svc := newPlazaService(nil, []Group{
		{ID: 1, Platform: PlatformOpenAI, ModelAllowlist: GroupModelAllowlist{Models: []string{"stale-ui-model"}}},
		{ID: 2, Platform: PlatformZhipu, ModelAllowlist: GroupModelAllowlist{Models: []string{"claude-stale"}}},
	}, nil)
	svc.accountRepo = &plazaAccountRepoStub{}
	out, err := svc.ListGroups(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.EqualValues(t, 1, out[0].ID)
	models := plazaModelsByName(out[0].Models)
	require.Contains(t, models, "gpt-5.6-sol")
	require.NotContains(t, models, "stale-ui-model")
}

func TestListPlazaGroups_AccountPlatformIsolationAndComposite(t *testing.T) {
	account := &Account{Platform: PlatformOpenAI, Credentials: map[string]any{"model_mapping": map[string]any{"gpt-custom": "upstream"}}}
	require.Empty(t, plazaAccountModels(&Group{Platform: PlatformDeepseek}, []*Account{account}, true))
	models := plazaAccountModels(&Group{Platform: PlatformComposite}, []*Account{account}, true)
	require.Len(t, models, 1)
	require.Equal(t, PlatformOpenAI, models[0].Platform)
}

func TestListPlazaGroups_MixedMappedAndUnmappedAccounts(t *testing.T) {
	svc := newPlazaService(nil, []Group{{ID: 1, Platform: PlatformOpenAI}}, nil)
	svc.accountRepo = &plazaAccountRepoStub{accounts: []Account{
		{Platform: PlatformOpenAI, GroupIDs: []int64{1}, Credentials: map[string]any{"model_mapping": map[string]any{"custom-gpt": "upstream"}}},
		{Platform: PlatformOpenAI, GroupIDs: []int64{1}},
		{Platform: PlatformOpenAI, GroupIDs: []int64{1}},
	}}
	out, err := svc.ListGroups(context.Background())
	require.NoError(t, err)
	models := plazaModelsByName(out[0].Models)
	require.Contains(t, models, "custom-gpt")
	require.Contains(t, models, "gpt-5.6-sol")
	require.Len(t, out[0].Models, len(models))
}

func TestListPlazaGroups_AccountReadFailure(t *testing.T) {
	sentinel := errors.New("catalog read failed")
	svc := newPlazaService(nil, nil, nil)
	svc.accountRepo = &plazaAccountRepoStub{err: sentinel}
	out, err := svc.ListGroups(context.Background())
	require.Nil(t, out)
	require.ErrorIs(t, err, sentinel)
}
