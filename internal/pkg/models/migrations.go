package models

import (
	"fmt"
	"ultahost-ai-gateway/internal/pkg/db"
)

func MigrateCustomerTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS customers (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS customer_users (
			id SERIAL PRIMARY KEY,
			customer_id INT REFERENCES customers(id) ON DELETE CASCADE,
			email TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS vps_instances (
			id SERIAL PRIMARY KEY,
			customer_id INT REFERENCES customers(id) ON DELETE CASCADE,
			hostname TEXT NOT NULL,
			ip_address INET NOT NULL,
			server_details JSONB,
			os TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,

		// Foreign keys

		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_customer_users_email ON customer_users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_vps_hostname ON vps_instances(hostname)`,
		`CREATE INDEX IF NOT EXISTS idx_vps_ip ON vps_instances(ip_address)`,
	}

	for _, q := range queries {
		if _, err := db.DB.Exec(q); err != nil {

			return err
		}
	}
	return nil
}

func MigrateAgentTables() error {

	queries := []string{
		`CREATE TABLE IF NOT EXISTS agents (
			id SERIAL PRIMARY KEY,
			vps_id INT REFERENCES vps_instances(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS agent_keys (
			id SERIAL PRIMARY KEY,
			agent_id INT REFERENCES agents(id) ON DELETE CASCADE,
			identity_token TEXT UNIQUE NOT NULL,
			signature_secret TEXT NOT NULL,
			fingerprint_sha256 TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS agent_certificates (
			id SERIAL PRIMARY KEY,
			agent_id INT REFERENCES agents(id) ON DELETE CASCADE,
			certificate TEXT NOT NULL,
			private_key TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS agent_heartbeats (
			id SERIAL PRIMARY KEY,
			agent_id INT REFERENCES agents(id) ON DELETE CASCADE,
			counter BIGINT DEFAULT 0,
			last_seen TIMESTAMP,
			created_at TIMESTAMP DEFAULT NOW()
		)`,

		// Foreign keys

		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_agent_keys_identity ON agent_keys(identity_token)`,
		`CREATE INDEX IF NOT EXISTS idx_heartbeats_agent ON agent_heartbeats(agent_id)`,
	}

	for _, q := range queries {
		if _, err := db.DB.Exec(q); err != nil {
			fmt.Println("  ------ query ------------ ", q)

			return err
		}
	}
	return nil
}

func MigrateTaskTables() error {

	queries := []string{
		`CREATE TABLE IF NOT EXISTS task_templates (
			id SERIAL PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id SERIAL PRIMARY KEY,
			task_template_id INT REFERENCES task_templates(id),
			agent_id INT REFERENCES agents(id),
			task_id TEXT UNIQUE NOT NULL,
			status TEXT DEFAULT 'pending',
			exit_code INT DEFAULT 0,
			stdout TEXT,
			stderr TEXT,
			started_at TIMESTAMP,
			finished_at TIMESTAMP,
			duration_sec BIGINT DEFAULT 0,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS system_checkpoints (
			id SERIAL PRIMARY KEY,
			task_id INT REFERENCES tasks(id),
			snapshot TEXT,
			created_at TIMESTAMP DEFAULT NOW()
		)`,

		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)`,
		`CREATE INDEX IF NOT EXISTS idx_checkpoints_task ON system_checkpoints(task_id)`,
	}

	for _, q := range queries {
		if _, err := db.DB.Exec(q); err != nil {

			return err
		}
	}
	return nil
}

func MigrateAITables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS chat_sessions (
			id SERIAL PRIMARY KEY,
			user_id INT REFERENCES customer_users(id),
			vps_id INT REFERENCES vps_instances(id),
			started_at TIMESTAMP DEFAULT NOW(),
			ended_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS chat_messages (
			id SERIAL PRIMARY KEY,
			session_id INT REFERENCES chat_sessions(id),
			sender TEXT NOT NULL,
			message TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS ai_intents (
			id SERIAL PRIMARY KEY,
			message_id INT REFERENCES chat_messages(id),
			intent TEXT NOT NULL,
			confidence FLOAT DEFAULT 0,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		// Foreign key constraints

		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_chat_sessions_user ON chat_sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_messages_session ON chat_messages(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ai_intents_message ON ai_intents(message_id)`,
	}

	for _, q := range queries {
		if _, err := db.DB.Exec(q); err != nil {
			fmt.Println("  ------ query ------------ ", q)

			return err
		}
	}
	return nil
}

func MigrateMonitoringTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS performance_metrics (
			id SERIAL PRIMARY KEY,
			agent_id INT REFERENCES agents(id) ON DELETE CASCADE,
			vps_id INT REFERENCES vps_instances(id) ON DELETE CASCADE,
			metric_name TEXT NOT NULL,
			value DOUBLE PRECISION NOT NULL,
			timestamp TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS alert_rules (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			metric_name TEXT NOT NULL,
			threshold DOUBLE PRECISION NOT NULL,
			operator TEXT NOT NULL,
			enabled BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT NOW(),
			last_triggered TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS alerts (
			id SERIAL PRIMARY KEY,
			rule_id INT REFERENCES alert_rules(id) ON DELETE CASCADE,
			agent_id INT REFERENCES agents(id) ON DELETE CASCADE,
			vps_id INT REFERENCES vps_instances(id) ON DELETE CASCADE,
			triggered_at TIMESTAMP DEFAULT NOW(),
			resolved_at TIMESTAMP,
			severity TEXT,
			message TEXT
		)`,
		// Foreign keys

		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_metrics_agent ON performance_metrics(agent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_metrics_vps ON performance_metrics(vps_id)`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_agent ON alerts(agent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_vps ON alerts(vps_id)`,
	}

	for _, q := range queries {
		if _, err := db.DB.Exec(q); err != nil {
			fmt.Println("  ------ query ------------ ", q)

			return err
		}
	}
	return nil
}

func MigrateSecurityTables() error {
	queries := []string{
		// Tables
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id SERIAL PRIMARY KEY,
			user_id INT,
			action TEXT NOT NULL,
			entity TEXT,
			entity_id INT,
			timestamp TIMESTAMP DEFAULT NOW(),
			details TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS security_events (
			id SERIAL PRIMARY KEY,
			event_type TEXT NOT NULL,
			severity TEXT NOT NULL,
			description TEXT,
			agent_id INT,
			timestamp TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS installation_tokens (
			id SERIAL PRIMARY KEY,
			token TEXT NOT NULL,
			user_id INT NOT NULL,
			vps_id INT NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			used BOOLEAN DEFAULT FALSE
		)`,
		`CREATE TABLE IF NOT EXISTS api_keys (
			id SERIAL PRIMARY KEY,
			key TEXT NOT NULL,
			user_id INT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			expires_at TIMESTAMP,
			revoked BOOLEAN DEFAULT FALSE
		)`,

		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_logs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_security_events_type ON security_events(event_type)`,
		`CREATE INDEX IF NOT EXISTS idx_tokens_expires ON installation_tokens(expires_at)`,
		`CREATE INDEX IF NOT EXISTS idx_api_keys_user ON api_keys(user_id)`,
	}

	for _, q := range queries {

		if _, err := db.DB.Exec(q); err != nil {
			fmt.Println("  ------ query ------------ ", q)

			return err
		}
	}
	return nil
}
