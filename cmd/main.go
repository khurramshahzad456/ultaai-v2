package main

import (
	"log"
	"ultahost-ai-gateway/internal/config"
	"ultahost-ai-gateway/internal/server"
)

func main() {
	// Load environment variables
	config.LoadConfig()

	// Initialize server
	s := server.NewServer()

	// Register all routes
	server.RegisterRoutes(s.Engine)

	log.Printf(" Server starting on port %s...\n", config.AppConfig.Port)
	if err := s.Start(); err != nil {
		log.Fatalf(" Server failed to start: %v", err)
	}
}
