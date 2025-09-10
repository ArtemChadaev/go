package handler

import (
	"github.com/ChadaevArtem/rest-go-for-vue/pkg/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	services *service.Service
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{services: services}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	auth := router.Group("/auth")
	{
		auth.POST("/sign-up", h.signUp)
		auth.POST("/sign-in", h.signIn)
	}

	//api := router.Group("/api", h.userIdentify)
	//{
	//	lists := api.Group("/lists")
	//	{
	//		lists.POST("/")
	//		lists.GET("/")
	//		lists.GET("/:id")
	//		lists.PUT("/:id")
	//		lists.DELETE("/:id")
	//	}
	//}

	return router
}
