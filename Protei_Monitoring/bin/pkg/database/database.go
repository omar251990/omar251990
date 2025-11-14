package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// DB wraps database connection and operations
type DB struct {
	conn   *sql.DB
	config *Config
}

// Config holds database configuration
type Config struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
	SSLMode  string
	MaxConns int
	MaxIdle  int
}

// New creates a new database connection
func New(config *Config) (*DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode)

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	conn.SetMaxOpenConns(config.MaxConns)
	conn.SetMaxIdleConns(config.MaxIdle)
	conn.SetConnMaxLifetime(time.Hour)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{
		conn:   conn,
		config: config,
	}

	// Run Liquibase migrations
	if err := db.RunMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// RunMigrations executes Liquibase-style migrations
func (db *DB) RunMigrations() error {
	// Create DATABASECHANGELOG table if not exists
	createChangeLogTable := `
	CREATE TABLE IF NOT EXISTS databasechangelog (
		id VARCHAR(255) NOT NULL,
		author VARCHAR(255) NOT NULL,
		filename VARCHAR(255) NOT NULL,
		dateexecuted TIMESTAMP NOT NULL,
		orderexecuted INTEGER NOT NULL,
		exectype VARCHAR(10) NOT NULL,
		md5sum VARCHAR(35),
		description VARCHAR(255),
		comments VARCHAR(255),
		tag VARCHAR(255),
		liquibase VARCHAR(20),
		contexts VARCHAR(255),
		labels VARCHAR(255),
		deployment_id VARCHAR(10)
	);
	CREATE TABLE IF NOT EXISTS databasechangeloglock (
		id INTEGER NOT NULL PRIMARY KEY,
		locked BOOLEAN NOT NULL,
		lockgranted TIMESTAMP,
		lockedby VARCHAR(255)
	);
	INSERT INTO databasechangeloglock (id, locked) VALUES (1, FALSE) ON CONFLICT DO NOTHING;
	`

	if _, err := db.conn.Exec(createChangeLogTable); err != nil {
		return fmt.Errorf("failed to create changelog table: %w", err)
	}

	// Run migrations
	migrations := []Migration{
		{
			ID:          "001-create-sessions-table",
			Author:      "protei",
			Description: "Create subscriber_sessions table",
			SQL: `
			CREATE TABLE IF NOT EXISTS subscriber_sessions (
				id BIGSERIAL PRIMARY KEY,
				tid VARCHAR(255) UNIQUE NOT NULL,
				imsi VARCHAR(15),
				msisdn VARCHAR(15),
				supi VARCHAR(20),
				procedure VARCHAR(50),
				start_time TIMESTAMP NOT NULL,
				end_time TIMESTAMP,
				duration_ms INTEGER,
				result VARCHAR(20),
				plmn VARCHAR(10),
				cell_id VARCHAR(50),
				apn VARCHAR(100),
				dnn VARCHAR(100),
				message_count INTEGER DEFAULT 0,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				INDEX idx_imsi (imsi),
				INDEX idx_msisdn (msisdn),
				INDEX idx_start_time (start_time),
				INDEX idx_plmn (plmn)
			);
			`,
		},
		{
			ID:          "002-create-transactions-table",
			Author:      "protei",
			Description: "Create transactions table",
			SQL: `
			CREATE TABLE IF NOT EXISTS transactions (
				id BIGSERIAL PRIMARY KEY,
				session_id BIGINT REFERENCES subscriber_sessions(id),
				protocol VARCHAR(50) NOT NULL,
				message_type VARCHAR(100),
				message_name VARCHAR(100),
				timestamp TIMESTAMP NOT NULL,
				direction VARCHAR(20),
				result VARCHAR(20),
				cause_code INTEGER,
				source_ip VARCHAR(45),
				dest_ip VARCHAR(45),
				details JSONB,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				INDEX idx_protocol (protocol),
				INDEX idx_timestamp (timestamp),
				INDEX idx_session_id (session_id)
			);
			`,
		},
		{
			ID:          "003-create-kpi-counters-table",
			Author:      "protei",
			Description: "Create KPI counters table",
			SQL: `
			CREATE TABLE IF NOT EXISTS kpi_counters (
				id BIGSERIAL PRIMARY KEY,
				time_bucket TIMESTAMP NOT NULL,
				procedure VARCHAR(50) NOT NULL,
				plmn VARCHAR(10),
				cell_id VARCHAR(50),
				total_count BIGINT DEFAULT 0,
				success_count BIGINT DEFAULT 0,
				failure_count BIGINT DEFAULT 0,
				timeout_count BIGINT DEFAULT 0,
				avg_latency_ms INTEGER,
				p95_latency_ms INTEGER,
				p99_latency_ms INTEGER,
				cause_codes JSONB,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				UNIQUE (time_bucket, procedure, plmn, cell_id)
			);
			CREATE INDEX IF NOT EXISTS idx_kpi_time_bucket ON kpi_counters(time_bucket);
			CREATE INDEX IF NOT EXISTS idx_kpi_procedure ON kpi_counters(procedure);
			`,
		},
		{
			ID:          "004-create-topology-table",
			Author:      "protei",
			Description: "Create topology mapping table",
			SQL: `
			CREATE TABLE IF NOT EXISTS topology (
				id SERIAL PRIMARY KEY,
				cell_id VARCHAR(50) UNIQUE NOT NULL,
				lac_tac VARCHAR(20),
				plmn VARCHAR(10),
				site_name VARCHAR(100),
				region VARCHAR(100),
				city VARCHAR(100),
				latitude DECIMAL(10, 8),
				longitude DECIMAL(11, 8),
				vendor VARCHAR(50),
				technology VARCHAR(20),
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS idx_topology_plmn ON topology(plmn);
			`,
		},
		{
			ID:          "005-create-dictionaries-table",
			Author:      "protei",
			Description: "Create dictionaries table",
			SQL: `
			CREATE TABLE IF NOT EXISTS dictionaries (
				id SERIAL PRIMARY KEY,
				type VARCHAR(50) NOT NULL,
				key VARCHAR(100) NOT NULL,
				value TEXT NOT NULL,
				description TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				UNIQUE (type, key)
			);
			INSERT INTO dictionaries (type, key, value, description) VALUES
				('mcc_mnc', '310410', 'AT&T', 'AT&T Mobility USA'),
				('mcc_mnc', '310260', 'T-Mobile', 'T-Mobile USA'),
				('mcc_mnc', '311480', 'Verizon', 'Verizon Wireless'),
				('country', '310', 'United States', 'USA'),
				('country', '234', 'United Kingdom', 'UK'),
				('country', '262', 'Germany', 'Germany')
			ON CONFLICT DO NOTHING;
			`,
		},
		{
			ID:          "006-create-alarms-table",
			Author:      "protei",
			Description: "Create alarms and alarm history",
			SQL: `
			CREATE TABLE IF NOT EXISTS alarms (
				id BIGSERIAL PRIMARY KEY,
				severity VARCHAR(20) NOT NULL,
				category VARCHAR(50) NOT NULL,
				procedure VARCHAR(50),
				description TEXT NOT NULL,
				details JSONB,
				threshold_value DECIMAL(10, 2),
				current_value DECIMAL(10, 2),
				acknowledged BOOLEAN DEFAULT FALSE,
				acknowledged_by VARCHAR(100),
				acknowledged_at TIMESTAMP,
				cleared BOOLEAN DEFAULT FALSE,
				cleared_at TIMESTAMP,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				INDEX idx_severity (severity),
				INDEX idx_acknowledged (acknowledged),
				INDEX idx_cleared (cleared)
			);
			`,
		},
		{
			ID:          "007-create-audit-log-table",
			Author:      "protei",
			Description: "Create audit log table",
			SQL: `
			CREATE TABLE IF NOT EXISTS audit_log (
				id BIGSERIAL PRIMARY KEY,
				timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				user_name VARCHAR(100) NOT NULL,
				user_ip VARCHAR(45),
				action VARCHAR(100) NOT NULL,
				category VARCHAR(50) NOT NULL,
				details JSONB,
				old_value TEXT,
				new_value TEXT,
				result VARCHAR(20),
				INDEX idx_timestamp (timestamp),
				INDEX idx_user_name (user_name),
				INDEX idx_action (action)
			);
			`,
		},
		{
			ID:          "008-create-user-accounts-table",
			Author:      "protei",
			Description: "Create user accounts table",
			SQL: `
			CREATE TABLE IF NOT EXISTS user_accounts (
				id SERIAL PRIMARY KEY,
				username VARCHAR(100) UNIQUE NOT NULL,
				password_hash VARCHAR(255) NOT NULL,
				full_name VARCHAR(200),
				email VARCHAR(200),
				role VARCHAR(50) NOT NULL,
				permissions JSONB,
				enabled BOOLEAN DEFAULT TRUE,
				last_login TIMESTAMP,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);
			INSERT INTO user_accounts (username, password_hash, full_name, role, permissions) VALUES
				('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Administrator', 'admin', '{"all": true}')
			ON CONFLICT DO NOTHING;
			`,
		},
		{
			ID:          "009-create-vip-subscribers-table",
			Author:      "protei",
			Description: "Create VIP subscribers table",
			SQL: `
			CREATE TABLE IF NOT EXISTS vip_subscribers (
				id SERIAL PRIMARY KEY,
				imsi VARCHAR(15) UNIQUE,
				msisdn VARCHAR(15) UNIQUE,
				name VARCHAR(200),
				priority INTEGER DEFAULT 1,
				alert_on_failure BOOLEAN DEFAULT TRUE,
				notification_emails TEXT[],
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);
			`,
		},
	}

	for _, migration := range migrations {
		if err := db.executeMigration(migration); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", migration.ID, err)
		}
	}

	return nil
}

// Migration represents a database migration
type Migration struct {
	ID          string
	Author      string
	Description string
	SQL         string
}

// executeMigration executes a single migration
func (db *DB) executeMigration(migration Migration) error {
	// Check if migration already executed
	var count int
	err := db.conn.QueryRow(
		"SELECT COUNT(*) FROM databasechangelog WHERE id = $1",
		migration.ID,
	).Scan(&count)

	if err != nil {
		return err
	}

	if count > 0 {
		// Already executed
		return nil
	}

	// Execute migration
	if _, err := db.conn.Exec(migration.SQL); err != nil {
		return err
	}

	// Record in changelog
	_, err = db.conn.Exec(`
		INSERT INTO databasechangelog (id, author, filename, dateexecuted, orderexecuted, exectype, description)
		VALUES ($1, $2, 'init', $3, (SELECT COALESCE(MAX(orderexecuted), 0) + 1 FROM databasechangelog), 'EXECUTED', $4)
	`, migration.ID, migration.Author, time.Now(), migration.Description)

	return err
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// GetConnection returns the underlying SQL connection
func (db *DB) GetConnection() *sql.DB {
	return db.conn
}
