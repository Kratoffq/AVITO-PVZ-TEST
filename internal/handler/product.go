package handler

import (
	"net/http"

	"github.com/avito/pvz/internal/metrics"
	"github.com/avito/pvz/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type addProductRequest struct {
	PVZID       uuid.UUID          `json:"pvz_id" binding:"required"`
	ProductType models.ProductType `json:"type" binding:"required,oneof=электроника одежда обувь"`
}

func (h *Handler) addProduct(c *gin.Context) {
	// Проверяем роль пользователя
	role := c.GetString("userRole")
	if role != string(models.RoleEmployee) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only employees can add products"})
		return
	}

	var req addProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.service.AddProduct(c.Request.Context(), req.PVZID, req.ProductType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	metrics.ProductsAddedTotal.Inc()

	c.JSON(http.StatusCreated, product)
}
