package server

import (
	"encoding/json"
	"net/http"

	"ultahost-ai-gateway/internal/api"
	"ultahost-ai-gateway/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRoutes(r *gin.Engine) {

	// Agent connect / register
	r.GET("/agent/connect", websocket.HandleAgentWebSocket)
	r.POST("/agent/register", api.InstallTokenMiddleware(), api.HandleAgentRegister)

	// Auth-protected
	r.Use(api.AuthMiddleware())
	r.POST("/chat", api.HandleChat)
	r.POST("/agent/enable", api.HandleEnableUltaAI)

	// Message routing by agent ID
	r.POST("/agents/:vpsId/send", func(c *gin.Context) {
		type req struct {
			Payload json.RawMessage `json:"payload" binding:"required"`
		}
		var body req
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		vpsId := c.Param("vpsId")
		if err := websocket.SendMessage(vpsId, body.Payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "queued"})
	})

	// Simple pool stats
	r.GET("/agents/pool/stats", func(c *gin.Context) {
		agents, totalMsgs, totalBytes := websocket.OfflineStats()
		c.JSON(http.StatusOK, gin.H{
			"active_connections": websocket.PoolCount(),
			"offline_agents":     agents,
			"offline_msgs":       totalMsgs,
			"offline_bytes":      totalBytes,
		})
	})

	// Prometheus metrics
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
