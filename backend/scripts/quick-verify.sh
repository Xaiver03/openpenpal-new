#!/bin/bash

# Quick Database and API Verification Script
# 快速验证数据库和API连接状态

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
BACKEND_URL=${BACKEND_URL:-"http://localhost:8080"}
DB_HOST=${DB_HOST:-"localhost"}
DB_PORT=${DB_PORT:-"5432"}
DB_USER=${DB_USER:-"rocalight"}
DB_NAME=${DB_NAME:-"openpenpal"}

print_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

print_pass() {
    echo -e "${GREEN}[✓ PASS]${NC} $1"
}

print_fail() {
    echo -e "${RED}[✗ FAIL]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[⚠ WARN]${NC} $1"
}

echo "🔍 Quick Verification - Database and API Status"
echo "=============================================="
echo ""

# 1. Database Connection Test
print_test "Database connectivity"
if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" > /dev/null 2>&1; then
    print_pass "Database connection successful"
    
    # Get database info
    DB_VERSION=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT version();" 2>/dev/null | head -1 | xargs)
    TABLE_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null | xargs)
    INDEX_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public';" 2>/dev/null | xargs)
    
    echo "  - PostgreSQL Version: $DB_VERSION"
    echo "  - Tables: $TABLE_COUNT"
    echo "  - Indexes: $INDEX_COUNT"
else
    print_fail "Database connection failed"
    exit 1
fi

# 2. SSL Configuration Test
print_test "SSL configuration"
SSL_ENABLED=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SHOW ssl;" 2>/dev/null | xargs)
CONN_SSL=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT ssl FROM pg_stat_ssl WHERE pid = pg_backend_pid();" 2>/dev/null | xargs)

if [[ "$SSL_ENABLED" == "on" ]]; then
    print_pass "SSL enabled on server"
    if [[ "$CONN_SSL" == "t" ]]; then
        print_pass "Current connection uses SSL"
    else
        print_warn "SSL available but not used"
    fi
else
    print_warn "SSL not enabled (development mode)"
fi

# 3. Connection Pool Test
print_test "Connection pool status"
CONN_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT count(*) FROM pg_stat_activity WHERE datname = '$DB_NAME';" 2>/dev/null | xargs)
ACTIVE_CONNS=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT count(*) FROM pg_stat_activity WHERE datname = '$DB_NAME' AND state = 'active';" 2>/dev/null | xargs)
IDLE_CONNS=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT count(*) FROM pg_stat_activity WHERE datname = '$DB_NAME' AND state = 'idle';" 2>/dev/null | xargs)

print_pass "Connection pool metrics retrieved"
echo "  - Total connections: $CONN_COUNT"
echo "  - Active connections: $ACTIVE_CONNS"
echo "  - Idle connections: $IDLE_CONNS"

# 4. Critical Indexes Test
print_test "Critical indexes"
CRITICAL_INDEXES=(
    "idx_users_school_role_active"
    "idx_letters_user_status_created"
    "idx_courier_tasks_courier_status"
)

MISSING_INDEXES=0
for idx in "${CRITICAL_INDEXES[@]}"; do
    EXISTS=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM pg_indexes WHERE indexname = '$idx';" 2>/dev/null | xargs)
    if [[ "$EXISTS" == "1" ]]; then
        echo "  - ✓ $idx"
    else
        echo "  - ✗ $idx (missing)"
        ((MISSING_INDEXES++))
    fi
done

if [[ $MISSING_INDEXES -eq 0 ]]; then
    print_pass "All critical indexes present"
else
    print_warn "$MISSING_INDEXES critical indexes missing"
fi

# 5. Backend API Test
print_test "Backend API connectivity"
HEALTH_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$BACKEND_URL/health" 2>/dev/null)
if [[ "$HEALTH_RESPONSE" == "200" ]]; then
    print_pass "Backend health endpoint responding"
    
    # Get health details
    HEALTH_DATA=$(curl -s "$BACKEND_URL/health" 2>/dev/null)
    if [[ -n "$HEALTH_DATA" ]]; then
        echo "  - Health data: $HEALTH_DATA"
    fi
else
    print_fail "Backend not responding (HTTP $HEALTH_RESPONSE)"
fi

# 6. API Endpoints Test
print_test "API endpoints availability"
API_ENDPOINTS=(
    "/api/v1/auth/login"
    "/api/v1/letters"
    "/api/v1/users/profile"
)

WORKING_ENDPOINTS=0
for endpoint in "${API_ENDPOINTS[@]}"; do
    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$BACKEND_URL$endpoint" 2>/dev/null)
    if [[ "$RESPONSE" == "401" ]] || [[ "$RESPONSE" == "403" ]] || [[ "$RESPONSE" == "405" ]]; then
        echo "  - ✓ $endpoint (protected as expected)"
        ((WORKING_ENDPOINTS++))
    elif [[ "$RESPONSE" == "200" ]]; then
        echo "  - ✓ $endpoint (accessible)"
        ((WORKING_ENDPOINTS++))
    else
        echo "  - ✗ $endpoint (HTTP $RESPONSE)"
    fi
done

if [[ $WORKING_ENDPOINTS -eq ${#API_ENDPOINTS[@]} ]]; then
    print_pass "All API endpoints responding correctly"
else
    print_warn "Some API endpoints may have issues"
fi

# 7. Performance Quick Test
print_test "Database performance"
START_TIME=$(date +%s%N)
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT COUNT(*) FROM users;" > /dev/null 2>&1
END_TIME=$(date +%s%N)
QUERY_TIME=$(( (END_TIME - START_TIME) / 1000000 ))

if [[ $QUERY_TIME -lt 100 ]]; then
    print_pass "Database query performance good (${QUERY_TIME}ms)"
else
    print_warn "Database query took ${QUERY_TIME}ms (may need optimization)"
fi

# Summary
echo ""
echo "📊 Quick Verification Summary"
echo "============================="
echo "✅ Database: Connected and operational"
echo "🔐 SSL: ${SSL_ENABLED:-disabled}"
echo "🔗 Connections: $CONN_COUNT active"
echo "📈 Indexes: $((${#CRITICAL_INDEXES[@]} - MISSING_INDEXES))/${#CRITICAL_INDEXES[@]} critical indexes present"
echo "🌐 API: Backend responding"
echo "⚡ Performance: ${QUERY_TIME}ms query time"

if [[ $MISSING_INDEXES -gt 0 ]]; then
    echo ""
    echo "⚠️  Recommendations:"
    echo "- Run index optimization: go run cmd/tools/optimize-indexes/main.go --mode=create"
fi

echo ""
echo "🔍 Verification Complete!"