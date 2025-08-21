package models

import "time"

// User chat session
type ChatSession struct {
	ID        int       `db:"id"`
	UserID    int       `db:"user_id"` // references customer_users.id
	VPSID     int       `db:"vps_id"`  // references vps_instances.id
	StartedAt time.Time `db:"started_at"`
	EndedAt   time.Time `db:"ended_at"`
}

// Individual messages in a session
type ChatMessage struct {
	ID        int       `db:"id"`
	SessionID int       `db:"session_id"` // references chat_sessions.id
	Sender    string    `db:"sender"`     // user or AI
	Message   string    `db:"message"`
	CreatedAt time.Time `db:"created_at"`
}

// AI detected intents per message
type AIIntent struct {
	ID         int       `db:"id"`
	MessageID  int       `db:"message_id"` // references chat_messages.id
	Intent     string    `db:"intent"`
	Confidence float64   `db:"confidence"`
	CreatedAt  time.Time `db:"created_at"`
}
