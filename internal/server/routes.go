// internal/server/routes.go
package server

import (
	"encoding/json"
	"net/http"
	"net/http/pprof" // NEW

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

	// Pool & offline stats
	r.GET("/agents/pool/stats", func(c *gin.Context) {
		agents, totalMsgs, totalBytes := websocket.OfflineStats()
		c.JSON(http.StatusOK, gin.H{
			"active_connections": websocket.PoolCount(),
			"offline_agents":     agents,
			"offline_msgs":       totalMsgs,
			"offline_bytes":      totalBytes,
		})
	})

	// --- NEW: health check for quick probes ---
	r.GET("/healthz", func(c *gin.Context) {
		agents, totalMsgs, totalBytes := websocket.OfflineStats()
		c.JSON(http.StatusOK, gin.H{
			"ok":                 true,
			"active_connections": websocket.PoolCount(),
			"offline_agents":     agents,
			"offline_msgs":       totalMsgs,
			"offline_bytes":      totalBytes,
		})
	})

	// Prometheus metrics
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// --- Optional: pprof (protect in prod) ---
	pp := gin.New()
	pp.GET("/debug/pprof/", gin.WrapF(pprof.Index))
	pp.GET("/debug/pprof/cmdline", gin.WrapF(pprof.Cmdline))
	pp.GET("/debug/pprof/profile", gin.WrapF(pprof.Profile))
	pp.GET("/debug/pprof/symbol", gin.WrapF(pprof.Symbol))
	pp.GET("/debug/pprof/trace", gin.WrapF(pprof.Trace))
	r.Any("/debug/pprof/*any", func(c *gin.Context) { pp.HandleContext(c) })
}
