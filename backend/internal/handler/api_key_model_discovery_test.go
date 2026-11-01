package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/antigravity"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyModelDiscoveryRequiresPanelIdentityAndValidID(t *testing.T) {
	for _, tc := range []struct {
		path          string
		authenticated bool
		status        int
	}{
		{"/keys/42/models", false, http.StatusUnauthorized},
		{"/keys/invalid/models", true, http.StatusBadRequest},
		{"/keys/0/models", true, http.StatusBadRequest},
		{"/keys/-1/models", true, http.StatusBadRequest},
	} {
		t.Run(tc.path, func(t *testing.T) {
			router := gin.New()
			h := &APIKeyHandler{} // A rejected request must never touch service/storage.
			if tc.authenticated {
				router.Use(func(c *gin.Context) {
					c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: 1})
					c.Next()
				})
			}
			called := false
			router.GET("/keys/:id/models", h.PrepareModelDiscovery, func(c *gin.Context) { called = true; c.Status(http.StatusOK) })
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))
			require.Equal(t, tc.status, recorder.Code)
			require.False(t, called)
			require.Equal(t, "no-store", recorder.Header().Get("Cache-Control"))
		})
	}
}

func TestAPIKeyModelDiscoveryPreservesAntigravityCatalogueFiltering(t *testing.T) {
	models := antigravity.DefaultModels()
	require.GreaterOrEqual(t, len(models), 2)
	allowed, denied := models[0].ID, models[1].ID
	key := &service.APIKey{Group: &service.Group{Platform: service.PlatformAntigravity, ModelAllowlist: service.GroupModelAllowlist{Enabled: true, Models: []string{allowed, denied}}}, User: &service.User{UserGroupDeniedModels: []string{denied}}}
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(string(middleware2.ContextKeyAPIKey), key); c.Next() })
	h := &GatewayHandler{}
	router.GET("/models", h.PanelModels)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/models", nil))
	require.Equal(t, http.StatusOK, recorder.Code)
	var body struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Len(t, body.Data, 1)
	require.Equal(t, allowed, body.Data[0].ID)
}
