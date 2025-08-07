package main

import (
	"log"

	"ultahost-ai-assistant/internal/config"
	"ultahost-ai-assistant/internal/server"
)

func main() {
	// Load environment variables
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	s := server.NewServer()
	if err := s.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
