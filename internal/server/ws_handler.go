package server

import (
	"ultahost-ai-gateway/internal/websocket"

	"github.com/gin-gonic/gin"
)

func WSHandle() gin.HandlerFunc {
	return websocket.HandleAgentWebSocket
}
