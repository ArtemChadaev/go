package handler

import (
	"github.com/ArtemChadaev/go/pkg/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	services *service.Service
	redis    *redis.Client
}

func NewHandler(services *service.Service, redis *redis.Client) *Handler {
	return &Handler{
		services: services,
		redis:    redis,
	}
}

// InitRoutes Машруты
func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	auth := router.Group("/auth", h.authRateLimiter)
	{
		auth.POST("/sign-up", h.signUp)
		auth.POST("/sign-in", h.signIn)
		auth.POST("/refresh", h.updateToken)

	}

	api := router.Group("/api", h.userIdentify, h.rateLimiter)
	{
		settings := api.Group("/settings")
		{
			settings.POST("/subscript")
			settings.POST("/dayCoin", h.dayCoin)
			settings.GET("/", h.getMySettings)
			settings.PUT("/", h.setNameIcon)
		}
	}

	return router
}
