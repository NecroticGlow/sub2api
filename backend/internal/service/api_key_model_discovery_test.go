package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type ccsModelKeyRepo struct {
	APIKeyRepository
	key *APIKey
	err error
}

func (r *ccsModelKeyRepo) GetByID(context.Context, int64) (*APIKey, error) { return r.key, r.err }

type ccsModelUserRepo struct {
	UserRepository
	user  *User
	reads int
}

func (r *ccsModelUserRepo) GetByID(context.Context, int64) (*User, error) {
	r.reads++
	return r.user, nil
}

type ccsModelGroupRepo struct {
	GroupRepository
	group *Group
	reads int
}

func (r *ccsModelGroupRepo) GetByIDLite(_ context.Context, id int64) (*Group, error) {
	r.reads++
	if r.group == nil || r.group.ID != id {
		return nil, ErrGroupNotFound
	}
	return r.group, nil
}

type ccsModelRestrictions struct{ UserGroupRateRepository }

func (*ccsModelRestrictions) GetRPMOverrideByUserAndGroup(context.Context, int64, int64) (*int, error) {
	return nil, nil
}
func (*ccsModelRestrictions) GetDeniedModelsByUserAndGroup(context.Context, int64, int64) ([]string, error) {
	return []string{"denied-model"}, nil
}

func TestAPIKeyModelDiscoveryOwnershipAndPermissions(t *testing.T) {
	for _, scenario := range []string{"allowed-exclusive", "foreign-owner", "missing-key", "lookup-error", "inactive-user", "revoked-group", "restricted-public-group", "subscription-without-access", "inactive-group", "deleted-group", "no-group", "multi-group-primary"} {
		t.Run(scenario, func(t *testing.T) {
			id := int64(42)
			key := &APIKey{ID: 7, UserID: 1, Key: "synthetic-private-key", GroupID: &id, User: &User{ID: 1, Status: StatusActive}}
			keys := &ccsModelKeyRepo{key: key}
			users := &ccsModelUserRepo{user: &User{ID: 1, Status: StatusActive, AllowedGroups: []int64{42}}}
			groups := &ccsModelGroupRepo{group: &Group{ID: 42, Status: StatusActive, IsExclusive: true}}
			owner := int64(1)
			var want error
			switch scenario {
			case "foreign-owner":
				owner = 2
				want = ErrAPIKeyNotFound
			case "missing-key":
				keys.key = nil
				want = ErrAPIKeyNotFound
			case "lookup-error":
				keys.err = errors.New("synthetic storage unavailable")
				want = keys.err
			case "inactive-user":
				users.user.Status = "disabled"
			case "revoked-group":
				users.user.AllowedGroups = nil
				want = ErrGroupNotAllowed
			case "restricted-public-group":
				groups.group.IsExclusive = false
				users.user.RestrictPublicGroups = true
				users.user.AllowedGroups = nil
				want = ErrGroupNotAllowed
			case "subscription-without-access":
				groups.group.SubscriptionType = "subscription"
				want = ErrGroupNotAllowed
			case "inactive-group":
				groups.group.Status = "inactive"
				want = ErrGroupNotAllowed
			case "deleted-group":
				groups.group = nil
				want = ErrGroupNotFound
			case "no-group":
				key.GroupID = nil
			case "multi-group-primary":
				key.GroupID = nil
				key.GroupIDs = []int64{42, 43}
			}
			svc := &APIKeyService{apiKeyRepo: keys, userRepo: users, groupRepo: groups, userGroupRateRepo: &ccsModelRestrictions{}}
			actual, err := svc.KeyForModelDiscovery(context.Background(), 7, owner)
			if want != nil {
				require.ErrorIs(t, err, want)
			} else if scenario == "inactive-user" || scenario == "no-group" {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, int64(42), *actual.GroupID)
				require.Same(t, groups.group, actual.Group)
				require.Equal(t, []string{"denied-model"}, actual.DeniedModelsInGroup())
				require.NotSame(t, actual, key)
				require.NotSame(t, actual.User, key.User)
				require.Nil(t, key.User.UserGroupDeniedModels, "must not mutate the source key/user")
			}
			if scenario == "foreign-owner" || scenario == "missing-key" || scenario == "lookup-error" {
				require.Zero(t, users.reads)
				require.Zero(t, groups.reads, "reject before accessing another owner's group")
			}
		})
	}
}
