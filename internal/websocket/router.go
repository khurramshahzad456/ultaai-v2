package websocket

import (
	"fmt"
	"time"
)

func RouteToIdentity(identityToken string, payload []byte) error {
	if a, ok := PoolGet(identityToken); ok {
		select {
		case a.Send <- payload:
			metricsEnqueued(1)
			return nil
		default:
			metricsDropped(1)
			close(a.Send)
			_ = a.Conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			return fmt.Errorf("agent disconnected: backpressure")
		}
	}
	dropped := OfflineEnqueue(identityToken, payload)
	if dropped > 0 {
		metricsDropped(dropped)
	}
	metricsOfflineBuffered(1)
	return nil
}
