-- ============================================================================
-- Protei_Bulk Advanced SMSC Routing & Multi-Gateway Management
-- Database Schema
-- ============================================================================

-- ============================================================================
-- 1. SMSC CONNECTIONS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS tbl_smsc_connections (
    smsc_id BIGSERIAL PRIMARY KEY,
    smsc_code VARCHAR(64) UNIQUE NOT NULL,
    smsc_name VARCHAR(255) NOT NULL,

    -- Connection Details
    connection_type VARCHAR(20) DEFAULT 'SMPP' CHECK (connection_type IN ('SMPP', 'HTTP', 'CUSTOM')),
    host VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL,

    -- Authentication
    username VARCHAR(128),
    password VARCHAR(255),
    system_id VARCHAR(128),
    system_type VARCHAR(64),

    -- SMPP Configuration
    bind_type VARCHAR(20) DEFAULT 'TRANSCEIVER' CHECK (bind_type IN ('TRANSMITTER', 'RECEIVER', 'TRANSCEIVER')),
    smpp_version VARCHAR(10) DEFAULT '3.4',
    encoding VARCHAR(20) DEFAULT 'GSM7' CHECK (encoding IN ('GSM7', 'UCS2', 'UTF8', 'ASCII')),
    ton INTEGER DEFAULT 1,
    npi INTEGER DEFAULT 1,

    -- Capacity & Limits
    max_tps INTEGER DEFAULT 100,
    max_connections INTEGER DEFAULT 10,
    window_size INTEGER DEFAULT 10,
    throughput_limit INTEGER,

    -- Classification
    country_code VARCHAR(10),
    region VARCHAR(64),
    network_operator VARCHAR(128),
    mcc_mnc JSONB DEFAULT '[]',

    -- Routing Behavior
    is_default_route BOOLEAN DEFAULT FALSE,
    route_mode VARCHAR(20) DEFAULT 'ACTIVE' CHECK (route_mode IN ('ACTIVE', 'STANDBY', 'DISABLED')),
    priority INTEGER DEFAULT 100,

    -- Cost (per message)
    cost_per_sms DECIMAL(10, 4) DEFAULT 0.0,
    currency VARCHAR(10) DEFAULT 'USD',

    -- DLR Settings
    dlr_callback_url VARCHAR(500),
    supports_dlr BOOLEAN DEFAULT TRUE,

    -- Connection Status
    status VARCHAR(20) DEFAULT 'DISCONNECTED' CHECK (status IN ('CONNECTED', 'DISCONNECTED', 'SUSPENDED', 'MAINTENANCE', 'ERROR')),
    last_bind_time TIMESTAMP,
    last_unbind_time TIMESTAMP,
    last_heartbeat TIMESTAMP,
    connection_errors INTEGER DEFAULT 0,

    -- Statistics
    total_messages_sent BIGINT DEFAULT 0,
    total_messages_received BIGINT DEFAULT 0,
    current_tps DECIMAL(10, 2) DEFAULT 0.0,
    avg_response_time_ms INTEGER DEFAULT 0,
    delivery_rate DECIMAL(5, 2) DEFAULT 0.0,

    -- Alerts
    alert_on_disconnect BOOLEAN DEFAULT TRUE,
    alert_on_high_error_rate BOOLEAN DEFAULT TRUE,
    error_rate_threshold DECIMAL(5, 2) DEFAULT 5.0,
    alert_emails JSONB DEFAULT '[]',

    -- Audit
    created_by VARCHAR(64),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    -- Metadata
    metadata JSONB DEFAULT '{}',
    notes TEXT
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_smsc_code ON tbl_smsc_connections(smsc_code);
CREATE INDEX IF NOT EXISTS idx_smsc_status ON tbl_smsc_connections(status);
CREATE INDEX IF NOT EXISTS idx_smsc_country ON tbl_smsc_connections(country_code);
CREATE INDEX IF NOT EXISTS idx_smsc_route_mode ON tbl_smsc_connections(route_mode);
CREATE INDEX IF NOT EXISTS idx_smsc_priority ON tbl_smsc_connections(priority);

-- ============================================================================
-- 2. ROUTING RULES TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS tbl_routing_rules (
    rule_id BIGSERIAL PRIMARY KEY,
    rule_code VARCHAR(64) UNIQUE NOT NULL,
    rule_name VARCHAR(255) NOT NULL,

    -- Scope
    customer_id BIGINT REFERENCES tbl_customers(customer_id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,

    -- Match Conditions
    condition_type VARCHAR(20) NOT NULL CHECK (condition_type IN ('PREFIX', 'SENDER', 'CUSTOMER', 'COUNTRY', 'MESSAGE_TYPE', 'REGEX', 'COMBINED')),
    condition_value VARCHAR(500) NOT NULL,

    -- Additional Filters
    msisdn_prefix VARCHAR(20),
    sender_id_pattern VARCHAR(100),
    message_type VARCHAR(20),
    country_code VARCHAR(10),
    mcc_mnc VARCHAR(10),
    regex_pattern VARCHAR(500),

    -- Combined Conditions (JSON)
    combined_conditions JSONB DEFAULT '{}',

    -- Routing Target
    smsc_id BIGINT REFERENCES tbl_smsc_connections(smsc_id) ON DELETE CASCADE,
    fallback_smsc_id BIGINT REFERENCES tbl_smsc_connections(smsc_id) ON DELETE SET NULL,

    -- Priority & Control
    priority INTEGER DEFAULT 100,
    is_active BOOLEAN DEFAULT TRUE,

    -- Time-Based Routing
    enable_time_based BOOLEAN DEFAULT FALSE,
    active_hours_start TIME,
    active_hours_end TIME,
    active_days JSONB DEFAULT '[]',
    timezone VARCHAR(50) DEFAULT 'UTC',

    -- Load Balancing
    enable_load_balance BOOLEAN DEFAULT FALSE,
    load_balance_smsc_ids JSONB DEFAULT '[]',
    load_balance_weights JSONB DEFAULT '{}',

    -- Cost Optimization
    enable_cost_routing BOOLEAN DEFAULT FALSE,
    max_cost_per_sms DECIMAL(10, 4),

    -- Statistics
    total_messages_routed BIGINT DEFAULT 0,
    total_fallbacks_used BIGINT DEFAULT 0,
    last_used_at TIMESTAMP,

    -- Audit
    created_by VARCHAR(64),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    -- Metadata
    metadata JSONB DEFAULT '{}',
    notes TEXT
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_routing_rule_code ON tbl_routing_rules(rule_code);
CREATE INDEX IF NOT EXISTS idx_routing_customer ON tbl_routing_rules(customer_id);
CREATE INDEX IF NOT EXISTS idx_routing_user ON tbl_routing_rules(user_id);
CREATE INDEX IF NOT EXISTS idx_routing_condition_type ON tbl_routing_rules(condition_type);
CREATE INDEX IF NOT EXISTS idx_routing_priority ON tbl_routing_rules(priority);
CREATE INDEX IF NOT EXISTS idx_routing_active ON tbl_routing_rules(is_active);
CREATE INDEX IF NOT EXISTS idx_routing_smsc ON tbl_routing_rules(smsc_id);
CREATE INDEX IF NOT EXISTS idx_routing_prefix ON tbl_routing_rules(msisdn_prefix);

-- ============================================================================
-- 3. ROUTING LOGS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS tbl_routing_logs (
    log_id BIGSERIAL PRIMARY KEY,

    -- Message Reference
    message_id BIGINT REFERENCES messages(id) ON DELETE CASCADE,
    campaign_id BIGINT REFERENCES campaigns(id) ON DELETE CASCADE,

    -- Routing Decision
    msisdn VARCHAR(20) NOT NULL,
    sender_id VARCHAR(20),
    message_type VARCHAR(20),

    -- Rule Applied
    rule_id BIGINT REFERENCES tbl_routing_rules(rule_id) ON DELETE SET NULL,
    rule_name VARCHAR(255),

    -- SMSC Selection
    selected_smsc_id BIGINT REFERENCES tbl_smsc_connections(smsc_id) ON DELETE SET NULL,
    smsc_name VARCHAR(255),

    -- Fallback Info
    is_fallback BOOLEAN DEFAULT FALSE,
    fallback_reason VARCHAR(255),
    original_smsc_id BIGINT,

    -- Result
    routing_status VARCHAR(20) CHECK (routing_status IN ('SUCCESS', 'FALLBACK', 'FAILED', 'NO_ROUTE')),
    routing_time_ms INTEGER,

    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),

    -- Metadata
    metadata JSONB DEFAULT '{}'
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_routing_log_message ON tbl_routing_logs(message_id);
CREATE INDEX IF NOT EXISTS idx_routing_log_campaign ON tbl_routing_logs(campaign_id);
CREATE INDEX IF NOT EXISTS idx_routing_log_rule ON tbl_routing_logs(rule_id);
CREATE INDEX IF NOT EXISTS idx_routing_log_smsc ON tbl_routing_logs(selected_smsc_id);
CREATE INDEX IF NOT EXISTS idx_routing_log_created ON tbl_routing_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_routing_log_msisdn ON tbl_routing_logs(msisdn);

-- Partition by month for scalability
CREATE TABLE IF NOT EXISTS tbl_routing_logs_partitioned (LIKE tbl_routing_logs INCLUDING ALL)
PARTITION BY RANGE (created_at);

-- ============================================================================
-- 4. COUNTRY CODES TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS tbl_country_codes (
    country_id BIGSERIAL PRIMARY KEY,
    country_code VARCHAR(10) UNIQUE NOT NULL,
    country_name VARCHAR(128) NOT NULL,
    iso_code_2 VARCHAR(2),
    iso_code_3 VARCHAR(3),

    -- Mobile Prefixes
    mobile_prefixes JSONB DEFAULT '[]',

    -- MCC-MNC Codes
    mcc VARCHAR(5),
    mnc_list JSONB DEFAULT '[]',

    -- Operators
    operators JSONB DEFAULT '[]',

    -- Routing Info
    default_smsc_id BIGINT REFERENCES tbl_smsc_connections(smsc_id) ON DELETE SET NULL,
    region VARCHAR(64),

    -- Metadata
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_country_code ON tbl_country_codes(country_code);
CREATE INDEX IF NOT EXISTS idx_country_iso2 ON tbl_country_codes(iso_code_2);
CREATE INDEX IF NOT EXISTS idx_country_region ON tbl_country_codes(region);

-- ============================================================================
-- 5. SMSC STATISTICS TABLE (Time-Series Data)
-- ============================================================================

CREATE TABLE IF NOT EXISTS tbl_smsc_statistics (
    stat_id BIGSERIAL PRIMARY KEY,
    smsc_id BIGINT REFERENCES tbl_smsc_connections(smsc_id) ON DELETE CASCADE,

    -- Time Period
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,
    period_type VARCHAR(20) DEFAULT 'HOURLY' CHECK (period_type IN ('MINUTE', 'HOURLY', 'DAILY', 'MONTHLY')),

    -- Traffic
    messages_submitted INTEGER DEFAULT 0,
    messages_delivered INTEGER DEFAULT 0,
    messages_failed INTEGER DEFAULT 0,
    messages_pending INTEGER DEFAULT 0,

    -- Performance
    avg_tps DECIMAL(10, 2) DEFAULT 0.0,
    peak_tps DECIMAL(10, 2) DEFAULT 0.0,
    avg_response_time_ms INTEGER DEFAULT 0,

    -- Delivery
    delivery_rate DECIMAL(5, 2) DEFAULT 0.0,
    error_rate DECIMAL(5, 2) DEFAULT 0.0,

    -- Errors
    bind_errors INTEGER DEFAULT 0,
    submit_errors INTEGER DEFAULT 0,
    timeout_errors INTEGER DEFAULT 0,

    -- Cost
    total_cost DECIMAL(15, 4) DEFAULT 0.0,

    -- Created
    created_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(smsc_id, period_start, period_type)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_smsc_stats_smsc ON tbl_smsc_statistics(smsc_id);
CREATE INDEX IF NOT EXISTS idx_smsc_stats_period ON tbl_smsc_statistics(period_start);
CREATE INDEX IF NOT EXISTS idx_smsc_stats_type ON tbl_smsc_statistics(period_type);

-- ============================================================================
-- 6. HELPER FUNCTIONS
-- ============================================================================

-- Function to find matching routing rule
CREATE OR REPLACE FUNCTION find_routing_rule(
    p_msisdn VARCHAR,
    p_sender_id VARCHAR,
    p_message_type VARCHAR,
    p_customer_id BIGINT
) RETURNS TABLE (
    rule_id BIGINT,
    rule_name VARCHAR,
    smsc_id BIGINT,
    fallback_smsc_id BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        r.rule_id,
        r.rule_name,
        r.smsc_id,
        r.fallback_smsc_id
    FROM tbl_routing_rules r
    WHERE r.is_active = TRUE
      AND (r.customer_id = p_customer_id OR r.customer_id IS NULL)
      AND (
          -- Prefix match
          (r.condition_type = 'PREFIX' AND p_msisdn LIKE r.condition_value || '%') OR
          -- Sender match
          (r.condition_type = 'SENDER' AND p_sender_id LIKE r.condition_value) OR
          -- Message type match
          (r.condition_type = 'MESSAGE_TYPE' AND p_message_type = r.condition_value) OR
          -- Customer match
          (r.condition_type = 'CUSTOMER' AND r.customer_id = p_customer_id)
      )
    ORDER BY r.priority ASC
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- Function to get country code from MSISDN
CREATE OR REPLACE FUNCTION get_country_from_msisdn(p_msisdn VARCHAR)
RETURNS VARCHAR AS $$
DECLARE
    v_country_code VARCHAR(10);
BEGIN
    -- Try to match against country prefixes (longest match first)
    SELECT country_code INTO v_country_code
    FROM tbl_country_codes
    WHERE p_msisdn LIKE country_code || '%'
    ORDER BY LENGTH(country_code) DESC
    LIMIT 1;

    RETURN v_country_code;
END;
$$ LANGUAGE plpgsql;

-- Function to get default SMSC
CREATE OR REPLACE FUNCTION get_default_smsc() RETURNS BIGINT AS $$
DECLARE
    v_smsc_id BIGINT;
BEGIN
    SELECT smsc_id INTO v_smsc_id
    FROM tbl_smsc_connections
    WHERE is_default_route = TRUE
      AND route_mode = 'ACTIVE'
      AND status = 'CONNECTED'
    ORDER BY priority ASC
    LIMIT 1;

    RETURN v_smsc_id;
END;
$$ LANGUAGE plpgsql;

-- Function to update SMSC statistics
CREATE OR REPLACE FUNCTION update_smsc_statistics(
    p_smsc_id BIGINT,
    p_status VARCHAR,
    p_response_time_ms INTEGER
) RETURNS VOID AS $$
BEGIN
    -- Update hourly statistics
    INSERT INTO tbl_smsc_statistics (
        smsc_id,
        period_start,
        period_end,
        period_type,
        messages_submitted,
        messages_delivered,
        messages_failed,
        avg_response_time_ms
    ) VALUES (
        p_smsc_id,
        DATE_TRUNC('hour', NOW()),
        DATE_TRUNC('hour', NOW()) + INTERVAL '1 hour',
        'HOURLY',
        1,
        CASE WHEN p_status = 'DELIVERED' THEN 1 ELSE 0 END,
        CASE WHEN p_status = 'FAILED' THEN 1 ELSE 0 END,
        p_response_time_ms
    )
    ON CONFLICT (smsc_id, period_start, period_type)
    DO UPDATE SET
        messages_submitted = tbl_smsc_statistics.messages_submitted + 1,
        messages_delivered = tbl_smsc_statistics.messages_delivered +
            CASE WHEN p_status = 'DELIVERED' THEN 1 ELSE 0 END,
        messages_failed = tbl_smsc_statistics.messages_failed +
            CASE WHEN p_status = 'FAILED' THEN 1 ELSE 0 END,
        avg_response_time_ms = (tbl_smsc_statistics.avg_response_time_ms + p_response_time_ms) / 2;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- 7. TRIGGERS
-- ============================================================================

-- Trigger to update routing rule statistics
CREATE OR REPLACE FUNCTION update_routing_rule_stats()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE tbl_routing_rules
    SET total_messages_routed = total_messages_routed + 1,
        last_used_at = NOW()
    WHERE rule_id = NEW.rule_id;

    IF NEW.is_fallback THEN
        UPDATE tbl_routing_rules
        SET total_fallbacks_used = total_fallbacks_used + 1
        WHERE rule_id = NEW.rule_id;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_routing_stats
AFTER INSERT ON tbl_routing_logs
FOR EACH ROW EXECUTE FUNCTION update_routing_rule_stats();

-- Trigger to update SMSC connection statistics
CREATE OR REPLACE FUNCTION update_smsc_connection_stats()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE tbl_smsc_connections
    SET total_messages_sent = total_messages_sent + 1,
        updated_at = NOW()
    WHERE smsc_id = NEW.selected_smsc_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_smsc_stats
AFTER INSERT ON tbl_routing_logs
FOR EACH ROW EXECUTE FUNCTION update_smsc_connection_stats();

-- ============================================================================
-- 8. SEED DATA - COMMON COUNTRY CODES
-- ============================================================================

INSERT INTO tbl_country_codes (country_code, country_name, iso_code_2, iso_code_3, region, mobile_prefixes, mcc) VALUES
('962', 'Jordan', 'JO', 'JOR', 'Middle East', '["7", "77", "78", "79"]', '416'),
('966', 'Saudi Arabia', 'SA', 'SAU', 'GCC', '["5"]', '420'),
('971', 'UAE', 'AE', 'ARE', 'GCC', '["5"]', '424'),
('965', 'Kuwait', 'KW', 'KWT', 'GCC', '["5", "6", "9"]', '419'),
('968', 'Oman', 'OM', 'OMN', 'GCC', '["7", "9"]', '422'),
('974', 'Qatar', 'QA', 'QAT', 'GCC', '["3", "5", "6", "7"]', '427'),
('973', 'Bahrain', 'BH', 'BHR', 'GCC', '["3"]', '426'),
('20', 'Egypt', 'EG', 'EGY', 'Middle East', '["1"]', '602'),
('961', 'Lebanon', 'LB', 'LBN', 'Middle East', '["3", "7"]', '415'),
('1', 'USA', 'US', 'USA', 'North America', '[]', '310'),
('44', 'UK', 'GB', 'GBR', 'Europe', '["7"]', '234')
ON CONFLICT (country_code) DO NOTHING;

COMMIT;
