-- ============================================================================
-- Protei_Bulk Subscriber Profiling, Segmentation & Privacy Engine
-- Database Schema
-- ============================================================================

-- ============================================================================
-- 1. ATTRIBUTE SCHEMA TABLE (Dynamic Field Definitions)
-- ============================================================================

CREATE TABLE IF NOT EXISTS tbl_attribute_schema (
    attribute_id BIGSERIAL PRIMARY KEY,
    attribute_name VARCHAR(64) UNIQUE NOT NULL,
    attribute_code VARCHAR(64) UNIQUE NOT NULL,
    display_name VARCHAR(128) NOT NULL,
    description TEXT,

    -- Data Type & Validation
    data_type VARCHAR(20) NOT NULL CHECK (data_type IN ('STRING', 'INTEGER', 'DECIMAL', 'BOOLEAN', 'ENUM', 'JSON', 'DATE', 'DATETIME')),
    allowed_values JSONB DEFAULT '[]',
    validation_regex VARCHAR(255),
    min_value DECIMAL(15, 4),
    max_value DECIMAL(15, 4),

    -- Field Configuration
    is_required BOOLEAN DEFAULT FALSE,
    is_searchable BOOLEAN DEFAULT TRUE,
    is_visible_to_cp BOOLEAN DEFAULT TRUE,
    is_encrypted BOOLEAN DEFAULT FALSE,

    -- Privacy & Security
    privacy_level VARCHAR(20) DEFAULT 'PUBLIC' CHECK (privacy_level IN ('PUBLIC', 'SENSITIVE', 'CONFIDENTIAL', 'RESTRICTED')),
    requires_permission VARCHAR(100),

    -- Display & UI
    display_order INTEGER DEFAULT 100,
    category VARCHAR(64),
    icon VARCHAR(64),
    help_text TEXT,

    -- Status
    is_active BOOLEAN DEFAULT TRUE,

    -- Audit
    created_by VARCHAR(64),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    -- Metadata
    metadata JSONB DEFAULT '{}'
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_attribute_name ON tbl_attribute_schema(attribute_name);
CREATE INDEX IF NOT EXISTS idx_attribute_code ON tbl_attribute_schema(attribute_code);
CREATE INDEX IF NOT EXISTS idx_attribute_searchable ON tbl_attribute_schema(is_searchable);
CREATE INDEX IF NOT EXISTS idx_attribute_active ON tbl_attribute_schema(is_active);

-- ============================================================================
-- 2. PROFILES TABLE (Main Subscriber Profiles)
-- ============================================================================

CREATE TABLE IF NOT EXISTS tbl_profiles (
    profile_id BIGSERIAL PRIMARY KEY,

    -- Identity (Privacy-Protected)
    msisdn_hash VARCHAR(64) UNIQUE NOT NULL,  -- SHA256 hash of MSISDN
    customer_id BIGINT REFERENCES tbl_customers(customer_id) ON DELETE CASCADE,

    -- Standard Attributes
    gender VARCHAR(20) CHECK (gender IN ('MALE', 'FEMALE', 'OTHER', 'UNKNOWN')),
    age INTEGER,
    date_of_birth DATE,
    language VARCHAR(20),

    -- Location
    country_code VARCHAR(10),
    region VARCHAR(64),
    city VARCHAR(64),
    postal_code VARCHAR(20),

    -- Device & Technology
    device_type VARCHAR(20) CHECK (device_type IN ('ANDROID', 'IOS', 'FEATURE_PHONE', 'OTHER', 'UNKNOWN')),
    device_model VARCHAR(128),
    os_version VARCHAR(64),

    -- Service Info
    plan_type VARCHAR(20) CHECK (plan_type IN ('PREPAID', 'POSTPAID', 'VIP', 'CORPORATE', 'OTHER')),
    subscription_date DATE,
    last_recharge_date DATE,

    -- Behavioral
    interests JSONB DEFAULT '[]',  -- Array of interest tags
    preferences JSONB DEFAULT '{}',  -- Key-value preferences

    -- Activity
    last_activity_date DATE,
    last_message_sent DATE,
    last_message_received DATE,
    total_messages_sent INTEGER DEFAULT 0,
    total_messages_received INTEGER DEFAULT 0,

    -- Status
    status VARCHAR(20) DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE', 'SUSPENDED', 'DELETED')),
    opt_in_marketing BOOLEAN DEFAULT FALSE,
    opt_in_sms BOOLEAN DEFAULT TRUE,

    -- Custom Attributes (Dynamic)
    custom_attributes JSONB DEFAULT '{}',

    -- Privacy & Consent
    consent_date TIMESTAMP,
    consent_version VARCHAR(20),
    gdpr_compliant BOOLEAN DEFAULT TRUE,
    data_retention_date DATE,

    -- Audit
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    imported_at TIMESTAMP,
    imported_by VARCHAR(64),

    -- Metadata
    metadata JSONB DEFAULT '{}',
    tags JSONB DEFAULT '[]'
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_profile_hash ON tbl_profiles(msisdn_hash);
CREATE INDEX IF NOT EXISTS idx_profile_customer ON tbl_profiles(customer_id);
CREATE INDEX IF NOT EXISTS idx_profile_gender ON tbl_profiles(gender);
CREATE INDEX IF NOT EXISTS idx_profile_age ON tbl_profiles(age);
CREATE INDEX IF NOT EXISTS idx_profile_region ON tbl_profiles(region);
CREATE INDEX IF NOT EXISTS idx_profile_city ON tbl_profiles(city);
CREATE INDEX IF NOT EXISTS idx_profile_device_type ON tbl_profiles(device_type);
CREATE INDEX IF NOT EXISTS idx_profile_plan_type ON tbl_profiles(plan_type);
CREATE INDEX IF NOT EXISTS idx_profile_status ON tbl_profiles(status);
CREATE INDEX IF NOT EXISTS idx_profile_last_activity ON tbl_profiles(last_activity_date);
CREATE INDEX IF NOT EXISTS idx_profile_opt_in ON tbl_profiles(opt_in_marketing);
CREATE INDEX IF NOT EXISTS idx_profile_created ON tbl_profiles(created_at);

-- GIN indexes for JSONB fields
CREATE INDEX IF NOT EXISTS idx_profile_interests ON tbl_profiles USING GIN(interests);
CREATE INDEX IF NOT EXISTS idx_profile_custom_attrs ON tbl_profiles USING GIN(custom_attributes);
CREATE INDEX IF NOT EXISTS idx_profile_tags ON tbl_profiles USING GIN(tags);

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_profile_gender_age ON tbl_profiles(gender, age);
CREATE INDEX IF NOT EXISTS idx_profile_region_city ON tbl_profiles(region, city);

-- ============================================================================
-- 3. PROFILE GROUPS (Segments/Audiences)
-- ============================================================================

CREATE TABLE IF NOT EXISTS tbl_profile_groups (
    group_id BIGSERIAL PRIMARY KEY,
    group_code VARCHAR(64) UNIQUE NOT NULL,
    group_name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Ownership
    customer_id BIGINT REFERENCES tbl_customers(customer_id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,

    -- Filter Definition
    filter_query JSONB NOT NULL,  -- JSON representation of filter conditions
    filter_sql TEXT,  -- Generated SQL WHERE clause (cached)

    -- Statistics
    record_count BIGINT DEFAULT 0,
    last_count_updated TIMESTAMP,

    -- Refresh Strategy
    is_dynamic BOOLEAN DEFAULT TRUE,  -- Auto-refresh vs static snapshot
    refresh_frequency VARCHAR(20) CHECK (refresh_frequency IN ('REALTIME', 'HOURLY', 'DAILY', 'WEEKLY', 'MONTHLY', 'MANUAL')),
    last_refreshed TIMESTAMP,
    next_refresh TIMESTAMP,

    -- Usage Tracking
    total_campaigns_sent INTEGER DEFAULT 0,
    total_messages_sent BIGINT DEFAULT 0,
    last_used_at TIMESTAMP,

    -- Visibility & Sharing
    visibility VARCHAR(20) DEFAULT 'PRIVATE' CHECK (visibility IN ('PRIVATE', 'SHARED', 'PUBLIC')),
    shared_with_users JSONB DEFAULT '[]',
    shared_with_customers JSONB DEFAULT '[]',

    -- Status
    is_active BOOLEAN DEFAULT TRUE,

    -- Audit
    created_by VARCHAR(64),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    -- Metadata
    metadata JSONB DEFAULT '{}',
    tags JSONB DEFAULT '[]'
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_group_code ON tbl_profile_groups(group_code);
CREATE INDEX IF NOT EXISTS idx_group_customer ON tbl_profile_groups(customer_id);
CREATE INDEX IF NOT EXISTS idx_group_user ON tbl_profile_groups(user_id);
CREATE INDEX IF NOT EXISTS idx_group_active ON tbl_profile_groups(is_active);
CREATE INDEX IF NOT EXISTS idx_group_visibility ON tbl_profile_groups(visibility);
CREATE INDEX IF NOT EXISTS idx_group_last_used ON tbl_profile_groups(last_used_at);

-- ============================================================================
-- 4. PROFILE GROUP MEMBERS (Cached Segment Members)
-- ============================================================================

CREATE TABLE IF NOT EXISTS tbl_profile_group_members (
    group_id BIGINT REFERENCES tbl_profile_groups(group_id) ON DELETE CASCADE,
    profile_id BIGINT REFERENCES tbl_profiles(profile_id) ON DELETE CASCADE,

    -- Timestamps
    added_at TIMESTAMP DEFAULT NOW(),

    PRIMARY KEY (group_id, profile_id)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_group_members_group ON tbl_profile_group_members(group_id);
CREATE INDEX IF NOT EXISTS idx_group_members_profile ON tbl_profile_group_members(profile_id);

-- ============================================================================
-- 5. PROFILE IMPORT JOBS (Bulk Import Tracking)
-- ============================================================================

CREATE TABLE IF NOT EXISTS tbl_profile_import_jobs (
    job_id BIGSERIAL PRIMARY KEY,
    job_code VARCHAR(64) UNIQUE NOT NULL,

    -- Job Details
    customer_id BIGINT REFERENCES tbl_customers(customer_id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,

    -- File Info
    file_name VARCHAR(255),
    file_size_bytes BIGINT,
    file_path VARCHAR(500),
    file_type VARCHAR(20) CHECK (file_type IN ('CSV', 'EXCEL', 'JSON', 'XML')),

    -- Column Mapping
    column_mapping JSONB,  -- Maps CSV columns to profile attributes

    -- Progress
    status VARCHAR(20) DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED', 'CANCELLED')),
    total_rows INTEGER DEFAULT 0,
    rows_processed INTEGER DEFAULT 0,
    rows_imported INTEGER DEFAULT 0,
    rows_updated INTEGER DEFAULT 0,
    rows_failed INTEGER DEFAULT 0,

    -- Timing
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    duration_seconds INTEGER,

    -- Errors
    error_message TEXT,
    error_rows JSONB DEFAULT '[]',  -- Sample of failed rows

    -- Options
    update_existing BOOLEAN DEFAULT TRUE,
    skip_duplicates BOOLEAN DEFAULT FALSE,
    hash_msisdn BOOLEAN DEFAULT TRUE,

    -- Audit
    created_by VARCHAR(64),
    created_at TIMESTAMP DEFAULT NOW(),

    -- Metadata
    metadata JSONB DEFAULT '{}'
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_import_job_code ON tbl_profile_import_jobs(job_code);
CREATE INDEX IF NOT EXISTS idx_import_customer ON tbl_profile_import_jobs(customer_id);
CREATE INDEX IF NOT EXISTS idx_import_user ON tbl_profile_import_jobs(user_id);
CREATE INDEX IF NOT EXISTS idx_import_status ON tbl_profile_import_jobs(status);
CREATE INDEX IF NOT EXISTS idx_import_created ON tbl_profile_import_jobs(created_at);

-- ============================================================================
-- 6. PROFILE QUERY LOG (Privacy Audit Trail)
-- ============================================================================

CREATE TABLE IF NOT EXISTS tbl_profile_query_log (
    query_id BIGSERIAL PRIMARY KEY,

    -- Query Details
    customer_id BIGINT REFERENCES tbl_customers(customer_id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,

    -- Query Info
    query_type VARCHAR(20) CHECK (query_type IN ('SEARCH', 'SEGMENT', 'EXPORT', 'COUNT', 'CAMPAIGN')),
    filter_query JSONB,
    result_count BIGINT,

    -- Group Reference
    group_id BIGINT REFERENCES tbl_profile_groups(group_id) ON DELETE SET NULL,
    group_name VARCHAR(255),

    -- Privacy Compliance
    includes_pii BOOLEAN DEFAULT FALSE,
    approval_required BOOLEAN DEFAULT FALSE,
    approved_by VARCHAR(64),
    approved_at TIMESTAMP,

    -- Performance
    query_time_ms INTEGER,
    cache_hit BOOLEAN DEFAULT FALSE,

    -- Source
    ip_address INET,
    user_agent TEXT,

    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),

    -- Metadata
    metadata JSONB DEFAULT '{}'
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_query_log_customer ON tbl_profile_query_log(customer_id);
CREATE INDEX IF NOT EXISTS idx_query_log_user ON tbl_profile_query_log(user_id);
CREATE INDEX IF NOT EXISTS idx_query_log_type ON tbl_profile_query_log(query_type);
CREATE INDEX IF NOT EXISTS idx_query_log_created ON tbl_profile_query_log(created_at);
CREATE INDEX IF NOT EXISTS idx_query_log_group ON tbl_profile_query_log(group_id);

-- ============================================================================
-- 7. PROFILE STATISTICS (Aggregated Data)
-- ============================================================================

CREATE TABLE IF NOT EXISTS tbl_profile_statistics (
    stat_id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT REFERENCES tbl_customers(customer_id) ON DELETE CASCADE,

    -- Time Period
    period_date DATE NOT NULL,
    period_type VARCHAR(20) DEFAULT 'DAILY' CHECK (period_type IN ('DAILY', 'WEEKLY', 'MONTHLY', 'YEARLY')),

    -- Profile Counts
    total_profiles BIGINT DEFAULT 0,
    active_profiles BIGINT DEFAULT 0,
    inactive_profiles BIGINT DEFAULT 0,
    new_profiles BIGINT DEFAULT 0,
    updated_profiles BIGINT DEFAULT 0,

    -- Demographics
    male_count BIGINT DEFAULT 0,
    female_count BIGINT DEFAULT 0,
    avg_age DECIMAL(5, 2),

    -- Device Distribution
    android_count BIGINT DEFAULT 0,
    ios_count BIGINT DEFAULT 0,
    feature_phone_count BIGINT DEFAULT 0,

    -- Plan Distribution
    prepaid_count BIGINT DEFAULT 0,
    postpaid_count BIGINT DEFAULT 0,

    -- Opt-In Rates
    opt_in_marketing_count BIGINT DEFAULT 0,
    opt_in_sms_count BIGINT DEFAULT 0,

    -- Created
    created_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(customer_id, period_date, period_type)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_profile_stats_customer ON tbl_profile_statistics(customer_id);
CREATE INDEX IF NOT EXISTS idx_profile_stats_period ON tbl_profile_statistics(period_date);
CREATE INDEX IF NOT EXISTS idx_profile_stats_type ON tbl_profile_statistics(period_type);

-- ============================================================================
-- 8. HELPER FUNCTIONS
-- ============================================================================

-- Function to hash MSISDN
CREATE OR REPLACE FUNCTION hash_msisdn(p_msisdn VARCHAR)
RETURNS VARCHAR AS $$
BEGIN
    RETURN encode(digest(p_msisdn, 'sha256'), 'hex');
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function to get profile by MSISDN (returns profile_id only for system use)
CREATE OR REPLACE FUNCTION get_profile_id_by_msisdn(p_msisdn VARCHAR)
RETURNS BIGINT AS $$
DECLARE
    v_hash VARCHAR(64);
    v_profile_id BIGINT;
BEGIN
    v_hash := hash_msisdn(p_msisdn);

    SELECT profile_id INTO v_profile_id
    FROM tbl_profiles
    WHERE msisdn_hash = v_hash
      AND status = 'ACTIVE';

    RETURN v_profile_id;
END;
$$ LANGUAGE plpgsql;

-- Function to count profiles by filter
CREATE OR REPLACE FUNCTION count_profiles_by_filter(
    p_customer_id BIGINT,
    p_filter JSONB
) RETURNS BIGINT AS $$
DECLARE
    v_count BIGINT;
    v_sql TEXT;
BEGIN
    -- This is a placeholder - actual implementation would parse p_filter
    -- and build dynamic SQL query
    SELECT COUNT(*) INTO v_count
    FROM tbl_profiles
    WHERE customer_id = p_customer_id
      AND status = 'ACTIVE';

    RETURN v_count;
END;
$$ LANGUAGE plpgsql;

-- Function to refresh profile group count
CREATE OR REPLACE FUNCTION refresh_profile_group_count(p_group_id BIGINT)
RETURNS VOID AS $$
DECLARE
    v_count BIGINT;
BEGIN
    SELECT COUNT(*) INTO v_count
    FROM tbl_profile_group_members
    WHERE group_id = p_group_id;

    UPDATE tbl_profile_groups
    SET record_count = v_count,
        last_count_updated = NOW()
    WHERE group_id = p_group_id;
END;
$$ LANGUAGE plpgsql;

-- Function to calculate profile statistics
CREATE OR REPLACE FUNCTION calculate_profile_statistics(
    p_customer_id BIGINT,
    p_period_date DATE
) RETURNS VOID AS $$
DECLARE
    v_total BIGINT;
    v_active BIGINT;
    v_inactive BIGINT;
    v_male BIGINT;
    v_female BIGINT;
    v_avg_age DECIMAL;
BEGIN
    -- Count profiles
    SELECT
        COUNT(*),
        SUM(CASE WHEN status = 'ACTIVE' THEN 1 ELSE 0 END),
        SUM(CASE WHEN status = 'INACTIVE' THEN 1 ELSE 0 END),
        SUM(CASE WHEN gender = 'MALE' THEN 1 ELSE 0 END),
        SUM(CASE WHEN gender = 'FEMALE' THEN 1 ELSE 0 END),
        AVG(age)
    INTO v_total, v_active, v_inactive, v_male, v_female, v_avg_age
    FROM tbl_profiles
    WHERE customer_id = p_customer_id;

    -- Insert or update statistics
    INSERT INTO tbl_profile_statistics (
        customer_id, period_date, period_type,
        total_profiles, active_profiles, inactive_profiles,
        male_count, female_count, avg_age
    ) VALUES (
        p_customer_id, p_period_date, 'DAILY',
        v_total, v_active, v_inactive,
        v_male, v_female, v_avg_age
    )
    ON CONFLICT (customer_id, period_date, period_type)
    DO UPDATE SET
        total_profiles = v_total,
        active_profiles = v_active,
        inactive_profiles = v_inactive,
        male_count = v_male,
        female_count = v_female,
        avg_age = v_avg_age,
        created_at = NOW();
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- 9. TRIGGERS
-- ============================================================================

-- Trigger to update profile updated_at timestamp
CREATE OR REPLACE FUNCTION update_profile_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_profile_update_timestamp
BEFORE UPDATE ON tbl_profiles
FOR EACH ROW EXECUTE FUNCTION update_profile_timestamp();

-- Trigger to update profile group timestamp
CREATE OR REPLACE FUNCTION update_profile_group_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_profile_group_update_timestamp
BEFORE UPDATE ON tbl_profile_groups
FOR EACH ROW EXECUTE FUNCTION update_profile_group_timestamp();

-- ============================================================================
-- 10. SEED DATA - DEFAULT ATTRIBUTES
-- ============================================================================

INSERT INTO tbl_attribute_schema (attribute_name, attribute_code, display_name, data_type, is_searchable, category, display_order) VALUES
('Gender', 'gender', 'Gender', 'ENUM', TRUE, 'Demographics', 10),
('Age', 'age', 'Age', 'INTEGER', TRUE, 'Demographics', 20),
('Date of Birth', 'date_of_birth', 'Date of Birth', 'DATE', FALSE, 'Demographics', 30),
('Language', 'language', 'Preferred Language', 'ENUM', TRUE, 'Preferences', 40),
('Country', 'country_code', 'Country', 'STRING', TRUE, 'Location', 50),
('Region', 'region', 'Region/State', 'STRING', TRUE, 'Location', 60),
('City', 'city', 'City', 'STRING', TRUE, 'Location', 70),
('Device Type', 'device_type', 'Device Type', 'ENUM', TRUE, 'Technology', 80),
('Plan Type', 'plan_type', 'Subscription Plan', 'ENUM', TRUE, 'Service', 90),
('Interests', 'interests', 'Interests', 'JSON', TRUE, 'Behavioral', 100),
('Opt-in Marketing', 'opt_in_marketing', 'Marketing Consent', 'BOOLEAN', TRUE, 'Privacy', 110),
('Last Activity', 'last_activity_date', 'Last Activity Date', 'DATE', TRUE, 'Activity', 120)
ON CONFLICT (attribute_code) DO NOTHING;

-- Update allowed values for enums
UPDATE tbl_attribute_schema SET allowed_values = '["MALE", "FEMALE", "OTHER", "UNKNOWN"]'::JSONB WHERE attribute_code = 'gender';
UPDATE tbl_attribute_schema SET allowed_values = '["ARABIC", "ENGLISH", "FRENCH", "SPANISH", "GERMAN"]'::JSONB WHERE attribute_code = 'language';
UPDATE tbl_attribute_schema SET allowed_values = '["ANDROID", "IOS", "FEATURE_PHONE", "OTHER", "UNKNOWN"]'::JSONB WHERE attribute_code = 'device_type';
UPDATE tbl_attribute_schema SET allowed_values = '["PREPAID", "POSTPAID", "VIP", "CORPORATE", "OTHER"]'::JSONB WHERE attribute_code = 'plan_type';

COMMIT;
