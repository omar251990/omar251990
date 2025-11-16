#!/bin/bash
# Comprehensive Performance Test for Protei_Bulk
# Tests: 5,000 TPS and 2,000 messages/second targets

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
API_HOST="${API_HOST:-http://localhost:8080}"
RESULTS_DIR="results/performance_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$RESULTS_DIR"

echo "╔═══════════════════════════════════════════════════════════╗"
echo "║       Protei_Bulk Performance & Load Testing             ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""
echo "Target: $API_HOST"
echo "Results: $RESULTS_DIR"
echo ""

# Check if Locust is installed
if ! command -v locust &> /dev/null; then
    echo -e "${RED}ERROR: Locust is not installed${NC}"
    echo "Install with: pip install locust"
    exit 1
fi

# ============================================
# TEST 1: BASELINE (100 users, 1 minute)
# ============================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Test 1: Baseline Load (100 users, 1 min)${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

locust -f tests/load/locustfile.py \
    --host="$API_HOST" \
    --users 100 \
    --spawn-rate 10 \
    --run-time 1m \
    --headless \
    --csv="$RESULTS_DIR/baseline" \
    --html="$RESULTS_DIR/baseline.html"

echo -e "${GREEN}✓ Baseline test complete${NC}"
echo ""

# ============================================
# TEST 2: MEDIUM LOAD (500 users, 3 minutes)
# ============================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Test 2: Medium Load (500 users, 3 min)${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

locust -f tests/load/locustfile.py \
    --host="$API_HOST" \
    --users 500 \
    --spawn-rate 25 \
    --run-time 3m \
    --headless \
    --csv="$RESULTS_DIR/medium" \
    --html="$RESULTS_DIR/medium.html"

echo -e "${GREEN}✓ Medium load test complete${NC}"
echo ""

# ============================================
# TEST 3: HIGH LOAD (2000 users, 5 minutes)
# ============================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Test 3: High Load (2000 users, 5 min)${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

locust -f tests/load/locustfile.py \
    --host="$API_HOST" \
    --users 2000 \
    --spawn-rate 100 \
    --run-time 5m \
    --headless \
    --csv="$RESULTS_DIR/high" \
    --html="$RESULTS_DIR/high.html"

echo -e "${GREEN}✓ High load test complete${NC}"
echo ""

# ============================================
# TEST 4: TARGET TPS (5000 TPS, 5 minutes)
# ============================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Test 4: Target TPS (5000 users, 5 min)${NC}"
echo -e "${BLUE}Target: 5,000 TPS${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

locust -f tests/load/locustfile.py \
    --host="$API_HOST" \
    --users 5000 \
    --spawn-rate 200 \
    --run-time 5m \
    --headless \
    --csv="$RESULTS_DIR/target_tps" \
    --html="$RESULTS_DIR/target_tps.html"

echo -e "${GREEN}✓ Target TPS test complete${NC}"
echo ""

# ============================================
# TEST 5: SPIKE TEST (10000 users, 2 minutes)
# ============================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Test 5: Spike Test (10000 users, 2 min)${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

locust -f tests/load/locustfile.py \
    --host="$API_HOST" \
    --users 10000 \
    --spawn-rate 500 \
    --run-time 2m \
    --headless \
    --csv="$RESULTS_DIR/spike" \
    --html="$RESULTS_DIR/spike.html"

echo -e "${GREEN}✓ Spike test complete${NC}"
echo ""

# ============================================
# TEST 6: SUSTAINED LOAD (2000 users, 30 min)
# ============================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Test 6: Sustained Load (2000 users, 30 min)${NC}"
echo -e "${BLUE}Target: 2,000 messages/second sustained${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

locust -f tests/load/locustfile.py \
    --host="$API_HOST" \
    --users 2000 \
    --spawn-rate 100 \
    --run-time 30m \
    --headless \
    --csv="$RESULTS_DIR/sustained" \
    --html="$RESULTS_DIR/sustained.html"

echo -e "${GREEN}✓ Sustained load test complete${NC}"
echo ""

# ============================================
# ANALYZE RESULTS
# ============================================
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║                   PERFORMANCE REPORT                      ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""

# Function to parse Locust CSV results
parse_results() {
    local test_name=$1
    local csv_file="$RESULTS_DIR/${test_name}_stats.csv"

    if [ -f "$csv_file" ]; then
        echo "─────────────────────────────────────────────"
        echo "$test_name Results:"
        echo "─────────────────────────────────────────────"

        # Extract key metrics from CSV
        tail -n 1 "$csv_file" | awk -F',' '{
            printf "  Requests: %s\n", $2
            printf "  Failures: %s (%.1f%%)\n", $3, ($3/$2)*100
            printf "  Avg Response Time: %.0f ms\n", $5
            printf "  Min Response Time: %.0f ms\n", $6
            printf "  Max Response Time: %.0f ms\n", $7
            printf "  Requests/sec: %.0f\n", $11
        }'
        echo ""
    fi
}

# Parse all results
parse_results "baseline"
parse_results "medium"
parse_results "high"
parse_results "target_tps"
parse_results "spike"
parse_results "sustained"

# ============================================
# PASS/FAIL CRITERIA
# ============================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "PASS/FAIL CRITERIA"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check if target TPS was achieved
TPS_RESULT=$(tail -n 1 "$RESULTS_DIR/target_tps_stats.csv" 2>/dev/null | awk -F',' '{print $11}')
TPS_THRESHOLD=5000

if [ ! -z "$TPS_RESULT" ]; then
    TPS_INT=$(echo "$TPS_RESULT" | awk '{print int($1)}')

    if [ "$TPS_INT" -ge "$TPS_THRESHOLD" ]; then
        echo -e "${GREEN}✅ TPS Target: PASS${NC}"
        echo "   Achieved: $TPS_INT TPS (target: $TPS_THRESHOLD TPS)"
    else
        echo -e "${RED}❌ TPS Target: FAIL${NC}"
        echo "   Achieved: $TPS_INT TPS (target: $TPS_THRESHOLD TPS)"
    fi
else
    echo -e "${YELLOW}⚠️  TPS Target: UNKNOWN (no data)${NC}"
fi
echo ""

# Check sustained throughput
SUSTAINED_RPS=$(tail -n 1 "$RESULTS_DIR/sustained_stats.csv" 2>/dev/null | awk -F',' '{print $11}')
SUSTAINED_THRESHOLD=2000

if [ ! -z "$SUSTAINED_RPS" ]; then
    SUSTAINED_INT=$(echo "$SUSTAINED_RPS" | awk '{print int($1)}')

    if [ "$SUSTAINED_INT" -ge "$SUSTAINED_THRESHOLD" ]; then
        echo -e "${GREEN}✅ Sustained Throughput: PASS${NC}"
        echo "   Achieved: $SUSTAINED_INT msgs/sec (target: $SUSTAINED_THRESHOLD msgs/sec)"
    else
        echo -e "${RED}❌ Sustained Throughput: FAIL${NC}"
        echo "   Achieved: $SUSTAINED_INT msgs/sec (target: $SUSTAINED_THRESHOLD msgs/sec)"
    fi
else
    echo -e "${YELLOW}⚠️  Sustained Throughput: UNKNOWN (no data)${NC}"
fi
echo ""

# Check response time (p95 < 200ms)
AVG_RESPONSE=$(tail -n 1 "$RESULTS_DIR/target_tps_stats.csv" 2>/dev/null | awk -F',' '{print $5}')

if [ ! -z "$AVG_RESPONSE" ]; then
    RESPONSE_INT=$(echo "$AVG_RESPONSE" | awk '{print int($1)}')

    if [ "$RESPONSE_INT" -lt 200 ]; then
        echo -e "${GREEN}✅ Response Time: PASS${NC}"
        echo "   Average: ${RESPONSE_INT}ms (target: <200ms)"
    else
        echo -e "${YELLOW}⚠️  Response Time: WARNING${NC}"
        echo "   Average: ${RESPONSE_INT}ms (target: <200ms)"
    fi
else
    echo -e "${YELLOW}⚠️  Response Time: UNKNOWN (no data)${NC}"
fi
echo ""

# ============================================
# SYSTEM METRICS DURING TEST
# ============================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "SYSTEM RESOURCE USAGE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "During peak load, check these metrics manually:"
echo ""
echo "1. CPU Usage:"
echo "   htop"
echo "   Target: <80% sustained"
echo ""
echo "2. Memory Usage:"
echo "   free -h"
echo "   Target: <80% of available RAM"
echo ""
echo "3. Database Connections:"
echo "   psql -c 'SELECT count(*) FROM pg_stat_activity;'"
echo "   Target: <100 connections"
echo ""
echo "4. Redis Memory:"
echo "   redis-cli INFO memory"
echo "   Target: <2GB"
echo ""
echo "5. Message Queue Depth:"
echo "   redis-cli LLEN message_queue"
echo "   Target: <10000 pending"
echo ""

# ============================================
# SUMMARY
# ============================================
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║                        SUMMARY                            ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""
echo "All test results saved to: $RESULTS_DIR"
echo ""
echo "View HTML reports:"
echo "  - Baseline: $RESULTS_DIR/baseline.html"
echo "  - Medium: $RESULTS_DIR/medium.html"
echo "  - High: $RESULTS_DIR/high.html"
echo "  - Target TPS: $RESULTS_DIR/target_tps.html"
echo "  - Spike: $RESULTS_DIR/spike.html"
echo "  - Sustained: $RESULTS_DIR/sustained.html"
echo ""
echo "Next Steps:"
echo "  1. Review HTML reports for detailed metrics"
echo "  2. Check application logs for errors during tests"
echo "  3. Analyze database slow query log"
echo "  4. Review system resource usage graphs"
echo "  5. Tune configuration if targets not met"
echo ""
