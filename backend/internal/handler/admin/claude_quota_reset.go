package admin

import (
	"context"
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// GET is a read-only probe; POST persists a shared, sanitized snapshot.
func (h *AccountHandler) QueryClaudeResetCredits(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	status, err := h.accountUsageService.QueryClaudeResetCredits(c.Request.Context(), id, c.Request.Method == "POST")
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, status)
}

func (h *AccountHandler) ResetClaudeQuota(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	var input service.ClaudeResetInput
	if c.ShouldBindJSON(&input) != nil {
		response.BadRequest(c, "Invalid reset confirmation")
		return
	}
	result, err := h.accountUsageService.ResetClaudeQuota(c.Request.Context(), id, input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	out := struct {
		*service.ClaudeResetResult
		Credits *service.ClaudeResetStatus `json:"credits"`
		Account *dto.Account               `json:"account,omitempty"`
		Warning bool                       `json:"refresh_warning"`
	}{ClaudeResetResult: result, Warning: result.RefreshWarning}
	// Consumption success is final even if local recovery/probing subsequently
	// fails. Never turn that into a retry-inducing 500 response.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(c.Request.Context()), 25*time.Second)
	defer cancel()
	out.Credits, err = h.accountUsageService.QueryClaudeResetCredits(ctx, id, true)
	if err != nil {
		out.Warning = true
	}
	if (result.Result == "reset" || result.Result == "already_used") && err == nil && h.accountUsageService.ClaudeQuotaCanResume(id) {
		if h.rateLimitService != nil {
			if err := h.rateLimitService.ClearRateLimit(ctx, id); err != nil {
				out.Warning = true
			}
		}
	}
	if a, err := h.adminService.GetAccount(ctx, id); err == nil {
		out.Account = dto.AccountForObserver(c.Request.Context(), dto.AccountFromService(a))
	} else {
		out.Warning = true
	}
	response.Success(c, out)
}
