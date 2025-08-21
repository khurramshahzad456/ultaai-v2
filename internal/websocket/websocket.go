package websocket

import (
	"fmt"
	"time"
	"ultahost-ai-gateway/internal/utils"
)

func SendMessage(vpsId string, payload []byte) error {
	CN := "Agent_" + vpsId
	keyInfo, exist := utils.GetAgentKeys(CN)
	if !exist {
		return fmt.Errorf("no keys")
	}

	// Try live connection first
	if a, ok := PoolGet(keyInfo.IdentityToken); ok {
		select {
		case a.Send <- payload:
			metricsEnqueued(1)
			return nil
		default:
			// Backpressure policy: disconnect or drop
			metricsDropped(1)
			close(a.Send)
			_ = a.Conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			return fmt.Errorf("agent disconnected: backpressure")
		}
	}

	// Offline: buffer
	dropped := OfflineEnqueue(keyInfo.IdentityToken, payload)
	if dropped > 0 {
		metricsDropped(dropped)
	}
	metricsOfflineBuffered(1)
	return nil
}
