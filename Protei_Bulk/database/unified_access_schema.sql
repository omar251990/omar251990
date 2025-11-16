-- ============================================================================
-- Protei_Bulk Unified Access & Submission Architecture
-- Database Schema Extensions
-- ============================================================================

-- ============================================================================
-- 1. UPDATE USERS TABLE FOR UNIFIED ACCESS
-- ============================================================================

-- Add unified access fields to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS api_key VARCHAR(64) UNIQUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS api_key_created_at TIMESTAMP;
ALTER TABLE users ADD COLUMN IF NOT EXISTS bind_type VARCHAR(20) DEFAULT 'BOTH' CHECK (bind_type IN ('SMPP', 'HTTP', 'BOTH', 'WEB_ONLY'));
ALTER TABLE users ADD COLUMN IF NOT EXISTS max_msg_per_day INTEGER DEFAULT 500000;
ALTER TABLE users ADD COLUMN IF NOT EXISTS allowed_smsc JSONB DEFAULT '[]';
ALTER TABLE users ADD COLUMN IF NOT EXISTS allowed_sender_ids JSONB DEFAULT '[]';
ALTER TABLE users ADD COLUMN IF NOT EXISTS can_use_smpp BOOLEAN DEFAULT TRUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS can_use_http BOOLEAN DEFAULT TRUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS can_use_api_bulk BOOLEAN DEFAULT TRUE;

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_api_key ON users(api_key);
CREATE INDEX IF NOT EXISTS idx_users_bind_type ON users(bind_type);

-- ============================================================================
-- 2. CAMPAIGNS TABLE (Extended)
-- ============================================================================

CREATE TABLE IF NOT EXISTS campaigns (
    id BIGSERIAL PRIMARY KEY,
    campaign_id VARCHAR(64) UNIQUE NOT NULL,
    customer_id BIGINT REFERENCES tbl_customers(customer_id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,

    -- Campaign Details
    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Submission Channel
    submission_channel VARCHAR(20) DEFAULT 'WEB' CHECK (submission_channel IN ('WEB', 'HTTP_API', 'SMPP', 'SCHEDULER')),
    submission_ip INET,

    -- Message Content
    sender_id VARCHAR(20) NOT NULL,
    message_content TEXT NOT NULL,
    encoding VARCHAR(20) DEFAULT 'GSM7' CHECK (encoding IN ('GSM7', 'UCS2', 'ASCII')),
    message_class VARCHAR(10) DEFAULT 'FLASH',

    -- Recipients
    total_recipients INTEGER NOT NULL,
    recipients_source VARCHAR(50), -- FILE, CONTACT_LIST, PROFILE, MANUAL, API

    -- Status
    status VARCHAR(20) DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'PENDING_APPROVAL', 'APPROVED', 'SCHEDULED', 'RUNNING', 'PAUSED', 'COMPLETED', 'FAILED', 'CANCELLED')),

    -- Schedule
    schedule_type VARCHAR(20) DEFAULT 'IMMEDIATE' CHECK (schedule_type IN ('IMMEDIATE', 'SCHEDULED', 'RECURRING')),
    scheduled_time TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,

    -- Limits
    max_messages_per_day INTEGER,
    priority VARCHAR(20) DEFAULT 'NORMAL' CHECK (priority IN ('CRITICAL', 'HIGH', 'NORMAL', 'LOW')),

    -- DLR Settings
    dlr_required BOOLEAN DEFAULT TRUE,
    dlr_callback_url VARCHAR(500),

    -- Maker-Checker
    created_by BIGINT REFERENCES users(id),
    approved_by BIGINT REFERENCES users(id),
    approved_at TIMESTAMP,

    -- Audit
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    -- Metadata
    metadata JSONB DEFAULT '{}'
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_campaigns_customer ON campaigns(customer_id);
CREATE INDEX IF NOT EXISTS idx_campaigns_user ON campaigns(user_id);
CREATE INDEX IF NOT EXISTS idx_campaigns_status ON campaigns(status);
CREATE INDEX IF NOT EXISTS idx_campaigns_scheduled ON campaigns(scheduled_time);
CREATE INDEX IF NOT EXISTS idx_campaigns_created ON campaigns(created_at);

-- ============================================================================
-- 3. MESSAGES TABLE (Extended for All Channels)
-- ============================================================================

CREATE TABLE IF NOT EXISTS messages (
    id BIGSERIAL PRIMARY KEY,
    message_id VARCHAR(64) UNIQUE NOT NULL,
    customer_id BIGINT REFERENCES tbl_customers(customer_id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    campaign_id BIGINT REFERENCES campaigns(id) ON DELETE CASCADE,

    -- Message Details
    from_addr VARCHAR(20) NOT NULL,
    to_addr VARCHAR(20) NOT NULL,
    message_text TEXT,
    encoding VARCHAR(20) DEFAULT 'GSM7',

    -- Submission
    submission_channel VARCHAR(20) DEFAULT 'WEB',
    submission_timestamp TIMESTAMP DEFAULT NOW(),
    submission_ip INET,

    -- SMPP Specific
    smpp_msg_id VARCHAR(64),
    smpp_system_id VARCHAR(50),
    esm_class INTEGER,
    protocol_id INTEGER,
    priority_flag INTEGER,
    data_coding INTEGER,

    -- Status
    status VARCHAR(20) DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'QUEUED', 'SENT', 'DELIVERED', 'FAILED', 'REJECTED', 'EXPIRED')),

    -- Routing
    smsc_id VARCHAR(50),
    route_id VARCHAR(50),

    -- Timestamps
    queued_at TIMESTAMP,
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,

    -- DLR
    dlr_status VARCHAR(50),
    dlr_timestamp TIMESTAMP,
    dlr_text TEXT,
    error_code VARCHAR(20),

    -- Billing
    cost DECIMAL(10, 4) DEFAULT 0.0,
    parts INTEGER DEFAULT 1,

    -- Metadata
    metadata JSONB DEFAULT '{}'
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_messages_customer ON messages(customer_id);
CREATE INDEX IF NOT EXISTS idx_messages_user ON messages(user_id);
CREATE INDEX IF NOT EXISTS idx_messages_campaign ON messages(campaign_id);
CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status);
CREATE INDEX IF NOT EXISTS idx_messages_to_addr ON messages(to_addr);
CREATE INDEX IF NOT EXISTS idx_messages_created ON messages(submission_timestamp);
CREATE INDEX IF NOT EXISTS idx_messages_smpp_msg_id ON messages(smpp_msg_id);

-- Partition by month for scalability
CREATE TABLE IF NOT EXISTS messages_partitioned (LIKE messages INCLUDING ALL)
PARTITION BY RANGE (submission_timestamp);

-- ============================================================================
-- 4. SMPP SESSIONS TABLE
-- ============================================================================

CREATE TABLE tbl_smpp_sessions (
    session_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    customer_id BIGINT REFERENCES tbl_customers(customer_id) ON DELETE CASCADE,

    -- Connection Details
    system_id VARCHAR(50) NOT NULL,
    bind_type VARCHAR(20) NOT NULL CHECK (bind_type IN ('TRANSMITTER', 'RECEIVER', 'TRANSCEIVER')),
    remote_ip INET NOT NULL,
    remote_port INTEGER,

    -- Session Info
    session_token VARCHAR(64) UNIQUE NOT NULL,
    smpp_version VARCHAR(10) DEFAULT '3.4',

    -- Status
    status VARCHAR(20) DEFAULT 'BOUND' CHECK (status IN ('BOUND', 'DISCONNECTED', 'SUSPENDED')),

    -- Throughput
    current_tps DECIMAL(10, 2) DEFAULT 0.0,
    messages_sent INTEGER DEFAULT 0,
    messages_received INTEGER DEFAULT 0,

    -- Timestamps
    bound_at TIMESTAMP DEFAULT NOW(),
    last_activity_at TIMESTAMP DEFAULT NOW(),
    disconnected_at TIMESTAMP,

    -- Metadata
    metadata JSONB DEFAULT '{}'
);

-- Create indexes
CREATE INDEX idx_smpp_sessions_user ON tbl_smpp_sessions(user_id);
CREATE INDEX idx_smpp_sessions_customer ON tbl_smpp_sessions(customer_id);
CREATE INDEX idx_smpp_sessions_status ON tbl_smpp_sessions(status);
CREATE INDEX idx_smpp_sessions_token ON tbl_smpp_sessions(session_token);

-- ============================================================================
-- 5. API REQUESTS LOG TABLE
-- ============================================================================

CREATE TABLE tbl_api_requests (
    request_id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT REFERENCES tbl_customers(customer_id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,

    -- Request Details
    endpoint VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    request_body TEXT,

    -- Authentication
    auth_type VARCHAR(20) CHECK (auth_type IN ('API_KEY', 'BASIC_AUTH', 'JWT', 'SMPP')),
    api_key VARCHAR(64),

    -- Source
    ip_address INET,
    user_agent TEXT,

    -- Response
    response_status INTEGER,
    response_body TEXT,
    response_time_ms INTEGER,

    -- Campaign Created
    campaign_id BIGINT REFERENCES campaigns(id),

    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),

    -- Metadata
    metadata JSONB DEFAULT '{}'
);

-- Create indexes
CREATE INDEX idx_api_requests_customer ON tbl_api_requests(customer_id);
CREATE INDEX idx_api_requests_user ON tbl_api_requests(user_id);
CREATE INDEX idx_api_requests_created ON tbl_api_requests(created_at);
CREATE INDEX idx_api_requests_endpoint ON tbl_api_requests(endpoint);

-- ============================================================================
-- 6. DLR (DELIVERY REPORTS) TABLE
-- ============================================================================

CREATE TABLE tbl_delivery_reports (
    dlr_id BIGSERIAL PRIMARY KEY,
    message_id BIGINT REFERENCES messages(id) ON DELETE CASCADE,
    customer_id BIGINT REFERENCES tbl_customers(customer_id) ON DELETE CASCADE,

    -- Message Reference
    external_msg_id VARCHAR(64),
    smpp_msg_id VARCHAR(64),

    -- Recipient
    msisdn VARCHAR(20) NOT NULL,

    -- DLR Details
    dlr_status VARCHAR(50) NOT NULL,
    dlr_code VARCHAR(20),
    dlr_text TEXT,

    -- Source
    smsc_id VARCHAR(50),
    received_from VARCHAR(100),

    -- Timestamps
    submit_time TIMESTAMP,
    done_time TIMESTAMP,
    received_at TIMESTAMP DEFAULT NOW(),

    -- Callback
    callback_sent BOOLEAN DEFAULT FALSE,
    callback_url VARCHAR(500),
    callback_response TEXT,
    callback_at TIMESTAMP,

    -- Metadata
    metadata JSONB DEFAULT '{}'
);

-- Create indexes
CREATE INDEX idx_dlr_message ON tbl_delivery_reports(message_id);
CREATE INDEX idx_dlr_customer ON tbl_delivery_reports(customer_id);
CREATE INDEX idx_dlr_external_msg_id ON tbl_delivery_reports(external_msg_id);
CREATE INDEX idx_dlr_smpp_msg_id ON tbl_delivery_reports(smpp_msg_id);
CREATE INDEX idx_dlr_received ON tbl_delivery_reports(received_at);

-- ============================================================================
-- 7. MESSAGE QUEUE TABLE
-- ============================================================================

CREATE TABLE tbl_message_queue (
    queue_id BIGSERIAL PRIMARY KEY,
    message_id BIGINT REFERENCES messages(id) ON DELETE CASCADE,
    campaign_id BIGINT REFERENCES campaigns(id) ON DELETE CASCADE,
    customer_id BIGINT REFERENCES tbl_customers(customer_id) ON DELETE CASCADE,

    -- Queue Details
    queue_priority INTEGER DEFAULT 5, -- 1=highest, 10=lowest
    queue_status VARCHAR(20) DEFAULT 'PENDING' CHECK (queue_status IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED', 'RETRY')),

    -- Retry Logic
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    next_retry_at TIMESTAMP,

    -- Processing
    worker_id VARCHAR(64),
    picked_at TIMESTAMP,
    processed_at TIMESTAMP,

    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    -- Error Tracking
    error_message TEXT,

    -- Metadata
    metadata JSONB DEFAULT '{}'
);

-- Create indexes
CREATE INDEX idx_queue_message ON tbl_message_queue(message_id);
CREATE INDEX idx_queue_campaign ON tbl_message_queue(campaign_id);
CREATE INDEX idx_queue_customer ON tbl_message_queue(customer_id);
CREATE INDEX idx_queue_status ON tbl_message_queue(queue_status);
CREATE INDEX idx_queue_priority ON tbl_message_queue(queue_priority);
CREATE INDEX idx_queue_next_retry ON tbl_message_queue(next_retry_at);

-- ============================================================================
-- 8. QUOTA TRACKING TABLE
-- ============================================================================

CREATE TABLE tbl_quota_usage (
    usage_id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT REFERENCES tbl_customers(customer_id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,

    -- Usage Period
    period_date DATE NOT NULL,
    period_type VARCHAR(20) DEFAULT 'DAILY' CHECK (period_type IN ('HOURLY', 'DAILY', 'MONTHLY')),

    -- Messages
    messages_sent INTEGER DEFAULT 0,
    messages_delivered INTEGER DEFAULT 0,
    messages_failed INTEGER DEFAULT 0,

    -- TPS Tracking
    peak_tps DECIMAL(10, 2) DEFAULT 0.0,
    avg_tps DECIMAL(10, 2) DEFAULT 0.0,

    -- Cost
    total_cost DECIMAL(15, 4) DEFAULT 0.0,

    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    UNIQUE (user_id, period_date, period_type)
);

-- Create indexes
CREATE INDEX idx_quota_customer ON tbl_quota_usage(customer_id);
CREATE INDEX idx_quota_user ON tbl_quota_usage(user_id);
CREATE INDEX idx_quota_period ON tbl_quota_usage(period_date);

-- ============================================================================
-- 9. HELPER FUNCTIONS
-- ============================================================================

-- Function to generate API key
CREATE OR REPLACE FUNCTION generate_api_key() RETURNS VARCHAR AS $$
BEGIN
    RETURN encode(gen_random_bytes(32), 'hex');
END;
$$ LANGUAGE plpgsql;

-- Function to check daily quota
CREATE OR REPLACE FUNCTION check_daily_quota(p_user_id BIGINT) RETURNS BOOLEAN AS $$
DECLARE
    v_max_msg_per_day INTEGER;
    v_messages_today INTEGER;
    v_today DATE := CURRENT_DATE;
BEGIN
    -- Get user's daily limit
    SELECT max_msg_per_day INTO v_max_msg_per_day
    FROM users
    WHERE id = p_user_id;

    -- Get messages sent today
    SELECT COALESCE(messages_sent, 0) INTO v_messages_today
    FROM tbl_quota_usage
    WHERE user_id = p_user_id
      AND period_date = v_today
      AND period_type = 'DAILY';

    -- Check if under quota
    RETURN (v_messages_today < v_max_msg_per_day);
END;
$$ LANGUAGE plpgsql;

-- Function to update quota usage
CREATE OR REPLACE FUNCTION update_quota_usage(
    p_user_id BIGINT,
    p_customer_id BIGINT,
    p_messages_count INTEGER DEFAULT 1
) RETURNS VOID AS $$
DECLARE
    v_today DATE := CURRENT_DATE;
BEGIN
    INSERT INTO tbl_quota_usage (customer_id, user_id, period_date, period_type, messages_sent)
    VALUES (p_customer_id, p_user_id, v_today, 'DAILY', p_messages_count)
    ON CONFLICT (user_id, period_date, period_type)
    DO UPDATE SET
        messages_sent = tbl_quota_usage.messages_sent + p_messages_count,
        updated_at = NOW();
END;
$$ LANGUAGE plpgsql;

-- Function to get campaign statistics
CREATE OR REPLACE FUNCTION get_campaign_stats(p_campaign_id BIGINT)
RETURNS TABLE (
    total INTEGER,
    pending INTEGER,
    queued INTEGER,
    sent INTEGER,
    delivered INTEGER,
    failed INTEGER,
    progress_percentage DECIMAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COUNT(*)::INTEGER as total,
        SUM(CASE WHEN status = 'PENDING' THEN 1 ELSE 0 END)::INTEGER as pending,
        SUM(CASE WHEN status = 'QUEUED' THEN 1 ELSE 0 END)::INTEGER as queued,
        SUM(CASE WHEN status = 'SENT' THEN 1 ELSE 0 END)::INTEGER as sent,
        SUM(CASE WHEN status = 'DELIVERED' THEN 1 ELSE 0 END)::INTEGER as delivered,
        SUM(CASE WHEN status = 'FAILED' THEN 1 ELSE 0 END)::INTEGER as failed,
        CASE
            WHEN COUNT(*) > 0 THEN
                (SUM(CASE WHEN status IN ('SENT', 'DELIVERED') THEN 1 ELSE 0 END)::DECIMAL / COUNT(*) * 100)
            ELSE 0
        END as progress_percentage
    FROM messages
    WHERE campaign_id = p_campaign_id;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- 10. TRIGGERS
-- ============================================================================

-- Trigger to update campaign statistics
CREATE OR REPLACE FUNCTION update_campaign_on_message()
RETURNS TRIGGER AS $$
BEGIN
    -- Update campaign updated_at
    UPDATE campaigns
    SET updated_at = NOW()
    WHERE id = NEW.campaign_id;

    -- Update quota
    PERFORM update_quota_usage(NEW.user_id, NEW.customer_id, 1);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_campaign_on_message
AFTER INSERT ON messages
FOR EACH ROW EXECUTE FUNCTION update_campaign_on_message();

-- Trigger to update SMPP session activity
CREATE OR REPLACE FUNCTION update_smpp_session_activity()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.smpp_system_id IS NOT NULL THEN
        UPDATE tbl_smpp_sessions
        SET last_activity_at = NOW(),
            messages_sent = messages_sent + 1
        WHERE system_id = NEW.smpp_system_id
          AND status = 'BOUND';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_smpp_activity
AFTER INSERT ON messages
FOR EACH ROW EXECUTE FUNCTION update_smpp_session_activity();

COMMIT;
