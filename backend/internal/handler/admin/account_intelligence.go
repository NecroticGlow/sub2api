package admin

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// TestIntelligence accepts no model/prompt overrides. Admin middleware protects
// this route; the service independently enforces the OAuth platform restriction.
func (h *AccountHandler) TestIntelligence(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	result, err := h.accountTestService.TestAccountIntelligence(c.Request.Context(), id)
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
	response.Success(c, result)
}
