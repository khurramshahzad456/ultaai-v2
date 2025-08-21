package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"ultahost-ai-gateway/internal/config"
	"ultahost-ai-gateway/internal/pkg/db"
	"ultahost-ai-gateway/internal/pkg/models"

	"github.com/gin-gonic/gin"
)

type Server struct {
	Engine *gin.Engine
	http   *http.Server
}

func NewServer() *Server {
	r := gin.Default()

	addr := fmt.Sprintf(":%s", config.AppConfig.Port)

	s := &Server{
		Engine: r,
		http: &http.Server{
			Addr:    addr,
			Handler: r,
		},
	}

	return s
}

func (s *Server) Start() error {
	log.Printf("🚀 Starting HTTP server on %s", s.http.Addr)
	return s.http.ListenAndServe()
}

// ✅ This is the missing method
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("🛑 Shutting down HTTP server...")
	return s.http.Shutdown(ctx)
}

func InitDatabase() error {
	connStr := "postgres://ultahost:ultahost@localhost:5432/ultaai_database?sslmode=disable"

	if err := db.Connect(connStr); err != nil {
		return err
	}

	// Run migrations
	if err := models.MigrateCustomerTables(); err != nil {
		return err
	}
	if err := models.MigrateAgentTables(); err != nil {
		return err
	}
	if err := models.MigrateTaskTables(); err != nil {
		return err
	}
	if err := models.MigrateAITables(); err != nil {
		return err
	}
	if err := models.MigrateMonitoringTables(); err != nil {
		return err
	}
	if err := models.MigrateSecurityTables(); err != nil {
		return err
	}

	fmt.Println("✅ All migrations applied successfully")
	return nil
}
