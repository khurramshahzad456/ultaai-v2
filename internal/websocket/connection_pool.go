package websocket

import "sync"

// identityToken -> *AgentConn
var ConnectedVPS sync.Map

// PoolPut replaces any existing connection for the same identity (auto-reconnect).
func PoolPut(identityToken string, a *AgentConn) {
	// If an old connection exists, close it first.
	if oldV, ok := ConnectedVPS.Load(identityToken); ok {
		old := oldV.(*AgentConn)
		// Close its Send channel; writePump will send a Close frame & exit.
		select {
		case <-old.closed:
			// already closed
		default:
			close(old.Send)
		}
	}
	ConnectedVPS.Store(identityToken, a)
}

func PoolGet(identityToken string) (*AgentConn, bool) {
	if v, ok := ConnectedVPS.Load(identityToken); ok {
		return v.(*AgentConn), true
	}
	return nil, false
}

func PoolDelete(identityToken string) {
	ConnectedVPS.Delete(identityToken)
}

func PoolCount() int {
	n := 0
	ConnectedVPS.Range(func(_, _ any) bool {
		n++
		return true
	})
	return n
}

func PoolRange(fn func(identity string, a *AgentConn) bool) {
	ConnectedVPS.Range(func(k, v any) bool {
		return fn(k.(string), v.(*AgentConn))
	})
}
