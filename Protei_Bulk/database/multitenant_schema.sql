-- ============================================================================
-- Protei_Bulk Multi-Tenant Permission & Control System
-- Database Schema for Multi-Tenancy
-- ============================================================================

-- ============================================================================
-- 1. CUSTOMERS (TENANTS) TABLE
-- ============================================================================

CREATE TABLE tbl_customers (
    customer_id BIGSERIAL PRIMARY KEY,
    customer_code VARCHAR(64) UNIQUE NOT NULL,
    customer_name VARCHAR(255) NOT NULL,
    company_name VARCHAR(255),

    -- Contact Information
    contact_person VARCHAR(255),
    contact_email VARCHAR(255),
    contact_phone VARCHAR(50),

    -- Address
    address TEXT,
    city VARCHAR(100),
    country VARCHAR(100),

    -- Status & License
    status VARCHAR(20) DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'SUSPENDED', 'EXPIRED', 'TRIAL')),
    license_key VARCHAR(255) UNIQUE,
    license_type VARCHAR(50) DEFAULT 'STANDARD' CHECK (license_type IN ('TRIAL', 'STANDARD', 'PREMIUM', 'ENTERPRISE')),
    expiry_date TIMESTAMP,

    -- Quotas & Limits
    max_users INTEGER DEFAULT 100,
    max_resellers INTEGER DEFAULT 10,
    max_tps INTEGER DEFAULT 1000,
    max_campaigns INTEGER DEFAULT 1000,
    storage_quota_gb INTEGER DEFAULT 100,

    -- Allowed Features (JSONB for flexibility)
    allowed_channels JSONB DEFAULT '["SMS", "EMAIL"]',
    allowed_modules JSONB DEFAULT '["campaigns", "templates", "reports"]',
    feature_flags JSONB DEFAULT '{}',

    -- Branding
    logo_url VARCHAR(500),
    primary_color VARCHAR(20),
    secondary_color VARCHAR(20),

    -- Configuration
    timezone VARCHAR(50) DEFAULT 'UTC',
    language VARCHAR(10) DEFAULT 'EN',
    date_format VARCHAR(20) DEFAULT 'YYYY-MM-DD',

    -- SMSC & Routing
    allowed_smsc_ids JSONB DEFAULT '[]',
    allowed_sender_ids JSONB DEFAULT '[]',
    routing_policy VARCHAR(50) DEFAULT 'AUTOMATIC',

    -- Billing
    billing_type VARCHAR(20) DEFAULT 'PREPAID' CHECK (billing_type IN ('PREPAID', 'POSTPAID')),
    balance DECIMAL(15, 4) DEFAULT 0.0,
    credit_limit DECIMAL(15, 4) DEFAULT 0.0,
    cost_per_sms DECIMAL(10, 4) DEFAULT 0.01,

    -- Alert Thresholds
    alert_balance_threshold DECIMAL(15, 4) DEFAULT 1000.0,
    alert_tps_threshold INTEGER DEFAULT 900,
    alert_emails JSONB DEFAULT '[]',

    -- Audit
    created_by VARCHAR(64),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_activity_at TIMESTAMP,

    -- Metadata
    metadata JSONB DEFAULT '{}',
    notes TEXT
);

-- Indexes for customers
CREATE INDEX idx_customers_status ON tbl_customers(status);
CREATE INDEX idx_customers_code ON tbl_customers(customer_code);
CREATE INDEX idx_customers_expiry ON tbl_customers(expiry_date);

-- ============================================================================
-- 2. ROLES TABLE
-- ============================================================================

CREATE TABLE tbl_roles (
    role_id BIGSERIAL PRIMARY KEY,
    role_name VARCHAR(100) NOT NULL,
    role_code VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,

    -- Hierarchy Level
    level INTEGER DEFAULT 100, -- Lower number = higher privilege
    -- 0 = Global Admin
    -- 10 = Customer Admin
    -- 20 = Reseller
    -- 30 = End User
    -- 40 = Approver
    -- 50 = Auditor
    -- 100 = Custom

    -- Scope
    scope VARCHAR(20) DEFAULT 'CUSTOMER' CHECK (scope IN ('GLOBAL', 'CUSTOMER', 'RESELLER')),

    -- Status
    is_system_role BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,

    -- Audit
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Insert predefined system roles
INSERT INTO tbl_roles (role_name, role_code, description, level, scope, is_system_role) VALUES
('Global System Admin', 'GLOBAL_ADMIN', 'Root administrator at Protei HQ level', 0, 'GLOBAL', TRUE),
('Customer Admin', 'CUSTOMER_ADMIN', 'Customer internal system administrator', 10, 'CUSTOMER', TRUE),
('Reseller', 'RESELLER', 'Sub-tenant under customer', 20, 'RESELLER', TRUE),
('End User', 'END_USER', 'Corporate client or end user', 30, 'CUSTOMER', TRUE),
('Approver', 'APPROVER', 'Campaign approver (maker-checker)', 40, 'CUSTOMER', TRUE),
('Auditor', 'AUDITOR', 'Read-only audit access', 50, 'CUSTOMER', TRUE);

-- ============================================================================
-- 3. PERMISSIONS TABLE
-- ============================================================================

CREATE TABLE tbl_permissions (
    permission_id BIGSERIAL PRIMARY KEY,
    permission_code VARCHAR(100) UNIQUE NOT NULL,
    module_name VARCHAR(100) NOT NULL,
    action_name VARCHAR(100) NOT NULL,
    description TEXT,

    -- Categorization
    category VARCHAR(50),

    -- Status
    is_system_permission BOOLEAN DEFAULT TRUE,
    is_active BOOLEAN DEFAULT TRUE,

    -- Audit
    created_at TIMESTAMP DEFAULT NOW()
);

-- Insert comprehensive permission set
INSERT INTO tbl_permissions (permission_code, module_name, action_name, description, category) VALUES
-- User Management
('USER_VIEW', 'users', 'view', 'View users', 'user_management'),
('USER_CREATE', 'users', 'create', 'Create new users', 'user_management'),
('USER_EDIT', 'users', 'edit', 'Edit user details', 'user_management'),
('USER_DELETE', 'users', 'delete', 'Delete users', 'user_management'),
('USER_ACTIVATE', 'users', 'activate', 'Activate/deactivate users', 'user_management'),
('USER_RESET_PASSWORD', 'users', 'reset_password', 'Reset user passwords', 'user_management'),

-- Permission Management
('PERMISSION_VIEW', 'permissions', 'view', 'View permissions', 'permission_management'),
('PERMISSION_ASSIGN', 'permissions', 'assign', 'Assign permissions to roles/users', 'permission_management'),
('PERMISSION_REVOKE', 'permissions', 'revoke', 'Revoke permissions', 'permission_management'),

-- Campaign Management
('CAMPAIGN_VIEW', 'campaigns', 'view', 'View campaigns', 'campaign_management'),
('CAMPAIGN_CREATE', 'campaigns', 'create', 'Create campaigns', 'campaign_management'),
('CAMPAIGN_EDIT', 'campaigns', 'edit', 'Edit campaigns', 'campaign_management'),
('CAMPAIGN_DELETE', 'campaigns', 'delete', 'Delete campaigns', 'campaign_management'),
('CAMPAIGN_SEND', 'campaigns', 'send', 'Send campaigns', 'campaign_management'),
('CAMPAIGN_PAUSE', 'campaigns', 'pause', 'Pause campaigns', 'campaign_management'),
('CAMPAIGN_RESUME', 'campaigns', 'resume', 'Resume campaigns', 'campaign_management'),
('CAMPAIGN_STOP', 'campaigns', 'stop', 'Stop campaigns', 'campaign_management'),
('CAMPAIGN_APPROVE', 'campaigns', 'approve', 'Approve campaigns', 'campaign_management'),
('CAMPAIGN_REJECT', 'campaigns', 'reject', 'Reject campaigns', 'campaign_management'),
('CAMPAIGN_DUPLICATE', 'campaigns', 'duplicate', 'Duplicate campaigns', 'campaign_management'),

-- Template Management
('TEMPLATE_VIEW', 'templates', 'view', 'View templates', 'template_management'),
('TEMPLATE_CREATE', 'templates', 'create', 'Create templates', 'template_management'),
('TEMPLATE_EDIT', 'templates', 'edit', 'Edit templates', 'template_management'),
('TEMPLATE_DELETE', 'templates', 'delete', 'Delete templates', 'template_management'),
('TEMPLATE_SHARE', 'templates', 'share', 'Share templates', 'template_management'),

-- Contact Management
('CONTACT_VIEW', 'contacts', 'view', 'View contact lists', 'contact_management'),
('CONTACT_CREATE', 'contacts', 'create', 'Create contact lists', 'contact_management'),
('CONTACT_EDIT', 'contacts', 'edit', 'Edit contact lists', 'contact_management'),
('CONTACT_DELETE', 'contacts', 'delete', 'Delete contact lists', 'contact_management'),
('CONTACT_IMPORT', 'contacts', 'import', 'Import contacts', 'contact_management'),
('CONTACT_EXPORT', 'contacts', 'export', 'Export contacts', 'contact_management'),

-- Report & Analytics
('REPORT_VIEW', 'reports', 'view', 'View reports', 'reporting'),
('REPORT_EXPORT', 'reports', 'export', 'Export reports', 'reporting'),
('REPORT_SCHEDULE', 'reports', 'schedule', 'Schedule reports', 'reporting'),
('ANALYTICS_VIEW', 'analytics', 'view', 'View analytics', 'reporting'),

-- SMSC & Routing
('SMSC_VIEW', 'smsc', 'view', 'View SMSC connections', 'routing'),
('SMSC_CREATE', 'smsc', 'create', 'Create SMSC connections', 'routing'),
('SMSC_EDIT', 'smsc', 'edit', 'Edit SMSC connections', 'routing'),
('SMSC_DELETE', 'smsc', 'delete', 'Delete SMSC connections', 'routing'),
('ROUTE_VIEW', 'routing', 'view', 'View routing rules', 'routing'),
('ROUTE_CREATE', 'routing', 'create', 'Create routing rules', 'routing'),
('ROUTE_EDIT', 'routing', 'edit', 'Edit routing rules', 'routing'),
('ROUTE_DELETE', 'routing', 'delete', 'Delete routing rules', 'routing'),

-- Quota & Balance
('QUOTA_VIEW', 'quota', 'view', 'View quotas', 'billing'),
('QUOTA_ASSIGN', 'quota', 'assign', 'Assign quotas', 'billing'),
('BALANCE_VIEW', 'balance', 'view', 'View balance', 'billing'),
('BALANCE_TOPUP', 'balance', 'topup', 'Top-up balance', 'billing'),
('BALANCE_TRANSFER', 'balance', 'transfer', 'Transfer balance', 'billing'),

-- System Configuration
('CONFIG_VIEW', 'config', 'view', 'View configuration', 'system'),
('CONFIG_EDIT', 'config', 'edit', 'Edit configuration', 'system'),
('BACKUP_CREATE', 'backup', 'create', 'Create backups', 'system'),
('BACKUP_RESTORE', 'backup', 'restore', 'Restore backups', 'system'),

-- Audit & Logs
('LOG_VIEW', 'logs', 'view', 'View logs', 'audit'),
('LOG_EXPORT', 'logs', 'export', 'Export logs', 'audit'),
('AUDIT_VIEW', 'audit', 'view', 'View audit trail', 'audit'),
('AUDIT_EXPORT', 'audit', 'export', 'Export audit trail', 'audit'),

-- Customer Management (Global Admin only)
('CUSTOMER_VIEW', 'customers', 'view', 'View customers', 'customer_management'),
('CUSTOMER_CREATE', 'customers', 'create', 'Create customers', 'customer_management'),
('CUSTOMER_EDIT', 'customers', 'edit', 'Edit customers', 'customer_management'),
('CUSTOMER_DELETE', 'customers', 'delete', 'Delete customers', 'customer_management'),
('CUSTOMER_SUSPEND', 'customers', 'suspend', 'Suspend customers', 'customer_management'),
('CUSTOMER_ACTIVATE', 'customers', 'activate', 'Activate customers', 'customer_management'),

-- API Access
('API_ACCESS', 'api', 'access', 'Access API', 'api'),
('API_KEY_MANAGE', 'api', 'manage_keys', 'Manage API keys', 'api'),

-- Testing
('SIMULATOR_ACCESS', 'simulator', 'access', 'Access SMS simulator', 'testing'),
('LOAD_TEST_ACCESS', 'loadtest', 'access', 'Access load testing', 'testing');

-- ============================================================================
-- 4. ROLE PERMISSIONS (DEFAULT MAPPING)
-- ============================================================================

CREATE TABLE tbl_role_permissions (
    role_id BIGINT REFERENCES tbl_roles(role_id) ON DELETE CASCADE,
    permission_id BIGINT REFERENCES tbl_permissions(permission_id) ON DELETE CASCADE,
    allow BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

-- Create indexes
CREATE INDEX idx_role_permissions_role ON tbl_role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission ON tbl_role_permissions(permission_id);

-- ============================================================================
-- 5. USER PERMISSIONS OVERRIDE
-- ============================================================================

CREATE TABLE tbl_user_permissions_override (
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    permission_id BIGINT REFERENCES tbl_permissions(permission_id) ON DELETE CASCADE,
    allow BOOLEAN DEFAULT TRUE,
    granted_by VARCHAR(64),
    granted_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    reason TEXT,
    PRIMARY KEY (user_id, permission_id)
);

-- Create indexes
CREATE INDEX idx_user_permissions_user ON tbl_user_permissions_override(user_id);
CREATE INDEX idx_user_permissions_permission ON tbl_user_permissions_override(permission_id);
CREATE INDEX idx_user_permissions_expires ON tbl_user_permissions_override(expires_at);

-- ============================================================================
-- 6. CUSTOMER CONFIGURATION
-- ============================================================================

CREATE TABLE tbl_customer_config (
    config_id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT REFERENCES tbl_customers(customer_id) ON DELETE CASCADE,
    config_key VARCHAR(100) NOT NULL,
    config_value TEXT,
    config_type VARCHAR(50) DEFAULT 'STRING',
    is_encrypted BOOLEAN DEFAULT FALSE,
    description TEXT,
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (customer_id, config_key)
);

-- Create index
CREATE INDEX idx_customer_config_customer ON tbl_customer_config(customer_id);
CREATE INDEX idx_customer_config_key ON tbl_customer_config(config_key);

-- ============================================================================
-- 7. UPDATE EXISTING TABLES FOR MULTI-TENANCY
-- ============================================================================

-- Add customer_id to users table
ALTER TABLE users ADD COLUMN customer_id BIGINT REFERENCES tbl_customers(customer_id) ON DELETE CASCADE;
ALTER TABLE users ADD COLUMN role_id BIGINT REFERENCES tbl_roles(role_id);
ALTER TABLE users ADD COLUMN parent_user_id BIGINT REFERENCES users(id);
ALTER TABLE users ADD COLUMN user_level INTEGER DEFAULT 30;

-- Create indexes
CREATE INDEX idx_users_customer ON users(customer_id);
CREATE INDEX idx_users_role ON users(role_id);
CREATE INDEX idx_users_parent ON users(parent_user_id);

-- Add customer_id to campaigns table
ALTER TABLE campaigns ADD COLUMN customer_id BIGINT REFERENCES tbl_customers(customer_id) ON DELETE CASCADE;
CREATE INDEX idx_campaigns_customer ON campaigns(customer_id);

-- Add customer_id to messages table
ALTER TABLE messages ADD COLUMN customer_id BIGINT REFERENCES tbl_customers(customer_id) ON DELETE CASCADE;
CREATE INDEX idx_messages_customer ON messages(customer_id);

-- Add customer_id to contact_lists table
ALTER TABLE contact_lists ADD COLUMN customer_id BIGINT REFERENCES tbl_customers(customer_id) ON DELETE CASCADE;
CREATE INDEX idx_contact_lists_customer ON contact_lists(customer_id);

-- Add customer_id to templates table
ALTER TABLE message_templates ADD COLUMN customer_id BIGINT REFERENCES tbl_customers(customer_id) ON DELETE CASCADE;
CREATE INDEX idx_templates_customer ON message_templates(customer_id);

-- ============================================================================
-- 8. PERMISSION AUDIT LOG
-- ============================================================================

CREATE TABLE tbl_permission_audit (
    audit_id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT REFERENCES tbl_customers(customer_id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    target_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    permission_id BIGINT REFERENCES tbl_permissions(permission_id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL, -- GRANTED, REVOKED, CHECKED
    result BOOLEAN,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

-- Create indexes
CREATE INDEX idx_permission_audit_customer ON tbl_permission_audit(customer_id);
CREATE INDEX idx_permission_audit_user ON tbl_permission_audit(user_id);
CREATE INDEX idx_permission_audit_created ON tbl_permission_audit(created_at);

-- ============================================================================
-- 9. HELPER FUNCTIONS
-- ============================================================================

-- Function to check user permission
CREATE OR REPLACE FUNCTION check_user_permission(
    p_user_id BIGINT,
    p_permission_code VARCHAR
) RETURNS BOOLEAN AS $$
DECLARE
    v_has_permission BOOLEAN := FALSE;
    v_permission_id BIGINT;
    v_role_id BIGINT;
BEGIN
    -- Get permission ID
    SELECT permission_id INTO v_permission_id
    FROM tbl_permissions
    WHERE permission_code = p_permission_code AND is_active = TRUE;

    IF v_permission_id IS NULL THEN
        RETURN FALSE;
    END IF;

    -- Check user-level override first
    SELECT allow INTO v_has_permission
    FROM tbl_user_permissions_override
    WHERE user_id = p_user_id
      AND permission_id = v_permission_id
      AND (expires_at IS NULL OR expires_at > NOW());

    IF FOUND THEN
        RETURN v_has_permission;
    END IF;

    -- Check role-level permission
    SELECT role_id INTO v_role_id FROM users WHERE id = p_user_id;

    SELECT allow INTO v_has_permission
    FROM tbl_role_permissions
    WHERE role_id = v_role_id AND permission_id = v_permission_id;

    RETURN COALESCE(v_has_permission, FALSE);
END;
$$ LANGUAGE plpgsql;

-- Function to get customer hierarchy
CREATE OR REPLACE FUNCTION get_customer_user_hierarchy(p_customer_id BIGINT)
RETURNS TABLE (
    user_id BIGINT,
    username VARCHAR,
    role_name VARCHAR,
    level INTEGER,
    parent_username VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    WITH RECURSIVE user_tree AS (
        -- Root users (no parent)
        SELECT
            u.id,
            u.username,
            r.role_name,
            u.user_level,
            u.parent_user_id,
            CAST(NULL AS VARCHAR) as parent_username
        FROM users u
        JOIN tbl_roles r ON u.role_id = r.role_id
        WHERE u.customer_id = p_customer_id AND u.parent_user_id IS NULL

        UNION ALL

        -- Child users
        SELECT
            u.id,
            u.username,
            r.role_name,
            u.user_level,
            u.parent_user_id,
            ut.username as parent_username
        FROM users u
        JOIN tbl_roles r ON u.role_id = r.role_id
        JOIN user_tree ut ON u.parent_user_id = ut.id
        WHERE u.customer_id = p_customer_id
    )
    SELECT id, username, role_name, level, parent_username FROM user_tree;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- 10. TRIGGERS
-- ============================================================================

-- Trigger to update customer last_activity_at
CREATE OR REPLACE FUNCTION update_customer_activity()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE tbl_customers
    SET last_activity_at = NOW()
    WHERE customer_id = NEW.customer_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_customer_activity_on_campaign
AFTER INSERT OR UPDATE ON campaigns
FOR EACH ROW EXECUTE FUNCTION update_customer_activity();

CREATE TRIGGER trigger_update_customer_activity_on_message
AFTER INSERT ON messages
FOR EACH ROW EXECUTE FUNCTION update_customer_activity();

-- ============================================================================
-- SEED DATA
-- ============================================================================

-- Create default Global Admin customer
INSERT INTO tbl_customers (
    customer_code, customer_name, company_name, status, license_type,
    max_users, max_tps, created_by
) VALUES (
    'PROTEI_HQ', 'Protei Headquarters', 'Protei Corporation',
    'ACTIVE', 'ENTERPRISE', 1000, 100000, 'SYSTEM'
);

-- Get the Global Admin customer ID
DO $$
DECLARE
    v_global_customer_id BIGINT;
    v_global_admin_role_id BIGINT;
BEGIN
    SELECT customer_id INTO v_global_customer_id FROM tbl_customers WHERE customer_code = 'PROTEI_HQ';
    SELECT role_id INTO v_global_admin_role_id FROM tbl_roles WHERE role_code = 'GLOBAL_ADMIN';

    -- Update existing admin user to be Global Admin
    UPDATE users
    SET customer_id = v_global_customer_id,
        role_id = v_global_admin_role_id,
        user_level = 0
    WHERE username = 'admin';
END $$;

COMMIT;
