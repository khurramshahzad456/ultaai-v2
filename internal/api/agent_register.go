package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"ultahost-ai-gateway/internal/utils"

	"github.com/gin-gonic/gin"
)

type AgentRegisterRequest struct {
	InstallToken string `json:"install_token" binding:"required"`
}

type AgentRegisterResponse struct {
	IdentityToken   string `json:"identity_token"`
	SignatureSecret string `json:"signature_secret"`
	Certificate     string `json:"certificate"`
	PrivateKey      string `json:"private_key"`
}

func generateSecret(length int) string {
	bytes := make([]byte, length)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func HandleAgentRegister(c *gin.Context) {
	var req AgentRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	tokenData, ok := utils.ConsumeInstallToken(req.InstallToken)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	identityToken := generateSecret(32)
	signatureSecret := generateSecret(32)
	certificate := base64.StdEncoding.EncodeToString([]byte("fake-cert"))
	privateKey := base64.StdEncoding.EncodeToString([]byte("fake-private-key"))

	utils.SaveAgentKeys(tokenData.VPSID, utils.AgentKeys{
		IdentityToken:   identityToken,
		SignatureSecret: signatureSecret,
		Certificate:     certificate,
		PrivateKey:      privateKey,
	})

	c.JSON(http.StatusOK, AgentRegisterResponse{
		IdentityToken:   identityToken,
		SignatureSecret: signatureSecret,
		Certificate:     certificate,
		PrivateKey:      privateKey,
	})
}
