package models

import "time"

// Audit trail
type AuditLog struct {
	ID        int       `db:"id"`
	UserID    int       `db:"user_id"`
	Action    string    `db:"action"`
	Entity    string    `db:"entity"`    // e.g., agent, vps, task
	EntityID  int       `db:"entity_id"` // related record id
	Timestamp time.Time `db:"timestamp"`
	Details   string    `db:"details"` // optional JSON or message
}

// Security events (threats, anomalies)
type SecurityEvent struct {
	ID          int       `db:"id"`
	EventType   string    `db:"event_type"`  // e.g., login_failure, unauthorized_access
	Severity    string    `db:"severity"`    // low, medium, high, critical
	Description string    `db:"description"` // details about event
	AgentID     int       `db:"agent_id"`    // optional
	Timestamp   time.Time `db:"timestamp"`
}

// Secure installation tokens
type InstallationToken struct {
	ID        int       `db:"id"`
	Token     string    `db:"token"` // securely generated string
	UserID    int       `db:"user_id"`
	VPSID     int       `db:"vps_id"`
	ExpiresAt time.Time `db:"expires_at"` // token validity
	Used      bool      `db:"used"`
}

// API access keys
type APIKey struct {
	ID        int        `db:"id"`
	Key       string     `db:"key"` // base64 or UUID
	UserID    int        `db:"user_id"`
	CreatedAt time.Time  `db:"created_at"`
	ExpiresAt *time.Time `db:"expires_at"` // optional
	Revoked   bool       `db:"revoked"`
}
