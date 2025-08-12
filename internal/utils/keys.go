package utils

import "sync"

type AgentKeys struct {
	IdentityToken   string
	SignatureSecret string
	Certificate     string
	PrivateKey      string
}

var (
	agentKeysStore   = make(map[uint]AgentKeys)
	agentKeysStoreMu sync.Mutex
)

func SaveAgentKeys(vpsID uint, keys AgentKeys) {
	agentKeysStoreMu.Lock()
	defer agentKeysStoreMu.Unlock()
	agentKeysStore[vpsID] = keys
}

func GetAgentKeys(vpsID uint) (AgentKeys, bool) {
	agentKeysStoreMu.Lock()
	defer agentKeysStoreMu.Unlock()
	keys, exists := agentKeysStore[vpsID]
	return keys, exists
}
