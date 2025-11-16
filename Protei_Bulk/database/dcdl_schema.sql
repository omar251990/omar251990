-- ============================================================================
-- Protei_Bulk Dynamic Campaign Data Loader (DCDL) Schema
-- Handles dynamic import, validation, caching, and mapping of campaign data
-- ============================================================================

-- ============================================================================
-- 1. DCDL DATASETS (Main Upload/Query Configurations)
-- ============================================================================

CREATE TABLE IF NOT EXISTS tbl_dcdl_datasets (
    dataset_id BIGSERIAL PRIMARY KEY,
    dataset_code VARCHAR(64) UNIQUE NOT NULL,
    dataset_name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Ownership
    customer_id BIGINT REFERENCES tbl_customers(customer_id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    campaign_id BIGINT REFERENCES campaigns(id) ON DELETE SET NULL,

    -- Data Source
    source_type VARCHAR(20) NOT NULL CHECK (source_type IN ('FILE_UPLOAD', 'DATABASE_QUERY', 'HYBRID', 'API')),

    -- File Upload Details
    file_name VARCHAR(255),
    file_type VARCHAR(20) CHECK (file_type IN ('CSV', 'EXCEL', 'JSON', 'XML')),
    file_size_bytes BIGINT,
    file_path VARCHAR(500),
    file_hash VARCHAR(64),  -- SHA256 hash for integrity

    -- Database Query Details
    query_id BIGINT REFERENCES tbl_dcdl_queries(query_id) ON DELETE SET NULL,
    connection_id VARCHAR(64),  -- Reference to external DB connection
    last_query_time TIMESTAMP,

    -- Data Statistics
    total_records INTEGER DEFAULT 0,
    valid_records INTEGER DEFAULT 0,
    invalid_records INTEGER DEFAULT 0,
    duplicate_records INTEGER DEFAULT 0,

    -- Column Schema
    columns JSONB DEFAULT '[]',  -- Array of column names
    column_types JSONB DEFAULT '{}',  -- Column name -> data type mapping
    sample_data JSONB DEFAULT '[]',  -- First 10 rows for preview

    -- Status & Lifecycle
    status VARCHAR(20) DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'VALIDATING', 'VALID', 'INVALID', 'CACHED', 'ACTIVE', 'EXPIRED', 'ARCHIVED')),
    validation_status VARCHAR(20),
    validation_message TEXT,

    -- Processing
    processing_started_at TIMESTAMP,
    processing_completed_at TIMESTAMP,
    processing_duration_ms INTEGER,

    -- Caching
    is_cached BOOLEAN DEFAULT FALSE,
    cache_key VARCHAR(128),
    cache_expiry TIMESTAMP,
    cache_size_bytes BIGINT,

    -- Access Control
    visibility VARCHAR(20) DEFAULT 'PRIVATE' CHECK (visibility IN ('PRIVATE', 'SHARED', 'PUBLIC')),
    shared_with_users JSONB DEFAULT '[]',
    shared_with_customers JSONB DEFAULT '[]',

    -- Audit
    created_by VARCHAR(64),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    uploaded_at TIMESTAMP,
    activated_at TIMESTAMP,

    -- Metadata
    metadata JSONB DEFAULT '{}',
    tags JSONB DEFAULT '[]',
    notes TEXT
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_dcdl_dataset_code ON tbl_dcdl_datasets(dataset_code);
CREATE INDEX IF NOT EXISTS idx_dcdl_customer ON tbl_dcdl_datasets(customer_id);
CREATE INDEX IF NOT EXISTS idx_dcdl_user ON tbl_dcdl_datasets(user_id);
CREATE INDEX IF NOT EXISTS idx_dcdl_campaign ON tbl_dcdl_datasets(campaign_id);
CREATE INDEX IF NOT EXISTS idx_dcdl_source_type ON tbl_dcdl_datasets(source_type);
CREATE INDEX IF NOT EXISTS idx_dcdl_status ON tbl_dcdl_datasets(status);
CREATE INDEX IF NOT EXISTS idx_dcdl_cached ON tbl_dcdl_datasets(is_cached);
CREATE INDEX IF NOT EXISTS idx_dcdl_created ON tbl_dcdl_datasets(created_at);

-- ============================================================================
-- 2. DCDL QUERIES (Database Query Configurations)
-- ============================================================================

CREATE TABLE IF NOT EXISTS tbl_dcdl_queries (
    query_id BIGSERIAL PRIMARY KEY,
    query_code VARCHAR(64) UNIQUE NOT NULL,
    query_name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Ownership
    customer_id BIGINT REFERENCES tbl_customers(customer_id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,

    -- Database Connection
    connection_id VARCHAR(64) NOT NULL,
    connection_name VARCHAR(255),
    db_type VARCHAR(20) CHECK (db_type IN ('POSTGRESQL', 'MYSQL', 'MARIADB', 'ORACLE', 'MSSQL', 'OTHER')),
    db_host VARCHAR(255),
    db_port INTEGER,
    db_name VARCHAR(128),
    db_username VARCHAR(128),
    db_password_encrypted TEXT,

    -- Query Definition
    query_sql TEXT NOT NULL,
    query_parameters JSONB DEFAULT '{}',  -- Named parameters
    query_timeout_seconds INTEGER DEFAULT 300,

    -- Expected Schema
    expected_columns JSONB DEFAULT '[]',
    column_mappings JSONB DEFAULT '{}',

    -- Execution
    last_executed_at TIMESTAMP,
    last_execution_duration_ms INTEGER,
    last_execution_status VARCHAR(20),
    last_execution_error TEXT,
    last_result_count INTEGER,

    -- Schedule
    enable_auto_refresh BOOLEAN DEFAULT FALSE,
    refresh_frequency VARCHAR(20) CHECK (refresh_frequency IN ('HOURLY', 'DAILY', 'WEEKLY', 'MONTHLY', 'MANUAL')),
    next_refresh_at TIMESTAMP,

    -- Cache Settings
    cache_results BOOLEAN DEFAULT TRUE,
    cache_ttl_seconds INTEGER DEFAULT 3600,

    -- Status
    is_active BOOLEAN DEFAULT TRUE,

    -- Audit
    created_by VARCHAR(64),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    -- Metadata
    metadata JSONB DEFAULT '{}',
    notes TEXT
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_dcdl_query_code ON tbl_dcdl_queries(query_code);
CREATE INDEX IF NOT EXISTS idx_dcdl_query_customer ON tbl_dcdl_queries(customer_id);
CREATE INDEX IF NOT EXISTS idx_dcdl_query_connection ON tbl_dcdl_queries(connection_id);
CREATE INDEX IF NOT EXISTS idx_dcdl_query_active ON tbl_dcdl_queries(is_active);

-- ============================================================================
-- 3. DCDL MAPPING (Column to Parameter Mappings)
-- ============================================================================

CREATE TABLE IF NOT EXISTS tbl_dcdl_mapping (
    mapping_id BIGSERIAL PRIMARY KEY,
    dataset_id BIGINT REFERENCES tbl_dcdl_datasets(dataset_id) ON DELETE CASCADE,
    campaign_id BIGINT REFERENCES campaigns(id) ON DELETE CASCADE,

    -- Mapping Details
    source_column VARCHAR(128) NOT NULL,
    parameter_name VARCHAR(128) NOT NULL,
    parameter_type VARCHAR(20) CHECK (parameter_type IN ('STRING', 'INTEGER', 'DECIMAL', 'DATE', 'BOOLEAN', 'MSISDN', 'SENDER_ID')),

    -- Transformation Rules
    transform_function VARCHAR(64),  -- e.g., 'UPPER', 'LOWER', 'TRIM', 'FORMAT_DATE'
    transform_params JSONB DEFAULT '{}',
    default_value TEXT,
    is_required BOOLEAN DEFAULT FALSE,

    -- Validation Rules
    validation_regex VARCHAR(255),
    min_length INTEGER,
    max_length INTEGER,
    min_value DECIMAL(15, 4),
    max_value DECIMAL(15, 4),
    allowed_values JSONB DEFAULT '[]',

    -- Usage
    usage_context VARCHAR(20) CHECK (usage_context IN ('MESSAGE_TEXT', 'SENDER_ID', 'ROUTING', 'METADATA', 'CUSTOM')),
    placeholder VARCHAR(128),  -- e.g., '${CREDITOR_NAME}'

    -- Status
    is_active BOOLEAN DEFAULT TRUE,

    -- Audit
    created_by VARCHAR(64),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    -- Metadata
    metadata JSONB DEFAULT '{}',

    UNIQUE(dataset_id, source_column, parameter_name)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_dcdl_mapping_dataset ON tbl_dcdl_mapping(dataset_id);
CREATE INDEX IF NOT EXISTS idx_dcdl_mapping_campaign ON tbl_dcdl_mapping(campaign_id);
CREATE INDEX IF NOT EXISTS idx_dcdl_mapping_param ON tbl_dcdl_mapping(parameter_name);
CREATE INDEX IF NOT EXISTS idx_dcdl_mapping_active ON tbl_dcdl_mapping(is_active);

-- ============================================================================
-- 4. DCDL DATA CACHE (Cached Dataset Records)
-- ============================================================================

CREATE TABLE IF NOT EXISTS tbl_dcdl_data_cache (
    cache_id BIGSERIAL PRIMARY KEY,
    dataset_id BIGINT REFERENCES tbl_dcdl_datasets(dataset_id) ON DELETE CASCADE,

    -- Record Data
    record_index INTEGER NOT NULL,
    record_data JSONB NOT NULL,  -- Complete record as JSON

    -- Quick Lookup Fields (indexed for performance)
    msisdn VARCHAR(20),
    sender_id VARCHAR(100),
    routing_code VARCHAR(64),

    -- Status
    is_valid BOOLEAN DEFAULT TRUE,
    validation_errors JSONB DEFAULT '[]',

    -- Timestamps
    cached_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,

    UNIQUE(dataset_id, record_index)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_dcdl_cache_dataset ON tbl_dcdl_data_cache(dataset_id);
CREATE INDEX IF NOT EXISTS idx_dcdl_cache_msisdn ON tbl_dcdl_data_cache(msisdn);
CREATE INDEX IF NOT EXISTS idx_dcdl_cache_sender ON tbl_dcdl_data_cache(sender_id);
CREATE INDEX IF NOT EXISTS idx_dcdl_cache_route ON tbl_dcdl_data_cache(routing_code);
CREATE INDEX IF NOT EXISTS idx_dcdl_cache_expires ON tbl_dcdl_data_cache(expires_at);

-- GIN index for JSONB data
CREATE INDEX IF NOT EXISTS idx_dcdl_cache_data ON tbl_dcdl_data_cache USING GIN(record_data);

-- ============================================================================
-- 5. DCDL VALIDATION ERRORS (Error Tracking)
-- ============================================================================

CREATE TABLE IF NOT EXISTS tbl_dcdl_validation_errors (
    error_id BIGSERIAL PRIMARY KEY,
    dataset_id BIGINT REFERENCES tbl_dcdl_datasets(dataset_id) ON DELETE CASCADE,

    -- Error Details
    row_number INTEGER,
    column_name VARCHAR(128),
    error_type VARCHAR(20) CHECK (error_type IN ('INVALID_HEADER', 'INVALID_FORMAT', 'DUPLICATE', 'MISSING_REQUIRED', 'OUT_OF_RANGE', 'INVALID_MSISDN', 'OTHER')),
    error_message TEXT,
    error_value TEXT,

    -- Sample Data
    row_data JSONB,

    -- Severity
    severity VARCHAR(20) DEFAULT 'ERROR' CHECK (severity IN ('WARNING', 'ERROR', 'CRITICAL')),

    -- Resolution
    is_resolved BOOLEAN DEFAULT FALSE,
    resolved_by VARCHAR(64),
    resolved_at TIMESTAMP,
    resolution_notes TEXT,

    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_dcdl_error_dataset ON tbl_dcdl_validation_errors(dataset_id);
CREATE INDEX IF NOT EXISTS idx_dcdl_error_type ON tbl_dcdl_validation_errors(error_type);
CREATE INDEX IF NOT EXISTS idx_dcdl_error_severity ON tbl_dcdl_validation_errors(severity);
CREATE INDEX IF NOT EXISTS idx_dcdl_error_resolved ON tbl_dcdl_validation_errors(is_resolved);

-- ============================================================================
-- 6. DCDL AUDIT LOG (Complete Change Tracking)
-- ============================================================================

CREATE TABLE IF NOT EXISTS tbl_dcdl_audit (
    audit_id BIGSERIAL PRIMARY KEY,

    -- Action Details
    action_type VARCHAR(20) CHECK (action_type IN ('UPLOAD', 'VALIDATE', 'MAP', 'CACHE', 'QUERY', 'DELETE', 'EXPIRE', 'SHARE')),
    dataset_id BIGINT REFERENCES tbl_dcdl_datasets(dataset_id) ON DELETE SET NULL,
    query_id BIGINT REFERENCES tbl_dcdl_queries(query_id) ON DELETE SET NULL,

    -- User Context
    customer_id BIGINT REFERENCES tbl_customers(customer_id) ON DELETE SET NULL,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    username VARCHAR(128),

    -- Action Description
    description TEXT,
    changes JSONB DEFAULT '{}',  -- Before/after values

    -- Result
    status VARCHAR(20) CHECK (status IN ('SUCCESS', 'FAILED', 'WARNING')),
    result_message TEXT,

    -- Source
    ip_address INET,
    user_agent TEXT,

    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),

    -- Metadata
    metadata JSONB DEFAULT '{}'
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_dcdl_audit_action ON tbl_dcdl_audit(action_type);
CREATE INDEX IF NOT EXISTS idx_dcdl_audit_dataset ON tbl_dcdl_audit(dataset_id);
CREATE INDEX IF NOT EXISTS idx_dcdl_audit_query ON tbl_dcdl_audit(query_id);
CREATE INDEX IF NOT EXISTS idx_dcdl_audit_customer ON tbl_dcdl_audit(customer_id);
CREATE INDEX IF NOT EXISTS idx_dcdl_audit_user ON tbl_dcdl_audit(user_id);
CREATE INDEX IF NOT EXISTS idx_dcdl_audit_created ON tbl_dcdl_audit(created_at);

-- ============================================================================
-- 7. DCDL PERFORMANCE STATISTICS
-- ============================================================================

CREATE TABLE IF NOT EXISTS tbl_dcdl_statistics (
    stat_id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT REFERENCES tbl_customers(customer_id) ON DELETE CASCADE,

    -- Time Period
    period_date DATE NOT NULL,
    period_type VARCHAR(20) DEFAULT 'DAILY' CHECK (period_type IN ('HOURLY', 'DAILY', 'WEEKLY', 'MONTHLY')),

    -- Upload Statistics
    total_uploads INTEGER DEFAULT 0,
    total_records_uploaded BIGINT DEFAULT 0,
    avg_upload_size_mb DECIMAL(10, 2),

    -- Validation Statistics
    total_validations INTEGER DEFAULT 0,
    successful_validations INTEGER DEFAULT 0,
    failed_validations INTEGER DEFAULT 0,
    avg_validation_time_ms INTEGER,

    -- Query Statistics
    total_queries INTEGER DEFAULT 0,
    successful_queries INTEGER DEFAULT 0,
    failed_queries INTEGER DEFAULT 0,
    avg_query_time_ms INTEGER,

    -- Cache Statistics
    cache_hits BIGINT DEFAULT 0,
    cache_misses BIGINT DEFAULT 0,
    cache_hit_rate DECIMAL(5, 2),

    -- Performance
    avg_record_processing_ms DECIMAL(10, 2),
    peak_tps DECIMAL(10, 2),

    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(customer_id, period_date, period_type)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_dcdl_stats_customer ON tbl_dcdl_statistics(customer_id);
CREATE INDEX IF NOT EXISTS idx_dcdl_stats_period ON tbl_dcdl_statistics(period_date);

-- ============================================================================
-- 8. HELPER FUNCTIONS
-- ============================================================================

-- Function to validate MSISDN format
CREATE OR REPLACE FUNCTION validate_msisdn(p_msisdn VARCHAR)
RETURNS BOOLEAN AS $$
BEGIN
    -- Basic MSISDN validation: digits only, 10-15 characters
    RETURN p_msisdn ~ '^[0-9]{10,15}$';
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function to hash dataset file
CREATE OR REPLACE FUNCTION hash_file_content(p_content TEXT)
RETURNS VARCHAR AS $$
BEGIN
    RETURN encode(digest(p_content, 'sha256'), 'hex');
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function to get dataset statistics
CREATE OR REPLACE FUNCTION get_dataset_statistics(p_dataset_id BIGINT)
RETURNS TABLE (
    total_records INTEGER,
    valid_records INTEGER,
    invalid_records INTEGER,
    cached_records INTEGER,
    validation_errors INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        d.total_records,
        d.valid_records,
        d.invalid_records,
        (SELECT COUNT(*) FROM tbl_dcdl_data_cache WHERE dataset_id = p_dataset_id)::INTEGER,
        (SELECT COUNT(*) FROM tbl_dcdl_validation_errors WHERE dataset_id = p_dataset_id AND is_resolved = FALSE)::INTEGER
    FROM tbl_dcdl_datasets d
    WHERE d.dataset_id = p_dataset_id;
END;
$$ LANGUAGE plpgsql;

-- Function to apply parameter mapping
CREATE OR REPLACE FUNCTION apply_parameter_mapping(
    p_dataset_id BIGINT,
    p_record_data JSONB
) RETURNS JSONB AS $$
DECLARE
    v_result JSONB := '{}';
    v_mapping RECORD;
    v_source_value TEXT;
    v_mapped_value TEXT;
BEGIN
    FOR v_mapping IN
        SELECT source_column, parameter_name, transform_function, default_value
        FROM tbl_dcdl_mapping
        WHERE dataset_id = p_dataset_id AND is_active = TRUE
    LOOP
        -- Get source value
        v_source_value := p_record_data->>v_mapping.source_column;

        -- Apply transformation if specified
        IF v_mapping.transform_function IS NOT NULL THEN
            CASE v_mapping.transform_function
                WHEN 'UPPER' THEN v_mapped_value := UPPER(v_source_value);
                WHEN 'LOWER' THEN v_mapped_value := LOWER(v_source_value);
                WHEN 'TRIM' THEN v_mapped_value := TRIM(v_source_value);
                ELSE v_mapped_value := v_source_value;
            END CASE;
        ELSE
            v_mapped_value := COALESCE(v_source_value, v_mapping.default_value);
        END IF;

        -- Add to result
        v_result := v_result || jsonb_build_object(v_mapping.parameter_name, v_mapped_value);
    END LOOP;

    RETURN v_result;
END;
$$ LANGUAGE plpgsql;

-- Function to expire old cached datasets
CREATE OR REPLACE FUNCTION expire_old_cache_datasets()
RETURNS INTEGER AS $$
DECLARE
    v_expired_count INTEGER;
BEGIN
    -- Delete expired cache records
    DELETE FROM tbl_dcdl_data_cache
    WHERE expires_at IS NOT NULL AND expires_at < NOW();

    GET DIAGNOSTICS v_expired_count = ROW_COUNT;

    -- Update dataset status
    UPDATE tbl_dcdl_datasets
    SET status = 'EXPIRED',
        is_cached = FALSE
    WHERE dataset_id IN (
        SELECT dataset_id FROM tbl_dcdl_datasets
        WHERE cache_expiry IS NOT NULL AND cache_expiry < NOW()
          AND status = 'CACHED'
    );

    RETURN v_expired_count;
END;
$$ LANGUAGE plpgsql;

-- Function to calculate cache statistics
CREATE OR REPLACE FUNCTION calculate_dcdl_statistics(p_date DATE)
RETURNS VOID AS $$
BEGIN
    INSERT INTO tbl_dcdl_statistics (
        customer_id, period_date, period_type,
        total_uploads, total_records_uploaded,
        total_validations, successful_validations, failed_validations
    )
    SELECT
        customer_id,
        DATE(created_at) as period_date,
        'DAILY',
        COUNT(*) as total_uploads,
        SUM(total_records) as total_records_uploaded,
        COUNT(*) as total_validations,
        COUNT(*) FILTER (WHERE status = 'VALID') as successful_validations,
        COUNT(*) FILTER (WHERE status = 'INVALID') as failed_validations
    FROM tbl_dcdl_datasets
    WHERE DATE(created_at) = p_date
    GROUP BY customer_id, period_date
    ON CONFLICT (customer_id, period_date, period_type)
    DO UPDATE SET
        total_uploads = EXCLUDED.total_uploads,
        total_records_uploaded = EXCLUDED.total_records_uploaded,
        total_validations = EXCLUDED.total_validations,
        successful_validations = EXCLUDED.successful_validations,
        failed_validations = EXCLUDED.failed_validations,
        created_at = NOW();
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- 9. TRIGGERS
-- ============================================================================

-- Trigger to update dataset timestamp
CREATE OR REPLACE FUNCTION update_dcdl_dataset_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_dcdl_dataset_update
BEFORE UPDATE ON tbl_dcdl_datasets
FOR EACH ROW EXECUTE FUNCTION update_dcdl_dataset_timestamp();

-- Trigger to log dataset changes
CREATE OR REPLACE FUNCTION log_dcdl_dataset_change()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO tbl_dcdl_audit (action_type, dataset_id, customer_id, user_id, description, status)
        VALUES ('UPLOAD', NEW.dataset_id, NEW.customer_id, NEW.user_id, 'Dataset created', 'SUCCESS');
    ELSIF TG_OP = 'UPDATE' AND OLD.status != NEW.status THEN
        INSERT INTO tbl_dcdl_audit (action_type, dataset_id, customer_id, user_id, description, status, changes)
        VALUES ('VALIDATE', NEW.dataset_id, NEW.customer_id, NEW.user_id,
                'Dataset status changed',
                CASE WHEN NEW.status IN ('VALID', 'CACHED', 'ACTIVE') THEN 'SUCCESS' ELSE 'FAILED' END,
                jsonb_build_object('old_status', OLD.status, 'new_status', NEW.status));
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_log_dataset_change
AFTER INSERT OR UPDATE ON tbl_dcdl_datasets
FOR EACH ROW EXECUTE FUNCTION log_dcdl_dataset_change();

-- ============================================================================
-- 10. INITIAL DATA - Common Transform Functions
-- ============================================================================

-- Store common transform functions metadata
CREATE TABLE IF NOT EXISTS tbl_dcdl_transform_functions (
    function_id SERIAL PRIMARY KEY,
    function_name VARCHAR(64) UNIQUE NOT NULL,
    description TEXT,
    input_type VARCHAR(20),
    output_type VARCHAR(20),
    example VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE
);

INSERT INTO tbl_dcdl_transform_functions (function_name, description, input_type, output_type, example) VALUES
('UPPER', 'Convert to uppercase', 'STRING', 'STRING', 'hello -> HELLO'),
('LOWER', 'Convert to lowercase', 'STRING', 'STRING', 'HELLO -> hello'),
('TRIM', 'Remove leading/trailing whitespace', 'STRING', 'STRING', '  text  -> text'),
('FORMAT_MSISDN', 'Format MSISDN to E.164', 'STRING', 'STRING', '0781234567 -> 962781234567'),
('FORMAT_DATE', 'Format date string', 'STRING', 'DATE', '2025-01-16 -> 16-01-2025'),
('EXTRACT_DIGITS', 'Extract only digits', 'STRING', 'STRING', 'abc123def -> 123'),
('TRUNCATE', 'Truncate to N characters', 'STRING', 'STRING', 'longtext -> long'),
('PAD_LEFT', 'Pad left with zeros', 'STRING', 'STRING', '123 -> 00123'),
('REPLACE', 'Replace substring', 'STRING', 'STRING', 'hello world -> hello universe'),
('CONCAT', 'Concatenate values', 'STRING', 'STRING', 'a, b -> ab')
ON CONFLICT (function_name) DO NOTHING;

COMMIT;
