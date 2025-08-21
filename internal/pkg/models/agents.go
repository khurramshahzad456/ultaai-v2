package models

import "time"

// Agents registry
type Agent struct {
	ID        int       `db:"id"`
	VPSID     int       `db:"vps_id"` // link to vps_instances.id
	Name      string    `db:"name"`   // CommonName from cert
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Agent security keys
type AgentKey struct {
	ID                int       `db:"id"`
	AgentID           int       `db:"agent_id"`       // references agents.id
	IdentityToken     string    `db:"identity_token"` // UNIQUE
	SignatureSecret   string    `db:"signature_secret"`
	FingerprintSHA256 string    `db:"fingerprint_sha256"` // UNIQUE
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"` // in case of rotation
}

// Agent certificates (PKI)
type AgentCertificate struct {
	ID          int       `db:"id"`
	AgentID     int       `db:"agent_id"`    // references agents.id
	Certificate string    `db:"certificate"` // PEM encoded
	PrivateKey  string    `db:"private_key"` // PEM encoded
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"` // if certificate rotates
}

// Agent heartbeat tracking
type AgentHeartbeat struct {
	ID        int       `db:"id"`
	AgentID   int       `db:"agent_id"`  // references agents.id
	Counter   uint64    `db:"counter"`   // heartbeat counter from agent
	LastSeen  time.Time `db:"last_seen"` // timestamp of last heartbeat
	Valid     bool      `db:"valid"`     // optional, if you want to mark invalid heartbeats
	CreatedAt time.Time `db:"created_at"`
}
