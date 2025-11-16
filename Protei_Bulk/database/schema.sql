-- =============================================
-- Protei_Bulk Database Schema (PostgreSQL)
-- Full Enterprise Messaging Platform
-- =============================================

-- ==================
-- 1. USER & ACCOUNT MANAGEMENT
-- ==================

CREATE TABLE account_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    hierarchy_level INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO account_types (name, description, hierarchy_level) VALUES
('SUPER_ADMIN', 'System super administrator', 0),
('ADMIN', 'Platform administrator', 1),
('RESELLER', 'Reseller account', 2),
('SELLER', 'Seller account', 3),
('END_USER', 'End user account', 4);

CREATE TABLE accounts (
    id BIGSERIAL PRIMARY KEY,
    account_id VARCHAR(64) UNIQUE NOT NULL,
    account_type_id INTEGER REFERENCES account_types(id),
    parent_account_id BIGINT REFERENCES accounts(id),
    company_name VARCHAR(255),
    business_name VARCHAR(255),
    account_status VARCHAR(20) DEFAULT 'ACTIVE',

    -- Billing & Credit
    billing_type VARCHAR(20) DEFAULT 'PREPAID', -- PREPAID, POSTPAID
    credit_limit DECIMAL(15,2) DEFAULT 0,
    current_balance DECIMAL(15,2) DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'USD',
    low_balance_threshold DECIMAL(15,2) DEFAULT 100,

    -- Sender Configuration
    free_sender BOOLEAN DEFAULT FALSE,
    allowed_sender_ids TEXT[], -- Array of allowed sender IDs
    blocked_sender_ids TEXT[], -- Array of blocked sender IDs
    default_sender_id VARCHAR(20),

    -- Routing & SMSC
    bound_smsc_ids INTEGER[], -- Specific SMSCs for this account
    routing_policy VARCHAR(50) DEFAULT 'ROUND_ROBIN',

    -- Throughput Limits
    max_tps INTEGER DEFAULT 100,
    max_concurrent_connections INTEGER DEFAULT 10,
    max_messages_per_day INTEGER,

    -- Quotas
    daily_quota INTEGER,
    monthly_quota INTEGER,
    used_today INTEGER DEFAULT 0,
    used_this_month INTEGER DEFAULT 0,

    -- Contact Info
    contact_name VARCHAR(255),
    contact_email VARCHAR(255),
    contact_phone VARCHAR(50),
    address TEXT,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,

    -- Metadata
    metadata JSONB,

    CONSTRAINT chk_billing_type CHECK (billing_type IN ('PREPAID', 'POSTPAID')),
    CONSTRAINT chk_account_status CHECK (account_status IN ('ACTIVE', 'SUSPENDED', 'BLOCKED', 'EXPIRED'))
);

CREATE INDEX idx_accounts_parent ON accounts(parent_account_id);
CREATE INDEX idx_accounts_status ON accounts(account_status);
CREATE INDEX idx_accounts_type ON accounts(account_type_id);

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(64) UNIQUE NOT NULL,
    account_id BIGINT REFERENCES accounts(id) ON DELETE CASCADE,
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,

    -- Profile
    full_name VARCHAR(255),
    phone VARCHAR(50),
    language VARCHAR(10) DEFAULT 'en',
    timezone VARCHAR(50) DEFAULT 'UTC',

    -- Status & Flags
    status VARCHAR(20) DEFAULT 'ACTIVE',
    email_verified BOOLEAN DEFAULT FALSE,
    phone_verified BOOLEAN DEFAULT FALSE,
    must_change_password BOOLEAN DEFAULT TRUE,

    -- Security
    two_factor_enabled BOOLEAN DEFAULT FALSE,
    two_factor_method VARCHAR(20), -- SMS, EMAIL, TOTP
    two_factor_secret VARCHAR(255),
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP,
    password_changed_at TIMESTAMP,
    password_expires_at TIMESTAMP,

    -- Session
    last_login_at TIMESTAMP,
    last_login_ip VARCHAR(50),
    last_activity_at TIMESTAMP,

    -- API Access
    api_enabled BOOLEAN DEFAULT FALSE,
    api_key VARCHAR(64) UNIQUE,
    api_key_expires_at TIMESTAMP,

    -- SMPP Access
    smpp_enabled BOOLEAN DEFAULT FALSE,
    smpp_username VARCHAR(100),
    smpp_password VARCHAR(255),
    smpp_system_type VARCHAR(20),

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT REFERENCES users(id),

    CONSTRAINT chk_user_status CHECK (status IN ('ACTIVE', 'SUSPENDED', 'LOCKED', 'DELETED')),
    CONSTRAINT chk_2fa_method CHECK (two_factor_method IN ('SMS', 'EMAIL', 'TOTP', NULL))
);

CREATE INDEX idx_users_account ON users(account_id);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_api_key ON users(api_key);
CREATE INDEX idx_users_status ON users(status);

-- ==================
-- 2. ROLES & PERMISSIONS (RBAC)
-- ==================

CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    role_name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    is_system_role BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO roles (role_name, display_name, description, is_system_role) VALUES
('SUPER_ADMIN', 'Super Administrator', 'Full system access', TRUE),
('ADMIN', 'Administrator', 'Platform administration', TRUE),
('RESELLER_ADMIN', 'Reseller Admin', 'Reseller management', TRUE),
('SELLER_ADMIN', 'Seller Admin', 'Seller operations', TRUE),
('USER', 'End User', 'Basic message sending', TRUE),
('APPROVER', 'Approver', 'Campaign approval authority', TRUE),
('VIEWER', 'Viewer', 'Read-only access', TRUE),
('API_USER', 'API User', 'API access only', TRUE);

CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    permission_key VARCHAR(100) UNIQUE NOT NULL,
    module VARCHAR(50),
    action VARCHAR(50),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert comprehensive permissions
INSERT INTO permissions (permission_key, module, action, description) VALUES
-- User Management
('users.create', 'users', 'create', 'Create new users'),
('users.read', 'users', 'read', 'View user details'),
('users.update', 'users', 'update', 'Edit user information'),
('users.delete', 'users', 'delete', 'Delete users'),
('users.manage_roles', 'users', 'manage_roles', 'Assign/remove roles'),

-- Account Management
('accounts.create', 'accounts', 'create', 'Create new accounts'),
('accounts.read', 'accounts', 'read', 'View account details'),
('accounts.update', 'accounts', 'update', 'Edit account information'),
('accounts.delete', 'accounts', 'delete', 'Delete accounts'),
('accounts.manage_credit', 'accounts', 'manage_credit', 'Manage account credits'),

-- Messages
('messages.send', 'messages', 'send', 'Send messages'),
('messages.send_bulk', 'messages', 'send_bulk', 'Send bulk messages'),
('messages.read', 'messages', 'read', 'View sent messages'),
('messages.cancel', 'messages', 'cancel', 'Cancel scheduled messages'),

-- Campaigns
('campaigns.create', 'campaigns', 'create', 'Create campaigns'),
('campaigns.read', 'campaigns', 'read', 'View campaigns'),
('campaigns.update', 'campaigns', 'update', 'Edit campaigns'),
('campaigns.delete', 'campaigns', 'delete', 'Delete campaigns'),
('campaigns.approve', 'campaigns', 'approve', 'Approve campaigns'),
('campaigns.start', 'campaigns', 'start', 'Start campaigns'),
('campaigns.pause', 'campaigns', 'pause', 'Pause campaigns'),
('campaigns.stop', 'campaigns', 'stop', 'Stop campaigns'),

-- Templates
('templates.create', 'templates', 'create', 'Create templates'),
('templates.read', 'templates', 'read', 'View templates'),
('templates.update', 'templates', 'update', 'Edit templates'),
('templates.delete', 'templates', 'delete', 'Delete templates'),

-- Reports
('reports.messages', 'reports', 'messages', 'View message reports'),
('reports.campaigns', 'reports', 'campaigns', 'View campaign reports'),
('reports.accounts', 'reports', 'accounts', 'View account reports'),
('reports.system', 'reports', 'system', 'View system reports'),
('reports.export', 'reports', 'export', 'Export reports'),

-- System
('system.config', 'system', 'config', 'System configuration'),
('system.monitoring', 'system', 'monitoring', 'System monitoring'),
('system.logs', 'system', 'logs', 'View system logs'),
('system.audit', 'system', 'audit', 'View audit logs');

CREATE TABLE role_permissions (
    role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INTEGER REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE user_roles (
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    assigned_by BIGINT REFERENCES users(id),
    PRIMARY KEY (user_id, role_id)
);

-- ==================
-- 3. SMSC & ROUTING
-- ==================

CREATE TABLE smsc_connections (
    id SERIAL PRIMARY KEY,
    smsc_id VARCHAR(50) UNIQUE NOT NULL,
    smsc_name VARCHAR(100),
    smsc_type VARCHAR(20), -- SMPP, UCP, HTTP, SIGTRAN

    -- Connection Details
    host VARCHAR(255),
    port INTEGER,
    system_id VARCHAR(100),
    password VARCHAR(255),
    system_type VARCHAR(20),
    interface_version VARCHAR(10) DEFAULT '3.4',

    -- Protocol Config
    bind_type VARCHAR(20) DEFAULT 'TRANSCEIVER', -- TRANSMITTER, RECEIVER, TRANSCEIVER
    enquire_link_interval INTEGER DEFAULT 30,
    request_timeout INTEGER DEFAULT 30,
    window_size INTEGER DEFAULT 100,

    -- Throughput
    max_tps INTEGER DEFAULT 1000,
    current_tps INTEGER DEFAULT 0,

    -- Status
    status VARCHAR(20) DEFAULT 'ACTIVE',
    connection_state VARCHAR(20) DEFAULT 'DISCONNECTED',
    last_connected_at TIMESTAMP,
    last_error TEXT,

    -- Priority & Weight
    priority INTEGER DEFAULT 100,
    weight INTEGER DEFAULT 1,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_smsc_status CHECK (status IN ('ACTIVE', 'INACTIVE', 'MAINTENANCE')),
    CONSTRAINT chk_smsc_state CHECK (connection_state IN ('CONNECTED', 'DISCONNECTED', 'CONNECTING', 'ERROR'))
);

CREATE TABLE routing_rules (
    id SERIAL PRIMARY KEY,
    rule_name VARCHAR(100) NOT NULL,
    rule_type VARCHAR(50), -- PREFIX, ACCOUNT, SENDER, TRAFFIC_TYPE
    priority INTEGER DEFAULT 100,

    -- Match Conditions
    msisdn_prefix VARCHAR(20),
    account_ids BIGINT[],
    sender_id_pattern VARCHAR(100),
    traffic_type VARCHAR(50), -- LOCAL, INTERNATIONAL, PREMIUM, OTP
    message_category VARCHAR(50),

    -- Target SMSC
    target_smsc_ids INTEGER[],
    routing_strategy VARCHAR(50) DEFAULT 'ROUND_ROBIN', -- ROUND_ROBIN, LEAST_LOAD, PRIORITY, FAILOVER

    -- Status
    enabled BOOLEAN DEFAULT TRUE,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_routing_rules_priority ON routing_rules(priority DESC);
CREATE INDEX idx_routing_rules_enabled ON routing_rules(enabled);

-- ==================
-- 4. MESSAGES & CAMPAIGNS
-- ==================

CREATE TABLE message_templates (
    id BIGSERIAL PRIMARY KEY,
    template_id VARCHAR(64) UNIQUE NOT NULL,
    account_id BIGINT REFERENCES accounts(id),
    user_id BIGINT REFERENCES users(id),

    name VARCHAR(255) NOT NULL,
    category VARCHAR(50), -- OTP, PROMO, ALERT, TRANSACTIONAL, MARKETING
    language VARCHAR(10) DEFAULT 'en',

    -- Content
    content TEXT NOT NULL,
    variables TEXT[], -- Array of variable names: [name, code, amount]
    encoding VARCHAR(20) DEFAULT 'GSM7', -- GSM7, UCS2, ASCII

    -- Metadata
    description TEXT,
    tags TEXT[],
    is_approved BOOLEAN DEFAULT FALSE,
    approved_by BIGINT REFERENCES users(id),
    approved_at TIMESTAMP,

    -- Usage Stats
    usage_count INTEGER DEFAULT 0,
    last_used_at TIMESTAMP,

    -- Status
    status VARCHAR(20) DEFAULT 'ACTIVE',

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_template_status CHECK (status IN ('ACTIVE', 'INACTIVE', 'DELETED'))
);

CREATE INDEX idx_templates_account ON message_templates(account_id);
CREATE INDEX idx_templates_category ON message_templates(category);
CREATE INDEX idx_templates_status ON message_templates(status);

CREATE TABLE msisdn_lists (
    id BIGSERIAL PRIMARY KEY,
    list_id VARCHAR(64) UNIQUE NOT NULL,
    account_id BIGINT REFERENCES accounts(id),
    user_id BIGINT REFERENCES users(id),

    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Privacy
    is_hidden BOOLEAN DEFAULT FALSE, -- Hidden lists: users can use but not view numbers

    -- Stats
    total_count INTEGER DEFAULT 0,
    active_count INTEGER DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE msisdn_list_entries (
    id BIGSERIAL PRIMARY KEY,
    list_id BIGINT REFERENCES msisdn_lists(id) ON DELETE CASCADE,
    msisdn VARCHAR(20) NOT NULL,

    -- Optional Fields
    name VARCHAR(255),
    attributes JSONB, -- Custom attributes: {age: 25, city: "Amman", ...}

    -- Status
    status VARCHAR(20) DEFAULT 'ACTIVE',

    -- Timestamps
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_entry_status CHECK (status IN ('ACTIVE', 'INACTIVE', 'BLACKLISTED'))
);

CREATE INDEX idx_list_entries_list ON msisdn_list_entries(list_id);
CREATE INDEX idx_list_entries_msisdn ON msisdn_list_entries(msisdn);
CREATE INDEX idx_list_entries_status ON msisdn_list_entries(status);

CREATE TABLE campaigns (
    id BIGSERIAL PRIMARY KEY,
    campaign_id VARCHAR(64) UNIQUE NOT NULL,
    account_id BIGINT REFERENCES accounts(id),
    user_id BIGINT REFERENCES users(id),

    -- Basic Info
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(50),

    -- Message Content
    sender_id VARCHAR(20),
    message_content TEXT,
    template_id BIGINT REFERENCES message_templates(id),
    encoding VARCHAR(20) DEFAULT 'GSM7',

    -- Recipients
    recipient_type VARCHAR(50), -- LIST, UPLOAD, PROFILE, MANUAL
    list_ids BIGINT[], -- Reference to msisdn_lists
    uploaded_file_path VARCHAR(500),
    profile_filter JSONB, -- Profile-based filter criteria
    total_recipients INTEGER DEFAULT 0,

    -- Scheduling
    schedule_type VARCHAR(50) DEFAULT 'IMMEDIATE', -- IMMEDIATE, SCHEDULED, RECURRING
    scheduled_at TIMESTAMP,
    recurring_pattern VARCHAR(100), -- Cron expression or pattern
    completed_at TIMESTAMP,

    -- Throttling & Control
    max_tps INTEGER DEFAULT 100,
    max_per_day_per_msisdn INTEGER DEFAULT 1, -- Prevent duplicate sends
    priority VARCHAR(20) DEFAULT 'NORMAL', -- CRITICAL, HIGH, NORMAL, LOW

    -- Status & Progress
    status VARCHAR(50) DEFAULT 'DRAFT',
    campaign_state VARCHAR(50) DEFAULT 'PENDING',

    -- Counters
    sent_count INTEGER DEFAULT 0,
    delivered_count INTEGER DEFAULT 0,
    failed_count INTEGER DEFAULT 0,
    pending_count INTEGER DEFAULT 0,

    -- Maker-Checker
    requires_approval BOOLEAN DEFAULT TRUE,
    approved BOOLEAN DEFAULT FALSE,
    approved_by BIGINT REFERENCES users(id),
    approved_at TIMESTAMP,
    rejection_reason TEXT,

    -- Cost
    estimated_cost DECIMAL(15,2),
    actual_cost DECIMAL(15,2),

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,

    CONSTRAINT chk_campaign_status CHECK (status IN ('DRAFT', 'PENDING_APPROVAL', 'APPROVED', 'REJECTED', 'SCHEDULED', 'RUNNING', 'PAUSED', 'COMPLETED', 'CANCELLED', 'FAILED')),
    CONSTRAINT chk_priority CHECK (priority IN ('CRITICAL', 'HIGH', 'NORMAL', 'LOW'))
);

CREATE INDEX idx_campaigns_account ON campaigns(account_id);
CREATE INDEX idx_campaigns_user ON campaigns(user_id);
CREATE INDEX idx_campaigns_status ON campaigns(status);
CREATE INDEX idx_campaigns_scheduled ON campaigns(scheduled_at);

CREATE TABLE messages (
    id BIGSERIAL PRIMARY KEY,
    message_id VARCHAR(64) UNIQUE NOT NULL,
    account_id BIGINT REFERENCES accounts(id),
    user_id BIGINT REFERENCES users(id),
    campaign_id BIGINT REFERENCES campaigns(id),

    -- Source
    source_type VARCHAR(20), -- API, SMPP, WEB, CAMPAIGN
    source_ip VARCHAR(50),

    -- Message Details
    sender_id VARCHAR(20),
    recipient VARCHAR(20) NOT NULL,
    message_text TEXT,
    encoding VARCHAR(20) DEFAULT 'GSM7',

    -- Classification
    message_type VARCHAR(50), -- OTP, PROMO, TRANSACTIONAL, ALERT
    priority VARCHAR(20) DEFAULT 'NORMAL',

    -- Routing
    smsc_id INTEGER REFERENCES smsc_connections(id),
    route_selected_by VARCHAR(100),

    -- Status & Delivery
    message_status VARCHAR(50) DEFAULT 'PENDING',
    dlr_status VARCHAR(50),
    dlr_code VARCHAR(10),
    dlr_message TEXT,
    error_code VARCHAR(20),
    error_message TEXT,

    -- Billing
    message_parts INTEGER DEFAULT 1,
    cost DECIMAL(10,4),
    currency VARCHAR(3) DEFAULT 'USD',

    -- Timestamps
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    queued_at TIMESTAMP,
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    failed_at TIMESTAMP,

    -- Retry
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    next_retry_at TIMESTAMP,

    -- Metadata
    metadata JSONB,

    CONSTRAINT chk_message_status CHECK (message_status IN ('PENDING', 'QUEUED', 'SENT', 'DELIVERED', 'FAILED', 'REJECTED', 'EXPIRED', 'CANCELLED'))
);

CREATE INDEX idx_messages_account ON messages(account_id);
CREATE INDEX idx_messages_user ON messages(user_id);
CREATE INDEX idx_messages_campaign ON messages(campaign_id);
CREATE INDEX idx_messages_recipient ON messages(recipient);
CREATE INDEX idx_messages_status ON messages(message_status);
CREATE INDEX idx_messages_submitted ON messages(submitted_at);
CREATE INDEX idx_messages_message_id ON messages(message_id);

-- ==================
-- 5. PROFILES & SEGMENTATION
-- ==================

CREATE TABLE subscriber_profiles (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT REFERENCES accounts(id),
    msisdn VARCHAR(20) NOT NULL,

    -- Profile Attributes
    attributes JSONB, -- Flexible JSON storage for any attributes

    -- Examples of attributes in JSONB:
    -- {
    --   "age": 25,
    --   "gender": "M",
    --   "city": "Amman",
    --   "subscription_type": "premium",
    --   "interests": ["sports", "news"],
    --   "last_purchase_date": "2025-01-15"
    -- }

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(account_id, msisdn)
);

CREATE INDEX idx_profiles_account ON subscriber_profiles(account_id);
CREATE INDEX idx_profiles_msisdn ON subscriber_profiles(msisdn);
CREATE INDEX idx_profiles_attributes ON subscriber_profiles USING GIN(attributes);

-- ==================
-- 6. DELIVERY REPORTS & CDR
-- ==================

CREATE TABLE delivery_reports (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGINT REFERENCES messages(id),

    -- DLR Details
    dlr_type VARCHAR(20), -- INTERMEDIATE, FINAL
    dlr_status VARCHAR(50),
    dlr_code VARCHAR(10),
    dlr_text TEXT,

    -- Source
    smsc_id INTEGER REFERENCES smsc_connections(id),
    smsc_message_id VARCHAR(100),

    -- Callback
    callback_url VARCHAR(500),
    callback_sent BOOLEAN DEFAULT FALSE,
    callback_response TEXT,

    -- Timestamp
    received_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_dlr_message ON delivery_reports(message_id);
CREATE INDEX idx_dlr_received ON delivery_reports(received_at);

CREATE TABLE cdr_records (
    id BIGSERIAL PRIMARY KEY,

    -- Message Reference
    message_id VARCHAR(64),
    campaign_id VARCHAR(64),
    account_id BIGINT,
    user_id BIGINT,

    -- Message Details
    sender_id VARCHAR(20),
    recipient VARCHAR(20),
    message_text TEXT,
    message_parts INTEGER DEFAULT 1,
    encoding VARCHAR(20),

    -- Routing & Delivery
    smsc_id INTEGER,
    smsc_name VARCHAR(100),
    protocol VARCHAR(20),

    -- Status
    status VARCHAR(50),
    dlr_status VARCHAR(50),
    dlr_code VARCHAR(10),
    error_code VARCHAR(20),

    -- Timing
    submitted_at TIMESTAMP,
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    processing_time_ms INTEGER,

    -- Billing
    cost DECIMAL(10,4),
    currency VARCHAR(3),

    -- Source
    source_type VARCHAR(20),
    source_ip VARCHAR(50),

    -- Metadata
    metadata JSONB,

    -- Timestamp
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Partition CDR table by month for better performance
CREATE INDEX idx_cdr_created ON cdr_records(created_at);
CREATE INDEX idx_cdr_message_id ON cdr_records(message_id);
CREATE INDEX idx_cdr_account ON cdr_records(account_id);
CREATE INDEX idx_cdr_recipient ON cdr_records(recipient);

-- ==================
-- 7. AUDIT & SECURITY
-- ==================

CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,

    -- User & Action
    user_id BIGINT REFERENCES users(id),
    username VARCHAR(100),
    account_id BIGINT REFERENCES accounts(id),

    -- Action Details
    action VARCHAR(100) NOT NULL,
    module VARCHAR(50),
    entity_type VARCHAR(50),
    entity_id VARCHAR(100),

    -- Request Details
    ip_address VARCHAR(50),
    user_agent TEXT,
    request_method VARCHAR(10),
    request_url TEXT,
    request_body JSONB,

    -- Response
    response_status INTEGER,
    response_body JSONB,

    -- Changes
    old_values JSONB,
    new_values JSONB,

    -- Result
    success BOOLEAN,
    error_message TEXT,

    -- Timestamp
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_user ON audit_logs(user_id);
CREATE INDEX idx_audit_account ON audit_logs(account_id);
CREATE INDEX idx_audit_action ON audit_logs(action);
CREATE INDEX idx_audit_created ON audit_logs(created_at);

CREATE TABLE blacklist (
    id BIGSERIAL PRIMARY KEY,

    -- Entry Details
    entry_type VARCHAR(20), -- MSISDN, IP, SENDER_ID, API_KEY
    value VARCHAR(255) NOT NULL,

    -- Reason
    reason TEXT,
    added_by BIGINT REFERENCES users(id),

    -- Status
    is_active BOOLEAN DEFAULT TRUE,

    -- Timestamps
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,

    UNIQUE(entry_type, value)
);

CREATE INDEX idx_blacklist_type_value ON blacklist(entry_type, value);
CREATE INDEX idx_blacklist_active ON blacklist(is_active);

-- ==================
-- 8. SYSTEM MONITORING & ALERTS
-- ==================

CREATE TABLE system_metrics (
    id BIGSERIAL PRIMARY KEY,

    -- Metric Details
    metric_type VARCHAR(50),
    metric_name VARCHAR(100),
    metric_value DECIMAL(15,2),
    metric_unit VARCHAR(20),

    -- Context
    node_id VARCHAR(50),
    component VARCHAR(50),

    -- Metadata
    tags JSONB,

    -- Timestamp
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_metrics_type ON system_metrics(metric_type);
CREATE INDEX idx_metrics_recorded ON system_metrics(recorded_at);

CREATE TABLE alerts (
    id BIGSERIAL PRIMARY KEY,
    alert_id VARCHAR(64) UNIQUE NOT NULL,

    -- Alert Details
    alert_type VARCHAR(50), -- LOW_BALANCE, HIGH_TPS, CONNECTION_FAILURE, THRESHOLD_BREACH
    severity VARCHAR(20), -- CRITICAL, HIGH, MEDIUM, LOW, INFO
    title VARCHAR(255),
    message TEXT,

    -- Target
    account_id BIGINT REFERENCES accounts(id),
    user_id BIGINT REFERENCES users(id),

    -- Status
    status VARCHAR(20) DEFAULT 'OPEN',
    acknowledged BOOLEAN DEFAULT FALSE,
    acknowledged_by BIGINT REFERENCES users(id),
    acknowledged_at TIMESTAMP,
    resolved BOOLEAN DEFAULT FALSE,
    resolved_at TIMESTAMP,

    -- Notification
    notification_sent BOOLEAN DEFAULT FALSE,
    notification_channels TEXT[], -- EMAIL, SMS, TELEGRAM, WEBHOOK

    -- Timestamp
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_alert_status CHECK (status IN ('OPEN', 'ACKNOWLEDGED', 'RESOLVED', 'CLOSED')),
    CONSTRAINT chk_alert_severity CHECK (severity IN ('CRITICAL', 'HIGH', 'MEDIUM', 'LOW', 'INFO'))
);

CREATE INDEX idx_alerts_account ON alerts(account_id);
CREATE INDEX idx_alerts_status ON alerts(status);
CREATE INDEX idx_alerts_severity ON alerts(severity);
CREATE INDEX idx_alerts_created ON alerts(created_at);

-- ==================
-- 9. CONFIGURATION
-- ==================

CREATE TABLE system_config (
    id SERIAL PRIMARY KEY,
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT,
    config_type VARCHAR(20) DEFAULT 'STRING', -- STRING, INTEGER, BOOLEAN, JSON
    description TEXT,
    is_encrypted BOOLEAN DEFAULT FALSE,

    -- Metadata
    category VARCHAR(50),
    is_editable BOOLEAN DEFAULT TRUE,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by BIGINT REFERENCES users(id)
);

-- ==================
-- 10. FUNCTIONS & TRIGGERS
-- ==================

-- Auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply to tables
CREATE TRIGGER update_accounts_updated_at BEFORE UPDATE ON accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_campaigns_updated_at BEFORE UPDATE ON campaigns
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_templates_updated_at BEFORE UPDATE ON message_templates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ==================
-- END OF SCHEMA
-- ==================
