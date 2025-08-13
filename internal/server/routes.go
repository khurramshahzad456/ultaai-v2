package server

import (
	"ultahost-ai-gateway/internal/api"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {

	r.GET("/agent/connect", api.HandleAgentWebSocket)
	// r.POST("/agent/register", api.HandleAgentRegister)
	r.POST("/agent/register", api.InstallTokenMiddleware(), api.HandleAgentRegister)
	r.Use(api.AuthMiddleware())

	r.POST("/chat", api.HandleChat)
	r.POST("/agent/enable", api.HandleEnableUltaAI)

}
