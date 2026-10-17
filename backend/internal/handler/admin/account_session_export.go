package admin

import (
	"net/http"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// ExportOpenAISession returns the upstream payload only after verifying account
// identity. Cookie persistence is explicit and encrypted outside account JSON.
func (h *AccountHandler) ExportOpenAISession(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 32<<10)
	var input struct {
		Cookie         string `json:"cookie"`
		RememberCookie bool   `json:"remember_cookie"`
	}
	if c.ShouldBindJSON(&input) != nil {
		response.BadRequest(c, "Invalid Session export request")
		return
	}
	account, err := h.adminService.GetAccount(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	session, err := h.openaiOAuthService.ExportSavedChatGPTSession(c.Request.Context(), account, input.Cookie, input.RememberCookie)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, session)
}

// GetOpenAISessionCookie returns presence and storage capability, never secrets.
func (h *AccountHandler) GetOpenAISessionCookie(c *gin.Context) {
	h.openAISessionCookieState(c, false)
}

func (h *AccountHandler) DeleteOpenAISessionCookie(c *gin.Context) {
	h.openAISessionCookieState(c, true)
}

func (h *AccountHandler) openAISessionCookieState(c *gin.Context, remove bool) {
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	account, err := h.adminService.GetAccount(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if remove {
		if err := h.openaiOAuthService.DeleteChatGPTSessionCookie(c.Request.Context(), account); err != nil {
			response.ErrorFrom(c, err)
			return
		}
		response.Success(c, gin.H{"deleted": true})
		return
	}
	state, err := h.openaiOAuthService.ChatGPTSessionCookieState(c.Request.Context(), account)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, state)
}
