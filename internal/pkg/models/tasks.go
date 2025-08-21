package models

import "time"

// Task definitions (templates)
type TaskTemplate struct {
	ID          int       `db:"id"`
	Name        string    `db:"name"`         // e.g., "install_wordpress"
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// Task execution tracking
type Task struct {
	ID          int       `db:"id"`
	TaskTemplateID int    `db:"task_template_id"` // references task_templates.id
	AgentID     int       `db:"agent_id"`          // references agents.id
	TaskID      string    `db:"task_id"`           // server-generated UUID (matches websocket)
	Status      string    `db:"status"`            // pending, running, success, failed
	ExitCode    int       `db:"exit_code"`
	Stdout      string    `db:"stdout"`
	Stderr      string    `db:"stderr"`
	StartedAt   time.Time `db:"started_at"`
	FinishedAt  time.Time `db:"finished_at"`
	DurationSec int64     `db:"duration_sec"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// Rollback system checkpoints
type SystemCheckpoint struct {
	ID        int       `db:"id"`
	TaskID    int       `db:"task_id"` // references tasks.id
	Snapshot  string    `db:"snapshot"` // e.g., file path or DB snapshot reference
	CreatedAt time.Time `db:"created_at"`
}
