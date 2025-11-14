#!/bin/bash
#
# Protei Monitoring v2.0 - Comprehensive Test Script
#
# This script tests all components of the Protei Monitoring system:
# - Database connectivity
# - Redis connectivity
# - Configuration validation
# - Application startup
# - Web API endpoints
# - Protocol decoders
# - AI features
# - Knowledge base
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Directories
INSTALL_DIR="/usr/protei/Protei_Monitoring"
CONFIG_DIR="$INSTALL_DIR/config"
LOG_DIR="$INSTALL_DIR/logs"

# Test results
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

# Print functions
print_header() {
    echo -e "${BLUE}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Protei Monitoring v2.0 - System Tests"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${NC}"
}

print_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

print_pass() {
    echo -e "${GREEN}  ✅ PASS${NC} $1"
    ((TESTS_PASSED++))
}

print_fail() {
    echo -e "${RED}  ❌ FAIL${NC} $1"
    ((TESTS_FAILED++))
}

print_skip() {
    echo -e "${YELLOW}  ⏭️  SKIP${NC} $1"
    ((TESTS_SKIPPED++))
}

print_info() {
    echo -e "${BLUE}  ℹ️  INFO${NC} $1"
}

# Load configuration
load_config() {
    if [ -f "$CONFIG_DIR/db.cfg" ]; then
        source "$CONFIG_DIR/db.cfg"
    fi
    if [ -f "$CONFIG_DIR/system.cfg" ]; then
        source "$CONFIG_DIR/system.cfg"
    fi
}

# Test 1: Configuration Files
test_configuration_files() {
    print_test "Configuration Files"

    local config_files=("license.cfg" "db.cfg" "protocols.cfg" "system.cfg" "trace.cfg" "paths.cfg" "security.cfg")
    local all_exist=true

    for cfg in "${config_files[@]}"; do
        if [ -f "$CONFIG_DIR/$cfg" ]; then
            # Validate syntax
            if bash -n "$CONFIG_DIR/$cfg" 2>/dev/null; then
                print_pass "$cfg exists and has valid syntax"
            else
                print_fail "$cfg has syntax errors"
                all_exist=false
            fi
        else
            print_fail "$cfg is missing"
            all_exist=false
        fi
    done

    if [ "$all_exist" = true ]; then
        print_pass "All configuration files present and valid"
    fi
}

# Test 2: Database Connectivity
test_database() {
    print_test "Database Connectivity"

    if [ -z "$DB_HOST" ] || [ -z "$DB_USER" ] || [ -z "$DB_NAME" ]; then
        print_fail "Database configuration not loaded"
        return
    fi

    export PGPASSWORD="$DB_PASSWORD"

    # Test connection
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" &>/dev/null; then
        print_pass "Database connection successful ($DB_NAME @ $DB_HOST:$DB_PORT)"
    else
        print_fail "Cannot connect to database"
        return
    fi

    # Test schema
    local tables=("users" "sessions" "messages" "subscribers" "issues" "kpis" "audit_log")
    local all_tables_exist=true

    for table in "${tables[@]}"; do
        if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "\d $table" &>/dev/null; then
            print_pass "Table '$table' exists"
        else
            print_fail "Table '$table' missing"
            all_tables_exist=false
        fi
    done

    # Test admin user
    local admin_count=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM users WHERE username='admin';")
    if [ "$admin_count" -gt 0 ]; then
        print_pass "Admin user exists"
    else
        print_fail "Admin user not found"
    fi
}

# Test 3: Redis Connectivity
test_redis() {
    print_test "Redis Connectivity"

    if [ "$REDIS_ENABLED" != "true" ]; then
        print_skip "Redis is disabled in configuration"
        return
    fi

    if redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" ping &>/dev/null; then
        print_pass "Redis connection successful ($REDIS_HOST:$REDIS_PORT)"

        # Test set/get
        redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" SET protei:test "test_value" &>/dev/null
        local value=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" GET protei:test)
        if [ "$value" = "test_value" ]; then
            print_pass "Redis read/write operations work"
        fi
        redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" DEL protei:test &>/dev/null
    else
        print_fail "Cannot connect to Redis"
    fi
}

# Test 4: Application Startup
test_application_startup() {
    print_test "Application Startup"

    # Check if already running
    if "$INSTALL_DIR/scripts/status" &>/dev/null; then
        print_pass "Application is running"
        APP_WAS_RUNNING=true
    else
        print_info "Application not running, attempting to start..."
        APP_WAS_RUNNING=false

        # Try to start
        if sudo "$INSTALL_DIR/scripts/start" &>/dev/null; then
            sleep 5  # Wait for startup
            if "$INSTALL_DIR/scripts/status" &>/dev/null; then
                print_pass "Application started successfully"
            else
                print_fail "Application failed to start"
                return
            fi
        else
            print_fail "Cannot start application"
            return
        fi
    fi

    # Check PID file
    if [ -f "$TMP_DIR/protei-monitoring.pid" ]; then
        local pid=$(cat "$TMP_DIR/protei-monitoring.pid")
        if ps -p "$pid" &>/dev/null; then
            print_pass "Process is running (PID: $pid)"
        else
            print_fail "PID file exists but process not running"
        fi
    fi
}

# Test 5: Web Server Availability
test_web_server() {
    print_test "Web Server Availability"

    local web_port=${WEB_PORT:-8080}

    # Check if port is listening
    if netstat -tulpn 2>/dev/null | grep -q ":$web_port "; then
        print_pass "Web server is listening on port $web_port"
    else
        print_fail "Web server not listening on port $web_port"
        return
    fi

    # Test health endpoint
    if curl -s "http://localhost:$web_port/health" &>/dev/null; then
        print_pass "Health endpoint responds"
    else
        print_fail "Health endpoint not responding"
    fi
}

# Test 6: API Endpoints
test_api_endpoints() {
    print_test "API Endpoints"

    local web_port=${WEB_PORT:-8080}
    local endpoints=(
        "/health"
        "/api/protocols"
        "/api/knowledge/standards"
        "/api/knowledge/protocols"
        "/api/analysis/issues"
        "/api/analysis/statistics"
        "/api/flows/templates"
        "/api/subscribers"
    )

    local endpoints_ok=0
    local endpoints_fail=0

    for endpoint in "${endpoints[@]}"; do
        local http_code=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:$web_port$endpoint")
        if [ "$http_code" = "200" ] || [ "$http_code" = "401" ]; then
            print_pass "Endpoint $endpoint responds (HTTP $http_code)"
            ((endpoints_ok++))
        else
            print_fail "Endpoint $endpoint failed (HTTP $http_code)"
            ((endpoints_fail++))
        fi
    done

    if [ $endpoints_ok -eq ${#endpoints[@]} ]; then
        print_pass "All API endpoints operational ($endpoints_ok/${#endpoints[@]})"
    fi
}

# Test 7: Knowledge Base
test_knowledge_base() {
    print_test "Knowledge Base"

    local web_port=${WEB_PORT:-8080}

    # Test standards endpoint
    local standards=$(curl -s "http://localhost:$web_port/api/knowledge/standards")
    if [ -n "$standards" ]; then
        local count=$(echo "$standards" | grep -o '"id"' | wc -l)
        if [ "$count" -ge 18 ]; then
            print_pass "Knowledge base contains 18+ standards"
        else
            print_fail "Knowledge base has only $count standards (expected 18+)"
        fi
    else
        print_fail "Cannot retrieve standards from knowledge base"
    fi

    # Test protocols endpoint
    local protocols=$(curl -s "http://localhost:$web_port/api/knowledge/protocols")
    if [ -n "$protocols" ]; then
        local count=$(echo "$protocols" | grep -o '"' | wc -l)
        if [ "$count" -ge 10 ]; then
            print_pass "Knowledge base contains 10+ protocols"
        else
            print_info "Knowledge base has protocols defined"
        fi
    fi
}

# Test 8: AI Analysis Engine
test_ai_analysis() {
    print_test "AI Analysis Engine"

    local web_port=${WEB_PORT:-8080}

    # Test statistics endpoint
    local stats=$(curl -s "http://localhost:$web_port/api/analysis/statistics")
    if [ -n "$stats" ]; then
        print_pass "AI analysis statistics available"
    else
        print_fail "Cannot retrieve AI analysis statistics"
    fi

    # Test issues endpoint
    local issues=$(curl -s "http://localhost:$web_port/api/analysis/issues")
    if [ -n "$issues" ]; then
        print_pass "AI analysis issues endpoint operational"
    else
        print_fail "Cannot retrieve AI analysis issues"
    fi
}

# Test 9: Flow Reconstructor
test_flow_reconstructor() {
    print_test "Flow Reconstructor"

    local web_port=${WEB_PORT:-8080}

    # Test templates endpoint
    local templates=$(curl -s "http://localhost:$web_port/api/flows/templates")
    if [ -n "$templates" ]; then
        local count=$(echo "$templates" | grep -o '"name"' | wc -l)
        if [ "$count" -ge 5 ]; then
            print_pass "Flow reconstructor has 5+ procedure templates"
        else
            print_info "Flow reconstructor has $count templates"
        fi
    else
        print_fail "Cannot retrieve flow templates"
    fi
}

# Test 10: Subscriber Correlation
test_subscriber_correlation() {
    print_test "Subscriber Correlation"

    local web_port=${WEB_PORT:-8080}

    # Test subscribers endpoint
    local response=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:$web_port/api/subscribers")
    if [ "$response" = "200" ]; then
        print_pass "Subscriber correlation endpoint operational"
    else
        print_fail "Subscriber correlation endpoint not responding correctly (HTTP $response)"
    fi
}

# Test 11: Log Files
test_log_files() {
    print_test "Log Files"

    local log_categories=("application" "system" "debug" "error" "access")
    local logs_ok=true

    for category in "${log_categories[@]}"; do
        if [ -d "$LOG_DIR/$category" ]; then
            print_pass "Log directory $category exists"
        else
            print_fail "Log directory $category missing"
            logs_ok=false
        fi
    done

    # Check if log files are being written
    if [ -f "$LOG_DIR/application/protei-monitoring.log" ]; then
        local log_size=$(stat -f%z "$LOG_DIR/application/protei-monitoring.log" 2>/dev/null || stat -c%s "$LOG_DIR/application/protei-monitoring.log" 2>/dev/null)
        if [ "$log_size" -gt 0 ]; then
            print_pass "Application log is being written"
        else
            print_info "Application log exists but is empty"
        fi
    fi
}

# Test 12: CDR Directories
test_cdr_directories() {
    print_test "CDR Output Directories"

    local protocols=("MAP" "CAP" "INAP" "Diameter" "GTP" "PFCP" "HTTP2" "NGAP" "S1AP" "NAS" "combined")
    local cdr_ok=true

    for protocol in "${protocols[@]}"; do
        if [ -d "$INSTALL_DIR/cdr/$protocol" ]; then
            print_pass "CDR directory $protocol exists"
        else
            print_fail "CDR directory $protocol missing"
            cdr_ok=false
        fi
    done
}

# Test 13: File Permissions
test_permissions() {
    print_test "File Permissions"

    # Check config directory permissions
    local config_perms=$(stat -c%a "$CONFIG_DIR" 2>/dev/null || stat -f%Lp "$CONFIG_DIR")
    if [ "$config_perms" = "750" ] || [ "$config_perms" = "755" ]; then
        print_pass "Config directory has correct permissions"
    else
        print_info "Config directory permissions: $config_perms (expected 750 or 755)"
    fi

    # Check sensitive files
    local sensitive_files=("db.cfg" "license.cfg" "security.cfg")
    for file in "${sensitive_files[@]}"; do
        if [ -f "$CONFIG_DIR/$file" ]; then
            local perms=$(stat -c%a "$CONFIG_DIR/$file" 2>/dev/null || stat -f%Lp "$CONFIG_DIR/$file")
            if [ "$perms" = "600" ] || [ "$perms" = "640" ]; then
                print_pass "$file has secure permissions ($perms)"
            else
                print_fail "$file has insecure permissions ($perms), should be 600 or 640"
            fi
        fi
    done
}

# Test 14: Protocol Decoders
test_protocol_decoders() {
    print_test "Protocol Decoders"

    local web_port=${WEB_PORT:-8080}

    # Get protocol list
    local protocols=$(curl -s "http://localhost:$web_port/api/protocols")
    if [ -n "$protocols" ]; then
        local expected_protocols=("MAP" "CAP" "INAP" "Diameter" "GTP" "PFCP" "HTTP/2" "NGAP" "S1AP" "NAS")
        local found=0

        for proto in "${expected_protocols[@]}"; do
            if echo "$protocols" | grep -q "$proto"; then
                ((found++))
            fi
        done

        if [ $found -eq ${#expected_protocols[@]} ]; then
            print_pass "All 10 protocol decoders registered"
        else
            print_fail "Only $found/10 protocol decoders found"
        fi
    else
        print_fail "Cannot retrieve protocol list"
    fi
}

# Test 15: Disk Space
test_disk_space() {
    print_test "Disk Space"

    local disk_usage=$(df -h "$INSTALL_DIR" | tail -1 | awk '{print $5}' | sed 's/%//')
    if [ "$disk_usage" -lt 80 ]; then
        print_pass "Disk usage is acceptable ($disk_usage%)"
    elif [ "$disk_usage" -lt 90 ]; then
        print_info "Disk usage is high ($disk_usage%)"
    else
        print_fail "Disk usage is critical ($disk_usage%)"
    fi
}

# Test 16: Memory Usage
test_memory() {
    print_test "Memory Usage"

    local mem_usage=$(free | grep Mem | awk '{print int($3/$2 * 100)}')
    if [ "$mem_usage" -lt 80 ]; then
        print_pass "Memory usage is acceptable ($mem_usage%)"
    elif [ "$mem_usage" -lt 90 ]; then
        print_info "Memory usage is high ($mem_usage%)"
    else
        print_fail "Memory usage is critical ($mem_usage%)"
    fi
}

# Print summary
print_summary() {
    local total_tests=$((TESTS_PASSED + TESTS_FAILED + TESTS_SKIPPED))

    echo ""
    echo -e "${BLUE}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Test Summary"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${NC}"
    echo ""
    echo "Total Tests: $total_tests"
    echo -e "${GREEN}Passed:      $TESTS_PASSED${NC}"
    echo -e "${RED}Failed:      $TESTS_FAILED${NC}"
    echo -e "${YELLOW}Skipped:     $TESTS_SKIPPED${NC}"
    echo ""

    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}✅ ALL TESTS PASSED!${NC}"
        echo ""
        echo "Protei Monitoring v2.0 is functioning correctly."
        return 0
    else
        echo -e "${RED}❌ SOME TESTS FAILED${NC}"
        echo ""
        echo "Please review the failed tests above and check:"
        echo "  - Configuration files in $CONFIG_DIR/"
        echo "  - Application logs in $LOG_DIR/"
        echo "  - Database connectivity and schema"
        echo ""
        return 1
    fi
}

# Cleanup (stop app if we started it)
cleanup() {
    if [ "$APP_WAS_RUNNING" = false ] && [ -n "$STOP_ON_EXIT" ]; then
        print_info "Stopping application (was started by test script)"
        sudo "$INSTALL_DIR/scripts/stop" &>/dev/null
    fi
}

# Main test execution
main() {
    print_header

    load_config

    # Run all tests
    test_configuration_files
    test_database
    test_redis
    test_application_startup
    test_web_server
    test_api_endpoints
    test_knowledge_base
    test_ai_analysis
    test_flow_reconstructor
    test_subscriber_correlation
    test_log_files
    test_cdr_directories
    test_permissions
    test_protocol_decoders
    test_disk_space
    test_memory

    # Print summary
    local result
    print_summary
    result=$?

    # Cleanup
    trap cleanup EXIT

    exit $result
}

# Run tests
main "$@"
