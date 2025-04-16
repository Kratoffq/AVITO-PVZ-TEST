package handler

import (
	"net/http"

	"github.com/avito/pvz/internal/metrics"
	"github.com/avito/pvz/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type createReceptionRequest struct {
	PVZID uuid.UUID `json:"pvz_id" binding:"required"`
}

func (h *Handler) createReception(c *gin.Context) {
	// Проверяем роль пользователя
	role := c.GetString("userRole")
	if role != string(models.RoleEmployee) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only employees can create receptions"})
		return
	}

	var req createReceptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reception, err := h.service.CreateReception(c.Request.Context(), req.PVZID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	metrics.ReceptionsCreatedTotal.Inc()

	c.JSON(http.StatusCreated, reception)
}
