package admin

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type intelligenceTestRequest struct {
	QuestionKind    string `json:"question_kind"`
	ReasoningEffort string `json:"reasoning_effort"`
}

// TestIntelligence accepts no model/prompt overrides. Admin middleware protects
// this route; the service independently enforces the OAuth platform restriction.
func (h *AccountHandler) TestIntelligence(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	var req intelligenceTestRequest
	// Empty bodies are kept compatible with the legacy iPhone knowledge test.
	if c.Request.Body != nil {
		if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
			response.BadRequest(c, "Invalid request: "+err.Error())
			return
		}
	}
	currentConcurrency := 0
	if h.concurrencyService != nil {
		if counts, countErr := h.concurrencyService.GetAccountConcurrencyBatch(c.Request.Context(), []int64{id}); countErr == nil {
			currentConcurrency = counts[id]
		}
	}
	result, err := h.accountTestService.TestAccountIntelligence(c.Request.Context(), id, req.QuestionKind, req.ReasoningEffort)
	if errors.Is(err, service.ErrIntelligenceAccountType) {
		response.BadRequest(c, err.Error())
		return
	}
	if errors.Is(err, service.ErrIntelligenceTestBusy) {
		response.Error(c, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	result.CurrentConcurrency = currentConcurrency
	if saveErr := h.accountTestService.SaveIntelligenceTestResult(c.Request.Context(), result); saveErr != nil {
		result.Error = strings.TrimSpace(strings.Join([]string{result.Error, "历史记录保存失败: " + saveErr.Error()}, " "))
	}
	response.Success(c, result)
}

// IntelligenceHistory returns the recent server-side results for an account.
func (h *AccountHandler) IntelligenceHistory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	history, err := h.accountTestService.GetIntelligenceTestHistory(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"items": history})
}
