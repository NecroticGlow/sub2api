package service

import (
	"context"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// KeyForModelDiscovery is a read-only panel lookup, not generation admission.
// It checks ownership before reading the full owner or any group metadata.
func (s *APIKeyService) KeyForModelDiscovery(ctx context.Context, id, ownerID int64) (*APIKey, error) {
	key, err := s.apiKeyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if key == nil || key.UserID != ownerID {
		return nil, ErrAPIKeyNotFound
	}
	// GetByID's shallow User edge does not include AllowedGroups. Reload the
	// full owner so exclusive/restricted groups are neither lost nor bypassed.
	user, err := s.userRepo.GetByID(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.IsActive() {
		return nil, infraerrors.Forbidden("USER_INACTIVE", "User account is not active")
	}
	copyKey := *key
	copyKey.User = user
	groupID := key.GroupID
	if groupID == nil {
		ids := key.ConfiguredGroupIDs()
		if len(ids) == 0 {
			return nil, infraerrors.BadRequest("GROUP_REQUIRED", "API key has no assigned group")
		}
		groupID = &ids[0]
	}
	// Pin reloads the bound group, checks status/permissions/subscription, and
	// loads per-group denied models on a request-local copy. No key is updated.
	return s.PinAPIKeyGroup(ctx, &copyKey, *groupID)
}
