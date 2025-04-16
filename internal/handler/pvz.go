package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/avito/pvz/internal/metrics"
	"github.com/avito/pvz/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type createPVZRequest struct {
	City string `json:"city" binding:"required"`
}

func (h *Handler) createPVZ(c *gin.Context) {
	// Проверяем роль пользователя
	role := c.GetString("userRole")
	if role != string(models.RoleEmployee) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only employees can create PVZ"})
		return
	}

	var req createPVZRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pvz, err := h.service.CreatePVZ(c.Request.Context(), req.City)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	metrics.PvzCreatedTotal.Inc()

	c.JSON(http.StatusCreated, pvz)
}

func (h *Handler) getPVZs(c *gin.Context) {
	// Проверяем роль пользователя
	role := c.GetString("userRole")
	if role != string(models.RoleEmployee) && role != string(models.RoleModerator) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only employees and moderators can view PVZ data"})
		return
	}

	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid startDate format"})
		return
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid endDate format"})
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 30 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
		return
	}

	pvzs, err := h.service.GetPVZsWithReceptions(c.Request.Context(), startDate, endDate, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pvzs)
}

func (h *Handler) closeLastReception(c *gin.Context) {
	// Проверяем роль пользователя
	role := c.GetString("userRole")
	if role != string(models.RoleEmployee) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only employees can close receptions"})
		return
	}

	pvzID, err := uuid.Parse(c.Param("pvzId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pvzId"})
		return
	}

	reception, err := h.service.CloseReception(c.Request.Context(), pvzID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reception)
}

func (h *Handler) deleteLastProduct(c *gin.Context) {
	// Проверяем роль пользователя
	role := c.GetString("userRole")
	if role != string(models.RoleEmployee) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only employees can delete products"})
		return
	}

	pvzID, err := uuid.Parse(c.Param("pvzId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pvzId"})
		return
	}

	if err := h.service.DeleteLastProduct(c.Request.Context(), pvzID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
