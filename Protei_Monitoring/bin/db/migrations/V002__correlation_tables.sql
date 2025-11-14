-- Liquibase Migration: Correlation Engine Tables
-- Version: 2.0
-- Description: Creates tables for multi-identifier correlation across protocols

-- ============================================================================
-- Correlation Sessions Table
-- ============================================================================
-- Stores correlated sessions across multiple interfaces and protocols
CREATE TABLE IF NOT EXISTS correlation_sessions (
    id VARCHAR(100) PRIMARY KEY,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    session_type VARCHAR(50), -- voice, data, sms, location_update, etc.

    -- Data usage
    bytes_uplink BIGINT DEFAULT 0,
    bytes_downlink BIGINT DEFAULT 0,

    -- Quality metrics
    success_rate NUMERIC(5,2) DEFAULT 100.0,
    avg_latency_ms INTEGER,
    error_count INTEGER DEFAULT 0,

    -- Cross-protocol references
    map_transaction_id VARCHAR(100),
    diameter_session_id VARCHAR(255),
    gtp_teid BIGINT,
    pfcp_seid BIGINT,
    ngap_ue_id BIGINT,
    s1ap_mme_id BIGINT,

    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_corr_sessions_start_time ON correlation_sessions(start_time DESC);
CREATE INDEX idx_corr_sessions_end_time ON correlation_sessions(end_time DESC);
CREATE INDEX idx_corr_sessions_status ON correlation_sessions(status);
CREATE INDEX idx_corr_sessions_type ON correlation_sessions(session_type);
CREATE INDEX idx_corr_sessions_map_txn ON correlation_sessions(map_transaction_id);
CREATE INDEX idx_corr_sessions_diameter_sid ON correlation_sessions(diameter_session_id);
CREATE INDEX idx_corr_sessions_gtp_teid ON correlation_sessions(gtp_teid);

-- ============================================================================
-- Correlation Identifiers Table
-- ============================================================================
-- Stores all identifiers associated with a correlation session
CREATE TABLE IF NOT EXISTS correlation_identifiers (
    id BIGSERIAL PRIMARY KEY,
    session_id VARCHAR(100) NOT NULL REFERENCES correlation_sessions(id) ON DELETE CASCADE,
    identifier_type VARCHAR(20) NOT NULL, -- IMSI, MSISDN, IMEI, TEID, SEID, IP, etc.
    identifier_value VARCHAR(100) NOT NULL,
    protocol VARCHAR(20) NOT NULL,
    first_seen TIMESTAMP NOT NULL,
    last_seen TIMESTAMP NOT NULL,
    confidence NUMERIC(3,2) DEFAULT 1.0, -- 0.0 to 1.0
    created_at TIMESTAMP DEFAULT NOW()
);

-- Unique constraint to prevent duplicate identifiers per session
CREATE UNIQUE INDEX idx_corr_identifiers_unique ON correlation_identifiers(session_id, identifier_type, identifier_value);

-- Indexes for fast lookup by identifier
CREATE INDEX idx_corr_identifiers_type ON correlation_identifiers(identifier_type);
CREATE INDEX idx_corr_identifiers_value ON correlation_identifiers(identifier_value);
CREATE INDEX idx_corr_identifiers_type_value ON correlation_identifiers(identifier_type, identifier_value);
CREATE INDEX idx_corr_identifiers_session ON correlation_identifiers(session_id);

-- ============================================================================
-- Correlation Transactions Table
-- ============================================================================
-- Maps transactions to correlation sessions
CREATE TABLE IF NOT EXISTS correlation_transactions (
    id BIGSERIAL PRIMARY KEY,
    session_id VARCHAR(100) NOT NULL REFERENCES correlation_sessions(id) ON DELETE CASCADE,
    transaction_id VARCHAR(100) NOT NULL UNIQUE,
    protocol VARCHAR(20) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    success BOOLEAN DEFAULT true,
    latency_ms INTEGER,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_corr_transactions_session ON correlation_transactions(session_id);
CREATE INDEX idx_corr_transactions_protocol ON correlation_transactions(protocol);
CREATE INDEX idx_corr_transactions_timestamp ON correlation_transactions(timestamp DESC);

-- ============================================================================
-- Location History Table
-- ============================================================================
-- Tracks subscriber location changes across sessions
CREATE TABLE IF NOT EXISTS correlation_location_history (
    id BIGSERIAL PRIMARY KEY,
    session_id VARCHAR(100) NOT NULL REFERENCES correlation_sessions(id) ON DELETE CASCADE,
    timestamp TIMESTAMP NOT NULL,
    protocol VARCHAR(20) NOT NULL,

    -- 2G/3G location
    mcc VARCHAR(3),
    mnc VARCHAR(3),
    lac VARCHAR(10),
    cell_id VARCHAR(20),

    -- 4G location
    tac VARCHAR(10),
    eutran_cgi VARCHAR(50),

    -- 5G location
    global_ran_id VARCHAR(50),

    -- Coordinates (if available from external sources)
    latitude NUMERIC(10,7),
    longitude NUMERIC(10,7),

    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_corr_location_session ON correlation_location_history(session_id);
CREATE INDEX idx_corr_location_timestamp ON correlation_location_history(timestamp DESC);
CREATE INDEX idx_corr_location_mcc_mnc ON correlation_location_history(mcc, mnc);
CREATE INDEX idx_corr_location_cell ON correlation_location_history(cell_id);

-- ============================================================================
-- Cross-Protocol Mapping Table
-- ============================================================================
-- Stores explicit mappings between protocol-specific identifiers
CREATE TABLE IF NOT EXISTS correlation_cross_protocol_mapping (
    id BIGSERIAL PRIMARY KEY,
    session_id VARCHAR(100) NOT NULL REFERENCES correlation_sessions(id) ON DELETE CASCADE,

    -- MAP identifiers
    map_invoke_id INTEGER,
    sccp_calling VARCHAR(50),
    sccp_called VARCHAR(50),

    -- Diameter identifiers
    diameter_hop_by_hop_id BIGINT,
    diameter_end_to_end_id BIGINT,

    -- GTP identifiers
    gtp_teid_control BIGINT,
    gtp_teid_user BIGINT,
    gtp_sequence_number INTEGER,

    -- PFCP identifiers
    pfcp_seid BIGINT,
    pfcp_node_id VARCHAR(100),

    -- NGAP identifiers
    ngap_amf_ue_id BIGINT,
    ngap_ran_ue_id INTEGER,
    guami VARCHAR(100),

    -- S1AP identifiers
    s1ap_mme_ue_id INTEGER,
    s1ap_enb_ue_id INTEGER,
    gummei VARCHAR(100),

    -- NAS identifiers
    nas_security_context VARCHAR(100),
    eps_bearer_id INTEGER,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_cross_protocol_session ON correlation_cross_protocol_mapping(session_id);
CREATE INDEX idx_cross_protocol_gtp_teid ON correlation_cross_protocol_mapping(gtp_teid_control);
CREATE INDEX idx_cross_protocol_pfcp_seid ON correlation_cross_protocol_mapping(pfcp_seid);

-- ============================================================================
-- Subscriber Correlation Index
-- ============================================================================
-- Fast lookup table for subscriber-to-session mapping
CREATE TABLE IF NOT EXISTS subscriber_correlation_index (
    imsi VARCHAR(15) PRIMARY KEY,
    current_session_id VARCHAR(100) REFERENCES correlation_sessions(id),
    active_sessions INTEGER DEFAULT 0,
    total_sessions BIGINT DEFAULT 0,
    first_seen TIMESTAMP DEFAULT NOW(),
    last_seen TIMESTAMP DEFAULT NOW(),
    last_location VARCHAR(100),
    last_protocol VARCHAR(20),
    total_bytes_uplink BIGINT DEFAULT 0,
    total_bytes_downlink BIGINT DEFAULT 0
);

CREATE INDEX idx_subscriber_corr_last_seen ON subscriber_correlation_index(last_seen DESC);
CREATE INDEX idx_subscriber_corr_active ON subscriber_correlation_index(active_sessions) WHERE active_sessions > 0;

-- ============================================================================
-- Correlation Statistics Table
-- ============================================================================
-- Stores aggregated correlation statistics
CREATE TABLE IF NOT EXISTS correlation_statistics (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL,
    time_period VARCHAR(20) NOT NULL, -- minute, hour, day

    -- Session counts
    total_sessions INTEGER DEFAULT 0,
    active_sessions INTEGER DEFAULT 0,
    completed_sessions INTEGER DEFAULT 0,

    -- Identifier counts
    total_identifiers INTEGER DEFAULT 0,
    imsi_count INTEGER DEFAULT 0,
    msisdn_count INTEGER DEFAULT 0,
    teid_count INTEGER DEFAULT 0,

    -- Protocol distribution
    map_sessions INTEGER DEFAULT 0,
    diameter_sessions INTEGER DEFAULT 0,
    gtp_sessions INTEGER DEFAULT 0,
    http2_sessions INTEGER DEFAULT 0,
    ngap_sessions INTEGER DEFAULT 0,
    s1ap_sessions INTEGER DEFAULT 0,

    -- Quality metrics
    avg_success_rate NUMERIC(5,2),
    avg_latency_ms INTEGER,
    total_errors INTEGER DEFAULT 0,

    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_corr_stats_timestamp ON correlation_statistics(timestamp DESC);
CREATE INDEX idx_corr_stats_period ON correlation_statistics(time_period);

-- ============================================================================
-- Views
-- ============================================================================

-- View: Active Correlation Sessions with All Identifiers
CREATE OR REPLACE VIEW v_active_correlation_sessions AS
SELECT
    cs.id,
    cs.start_time,
    cs.end_time,
    cs.status,
    cs.session_type,
    cs.bytes_uplink,
    cs.bytes_downlink,
    cs.success_rate,
    cs.error_count,
    -- Aggregate identifiers
    STRING_AGG(DISTINCT CASE WHEN ci.identifier_type = 'IMSI' THEN ci.identifier_value END, ',') as imsi_list,
    STRING_AGG(DISTINCT CASE WHEN ci.identifier_type = 'MSISDN' THEN ci.identifier_value END, ',') as msisdn_list,
    STRING_AGG(DISTINCT CASE WHEN ci.identifier_type = 'IMEI' THEN ci.identifier_value END, ',') as imei_list,
    STRING_AGG(DISTINCT CASE WHEN ci.identifier_type = 'IP' THEN ci.identifier_value END, ',') as ip_list,
    -- Aggregate protocols
    STRING_AGG(DISTINCT ct.protocol, ',') as protocols,
    -- Transaction count
    COUNT(DISTINCT ct.transaction_id) as transaction_count,
    -- Current location
    (SELECT mcc || '-' || mnc || '-' || COALESCE(lac, tac) || '-' || COALESCE(cell_id, eutran_cgi, global_ran_id)
     FROM correlation_location_history clh
     WHERE clh.session_id = cs.id
     ORDER BY clh.timestamp DESC
     LIMIT 1) as current_location
FROM correlation_sessions cs
LEFT JOIN correlation_identifiers ci ON cs.id = ci.session_id
LEFT JOIN correlation_transactions ct ON cs.id = ct.session_id
WHERE cs.status = 'active'
GROUP BY cs.id;

-- View: Subscriber Timeline
CREATE OR REPLACE VIEW v_subscriber_timeline AS
SELECT
    ci.identifier_value as imsi,
    cs.id as session_id,
    cs.start_time,
    cs.end_time,
    cs.session_type,
    cs.bytes_uplink,
    cs.bytes_downlink,
    cs.success_rate,
    STRING_AGG(DISTINCT ct.protocol, ',') as protocols,
    COUNT(DISTINCT ct.transaction_id) as transaction_count
FROM correlation_identifiers ci
JOIN correlation_sessions cs ON ci.session_id = cs.id
LEFT JOIN correlation_transactions ct ON cs.id = ct.session_id
WHERE ci.identifier_type = 'IMSI'
GROUP BY ci.identifier_value, cs.id
ORDER BY cs.start_time DESC;

-- ============================================================================
-- Functions
-- ============================================================================

-- Function: Get Correlation Session Summary
CREATE OR REPLACE FUNCTION get_correlation_session_summary(p_session_id VARCHAR)
RETURNS TABLE (
    session_id VARCHAR,
    protocols TEXT,
    identifiers JSONB,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    duration_seconds INTEGER,
    transaction_count BIGINT,
    data_usage_mb NUMERIC,
    success_rate NUMERIC,
    location_changes INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        cs.id,
        STRING_AGG(DISTINCT ct.protocol, ', ') as protocols,
        JSONB_OBJECT_AGG(ci.identifier_type, ARRAY_AGG(DISTINCT ci.identifier_value)) as identifiers,
        cs.start_time,
        cs.end_time,
        EXTRACT(EPOCH FROM (cs.end_time - cs.start_time))::INTEGER as duration_seconds,
        COUNT(DISTINCT ct.transaction_id) as transaction_count,
        ROUND((cs.bytes_uplink + cs.bytes_downlink)::NUMERIC / 1048576, 2) as data_usage_mb,
        cs.success_rate,
        (SELECT COUNT(*) FROM correlation_location_history WHERE session_id = cs.id) as location_changes
    FROM correlation_sessions cs
    LEFT JOIN correlation_identifiers ci ON cs.id = ci.session_id
    LEFT JOIN correlation_transactions ct ON cs.id = ct.session_id
    WHERE cs.id = p_session_id
    GROUP BY cs.id;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Triggers
-- ============================================================================

-- Trigger: Update correlation_sessions.updated_at on modification
CREATE OR REPLACE FUNCTION update_correlation_session_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_correlation_session_timestamp
    BEFORE UPDATE ON correlation_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_correlation_session_timestamp();

-- Trigger: Update subscriber correlation index
CREATE OR REPLACE FUNCTION update_subscriber_correlation_index()
RETURNS TRIGGER AS $$
BEGIN
    -- Update subscriber index when new IMSI identifier is added
    IF NEW.identifier_type = 'IMSI' THEN
        INSERT INTO subscriber_correlation_index (
            imsi, current_session_id, active_sessions, total_sessions, last_seen
        ) VALUES (
            NEW.identifier_value, NEW.session_id, 1, 1, NEW.last_seen
        )
        ON CONFLICT (imsi) DO UPDATE SET
            current_session_id = EXCLUDED.current_session_id,
            active_sessions = subscriber_correlation_index.active_sessions + 1,
            total_sessions = subscriber_correlation_index.total_sessions + 1,
            last_seen = EXCLUDED.last_seen;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_subscriber_correlation_index
    AFTER INSERT ON correlation_identifiers
    FOR EACH ROW
    EXECUTE FUNCTION update_subscriber_correlation_index();

-- ============================================================================
-- Initial Data
-- ============================================================================

-- Insert initial correlation statistics record
INSERT INTO correlation_statistics (
    timestamp, time_period,
    total_sessions, active_sessions, completed_sessions
) VALUES (
    NOW(), 'hour', 0, 0, 0
)
ON CONFLICT DO NOTHING;

-- ============================================================================
-- Comments
-- ============================================================================

COMMENT ON TABLE correlation_sessions IS 'Stores correlated sessions across multiple protocols and interfaces';
COMMENT ON TABLE correlation_identifiers IS 'Stores all subscriber identifiers (IMSI, MSISDN, IMEI, TEID, etc.) associated with sessions';
COMMENT ON TABLE correlation_transactions IS 'Maps individual transactions to correlation sessions';
COMMENT ON TABLE correlation_location_history IS 'Tracks subscriber location changes across sessions';
COMMENT ON TABLE correlation_cross_protocol_mapping IS 'Stores explicit mappings between protocol-specific identifiers';
COMMENT ON TABLE subscriber_correlation_index IS 'Fast lookup index for subscriber-to-session mapping';
COMMENT ON TABLE correlation_statistics IS 'Aggregated correlation statistics for monitoring and reporting';

COMMENT ON VIEW v_active_correlation_sessions IS 'Consolidated view of active sessions with all identifiers and metadata';
COMMENT ON VIEW v_subscriber_timeline IS 'Timeline view of all sessions for each subscriber';

-- End of migration
