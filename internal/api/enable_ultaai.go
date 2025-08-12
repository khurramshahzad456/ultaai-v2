package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
	"ultahost-ai-gateway/internal/utils"

	"github.com/gin-gonic/gin"
)

type EnableUltaAIRequest struct {
	UserID uint `json:"user_id" binding:"required"`
	VPSID  uint `json:"vps_id" binding:"required"`
}

// Generate secure random token
func generateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func HandleEnableUltaAI(c *gin.Context) {
	var req EnableUltaAIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusBadRequest, "Invalid request")
		return
	}

	// Generate a unique one-time token
	token, err := generateRandomToken(16)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Store token in memory for 15 minutes
	utils.SaveInstallToken(token, req.UserID, req.VPSID, 15*time.Minute)

	// Return only the curl command with .sh file
	curlCmd := fmt.Sprintf(
		`curl -s https://193.109.193.72/install.sh | bash -s -- --token=%s`,
		token,
	)
	c.String(http.StatusOK, curlCmd)
}
