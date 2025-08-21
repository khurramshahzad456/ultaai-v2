package server

import "ultahost-ai-gateway/internal/websocket"

// Iterates pool and closes Send to trigger writePump close frames.
func CloseAllAgentConnections() {
	websocket.PoolRange(func(_ string, a *websocket.AgentConn) bool {
		select {
		case <-a.Closed():
			// already closed
		default:
			close(a.Send)
		}
		return true
	})
}
