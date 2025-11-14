--liquibase formatted sql
--changeset protei:1

-- ==============================================================================
-- Protei Monitoring v2.0 - Database Schema
-- ==============================================================================
-- This schema supports:
-- - Multi-generation (2G/3G/4G/5G) protocol monitoring
-- - Subscriber correlation across all identifiers
-- - Transaction tracking (MAP/CAP/INAP/Diameter/HTTP/GTP)
-- - KPI aggregation and historical analysis
-- - Location tracking and topology
-- - Alarm and alert management
-- - Complete audit trail
-- ==============================================================================

-- ------------------------------------------------------------------------------
-- 1. Users and Authentication
-- ------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    email VARCHAR(100),
    role VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'noc_viewer', 'engineer', 'security_auditor')),
    enabled BOOLEAN DEFAULT true,
    ldap_enabled BOOLEAN DEFAULT false,
    ldap_dn VARCHAR(255),
    require_password_change BOOLEAN DEFAULT false,
    password_expiry_date DATE,
    failed_login_attempts INTEGER DEFAULT 0,
    account_locked_until TIMESTAMP,
    last_login TIMESTAMP,
    last_password_change TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    created_by VARCHAR(50),
    CONSTRAINT chk_email CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_enabled ON users(enabled);

COMMENT ON TABLE users IS 'User accounts with RBAC support and LDAP integration';

-- ------------------------------------------------------------------------------
-- 2. User Sessions
-- ------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS user_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    ip_address INET,
    user_agent TEXT,
    login_time TIMESTAMP DEFAULT NOW(),
    last_activity TIMESTAMP DEFAULT NOW(),
    logout_time TIMESTAMP,
    session_active BOOLEAN DEFAULT true,
    CONSTRAINT fk_user_sessions_user FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_user_sessions_token ON user_sessions(session_token);
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_active ON user_sessions(session_active);

-- ------------------------------------------------------------------------------
-- 3. Permissions and Roles
-- ------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    permission_name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    category VARCHAR(50), -- 'protocol', 'network', 'config', 'system'
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role VARCHAR(20) NOT NULL,
    permission_id INTEGER REFERENCES permissions(id) ON DELETE CASCADE,
    granted_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (role, permission_id)
);

-- ------------------------------------------------------------------------------
-- 4. Dictionaries - MCC/MNC, Countries, Operators
-- ------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS countries (
    mcc INTEGER PRIMARY KEY,
    country_code VARCHAR(3) NOT NULL, -- ISO 3166-1 alpha-3
    country_name VARCHAR(100) NOT NULL,
    region VARCHAR(50), -- 'Europe', 'Asia', 'Americas', etc.
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS operators (
    id SERIAL PRIMARY KEY,
    mcc INTEGER NOT NULL,
    mnc INTEGER NOT NULL,
    operator_name VARCHAR(100) NOT NULL,
    operator_type VARCHAR(20), -- 'MNO', 'MVNO'
    technology VARCHAR(50), -- '2G,3G,4G,5G'
    country VARCHAR(100),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(mcc, mnc)
);

CREATE INDEX idx_operators_mcc_mnc ON operators(mcc, mnc);
CREATE INDEX idx_operators_country ON operators(country);

COMMENT ON TABLE operators IS 'Mobile network operators worldwide (MCC/MNC mapping)';

-- ------------------------------------------------------------------------------
-- 5. Network Topology - Cells, Sites, Regions
-- ------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS network_topology (
    id BIGSERIAL PRIMARY KEY,
    cell_id VARCHAR(50),
    lac INTEGER,
    tac INTEGER,
    rac INTEGER,
    sac INTEGER,
    enb_id INTEGER,
    gnb_id INTEGER,
    ecgi VARCHAR(50), -- E-UTRAN Cell Global Identifier
    ncgi VARCHAR(50), -- NR Cell Global Identifier
    site_name VARCHAR(100),
    site_id VARCHAR(50),
    region VARCHAR(100),
    city VARCHAR(100),
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    technology VARCHAR(20), -- '2G', '3G', '4G', '5G'
    vendor VARCHAR(50), -- 'Ericsson', 'Huawei', 'Nokia', etc.
    status VARCHAR(20) DEFAULT 'active', -- 'active', 'inactive', 'maintenance'
    capacity_mbps INTEGER,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_topology_cell_id ON network_topology(cell_id);
CREATE INDEX idx_topology_lac ON network_topology(lac);
CREATE INDEX idx_topology_tac ON network_topology(tac);
CREATE INDEX idx_topology_site_id ON network_topology(site_id);
CREATE INDEX idx_topology_region ON network_topology(region);
CREATE INDEX idx_topology_technology ON network_topology(technology);

COMMENT ON TABLE network_topology IS 'Network topology mapping for location tracking';

-- ------------------------------------------------------------------------------
-- 6. Subscribers - Master Table
-- ------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS subscribers (
    id BIGSERIAL PRIMARY KEY,
    imsi VARCHAR(15) UNIQUE,
    msisdn VARCHAR(15),
    imei VARCHAR(15),
    imei_sv VARCHAR(16),
    home_mcc INTEGER,
    home_mnc INTEGER,
    subscriber_type VARCHAR(20) DEFAULT 'regular', -- 'regular', 'vip', 'test'
    vip_priority INTEGER, -- 1 (highest) to 10 (lowest) for VIP subscribers

    -- Current status
    current_status VARCHAR(20), -- 'attached', 'detached', 'suspended', 'unknown'
    current_cell_id VARCHAR(50),
    current_lac INTEGER,
    current_tac INTEGER,
    current_location VARCHAR(100), -- Human-readable location
    current_rat VARCHAR(10), -- 'GSM', 'UMTS', 'LTE', 'NR'

    -- Statistics
    total_sessions INTEGER DEFAULT 0,
    total_data_mb BIGINT DEFAULT 0,
    total_voice_minutes INTEGER DEFAULT 0,
    total_sms INTEGER DEFAULT 0,
    failed_sessions INTEGER DEFAULT 0,

    -- Timestamps
    first_seen TIMESTAMP DEFAULT NOW(),
    last_seen TIMESTAMP,
    last_attach_time TIMESTAMP,
    last_detach_time TIMESTAMP,

    -- Metadata
    metadata JSONB,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_subscribers_imsi ON subscribers(imsi);
CREATE INDEX idx_subscribers_msisdn ON subscribers(msisdn);
CREATE INDEX idx_subscribers_imei ON subscribers(imei);
CREATE INDEX idx_subscribers_type ON subscribers(subscriber_type);
CREATE INDEX idx_subscribers_vip ON subscribers(vip_priority) WHERE vip_priority IS NOT NULL;
CREATE INDEX idx_subscribers_current_cell ON subscribers(current_cell_id);
CREATE INDEX idx_subscribers_last_seen ON subscribers(last_seen DESC);

COMMENT ON TABLE subscribers IS 'Master subscriber table with current status and statistics';

-- ------------------------------------------------------------------------------
-- 7. Subscriber Location History
-- ------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS subscriber_location_history (
    id BIGSERIAL PRIMARY KEY,
    subscriber_id BIGINT REFERENCES subscribers(id) ON DELETE CASCADE,
    imsi VARCHAR(15) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    cell_id VARCHAR(50),
    lac INTEGER,
    tac INTEGER,
    location VARCHAR(100),
    rat VARCHAR(10), -- 'GSM', 'UMTS', 'LTE', 'NR'
    event_type VARCHAR(50), -- 'attach', 'location_update', 'handover', 'detach'
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_location_history_subscriber_id ON subscriber_location_history(subscriber_id);
CREATE INDEX idx_location_history_imsi ON subscriber_location_history(imsi);
CREATE INDEX idx_location_history_timestamp ON subscriber_location_history(timestamp DESC);
CREATE INDEX idx_location_history_cell_id ON subscriber_location_history(cell_id);

-- Partition by month for better performance
-- ALTER TABLE subscriber_location_history PARTITION BY RANGE (timestamp);

COMMENT ON TABLE subscriber_location_history IS 'Historical location tracking for subscribers';

-- ------------------------------------------------------------------------------
-- 8. Transactions - MAP/CAP/INAP/Diameter/HTTP
-- ------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,
    transaction_id VARCHAR(100) UNIQUE NOT NULL, -- TID, Dialog ID, Session ID, Call ID
    protocol VARCHAR(20) NOT NULL, -- 'MAP', 'CAP', 'INAP', 'Diameter', 'HTTP', 'SIP'
    sub_protocol VARCHAR(50), -- 'S6a', 'Gx', 'Gy', 'SWx', etc. for Diameter

    -- Identifiers
    imsi VARCHAR(15),
    msisdn VARCHAR(15),
    imei VARCHAR(15),

    -- Network elements
    source_node VARCHAR(100), -- MME, HSS, SGSN, MSC, etc.
    dest_node VARCHAR(100),
    source_ip INET,
    dest_ip INET,
    source_port INTEGER,
    dest_port INTEGER,

    -- Transaction details
    operation VARCHAR(100), -- 'UpdateLocation', 'AuthInfo', 'CreateSession', etc.
    direction VARCHAR(10), -- 'request', 'response', 'notification'
    result VARCHAR(20), -- 'success', 'failure', 'timeout', 'abort'
    result_code INTEGER, -- Diameter Result-Code, MAP error, GTP cause, HTTP status
    error_description TEXT,

    -- Timing
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration_ms INTEGER,

    -- Location
    cell_id VARCHAR(50),
    lac INTEGER,
    tac INTEGER,
    mcc INTEGER,
    mnc INTEGER,

    -- Flags
    is_roaming BOOLEAN DEFAULT false,
    is_inbound BOOLEAN,
    is_test BOOLEAN DEFAULT false,

    -- Raw data
    request_data JSONB,
    response_data JSONB,

    -- Metadata
    metadata JSONB,

    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_transactions_transaction_id ON transactions(transaction_id);
CREATE INDEX idx_transactions_protocol ON transactions(protocol);
CREATE INDEX idx_transactions_sub_protocol ON transactions(sub_protocol);
CREATE INDEX idx_transactions_imsi ON transactions(imsi);
CREATE INDEX idx_transactions_msisdn ON transactions(msisdn);
CREATE INDEX idx_transactions_start_time ON transactions(start_time DESC);
CREATE INDEX idx_transactions_result ON transactions(result);
CREATE INDEX idx_transactions_result_code ON transactions(result_code);
CREATE INDEX idx_transactions_operation ON transactions(operation);
CREATE INDEX idx_transactions_roaming ON transactions(is_roaming);

COMMENT ON TABLE transactions IS 'All signaling transactions across protocols';

-- ------------------------------------------------------------------------------
-- 9. GTP Sessions - User Plane
-- ------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS gtp_sessions (
    id BIGSERIAL PRIMARY KEY,
    session_id VARCHAR(100) UNIQUE NOT NULL,

    -- Identifiers
    imsi VARCHAR(15) NOT NULL,
    msisdn VARCHAR(15),
    imei VARCHAR(15),

    -- GTP specific
    teid_control BIGINT, -- Control plane TEID
    teid_user BIGINT, -- User plane TEID
    apn VARCHAR(100),
    pdn_type VARCHAR(20), -- 'IPv4', 'IPv6', 'IPv4v6'
    ue_ip_address INET,

    -- QoS
    qci INTEGER, -- QoS Class Identifier (4G)
    arp INTEGER, -- Allocation and Retention Priority
    five_qi INTEGER, -- 5QI (5G)
    mbr_uplink_kbps INTEGER, -- Maximum Bit Rate
    mbr_downlink_kbps INTEGER,
    gbr_uplink_kbps INTEGER, -- Guaranteed Bit Rate
    gbr_downlink_kbps INTEGER,

    -- Network nodes
    sgw_ip INET,
    pgw_ip INET,
    enb_ip INET,
    mme_ip INET,

    -- Session lifecycle
    session_state VARCHAR(20), -- 'active', 'suspended', 'terminated'
    create_time TIMESTAMP NOT NULL,
    modify_time TIMESTAMP,
    delete_time TIMESTAMP,
    duration_seconds INTEGER,

    -- Location
    cell_id VARCHAR(50),
    tac INTEGER,
    ecgi VARCHAR(50),

    -- Traffic counters
    uplink_bytes BIGINT DEFAULT 0,
    downlink_bytes BIGINT DEFAULT 0,
    uplink_packets BIGINT DEFAULT 0,
    downlink_packets BIGINT DEFAULT 0,

    -- Termination cause
    termination_cause INTEGER,
    termination_reason TEXT,

    -- Metadata
    metadata JSONB,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_gtp_sessions_session_id ON gtp_sessions(session_id);
CREATE INDEX idx_gtp_sessions_imsi ON gtp_sessions(imsi);
CREATE INDEX idx_gtp_sessions_msisdn ON gtp_sessions(msisdn);
CREATE INDEX idx_gtp_sessions_apn ON gtp_sessions(apn);
CREATE INDEX idx_gtp_sessions_state ON gtp_sessions(session_state);
CREATE INDEX idx_gtp_sessions_create_time ON gtp_sessions(create_time DESC);
CREATE INDEX idx_gtp_sessions_teid_control ON gtp_sessions(teid_control);
CREATE INDEX idx_gtp_sessions_cell_id ON gtp_sessions(cell_id);

COMMENT ON TABLE gtp_sessions IS 'GTP user plane sessions with traffic counters';

-- ------------------------------------------------------------------------------
-- 10. Messages - Detailed Protocol Messages
-- ------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS messages (
    id BIGSERIAL PRIMARY KEY,
    transaction_id BIGINT REFERENCES transactions(id) ON DELETE CASCADE,
    session_id VARCHAR(100), -- Link to GTP session if applicable

    -- Message identification
    timestamp TIMESTAMP NOT NULL,
    protocol VARCHAR(20) NOT NULL,
    message_type VARCHAR(100) NOT NULL,
    direction VARCHAR(10) NOT NULL, -- 'sent', 'received'

    -- Network addressing
    source VARCHAR(100),
    destination VARCHAR(100),
    source_ip INET,
    dest_ip INET,

    -- Message content
    decoded_data JSONB, -- Fully decoded message as JSON
    raw_data BYTEA, -- Raw packet bytes

    -- Identifiers extracted from message
    imsi VARCHAR(15),
    msisdn VARCHAR(15),

    -- Message size
    message_size_bytes INTEGER,

    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_messages_transaction_id ON messages(transaction_id);
CREATE INDEX idx_messages_session_id ON messages(session_id);
CREATE INDEX idx_messages_timestamp ON messages(timestamp DESC);
CREATE INDEX idx_messages_protocol ON messages(protocol);
CREATE INDEX idx_messages_message_type ON messages(message_type);
CREATE INDEX idx_messages_imsi ON messages(imsi);

-- Partition by timestamp for better performance
-- ALTER TABLE messages PARTITION BY RANGE (timestamp);

COMMENT ON TABLE messages IS 'Detailed protocol messages with full decode';

-- ------------------------------------------------------------------------------
-- 11. KPI Counters - Aggregated Metrics
-- ------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS kpi_counters (
    id BIGSERIAL PRIMARY KEY,

    -- Time bucket
    timestamp TIMESTAMP NOT NULL,
    bucket_size VARCHAR(10) NOT NULL, -- '5min', '1hour', '1day'

    -- Dimensions
    protocol VARCHAR(20),
    sub_protocol VARCHAR(50),
    operation VARCHAR(100),
    cell_id VARCHAR(50),
    lac INTEGER,
    tac INTEGER,
    apn VARCHAR(100),
    mcc INTEGER,
    mnc INTEGER,
    network_element VARCHAR(100),

    -- Counters
    total_attempts INTEGER DEFAULT 0,
    total_successes INTEGER DEFAULT 0,
    total_failures INTEGER DEFAULT 0,
    total_timeouts INTEGER DEFAULT 0,
    total_aborts INTEGER DEFAULT 0,

    -- Latency stats
    avg_duration_ms NUMERIC(10, 2),
    min_duration_ms INTEGER,
    max_duration_ms INTEGER,
    p50_duration_ms INTEGER,
    p95_duration_ms INTEGER,
    p99_duration_ms INTEGER,

    -- Traffic volume (for GTP)
    total_uplink_mb NUMERIC(12, 2) DEFAULT 0,
    total_downlink_mb NUMERIC(12, 2) DEFAULT 0,

    -- Result codes distribution (JSONB for flexibility)
    result_code_distribution JSONB,

    created_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(timestamp, bucket_size, protocol, sub_protocol, operation, cell_id, apn, network_element)
);

CREATE INDEX idx_kpi_timestamp ON kpi_counters(timestamp DESC);
CREATE INDEX idx_kpi_bucket ON kpi_counters(bucket_size);
CREATE INDEX idx_kpi_protocol ON kpi_counters(protocol);
CREATE INDEX idx_kpi_cell_id ON kpi_counters(cell_id);
CREATE INDEX idx_kpi_apn ON kpi_counters(apn);

COMMENT ON TABLE kpi_counters IS 'Aggregated KPI metrics for dashboards and reporting';

-- ------------------------------------------------------------------------------
-- 12. Alarms and Alerts
-- ------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS alarm_rules (
    id SERIAL PRIMARY KEY,
    rule_name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    enabled BOOLEAN DEFAULT true,

    -- Rule conditions
    metric_name VARCHAR(100) NOT NULL, -- 'attach_success_rate', 'diameter_5xx_rate', etc.
    threshold_operator VARCHAR(10) NOT NULL, -- '>', '<', '>=', '<=', '=='
    threshold_value NUMERIC(10, 2) NOT NULL,
    evaluation_window_minutes INTEGER DEFAULT 5,

    -- Dimensions (filter)
    protocol VARCHAR(20),
    cell_id VARCHAR(50),
    apn VARCHAR(100),
    network_element VARCHAR(100),

    -- Severity and actions
    severity VARCHAR(20) NOT NULL, -- 'critical', 'major', 'minor', 'warning'
    notification_emails TEXT[], -- Array of email addresses
    notification_sms TEXT[],
    webhook_url TEXT,
    snmp_trap_enabled BOOLEAN DEFAULT false,

    -- Escalation
    escalation_delay_minutes INTEGER,
    escalation_emails TEXT[],

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    created_by VARCHAR(50)
);

CREATE TABLE IF NOT EXISTS alarms (
    id BIGSERIAL PRIMARY KEY,
    alarm_rule_id INTEGER REFERENCES alarm_rules(id) ON DELETE CASCADE,

    -- Alarm details
    alarm_name VARCHAR(100) NOT NULL,
    description TEXT,
    severity VARCHAR(20) NOT NULL,

    -- Context
    metric_name VARCHAR(100),
    metric_value NUMERIC(10, 2),
    threshold_value NUMERIC(10, 2),
    protocol VARCHAR(20),
    cell_id VARCHAR(50),
    apn VARCHAR(100),
    network_element VARCHAR(100),
    affected_subscribers INTEGER,

    -- Lifecycle
    alarm_state VARCHAR(20) DEFAULT 'active', -- 'active', 'acknowledged', 'cleared'
    first_occurrence TIMESTAMP DEFAULT NOW(),
    last_occurrence TIMESTAMP DEFAULT NOW(),
    occurrence_count INTEGER DEFAULT 1,

    -- Management
    acknowledged_at TIMESTAMP,
    acknowledged_by VARCHAR(50),
    cleared_at TIMESTAMP,
    cleared_reason TEXT,

    -- Additional context
    related_transaction_ids TEXT[],
    root_cause_analysis TEXT,
    recommendations TEXT,

    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_alarms_rule_id ON alarms(alarm_rule_id);
CREATE INDEX idx_alarms_severity ON alarms(severity);
CREATE INDEX idx_alarms_state ON alarms(alarm_state);
CREATE INDEX idx_alarms_first_occurrence ON alarms(first_occurrence DESC);
CREATE INDEX idx_alarms_protocol ON alarms(protocol);
CREATE INDEX idx_alarms_cell_id ON alarms(cell_id);

COMMENT ON TABLE alarms IS 'Active and historical alarms with management workflow';

-- ------------------------------------------------------------------------------
-- 13. AI Analysis - Detected Issues
-- ------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS ai_detected_issues (
    id BIGSERIAL PRIMARY KEY,

    -- Issue classification
    issue_type VARCHAR(50) NOT NULL, -- 'repeated_failure', 'anomaly', 'deviation', etc.
    severity VARCHAR(20) NOT NULL,
    category VARCHAR(50), -- 'protocol_error', 'network_issue', 'performance', etc.

    -- Description
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    root_cause TEXT,

    -- Context
    protocol VARCHAR(20),
    affected_subscribers INTEGER DEFAULT 0,
    affected_cells TEXT[],
    affected_apns TEXT[],

    -- Pattern details
    pattern_description TEXT,
    confidence_score NUMERIC(3, 2), -- 0.00 to 1.00

    -- Recommendations
    recommendations JSONB, -- Array of recommendation objects

    -- 3GPP/Standard reference
    standard_reference VARCHAR(100), -- e.g., "3GPP TS 29.272 Section 7.4.3"
    related_error_codes INTEGER[],

    -- Timestamps
    first_detected TIMESTAMP DEFAULT NOW(),
    last_detected TIMESTAMP DEFAULT NOW(),
    occurrence_count INTEGER DEFAULT 1,

    -- Resolution
    resolved BOOLEAN DEFAULT false,
    resolved_at TIMESTAMP,
    resolved_by VARCHAR(50),
    resolution_notes TEXT,

    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_ai_issues_type ON ai_detected_issues(issue_type);
CREATE INDEX idx_ai_issues_severity ON ai_detected_issues(severity);
CREATE INDEX idx_ai_issues_protocol ON ai_detected_issues(protocol);
CREATE INDEX idx_ai_issues_resolved ON ai_detected_issues(resolved);
CREATE INDEX idx_ai_issues_first_detected ON ai_detected_issues(first_detected DESC);

COMMENT ON TABLE ai_detected_issues IS 'AI-detected issues with root cause and recommendations';

-- ------------------------------------------------------------------------------
-- 14. Audit Log
-- ------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS audit_log (
    id BIGSERIAL PRIMARY KEY,

    -- Who
    username VARCHAR(50) NOT NULL,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,

    -- What
    action VARCHAR(100) NOT NULL, -- 'login', 'config_change', 'alarm_ack', etc.
    resource VARCHAR(100), -- Resource affected
    operation VARCHAR(20), -- 'create', 'read', 'update', 'delete'

    -- Details
    old_value JSONB,
    new_value JSONB,
    details TEXT,

    -- When & Where
    timestamp TIMESTAMP DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT,

    -- Result
    success BOOLEAN DEFAULT true,
    error_message TEXT
);

CREATE INDEX idx_audit_username ON audit_log(username);
CREATE INDEX idx_audit_action ON audit_log(action);
CREATE INDEX idx_audit_timestamp ON audit_log(timestamp DESC);
CREATE INDEX idx_audit_resource ON audit_log(resource);

COMMENT ON TABLE audit_log IS 'Complete audit trail of all user actions';

-- ------------------------------------------------------------------------------
-- 15. System Configuration (Runtime Config Snapshot)
-- ------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS system_config (
    id SERIAL PRIMARY KEY,
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT,
    value_type VARCHAR(20), -- 'string', 'integer', 'boolean', 'json'
    category VARCHAR(50), -- 'network', 'protocol', 'capture', 'alarm', etc.
    description TEXT,

    -- Versioning
    version INTEGER DEFAULT 1,
    previous_value TEXT,

    -- Change management
    last_modified_at TIMESTAMP DEFAULT NOW(),
    last_modified_by VARCHAR(50),

    -- Validation
    validation_regex VARCHAR(255),
    allowed_values TEXT[], -- For enum-like configs

    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_config_key ON system_config(config_key);
CREATE INDEX idx_config_category ON system_config(category);

COMMENT ON TABLE system_config IS 'Runtime configuration managed from web UI';

-- ------------------------------------------------------------------------------
-- 16. CDR Metadata (Track CDR Files)
-- ------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS cdr_files (
    id BIGSERIAL PRIMARY KEY,
    file_name VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    protocol VARCHAR(20) NOT NULL,

    -- Time range covered
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,

    -- Statistics
    record_count INTEGER DEFAULT 0,
    file_size_bytes BIGINT,

    -- Status
    status VARCHAR(20) DEFAULT 'active', -- 'active', 'archived', 'deleted'
    compressed BOOLEAN DEFAULT false,

    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_cdr_files_protocol ON cdr_files(protocol);
CREATE INDEX idx_cdr_files_start_time ON cdr_files(start_time DESC);
CREATE INDEX idx_cdr_files_status ON cdr_files(status);

COMMENT ON TABLE cdr_files IS 'Metadata for CDR files generated per protocol';

-- ------------------------------------------------------------------------------
-- 17. Saved Filters and Views (User Preferences)
-- ------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS saved_filters (
    id SERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    filter_name VARCHAR(100) NOT NULL,
    description TEXT,

    -- Filter criteria (stored as JSON)
    filter_criteria JSONB NOT NULL,

    -- Scope
    scope VARCHAR(20) DEFAULT 'private', -- 'private', 'shared', 'team'

    -- Usage stats
    usage_count INTEGER DEFAULT 0,
    last_used_at TIMESTAMP,

    created_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(user_id, filter_name)
);

CREATE INDEX idx_saved_filters_user_id ON saved_filters(user_id);
CREATE INDEX idx_saved_filters_scope ON saved_filters(scope);

COMMENT ON TABLE saved_filters IS 'User-defined saved filters and search queries';

-- ==============================================================================
-- Initial Data - Basic dictionaries
-- ==============================================================================

-- Insert default permissions
INSERT INTO permissions (permission_name, description, category) VALUES
    ('view_dashboard', 'View main dashboard', 'ui'),
    ('view_transactions', 'View transaction details', 'data'),
    ('view_subscribers', 'View subscriber information', 'data'),
    ('view_kpis', 'View KPI dashboards', 'data'),
    ('export_data', 'Export data and CDRs', 'data'),
    ('download_pcap', 'Download PCAP files', 'data'),
    ('manage_alarms', 'Acknowledge and manage alarms', 'alarm'),
    ('configure_alarms', 'Create and edit alarm rules', 'config'),
    ('view_config', 'View system configuration', 'config'),
    ('edit_config', 'Modify system configuration', 'config'),
    ('manage_users', 'Create and manage users', 'admin'),
    ('view_audit_log', 'View audit trail', 'security'),
    ('system_control', 'Start/stop/restart services', 'admin'),
    ('view_all_networks', 'Access all networks and PLMNs', 'data'),
    ('mask_subscriber_data', 'See masked IMSI/MSISDN', 'security')
ON CONFLICT (permission_name) DO NOTHING;

-- Sample MCC/MNC data (top operators)
INSERT INTO countries (mcc, country_code, country_name, region) VALUES
    (250, 'RUS', 'Russian Federation', 'Europe'),
    (310, 'USA', 'United States', 'Americas'),
    (262, 'DEU', 'Germany', 'Europe'),
    (234, 'GBR', 'United Kingdom', 'Europe'),
    (208, 'FRA', 'France', 'Europe'),
    (222, 'ITA', 'Italy', 'Europe'),
    (460, 'CHN', 'China', 'Asia'),
    (440, 'JPN', 'Japan', 'Asia'),
    (450, 'KOR', 'South Korea', 'Asia'),
    (505, 'AUS', 'Australia', 'Oceania')
ON CONFLICT (mcc) DO NOTHING;

INSERT INTO operators (mcc, mnc, operator_name, operator_type, technology, country) VALUES
    (250, 1, 'MTS Russia', 'MNO', '2G,3G,4G,5G', 'Russian Federation'),
    (250, 2, 'MegaFon', 'MNO', '2G,3G,4G,5G', 'Russian Federation'),
    (250, 99, 'Beeline', 'MNO', '2G,3G,4G,5G', 'Russian Federation'),
    (310, 260, 'T-Mobile USA', 'MNO', '2G,3G,4G,5G', 'United States'),
    (310, 410, 'AT&T', 'MNO', '2G,3G,4G,5G', 'United States'),
    (310, 120, 'Verizon', 'MNO', '3G,4G,5G', 'United States'),
    (262, 1, 'Deutsche Telekom', 'MNO', '2G,3G,4G,5G', 'Germany'),
    (262, 2, 'Vodafone Germany', 'MNO', '2G,3G,4G,5G', 'Germany'),
    (262, 3, 'Telefonica Germany', 'MNO', '2G,3G,4G,5G', 'Germany'),
    (234, 15, 'Vodafone UK', 'MNO', '2G,3G,4G,5G', 'United Kingdom'),
    (234, 20, 'Three UK', 'MNO', '3G,4G,5G', 'United Kingdom'),
    (234, 30, 'EE (Everything Everywhere)', 'MNO', '2G,3G,4G,5G', 'United Kingdom')
ON CONFLICT (mcc, mnc) DO NOTHING;

-- ==============================================================================
-- Views for Common Queries
-- ==============================================================================

-- Active subscribers with current location
CREATE OR REPLACE VIEW v_active_subscribers AS
SELECT
    s.id,
    s.imsi,
    s.msisdn,
    s.imei,
    s.subscriber_type,
    s.current_status,
    s.current_cell_id,
    s.current_location,
    s.current_rat,
    s.last_seen,
    nt.site_name,
    nt.region,
    nt.city,
    o.operator_name
FROM subscribers s
LEFT JOIN network_topology nt ON s.current_cell_id = nt.cell_id
LEFT JOIN operators o ON s.home_mcc = o.mcc AND s.home_mnc = o.mnc
WHERE s.current_status = 'attached'
ORDER BY s.last_seen DESC;

-- KPI summary (last hour)
CREATE OR REPLACE VIEW v_kpi_summary_1h AS
SELECT
    protocol,
    sub_protocol,
    SUM(total_attempts) as total_attempts,
    SUM(total_successes) as total_successes,
    SUM(total_failures) as total_failures,
    CASE
        WHEN SUM(total_attempts) > 0
        THEN ROUND((SUM(total_successes)::numeric / SUM(total_attempts) * 100), 2)
        ELSE 0
    END as success_rate,
    AVG(avg_duration_ms) as avg_duration_ms
FROM kpi_counters
WHERE timestamp >= NOW() - INTERVAL '1 hour'
    AND bucket_size = '5min'
GROUP BY protocol, sub_protocol
ORDER BY protocol, sub_protocol;

-- Active alarms
CREATE OR REPLACE VIEW v_active_alarms AS
SELECT
    a.id,
    a.alarm_name,
    a.severity,
    a.description,
    a.protocol,
    a.cell_id,
    a.network_element,
    a.affected_subscribers,
    a.first_occurrence,
    a.last_occurrence,
    a.occurrence_count,
    a.alarm_state,
    ar.rule_name,
    ar.threshold_value
FROM alarms a
JOIN alarm_rules ar ON a.alarm_rule_id = ar.id
WHERE a.alarm_state IN ('active', 'acknowledged')
ORDER BY
    CASE a.severity
        WHEN 'critical' THEN 1
        WHEN 'major' THEN 2
        WHEN 'minor' THEN 3
        WHEN 'warning' THEN 4
    END,
    a.first_occurrence DESC;

-- ==============================================================================
-- Functions and Triggers
-- ==============================================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply to relevant tables
CREATE TRIGGER update_subscribers_updated_at BEFORE UPDATE ON subscribers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_gtp_sessions_updated_at BEFORE UPDATE ON gtp_sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ==============================================================================
-- Schema Version
-- ==============================================================================

CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    description TEXT,
    applied_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO schema_version (version, description) VALUES
    (1, 'Initial schema - Complete multi-G monitoring database')
ON CONFLICT (version) DO NOTHING;

-- ==============================================================================
-- End of Schema
-- ==============================================================================

COMMENT ON SCHEMA public IS 'Protei Monitoring v2.0 - Multi-generation telecom monitoring system';
