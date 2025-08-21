package models

import "time"

// System metrics
type PerformanceMetric struct {
	ID         int       `db:"id"`
	AgentID    int       `db:"agent_id"`    // optional, can be NULL for system-wide
	MetricName string    `db:"metric_name"` // e.g., cpu_usage, memory_free
	Value      float64   `db:"value"`
	Timestamp  time.Time `db:"timestamp"`
}

// Alert rules
type AlertRule struct {
	ID            int       `db:"id"`
	Name          string    `db:"name"`
	MetricName    string    `db:"metric_name"`
	Threshold     float64   `db:"threshold"`
	Operator      string    `db:"operator"` // e.g., ">", "<="
	Enabled       bool      `db:"enabled"`
	CreatedAt     time.Time `db:"created_at"`
	LastTriggered time.Time `db:"last_triggered"`
}

// Active alerts / incidents
type Alert struct {
	ID          int        `db:"id"`
	RuleID      int        `db:"rule_id"`  // references alert_rules.id
	AgentID     int        `db:"agent_id"` // optional, NULL for system-wide
	TriggeredAt time.Time  `db:"triggered_at"`
	ResolvedAt  *time.Time `db:"resolved_at"` // NULL if still active
	Severity    string     `db:"severity"`    // e.g., info, warning, critical
	Message     string     `db:"message"`
}
