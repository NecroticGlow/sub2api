package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func (h *AccountHandler) GetPelicanManualHistory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	items, err := h.accountTestService.GetPelicanManualHistory(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

func (h *AccountHandler) SavePelicanManualRecord(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 256<<10)
	var record service.PelicanManualRecord
	if c.ShouldBindJSON(&record) != nil {
		response.BadRequest(c, "Invalid or oversized manual test record")
		return
	}
	if err := service.ValidatePelicanManualRecord(&record); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.accountTestService.SavePelicanManualRecord(c.Request.Context(), id, &record); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"saved": true})
}
