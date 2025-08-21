package websocket

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"ultahost-ai-gateway/internal/utils"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:    1024, // small to reduce per-conn mem
	WriteBufferSize:   1024,
	EnableCompression: false,
	CheckOrigin:       func(r *http.Request) bool { return true },
}

type AgentConn struct {
	Conn                 *ws.Conn
	IdentityToken        string
	LastHeartbeatCounter uint64
	LastSeen             time.Time

	Send   chan []byte   // bounded outbound queue
	closed chan struct{} // closed when writer exits

	mu sync.Mutex
}

const (
	readTimeout  = 60 * time.Second
	writeTimeout = 10 * time.Second
	pongWait     = 70 * time.Second
	pingPeriod   = 30 * time.Second
)

func HandleAgentWebSocket(c *gin.Context) {
	if c.Request.TLS == nil || len(c.Request.TLS.PeerCertificates) == 0 {
		c.String(http.StatusUnauthorized, "Client certificate required")
		return
	}
	clientCert := c.Request.TLS.PeerCertificates[0]
	cn := clientCert.Subject.CommonName

	// Upgrade to WS
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("upgrade error: %v", err)
		return
	}

	// Lookup issued keys
	keyInfo, exist := utils.GetAgentKeys(cn)
	if !exist {
		_ = conn.WriteMessage(ws.TextMessage, []byte("agent not enrolled"))
		_ = conn.Close()
		return
	}

	// Fingerprint check
	presentedFP := sha256.Sum256(clientCert.Raw)
	if hex.EncodeToString(presentedFP[:]) != keyInfo.FingerprintSHA256 {
		_ = conn.WriteMessage(ws.TextMessage, []byte("certificate fingerprint mismatch"))
		_ = conn.Close()
		return
	}

	// Build connection
	agentConn := &AgentConn{
		Conn:          conn,
		IdentityToken: keyInfo.IdentityToken,
		LastSeen:      time.Now(),
		Send:          make(chan []byte, 128), // lower to save RAM; tune per traffic
		closed:        make(chan struct{}),
	}

	// Auto-reconnect: replace any existing for this identity
	PoolPut(keyInfo.IdentityToken, agentConn)
	metricsIncActive()

	// WS read settings
	conn.SetReadLimit(1 << 20) // 1MiB
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		agentConn.mu.Lock()
		agentConn.LastSeen = time.Now()
		agentConn.mu.Unlock()
		return nil
	})

	// Optional: custom close handler to mark offline quickly
	conn.SetCloseHandler(func(code int, text string) error {
		return nil // default behavior ok; read loop's defer handles cleanup
	})

	// Start loops
	go writePump(agentConn)
	go handleAgentReadLoop(agentConn, keyInfo)

	// After connection established, flush any offline-buffered messages
	go func(identity string, a *AgentConn) {
		queued := OfflineDrain(identity)
		if len(queued) > 0 {
			flushed := 0
			for _, m := range queued {
				select {
				case a.Send <- m:
					flushed++
				default:
					// backpressure: stop flushing to avoid OOM
					metricsDropped(1)
					return
				}
			}
			metricsOfflineFlushed(flushed)
		}
	}(keyInfo.IdentityToken, agentConn)

	log.Printf("Agent connected: CN=%s, IdentityToken=%s", cn, keyInfo.IdentityToken)
}

func writePump(a *AgentConn) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = a.Conn.Close()
		close(a.closed)
		metricsDecActive()
	}()

	for {
		select {
		case msg, ok := <-a.Send:
			a.Conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if !ok {
				_ = a.Conn.WriteMessage(ws.CloseMessage, []byte{})
				return
			}
			w, err := a.Conn.NextWriter(ws.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(msg); err != nil {
				_ = w.Close()
				return
			}

			// batch a small burst
			drain := 0
			for drain < 16 {
				select {
				case more := <-a.Send:
					if _, err := w.Write([]byte{'\n'}); err != nil {
						_ = w.Close()
						return
					}
					if _, err := w.Write(more); err != nil {
						_ = w.Close()
						return
					}
					drain++
				default:
					drain = 16
				}
			}
			if err := w.Close(); err != nil {
				return
			}
			metricsEnqueued(1 + drain)

		case <-ticker.C:
			a.Conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := a.Conn.WriteMessage(ws.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func handleAgentReadLoop(a *AgentConn, keyInfo utils.AgentKeys) {
	defer func() {
		PoolDelete(a.IdentityToken)
		failPendingForAgent(keyInfo.IdentityToken, "connection closed")
		// Stop writer
		select {
		case <-a.closed:
			// already closed
		default:
			close(a.Send)
		}
		_ = a.Conn.Close()
	}()

	for {
		mt, msg, err := a.Conn.ReadMessage()
		if err != nil {
			log.Printf("read error (%s): %v", a.IdentityToken, err)
			return
		}
		a.mu.Lock()
		a.LastSeen = time.Now()
		a.mu.Unlock()

		if mt == ws.TextMessage {
			var generic map[string]interface{}
			if err := json.Unmarshal(msg, &generic); err == nil {
				if t, ok := generic["type"].(string); ok {
					switch t {
					case "heartbeat":
						if err := verifyHeartbeat(msg, keyInfo); err != nil {
							log.Printf("heartbeat verification failed (%s): %v", a.IdentityToken, err)
							return
						}
						continue
					case "task_result":
						var tr TaskResult
						if err := json.Unmarshal(msg, &tr); err != nil {
							log.Printf("invalid task_result from %s: %v", keyInfo.IdentityToken, err)
							continue
						}
						if resolved := resolvePending(tr.TaskID, tr); !resolved {
							log.Printf("unknown task_id %s (agent %s)", tr.TaskID, keyInfo.IdentityToken)
						}
						continue
					}
				}
			}
			log.Printf("msg from %s: %s", a.IdentityToken, string(msg))
		}
	}
}

// verifyHeartbeat unchanged (uses PoolGet)
func verifyHeartbeat(msg []byte, keyInfo utils.AgentKeys) error {
	type hb struct {
		Type      string `json:"type"`
		Version   int    `json:"version"`
		AgentID   string `json:"agent_id"`
		Counter   uint64 `json:"counter"`
		Nonce     string `json:"nonce"`
		Timestamp string `json:"timestamp"`
		Signature string `json:"signature"`
	}
	var h hb
	if err := json.Unmarshal(msg, &h); err != nil {
		return err
	}

	maxSkew := 5 * time.Minute
	ht, err := time.Parse(time.RFC3339Nano, h.Timestamp)
	if err != nil {
		ht, err = time.Parse(time.RFC3339, h.Timestamp)
		if err != nil {
			return fmt.Errorf("invalid heartbeat timestamp: %w", err)
		}
	}
	delta := time.Since(ht.UTC())
	if delta < 0 {
		delta = -delta
	}
	if delta > maxSkew {
		return errors.New("heartbeat timestamp outside allowed skew")
	}

	canon := fmt.Sprintf("%s|%s|%d|%s|%s", h.Version, h.AgentID, h.Counter, h.Nonce, h.Timestamp)
	expected := utils.HMACSHA256Base64([]byte(keyInfo.SignatureSecret), canon)
	if expected != h.Signature {
		return errors.New("invalid signature")
	}

	if aConn, ok := PoolGet(keyInfo.IdentityToken); ok {
		if h.Counter <= aConn.LastHeartbeatCounter {
			return errors.New("replay or old counter")
		}
		aConn.LastHeartbeatCounter = h.Counter
	}
	return nil
}

func (a *AgentConn) Closed() <-chan struct{} { return a.closed }
