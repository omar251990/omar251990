-- =============================================
-- Protei_Bulk Seed Data
-- Initial data for testing and development
-- =============================================

-- Insert default system configuration
INSERT INTO system_config (config_key, config_value, config_type, description, category, is_editable) VALUES
('system.name', 'Protei_Bulk', 'STRING', 'System name', 'system', FALSE),
('system.version', '1.0.0', 'STRING', 'System version', 'system', FALSE),
('system.max_tps', '10000', 'INTEGER', 'Maximum system TPS', 'performance', TRUE),
('system.default_timezone', 'UTC', 'STRING', 'Default timezone', 'system', TRUE),
('system.working_hours_start', '00:00', 'STRING', 'System working hours start', 'operation', TRUE),
('system.working_hours_end', '23:59', 'STRING', 'System working hours end', 'operation', TRUE),
('message.max_length', '1600', 'INTEGER', 'Maximum message length', 'messaging', TRUE),
('message.default_encoding', 'GSM7', 'STRING', 'Default message encoding', 'messaging', TRUE),
('message.default_priority', 'NORMAL', 'STRING', 'Default message priority', 'messaging', TRUE),
('campaign.require_approval', 'true', 'BOOLEAN', 'Require campaign approval', 'campaign', TRUE),
('campaign.max_recipients', '1000000', 'INTEGER', 'Maximum recipients per campaign', 'campaign', TRUE),
('alert.low_balance_threshold', '100', 'INTEGER', 'Low balance alert threshold', 'alerts', TRUE),
('alert.email_enabled', 'true', 'BOOLEAN', 'Enable email alerts', 'alerts', TRUE),
('alert.sms_enabled', 'false', 'BOOLEAN', 'Enable SMS alerts', 'alerts', TRUE),
('security.password_min_length', '12', 'INTEGER', 'Minimum password length', 'security', TRUE),
('security.password_expiry_days', '90', 'INTEGER', 'Password expiry in days', 'security', TRUE),
('security.max_failed_attempts', '5', 'INTEGER', 'Maximum failed login attempts', 'security', TRUE),
('security.lockout_duration_minutes', '30', 'INTEGER', 'Account lockout duration', 'security', TRUE),
('security.session_timeout_minutes', '60', 'INTEGER', 'Session timeout in minutes', 'security', TRUE),
('cdr.retention_days', '365', 'INTEGER', 'CDR retention period', 'cdr', TRUE),
('cdr.auto_archive', 'true', 'BOOLEAN', 'Auto-archive old CDRs', 'cdr', TRUE),
('logs.retention_days', '90', 'INTEGER', 'Log retention period', 'logging', TRUE),
('logs.auto_rotation', 'true', 'BOOLEAN', 'Enable automatic log rotation', 'logging', TRUE);

-- Grant all permissions to Super Admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT 1, id FROM permissions;

-- Grant read/write permissions to Admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT 2, id FROM permissions WHERE permission_key NOT IN ('system.config');

-- Grant campaign and message permissions to User role
INSERT INTO role_permissions (role_id, permission_id)
SELECT 5, id FROM permissions WHERE module IN ('messages', 'campaigns', 'templates') AND action IN ('create', 'read');

-- Grant approval permissions to Approver role
INSERT INTO role_permissions (role_id, permission_id)
SELECT 6, id FROM permissions WHERE permission_key IN ('campaigns.approve', 'campaigns.read', 'messages.read');

-- Grant read-only permissions to Viewer role
INSERT INTO role_permissions (role_id, permission_id)
SELECT 7, id FROM permissions WHERE action = 'read';

-- Create default Super Admin account
INSERT INTO accounts (
    account_id,
    account_type_id,
    company_name,
    business_name,
    account_status,
    billing_type,
    credit_limit,
    current_balance,
    free_sender,
    max_tps,
    max_concurrent_connections,
    contact_name,
    contact_email,
    contact_phone
) VALUES (
    'ACC_SUPER_ADMIN',
    1, -- SUPER_ADMIN
    'Protei Corporation',
    'System Administration',
    'ACTIVE',
    'POSTPAID',
    999999999.00,
    999999999.00,
    TRUE,
    10000,
    1000,
    'System Administrator',
    'admin@protei.com',
    '+1234567890'
);

-- Create default admin user
-- Password: Admin@123 (bcrypt hash)
INSERT INTO users (
    user_id,
    account_id,
    username,
    email,
    password_hash,
    full_name,
    status,
    email_verified,
    must_change_password,
    two_factor_enabled,
    api_enabled,
    smpp_enabled
) VALUES (
    'USR_ADMIN',
    (SELECT id FROM accounts WHERE account_id = 'ACC_SUPER_ADMIN'),
    'admin',
    'admin@protei.com',
    '$2b$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYqNqkqGQm6', -- Admin@123
    'System Administrator',
    'ACTIVE',
    TRUE,
    TRUE, -- Must change password on first login
    FALSE,
    TRUE,
    TRUE
);

-- Assign Super Admin role to admin user
INSERT INTO user_roles (user_id, role_id)
VALUES (
    (SELECT id FROM users WHERE username = 'admin'),
    (SELECT id FROM roles WHERE role_name = 'SUPER_ADMIN')
);

-- Generate API key for admin user
UPDATE users SET api_key = 'ADMIN_API_KEY_' || MD5(RANDOM()::TEXT || CLOCK_TIMESTAMP()::TEXT)
WHERE username = 'admin';

-- Create demo SMSC connection
INSERT INTO smsc_connections (
    smsc_id,
    smsc_name,
    smsc_type,
    host,
    port,
    system_id,
    password,
    system_type,
    interface_version,
    bind_type,
    enquire_link_interval,
    request_timeout,
    window_size,
    max_tps,
    status,
    connection_state,
    priority,
    weight
) VALUES (
    'SMSC_DEMO_001',
    'Demo SMSC Connection',
    'SMPP',
    'localhost',
    2775,
    'demo_user',
    'demo_pass',
    'SMPP',
    '3.4',
    'TRANSCEIVER',
    30,
    30,
    100,
    1000,
    'ACTIVE',
    'DISCONNECTED',
    100,
    1
);

-- Create default routing rule
INSERT INTO routing_rules (
    rule_name,
    rule_type,
    priority,
    target_smsc_ids,
    routing_strategy,
    enabled
) VALUES (
    'Default Routing Rule',
    'DEFAULT',
    100,
    ARRAY[(SELECT id FROM smsc_connections WHERE smsc_id = 'SMSC_DEMO_001')],
    'ROUND_ROBIN',
    TRUE
);

-- Create sample message templates
INSERT INTO message_templates (
    template_id,
    account_id,
    user_id,
    name,
    category,
    language,
    content,
    variables,
    encoding,
    is_approved,
    status
) VALUES
(
    'TPL_OTP_001',
    (SELECT id FROM accounts WHERE account_id = 'ACC_SUPER_ADMIN'),
    (SELECT id FROM users WHERE username = 'admin'),
    'OTP Verification',
    'OTP',
    'en',
    'Your verification code is {code}. Valid for {validity} minutes.',
    ARRAY['code', 'validity'],
    'GSM7',
    TRUE,
    'ACTIVE'
),
(
    'TPL_WELCOME_001',
    (SELECT id FROM accounts WHERE account_id = 'ACC_SUPER_ADMIN'),
    (SELECT id FROM users WHERE username = 'admin'),
    'Welcome Message',
    'TRANSACTIONAL',
    'en',
    'Welcome {name}! Thank you for joining {company}.',
    ARRAY['name', 'company'],
    'GSM7',
    TRUE,
    'ACTIVE'
),
(
    'TPL_PROMO_001',
    (SELECT id FROM accounts WHERE account_id = 'ACC_SUPER_ADMIN'),
    (SELECT id FROM users WHERE username = 'admin'),
    'Promotional Offer',
    'PROMO',
    'en',
    'Hi {name}, get {discount}% off on your next purchase! Code: {code}',
    ARRAY['name', 'discount', 'code'],
    'GSM7',
    TRUE,
    'ACTIVE'
),
(
    'TPL_ALERT_001',
    (SELECT id FROM accounts WHERE account_id = 'ACC_SUPER_ADMIN'),
    (SELECT id FROM users WHERE username = 'admin'),
    'System Alert',
    'ALERT',
    'en',
    'ALERT: {message}. Time: {timestamp}',
    ARRAY['message', 'timestamp'],
    'GSM7',
    TRUE,
    'ACTIVE'
);

-- Create sample MSISDN list
INSERT INTO msisdn_lists (
    list_id,
    account_id,
    user_id,
    name,
    description,
    is_hidden,
    total_count,
    active_count
) VALUES (
    'LIST_DEMO_001',
    (SELECT id FROM accounts WHERE account_id = 'ACC_SUPER_ADMIN'),
    (SELECT id FROM users WHERE username = 'admin'),
    'Demo Contact List',
    'Sample contact list for testing',
    FALSE,
    5,
    5
);

-- Add sample MSISDNs to the list
INSERT INTO msisdn_list_entries (list_id, msisdn, name, attributes, status)
VALUES
(
    (SELECT id FROM msisdn_lists WHERE list_id = 'LIST_DEMO_001'),
    '1234567890',
    'John Doe',
    '{"age": 25, "city": "Amman", "gender": "M"}'::JSONB,
    'ACTIVE'
),
(
    (SELECT id FROM msisdn_lists WHERE list_id = 'LIST_DEMO_001'),
    '1234567891',
    'Jane Smith',
    '{"age": 30, "city": "Dubai", "gender": "F"}'::JSONB,
    'ACTIVE'
),
(
    (SELECT id FROM msisdn_lists WHERE list_id = 'LIST_DEMO_001'),
    '1234567892',
    'Bob Johnson',
    '{"age": 35, "city": "Riyadh", "gender": "M"}'::JSONB,
    'ACTIVE'
),
(
    (SELECT id FROM msisdn_lists WHERE list_id = 'LIST_DEMO_001'),
    '1234567893',
    'Alice Williams',
    '{"age": 28, "city": "Cairo", "gender": "F"}'::JSONB,
    'ACTIVE'
),
(
    (SELECT id FROM msisdn_lists WHERE list_id = 'LIST_DEMO_001'),
    '1234567894',
    'Charlie Brown',
    '{"age": 32, "city": "Amman", "gender": "M"}'::JSONB,
    'ACTIVE'
);

-- Create sample subscriber profiles
INSERT INTO subscriber_profiles (account_id, msisdn, attributes)
VALUES
(
    (SELECT id FROM accounts WHERE account_id = 'ACC_SUPER_ADMIN'),
    '1234567890',
    '{"age": 25, "gender": "M", "city": "Amman", "interests": ["sports", "technology"], "subscription_type": "premium"}'::JSONB
),
(
    (SELECT id FROM accounts WHERE account_id = 'ACC_SUPER_ADMIN'),
    '1234567891',
    '{"age": 30, "gender": "F", "city": "Dubai", "interests": ["fashion", "travel"], "subscription_type": "standard"}'::JSONB
),
(
    (SELECT id FROM accounts WHERE account_id = 'ACC_SUPER_ADMIN'),
    '1234567892',
    '{"age": 35, "gender": "M", "city": "Riyadh", "interests": ["business", "finance"], "subscription_type": "premium"}'::JSONB
);

-- Approve admin user's templates
UPDATE message_templates
SET approved_by = (SELECT id FROM users WHERE username = 'admin'),
    approved_at = CURRENT_TIMESTAMP
WHERE template_id LIKE 'TPL_%';

-- Log installation in audit
INSERT INTO audit_logs (
    user_id,
    username,
    action,
    module,
    entity_type,
    success,
    created_at
) VALUES (
    (SELECT id FROM users WHERE username = 'admin'),
    'admin',
    'SYSTEM_INSTALL',
    'system',
    'installation',
    TRUE,
    CURRENT_TIMESTAMP
);

-- Create initial system alert
INSERT INTO alerts (
    alert_id,
    alert_type,
    severity,
    title,
    message,
    status,
    notification_sent
) VALUES (
    'ALT_INSTALL_' || MD5(RANDOM()::TEXT),
    'SYSTEM',
    'INFO',
    'System Installation Complete',
    'Protei_Bulk has been successfully installed and initialized.',
    'OPEN',
    FALSE
);

-- Output summary
DO $$
DECLARE
    user_count INTEGER;
    account_count INTEGER;
    template_count INTEGER;
    list_count INTEGER;
    smsc_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO user_count FROM users;
    SELECT COUNT(*) INTO account_count FROM accounts;
    SELECT COUNT(*) INTO template_count FROM message_templates;
    SELECT COUNT(*) INTO list_count FROM msisdn_lists;
    SELECT COUNT(*) INTO smsc_count FROM smsc_connections;

    RAISE NOTICE '';
    RAISE NOTICE '╔════════════════════════════════════════════════════════════════╗';
    RAISE NOTICE '║           Protei_Bulk Seed Data Loaded Successfully           ║';
    RAISE NOTICE '╚════════════════════════════════════════════════════════════════╝';
    RAISE NOTICE '';
    RAISE NOTICE 'Created:';
    RAISE NOTICE '  • % user account(s)', user_count;
    RAISE NOTICE '  • % company account(s)', account_count;
    RAISE NOTICE '  • % message template(s)', template_count;
    RAISE NOTICE '  • % MSISDN list(s)', list_count;
    RAISE NOTICE '  • % SMSC connection(s)', smsc_count;
    RAISE NOTICE '';
    RAISE NOTICE 'Default Credentials:';
    RAISE NOTICE '  Username: admin';
    RAISE NOTICE '  Password: Admin@123';
    RAISE NOTICE '  (CHANGE ON FIRST LOGIN)';
    RAISE NOTICE '';
END $$;
