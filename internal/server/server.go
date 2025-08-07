package server

import (
	"fmt"

	"ultahost-ai-assistant/internal/config"
	"ultahost-ai-assistant/internal/handler"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	Engine *gin.Engine
	Port   int
}

func NewServer() *Server {
	router := gin.Default()

	router.Use(cors.Default())

	return &Server{
		Engine: router,
		Port:   config.Get().Port, // Loaded from .env
	}
}

func (s *Server) RegisterRoutes() {
	h := handler.NewHandler()
	// /sh := handler.NewStepHandler()
	s.Engine.POST("/run-command", h.RunCommand)
	// s.Engine.POST("/run-steps", h.RunSteps)

}

func (s *Server) Start() error {
	s.RegisterRoutes()
	address := fmt.Sprintf(":%d", s.Port)
	return s.Engine.Run(address)
}
