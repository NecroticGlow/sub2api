package handler

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// PrepareModelDiscovery exposes the normal model catalogue through the signed-in
// panel, not through an API URL that a reverse proxy may redirect cross-origin.
// It is metadata-only: no generated request, usage write, or API-key forwarding.
// The caller must still own the key and retain access to its current group.
func (h *APIKeyHandler) PrepareModelDiscovery(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		c.Abort()
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid key ID")
		c.Abort()
		return
	}
	key, err := h.apiKeyService.KeyForModelDiscovery(c.Request.Context(), id, subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}
	c.Set(string(middleware2.ContextKeyAPIKey), key)
	c.Next()
}

// PanelModels keeps the same Antigravity catalogue used by the legacy import
// endpoint; other platforms use their normal group-filtered model list.
func (h *GatewayHandler) PanelModels(c *gin.Context) {
	key, ok := middleware2.GetAPIKeyFromContext(c)
	if ok && key != nil && key.Group != nil && key.Group.Platform == service.PlatformAntigravity {
		h.AntigravityModels(c)
		return
	}
	h.Models(c)
}
