-- ============================================================================
-- Protei_Bulk Comprehensive CDR (Call Detail Record) Schema
-- Handles all channels: SMS, USSD, WhatsApp, Telegram, Email, Push
-- ============================================================================

-- ============================================================================
-- 1. MAIN CDR TABLE (Partitioned by Month)
-- ============================================================================

CREATE TABLE IF NOT EXISTS tbl_cdr_records (
    cdr_id BIGSERIAL NOT NULL,

    -- Timestamps (Critical for partitioning)
    submission_timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    queued_timestamp TIMESTAMP,
    sent_timestamp TIMESTAMP,
    delivered_timestamp TIMESTAMP,
    failed_timestamp TIMESTAMP,

    -- Identifiers
    message_id VARCHAR(64) NOT NULL,
    campaign_id VARCHAR(64),
    customer_id BIGINT NOT NULL,
    user_id BIGINT,
    account_id BIGINT,

    -- Channel Information
    channel VARCHAR(20) NOT NULL CHECK (channel IN ('SMS', 'USSD', 'WHATSAPP', 'TELEGRAM', 'EMAIL', 'PUSH', 'OTHER')),
    message_type VARCHAR(20) CHECK (message_type IN ('P2P', 'A2P', 'P2A', 'BULK', 'OTP', 'TRANSACTIONAL', 'PROMOTIONAL')),

    -- Source & Destination
    sender_id VARCHAR(100),
    destination VARCHAR(255) NOT NULL,
    destination_type VARCHAR(20) DEFAULT 'MSISDN' CHECK (destination_type IN ('MSISDN', 'EMAIL', 'DEVICE_TOKEN', 'USER_ID', 'GROUP_ID')),

    -- Routing Information
    route_id VARCHAR(64),
    smsc_id VARCHAR(64),
    smsc_name VARCHAR(255),
    gateway_id VARCHAR(64),
    is_fallback BOOLEAN DEFAULT FALSE,
    routing_priority INTEGER DEFAULT 5,

    -- Message Content
    message_text TEXT,
    message_length INTEGER,
    message_parts INTEGER DEFAULT 1,
    encoding VARCHAR(20),

    -- Status & Result
    status VARCHAR(20) NOT NULL CHECK (status IN ('PENDING', 'QUEUED', 'SENT', 'DELIVERED', 'FAILED', 'REJECTED', 'EXPIRED', 'BLACKLISTED')),
    dlr_status VARCHAR(50),
    error_code VARCHAR(20),
    error_message TEXT,
    reject_reason VARCHAR(255),

    -- DLR (Delivery Report) Details
    dlr_received_at TIMESTAMP,
    dlr_text VARCHAR(500),
    smsc_message_id VARCHAR(100),

    -- Performance Metrics
    routing_time_ms INTEGER,
    queue_time_ms INTEGER,
    sending_time_ms INTEGER,
    delivery_time_ms INTEGER,
    total_time_ms INTEGER,

    -- Billing & Cost
    cost DECIMAL(10, 4) DEFAULT 0.0,
    currency VARCHAR(10) DEFAULT 'USD',
    billing_units INTEGER DEFAULT 1,

    -- Quality Metrics
    tps_at_submission DECIMAL(10, 2),
    queue_depth_at_submission INTEGER,

    -- Security & Compliance
    submission_ip INET,
    submission_user_agent TEXT,
    api_key_used VARCHAR(64),

    -- Channel-Specific Data (JSON for flexibility)
    channel_metadata JSONB DEFAULT '{}',

    -- Audit
    created_at TIMESTAMP DEFAULT NOW(),

    PRIMARY KEY (cdr_id, submission_timestamp)
) PARTITION BY RANGE (submission_timestamp);

-- Create indexes on main table (inherited by partitions)
CREATE INDEX IF NOT EXISTS idx_cdr_submission ON tbl_cdr_records(submission_timestamp);
CREATE INDEX IF NOT EXISTS idx_cdr_message_id ON tbl_cdr_records(message_id);
CREATE INDEX IF NOT EXISTS idx_cdr_campaign ON tbl_cdr_records(campaign_id);
CREATE INDEX IF NOT EXISTS idx_cdr_customer ON tbl_cdr_records(customer_id);
CREATE INDEX IF NOT EXISTS idx_cdr_user ON tbl_cdr_records(user_id);
CREATE INDEX IF NOT EXISTS idx_cdr_channel ON tbl_cdr_records(channel);
CREATE INDEX IF NOT EXISTS idx_cdr_status ON tbl_cdr_records(status);
CREATE INDEX IF NOT EXISTS idx_cdr_destination ON tbl_cdr_records(destination);
CREATE INDEX IF NOT EXISTS idx_cdr_smsc ON tbl_cdr_records(smsc_id);
CREATE INDEX IF NOT EXISTS idx_cdr_created ON tbl_cdr_records(created_at);

-- GIN index for channel metadata
CREATE INDEX IF NOT EXISTS idx_cdr_metadata ON tbl_cdr_records USING GIN(channel_metadata);

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_cdr_customer_timestamp ON tbl_cdr_records(customer_id, submission_timestamp);
CREATE INDEX IF NOT EXISTS idx_cdr_campaign_status ON tbl_cdr_records(campaign_id, status);
CREATE INDEX IF NOT EXISTS idx_cdr_channel_timestamp ON tbl_cdr_records(channel, submission_timestamp);

-- ============================================================================
-- 2. CREATE INITIAL PARTITIONS (Last 3 months + Next 3 months)
-- ============================================================================

-- Create partitions for current and upcoming months
DO $$
DECLARE
    partition_date DATE;
    partition_name TEXT;
    start_date DATE;
    end_date DATE;
BEGIN
    -- Create partitions from 3 months ago to 3 months in future
    FOR i IN -3..3 LOOP
        partition_date := DATE_TRUNC('month', NOW() + (i || ' months')::INTERVAL);
        partition_name := 'tbl_cdr_records_' || TO_CHAR(partition_date, 'YYYY_MM');
        start_date := partition_date;
        end_date := partition_date + INTERVAL '1 month';

        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF tbl_cdr_records FOR VALUES FROM (%L) TO (%L)',
            partition_name, start_date, end_date
        );
    END LOOP;
END $$;

-- ============================================================================
-- 3. DAILY CDR SUMMARY (Pre-Aggregated for Fast Reporting)
-- ============================================================================

CREATE TABLE IF NOT EXISTS tbl_cdr_daily_summary (
    summary_id BIGSERIAL PRIMARY KEY,

    -- Period
    report_date DATE NOT NULL,
    customer_id BIGINT NOT NULL,

    -- Grouping Dimensions
    campaign_id VARCHAR(64),
    channel VARCHAR(20),
    message_type VARCHAR(20),
    smsc_id VARCHAR(64),
    status VARCHAR(20),

    -- Aggregated Counts
    total_submitted INTEGER DEFAULT 0,
    total_queued INTEGER DEFAULT 0,
    total_sent INTEGER DEFAULT 0,
    total_delivered INTEGER DEFAULT 0,
    total_failed INTEGER DEFAULT 0,
    total_rejected INTEGER DEFAULT 0,
    total_expired INTEGER DEFAULT 0,
    total_blacklisted INTEGER DEFAULT 0,

    -- Performance Metrics
    avg_routing_time_ms DECIMAL(10, 2),
    avg_queue_time_ms DECIMAL(10, 2),
    avg_sending_time_ms DECIMAL(10, 2),
    avg_delivery_time_ms DECIMAL(10, 2),
    max_delivery_time_ms INTEGER,
    min_delivery_time_ms INTEGER,

    -- Quality Metrics
    delivery_rate DECIMAL(5, 2),  -- Percentage
    failure_rate DECIMAL(5, 2),   -- Percentage
    avg_tps DECIMAL(10, 2),
    peak_tps DECIMAL(10, 2),

    -- Billing
    total_cost DECIMAL(15, 4) DEFAULT 0.0,
    total_billing_units INTEGER DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(report_date, customer_id, campaign_id, channel, message_type, smsc_id, status)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_summary_date ON tbl_cdr_daily_summary(report_date);
CREATE INDEX IF NOT EXISTS idx_summary_customer ON tbl_cdr_daily_summary(customer_id);
CREATE INDEX IF NOT EXISTS idx_summary_campaign ON tbl_cdr_daily_summary(campaign_id);
CREATE INDEX IF NOT EXISTS idx_summary_channel ON tbl_cdr_daily_summary(channel);
CREATE INDEX IF NOT EXISTS idx_summary_composite ON tbl_cdr_daily_summary(report_date, customer_id, channel);

-- ============================================================================
-- 4. HOURLY CDR STATISTICS (Real-Time Monitoring)
-- ============================================================================

CREATE TABLE IF NOT EXISTS tbl_cdr_hourly_stats (
    stat_id BIGSERIAL PRIMARY KEY,

    -- Time Period
    period_hour TIMESTAMP NOT NULL,
    customer_id BIGINT NOT NULL,
    channel VARCHAR(20),

    -- Counts
    messages_submitted INTEGER DEFAULT 0,
    messages_delivered INTEGER DEFAULT 0,
    messages_failed INTEGER DEFAULT 0,

    -- Performance
    avg_tps DECIMAL(10, 2) DEFAULT 0.0,
    peak_tps DECIMAL(10, 2) DEFAULT 0.0,
    avg_response_time_ms INTEGER DEFAULT 0,

    -- Quality
    delivery_rate DECIMAL(5, 2) DEFAULT 0.0,
    error_rate DECIMAL(5, 2) DEFAULT 0.0,

    -- Created
    created_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(period_hour, customer_id, channel)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_hourly_period ON tbl_cdr_hourly_stats(period_hour);
CREATE INDEX IF NOT EXISTS idx_hourly_customer ON tbl_cdr_hourly_stats(customer_id);
CREATE INDEX IF NOT EXISTS idx_hourly_channel ON tbl_cdr_hourly_stats(channel);

-- ============================================================================
-- 5. CHANNEL-SPECIFIC EXTENDED CDR TABLES
-- ============================================================================

-- SMS Extended CDR
CREATE TABLE IF NOT EXISTS tbl_cdr_sms_extended (
    cdr_id BIGINT PRIMARY KEY,

    -- SMPP Specific
    smpp_version VARCHAR(10),
    esm_class INTEGER,
    protocol_id INTEGER,
    priority_flag INTEGER,
    data_coding INTEGER,
    registered_delivery INTEGER,

    -- Concatenated Messages
    is_concatenated BOOLEAN DEFAULT FALSE,
    concat_ref INTEGER,
    concat_total_parts INTEGER,
    concat_part_number INTEGER,

    -- Additional Fields
    validity_period INTEGER,
    service_type VARCHAR(20),
    source_ton INTEGER,
    source_npi INTEGER,
    dest_ton INTEGER,
    dest_npi INTEGER,

    -- Operator Info
    mcc VARCHAR(5),
    mnc VARCHAR(5),
    network_operator VARCHAR(100),

    -- Created
    created_at TIMESTAMP DEFAULT NOW()
);

-- WhatsApp Extended CDR
CREATE TABLE IF NOT EXISTS tbl_cdr_whatsapp_extended (
    cdr_id BIGINT PRIMARY KEY,

    -- WhatsApp Specific
    wa_message_id VARCHAR(128),
    template_name VARCHAR(128),
    template_language VARCHAR(10),

    -- Media
    has_media BOOLEAN DEFAULT FALSE,
    media_type VARCHAR(20),
    media_url VARCHAR(500),

    -- Interactive
    has_buttons BOOLEAN DEFAULT FALSE,
    button_clicked VARCHAR(128),

    -- Status Details
    wa_status VARCHAR(50),
    wa_timestamp TIMESTAMP,

    -- Created
    created_at TIMESTAMP DEFAULT NOW()
);

-- Email Extended CDR
CREATE TABLE IF NOT EXISTS tbl_cdr_email_extended (
    cdr_id BIGINT PRIMARY KEY,

    -- Email Specific
    subject VARCHAR(500),
    from_email VARCHAR(255),
    reply_to VARCHAR(255),

    -- Tracking
    opened BOOLEAN DEFAULT FALSE,
    opened_at TIMESTAMP,
    clicked BOOLEAN DEFAULT FALSE,
    clicked_at TIMESTAMP,

    -- Bounces
    bounce_type VARCHAR(20),
    bounce_reason TEXT,

    -- SMTP Info
    smtp_server VARCHAR(255),
    smtp_response_code INTEGER,

    -- Created
    created_at TIMESTAMP DEFAULT NOW()
);

-- Push Notification Extended CDR
CREATE TABLE IF NOT EXISTS tbl_cdr_push_extended (
    cdr_id BIGINT PRIMARY KEY,

    -- Push Specific
    platform VARCHAR(20) CHECK (platform IN ('FCM', 'APNS', 'HUAWEI')),
    device_token VARCHAR(255),

    -- Notification
    title VARCHAR(255),
    badge INTEGER,
    sound VARCHAR(64),

    -- Data Payload
    data_payload JSONB,

    -- Status
    notification_id VARCHAR(128),
    fcm_message_id VARCHAR(128),
    apns_id VARCHAR(128),

    -- Created
    created_at TIMESTAMP DEFAULT NOW()
);

-- ============================================================================
-- 6. FUNCTIONS FOR CDR PROCESSING
-- ============================================================================

-- Function to insert CDR with automatic partition routing
CREATE OR REPLACE FUNCTION insert_cdr(
    p_message_id VARCHAR,
    p_customer_id BIGINT,
    p_channel VARCHAR,
    p_sender_id VARCHAR,
    p_destination VARCHAR,
    p_status VARCHAR,
    p_cost DECIMAL DEFAULT 0.0
) RETURNS BIGINT AS $$
DECLARE
    v_cdr_id BIGINT;
BEGIN
    INSERT INTO tbl_cdr_records (
        message_id, customer_id, channel, sender_id, destination, status, cost
    ) VALUES (
        p_message_id, p_customer_id, p_channel, p_sender_id, p_destination, p_status, p_cost
    ) RETURNING cdr_id INTO v_cdr_id;

    RETURN v_cdr_id;
END;
$$ LANGUAGE plpgsql;

-- Function to update CDR status
CREATE OR REPLACE FUNCTION update_cdr_status(
    p_message_id VARCHAR,
    p_status VARCHAR,
    p_dlr_status VARCHAR DEFAULT NULL,
    p_error_code VARCHAR DEFAULT NULL,
    p_error_message TEXT DEFAULT NULL
) RETURNS VOID AS $$
BEGIN
    UPDATE tbl_cdr_records
    SET status = p_status,
        dlr_status = COALESCE(p_dlr_status, dlr_status),
        error_code = COALESCE(p_error_code, error_code),
        error_message = COALESCE(p_error_message, error_message),
        delivered_timestamp = CASE WHEN p_status = 'DELIVERED' THEN NOW() ELSE delivered_timestamp END,
        failed_timestamp = CASE WHEN p_status IN ('FAILED', 'REJECTED', 'EXPIRED') THEN NOW() ELSE failed_timestamp END
    WHERE message_id = p_message_id;
END;
$$ LANGUAGE plpgsql;

-- Function to aggregate daily CDR summary
CREATE OR REPLACE FUNCTION aggregate_daily_cdr_summary(p_date DATE)
RETURNS VOID AS $$
BEGIN
    INSERT INTO tbl_cdr_daily_summary (
        report_date, customer_id, campaign_id, channel, message_type, smsc_id, status,
        total_submitted, total_delivered, total_failed, total_rejected,
        avg_delivery_time_ms, delivery_rate, total_cost
    )
    SELECT
        DATE(submission_timestamp) as report_date,
        customer_id,
        campaign_id,
        channel,
        message_type,
        smsc_id,
        status,
        COUNT(*) as total_submitted,
        COUNT(*) FILTER (WHERE status = 'DELIVERED') as total_delivered,
        COUNT(*) FILTER (WHERE status = 'FAILED') as total_failed,
        COUNT(*) FILTER (WHERE status = 'REJECTED') as total_rejected,
        AVG(delivery_time_ms) as avg_delivery_time_ms,
        CASE
            WHEN COUNT(*) > 0 THEN (COUNT(*) FILTER (WHERE status = 'DELIVERED')::DECIMAL / COUNT(*) * 100)
            ELSE 0
        END as delivery_rate,
        SUM(cost) as total_cost
    FROM tbl_cdr_records
    WHERE DATE(submission_timestamp) = p_date
    GROUP BY report_date, customer_id, campaign_id, channel, message_type, smsc_id, status
    ON CONFLICT (report_date, customer_id, campaign_id, channel, message_type, smsc_id, status)
    DO UPDATE SET
        total_submitted = EXCLUDED.total_submitted,
        total_delivered = EXCLUDED.total_delivered,
        total_failed = EXCLUDED.total_failed,
        total_rejected = EXCLUDED.total_rejected,
        avg_delivery_time_ms = EXCLUDED.avg_delivery_time_ms,
        delivery_rate = EXCLUDED.delivery_rate,
        total_cost = EXCLUDED.total_cost,
        updated_at = NOW();
END;
$$ LANGUAGE plpgsql;

-- Function to create next month's partition
CREATE OR REPLACE FUNCTION create_next_cdr_partition()
RETURNS TEXT AS $$
DECLARE
    partition_name TEXT;
    start_date DATE;
    end_date DATE;
BEGIN
    -- Next month
    start_date := DATE_TRUNC('month', NOW() + INTERVAL '1 month');
    end_date := start_date + INTERVAL '1 month';
    partition_name := 'tbl_cdr_records_' || TO_CHAR(start_date, 'YYYY_MM');

    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF tbl_cdr_records FOR VALUES FROM (%L) TO (%L)',
        partition_name, start_date, end_date
    );

    RETURN partition_name;
END;
$$ LANGUAGE plpgsql;

-- Function to drop old partitions (older than 12 months)
CREATE OR REPLACE FUNCTION drop_old_cdr_partitions(p_months_to_keep INTEGER DEFAULT 12)
RETURNS TEXT[] AS $$
DECLARE
    partition_name TEXT;
    dropped_partitions TEXT[] := '{}';
    cutoff_date DATE;
BEGIN
    cutoff_date := DATE_TRUNC('month', NOW() - (p_months_to_keep || ' months')::INTERVAL);

    FOR partition_name IN
        SELECT tablename
        FROM pg_tables
        WHERE schemaname = 'public'
          AND tablename LIKE 'tbl_cdr_records_%'
          AND tablename != 'tbl_cdr_records'
          AND TO_DATE(SUBSTRING(tablename FROM 'tbl_cdr_records_(.*)'), 'YYYY_MM') < cutoff_date
    LOOP
        EXECUTE 'DROP TABLE IF EXISTS ' || partition_name;
        dropped_partitions := array_append(dropped_partitions, partition_name);
    END LOOP;

    RETURN dropped_partitions;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- 7. TRIGGERS
-- ============================================================================

-- Trigger to calculate delivery time
CREATE OR REPLACE FUNCTION calculate_cdr_delivery_time()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'DELIVERED' AND NEW.delivered_timestamp IS NOT NULL THEN
        NEW.delivery_time_ms := EXTRACT(EPOCH FROM (NEW.delivered_timestamp - NEW.submission_timestamp)) * 1000;
        NEW.total_time_ms := NEW.delivery_time_ms;
    ELSIF NEW.status IN ('FAILED', 'REJECTED') AND NEW.failed_timestamp IS NOT NULL THEN
        NEW.total_time_ms := EXTRACT(EPOCH FROM (NEW.failed_timestamp - NEW.submission_timestamp)) * 1000;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_calculate_delivery_time
BEFORE INSERT OR UPDATE ON tbl_cdr_records
FOR EACH ROW EXECUTE FUNCTION calculate_cdr_delivery_time();

-- ============================================================================
-- 8. VIEWS FOR REPORTING
-- ============================================================================

-- Materialized view for last 7 days summary
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_cdr_weekly_summary AS
SELECT
    DATE(submission_timestamp) as report_date,
    customer_id,
    channel,
    COUNT(*) as total_messages,
    COUNT(*) FILTER (WHERE status = 'DELIVERED') as delivered,
    COUNT(*) FILTER (WHERE status = 'FAILED') as failed,
    ROUND(AVG(delivery_time_ms), 2) as avg_delivery_time_ms,
    ROUND((COUNT(*) FILTER (WHERE status = 'DELIVERED')::DECIMAL / COUNT(*) * 100), 2) as delivery_rate,
    SUM(cost) as total_cost
FROM tbl_cdr_records
WHERE submission_timestamp >= NOW() - INTERVAL '7 days'
GROUP BY report_date, customer_id, channel
ORDER BY report_date DESC, customer_id, channel;

CREATE UNIQUE INDEX ON mv_cdr_weekly_summary (report_date, customer_id, channel);

-- Refresh weekly summary every 5 minutes
-- SELECT cron.schedule('refresh-weekly-summary', '*/5 * * * *', 'REFRESH MATERIALIZED VIEW CONCURRENTLY mv_cdr_weekly_summary');

-- View for current day real-time statistics
CREATE OR REPLACE VIEW v_cdr_today AS
SELECT
    customer_id,
    channel,
    COUNT(*) as total_today,
    COUNT(*) FILTER (WHERE status = 'DELIVERED') as delivered_today,
    COUNT(*) FILTER (WHERE status = 'FAILED') as failed_today,
    COUNT(*) FILTER (WHERE status = 'PENDING') as pending_today,
    ROUND(AVG(delivery_time_ms), 2) as avg_delivery_time_ms,
    ROUND((COUNT(*) FILTER (WHERE status = 'DELIVERED')::DECIMAL / NULLIF(COUNT(*), 0) * 100), 2) as delivery_rate,
    SUM(cost) as cost_today
FROM tbl_cdr_records
WHERE DATE(submission_timestamp) = CURRENT_DATE
GROUP BY customer_id, channel;

COMMIT;
