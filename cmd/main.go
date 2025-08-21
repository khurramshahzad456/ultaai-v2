package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ultahost-ai-gateway/internal/config"
	"ultahost-ai-gateway/internal/server"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadConfig()
	if err := server.InitDatabase(); err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}

	// Primary HTTP server (your existing server.NewServer())
	s := server.NewServer()
	server.RegisterRoutes(s.Engine)

	// Start TLS WS server (separate port)
	wsSrv := startWSTLSServer()

	// Run primary
	go func() {
		log.Printf("HTTP server starting on port %s ...", config.AppConfig.Port)
		if err := s.Start(); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Signals
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)

	<-sigch
	log.Println("Shutting down...")

	// Graceful shutdown both servers
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Printf("HTTP server graceful shutdown error: %v", err)
	}
	shutdownWSTLSServer(wsSrv, ctx)

	// Close all agent connections
	gracefulCloseAllAgents()

	log.Println("Bye.")
}

func startWSTLSServer() *server.WSTLSServer {
	// Create a gin router dedicated to WS
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/agent/connect", server.WSHandle())

	// Load CA and TLS settings (inside helper)
	return server.StartWSTLSServer(":8443", r)
}

func shutdownWSTLSServer(srv *server.WSTLSServer, ctx context.Context) {
	if srv != nil {
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("WS TLS server shutdown error: %v", err)
		}
	}
}

// Close all agent connections cleanly (signals writePump to exit)
func gracefulCloseAllAgents() {
	server.CloseAllAgentConnections()
}
