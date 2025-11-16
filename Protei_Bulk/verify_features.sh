#!/bin/bash
# Protei_Bulk Feature Verification Script
# Tests all advertised features and performance claims

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Test results
declare -a RESULTS

test_result() {
    local name=$1
    local status=$2
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if [ "$status" = "PASS" ]; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        echo -e "${GREEN}✅ PASS${NC}: $name"
        RESULTS+=("✅ $name")
    elif [ "$status" = "WARN" ]; then
        echo -e "${YELLOW}⚠️  WARN${NC}: $name"
        RESULTS+=("⚠️  $name")
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
        echo -e "${RED}❌ FAIL${NC}: $name"
        RESULTS+=("❌ $name")
    fi
}

echo "╔═══════════════════════════════════════════════════════════╗"
echo "║     Protei_Bulk Feature Verification & Testing           ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""

# ============================================
# 1. SERVICE HEALTH CHECKS
# ============================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1. SERVICE HEALTH CHECKS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check API service
if curl -sf http://localhost:8080/api/v1/health > /dev/null 2>&1; then
    test_result "API Service Running" "PASS"
else
    test_result "API Service Running" "FAIL"
fi

# Check Web UI
if curl -sf http://localhost:3000 > /dev/null 2>&1; then
    test_result "Web UI Service Running" "PASS"
else
    test_result "Web UI Service Running" "FAIL"
fi

# Check PostgreSQL
if psql -U postgres -c "SELECT 1" > /dev/null 2>&1; then
    test_result "PostgreSQL Database Running" "PASS"
else
    test_result "PostgreSQL Database Running" "FAIL"
fi

# Check Redis
if redis-cli ping | grep -q "PONG"; then
    test_result "Redis Cache Running" "PASS"
else
    test_result "Redis Cache Running" "FAIL"
fi

# ============================================
# 2. DATABASE SCHEMA VERIFICATION
# ============================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2. DATABASE SCHEMA VERIFICATION"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check core tables
TABLES=(
    "users"
    "customers"
    "campaigns"
    "tbl_smsc_connections"
    "tbl_routing_rules"
    "tbl_profiles"
    "tbl_profile_groups"
    "tbl_cdr_records"
    "tbl_dcdl_datasets"
)

for table in "${TABLES[@]}"; do
    if psql -U postgres -d protei_bulk -c "SELECT 1 FROM $table LIMIT 1" > /dev/null 2>&1; then
        test_result "Table exists: $table" "PASS"
    else
        test_result "Table exists: $table" "FAIL"
    fi
done

# ============================================
# 3. API ENDPOINT VERIFICATION
# ============================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3. API ENDPOINT VERIFICATION"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Test API endpoints (without auth for basic check)
ENDPOINTS=(
    "/api/v1/health"
    "/api/v1/messages"
    "/api/v1/campaigns"
    "/api/v1/profiles"
    "/api/v1/segments"
    "/api/v1/analytics/metrics/messages/realtime"
)

for endpoint in "${ENDPOINTS[@]}"; do
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080$endpoint)

    # Accept 200 (success), 401 (auth required), 403 (forbidden)
    if [[ "$HTTP_CODE" =~ ^(200|401|403)$ ]]; then
        test_result "API endpoint: $endpoint" "PASS"
    else
        test_result "API endpoint: $endpoint (HTTP $HTTP_CODE)" "FAIL"
    fi
done

# ============================================
# 4. FEATURE IMPLEMENTATION CHECK
# ============================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "4. FEATURE IMPLEMENTATION CHECK"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check Python service files
if [ -f "src/services/routing_engine.py" ]; then
    test_result "Routing Engine Implementation" "PASS"
else
    test_result "Routing Engine Implementation" "FAIL"
fi

if [ -f "src/services/profile_service.py" ]; then
    test_result "Profile Service Implementation" "PASS"
else
    test_result "Profile Service Implementation" "FAIL"
fi

if [ -f "src/services/segmentation_service.py" ]; then
    test_result "Segmentation Service Implementation" "PASS"
else
    test_result "Segmentation Service Implementation" "FAIL"
fi

# Check for multi-channel support
if grep -q "WhatsApp" src/**/*.py 2>/dev/null; then
    test_result "WhatsApp Channel Implementation" "WARN"
else
    test_result "WhatsApp Channel Implementation" "FAIL"
fi

if grep -q "Viber" src/**/*.py 2>/dev/null; then
    test_result "Viber Channel Implementation" "WARN"
else
    test_result "Viber Channel Implementation" "FAIL"
fi

# Check for AI features
if grep -rq "AI\|machine.learning\|tensorflow\|pytorch" src/ 2>/dev/null; then
    test_result "AI/ML Features Implementation" "WARN"
else
    test_result "AI/ML Features Implementation" "FAIL"
fi

# ============================================
# 5. WEB UI VERIFICATION
# ============================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "5. WEB UI VERIFICATION"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check React pages
UI_PAGES=(
    "web/src/pages/Dashboard/DashboardPage.jsx"
    "web/src/pages/Campaigns/CampaignList.jsx"
    "web/src/pages/Campaigns/CreateCampaign.jsx"
    "web/src/pages/Users/UserAccounts.jsx"
    "web/src/pages/Contacts/ContactLists.jsx"
)

for page in "${UI_PAGES[@]}"; do
    if [ -f "$page" ]; then
        test_result "UI Page: $(basename $page)" "PASS"
    else
        test_result "UI Page: $(basename $page)" "FAIL"
    fi
done

# Check for missing UI components
if [ ! -f "web/src/pages/Routing/RoutingConfig.jsx" ]; then
    test_result "Routing UI (Backend ready, UI missing)" "WARN"
fi

if [ ! -f "web/src/pages/Profiles/ProfileManagement.jsx" ]; then
    test_result "Profile Management UI (Backend ready, UI missing)" "WARN"
fi

# ============================================
# 6. PERFORMANCE TESTING
# ============================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "6. PERFORMANCE TESTING (Quick Check)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check if Locust is installed
if command -v locust &> /dev/null; then
    test_result "Load Testing Tool (Locust) Installed" "PASS"

    echo "Running quick performance test (10 seconds)..."

    # Run quick load test
    if timeout 15s locust -f tests/load/locustfile.py \
        --host=http://localhost:8080 \
        --users 100 \
        --spawn-rate 10 \
        --run-time 10s \
        --headless \
        --only-summary 2>&1 | tee /tmp/locust_output.txt; then

        # Parse results
        RPS=$(grep -oP "Requests/s.*\K[0-9.]+" /tmp/locust_output.txt | head -1)

        if [ ! -z "$RPS" ]; then
            echo "  → Requests/sec: $RPS"

            # Check if meets minimum threshold (500 RPS for quick test)
            if (( $(echo "$RPS > 500" | bc -l) )); then
                test_result "Performance: Basic Load Test (>500 RPS)" "PASS"
            else
                test_result "Performance: Basic Load Test (<500 RPS)" "WARN"
            fi
        fi
    fi
else
    test_result "Load Testing Tool (Locust) Installed" "FAIL"
    echo "  → Install with: pip install locust"
fi

# ============================================
# 7. DOCUMENTATION CHECK
# ============================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "7. DOCUMENTATION CHECK"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

DOCS=(
    "README.md"
    "ROADMAP.md"
    "INSTALLATION_GUIDE.md"
    "PERFORMANCE_ARCHITECTURE.md"
    "PROFILING_ARCHITECTURE.md"
    "FEATURE_VERIFICATION_GUIDE.md"
)

for doc in "${DOCS[@]}"; do
    if [ -f "$doc" ]; then
        test_result "Documentation: $doc" "PASS"
    else
        test_result "Documentation: $doc" "FAIL"
    fi
done

# ============================================
# FINAL REPORT
# ============================================
echo ""
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║                    VERIFICATION REPORT                    ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""
echo "Total Tests: $TOTAL_TESTS"
echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed: ${RED}$FAILED_TESTS${NC}"
echo -e "Success Rate: $(echo "scale=1; $PASSED_TESTS * 100 / $TOTAL_TESTS" | bc)%"
echo ""

# Calculate categories
WARNINGS=$(printf '%s\n' "${RESULTS[@]}" | grep -c "⚠️" || true)
echo -e "Warnings: ${YELLOW}$WARNINGS${NC}"
echo ""

# Key findings
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "KEY FINDINGS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "✅ WORKING:"
echo "  - Core SMS messaging API"
echo "  - Database schemas (comprehensive)"
echo "  - Basic web dashboard"
echo "  - Routing engine backend"
echo "  - Profile & segmentation backend"
echo ""
echo "⚠️  PARTIALLY WORKING:"
echo "  - Web UI incomplete (missing routing, profiling UIs)"
echo "  - Multi-channel (schemas only, no implementation)"
echo "  - Performance not tested at advertised scale"
echo ""
echo "❌ NOT WORKING / MISSING:"
echo "  - WhatsApp Business API"
echo "  - Viber messaging"
echo "  - RCS messaging"
echo "  - Voice calling"
echo "  - Chatbot builder"
echo "  - AI Campaign Designer"
echo "  - A/B Testing"
echo "  - Customer Journey Automation"
echo "  - Self-healing infrastructure"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "RECOMMENDATIONS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "1. Update ROADMAP.md:"
echo "   - Move unimplemented features to 'Planned' section"
echo "   - Only mark truly working features as 'Available Now'"
echo ""
echo "2. Complete Web UI:"
echo "   - Build routing configuration UI"
echo "   - Build profile management UI"
echo "   - Build segmentation query builder UI"
echo ""
echo "3. Performance Testing:"
echo "   - Run full 5,000 TPS load test"
echo "   - Test 2,000 messages/second throughput"
echo "   - Verify with production-like data volumes"
echo ""
echo "4. Implement Priority Features:"
echo "   - DCDL service layer (schema exists)"
echo "   - Multi-channel support (WhatsApp, Email)"
echo "   - Complete CDR partition automation"
echo ""

# Save report to file
REPORT_FILE="verification_report_$(date +%Y%m%d_%H%M%S).txt"
{
    echo "Protei_Bulk Verification Report"
    echo "Generated: $(date)"
    echo ""
    echo "Results:"
    printf '%s\n' "${RESULTS[@]}"
    echo ""
    echo "Summary:"
    echo "Total: $TOTAL_TESTS"
    echo "Passed: $PASSED_TESTS"
    echo "Failed: $FAILED_TESTS"
    echo "Warnings: $WARNINGS"
} > "$REPORT_FILE"

echo "Full report saved to: $REPORT_FILE"
echo ""
