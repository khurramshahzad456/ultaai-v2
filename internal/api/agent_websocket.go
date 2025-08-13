package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Example: map VPSID (or agent ID) to signatureSecret; replace with your DB/store
var agentSecrets = map[string]string{
	"agent1": "your_signature_secret_here_in_hex_or_raw",
}

type HeartbeatMessage struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

func verifySignature(message, signature, signatureSecret string) bool {
	mac := hmac.New(sha256.New, []byte(signatureSecret))
	mac.Write([]byte(message))
	expectedMAC := mac.Sum(nil)
	expectedSig := hex.EncodeToString(expectedMAC)
	return hmac.Equal([]byte(expectedSig), []byte(signature))
}

func HandleAgentWebSocket(c *gin.Context) {
	// You must identify agent from context or params, for demo using fixed agent ID
	agentID := "agent1"
	secret, ok := agentSecrets[agentID]
	if !ok {
		c.String(http.StatusUnauthorized, "Unknown agent")
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("WebSocket read error:", err)
			break
		}

		var hb HeartbeatMessage
		err = json.Unmarshal(msgBytes, &hb)
		if err != nil {
			fmt.Println("Invalid JSON:", err)
			continue
		}

		if verifySignature(hb.Message, hb.Signature, secret) {
			fmt.Println("Valid signature from agent:", hb.Message)
			conn.WriteMessage(websocket.TextMessage, []byte("ACK: valid heartbeat"))
		} else {
			fmt.Println("Invalid signature, rejecting message")
			conn.WriteMessage(websocket.TextMessage, []byte("NACK: invalid signature"))
			// Optionally: close connection or handle as needed
		}
	}
}
