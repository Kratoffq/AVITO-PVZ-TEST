package handler

import (
	"github.com/avito/pvz/internal/config"
	"github.com/avito/pvz/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service service.Service
	config  *config.Config
}

func NewHandler(service service.Service, config *config.Config) *Handler {
	return &Handler{
		service: service,
		config:  config,
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.Default()

	auth := router.Group("/auth")
	{
		auth.POST("/register", h.register)
		auth.POST("/login", h.login)
		auth.POST("/dummyLogin", h.dummyLogin)
	}

	api := router.Group("/api", h.authMiddleware)
	{
		pvz := api.Group("/pvz")
		{
			pvz.POST("", h.createPVZ)
			pvz.GET("", h.getPVZs)
			pvz.POST("/:pvzId/close_last_reception", h.closeLastReception)
			pvz.POST("/:pvzId/delete_last_product", h.deleteLastProduct)
		}

		receptions := api.Group("/receptions")
		{
			receptions.POST("", h.createReception)
		}

		products := api.Group("/products")
		{
			products.POST("", h.addProduct)
		}
	}

	return router
}
