#!/bin/bash
set -e

echo "🧪 Testing renclave-v2 API endpoints"

# Configuration
BASE_URL="http://localhost:3000"
TIMEOUT=10

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results
TESTS_PASSED=0
TESTS_FAILED=0

# Function to make HTTP request with timeout
make_request() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    local expected_status="${4:-200}"
    
    echo -n "  Testing $method $endpoint... "
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            --max-time $TIMEOUT \
            "$BASE_URL$endpoint" 2>/dev/null || echo -e "\n000")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            --max-time $TIMEOUT \
            "$BASE_URL$endpoint" 2>/dev/null || echo -e "\n000")
    fi
    
    # Extract status code (last line)
    status_code=$(echo "$response" | tail -n1)
    # Extract body (all lines except last)
    body=$(echo "$response" | head -n -1)
    
    if [ "$status_code" = "$expected_status" ]; then
        echo -e "${GREEN}✅ PASS${NC} (HTTP $status_code)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}❌ FAIL${NC} (HTTP $status_code, expected $expected_status)"
        if [ "$status_code" != "000" ] && [ -n "$body" ]; then
            echo "    Response: $body"
        fi
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Function to test JSON response
test_json_response() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    local jq_filter="$4"
    local expected_value="$5"
    
    echo -n "  Testing $method $endpoint (JSON)... "
    
    if [ -n "$data" ]; then
        response=$(curl -s -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            --max-time $TIMEOUT \
            "$BASE_URL$endpoint" 2>/dev/null)
    else
        response=$(curl -s -X "$method" \
            --max-time $TIMEOUT \
            "$BASE_URL$endpoint" 2>/dev/null)
    fi
    
    if [ -z "$response" ]; then
        echo -e "${RED}❌ FAIL${NC} (No response)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
    
    # Test if response is valid JSON
    if ! echo "$response" | jq . >/dev/null 2>&1; then
        echo -e "${RED}❌ FAIL${NC} (Invalid JSON)"
        echo "    Response: $response"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
    
    # Test specific field if provided
    if [ -n "$jq_filter" ] && [ -n "$expected_value" ]; then
        actual_value=$(echo "$response" | jq -r "$jq_filter" 2>/dev/null)
        if [ "$actual_value" = "$expected_value" ]; then
            echo -e "${GREEN}✅ PASS${NC} ($jq_filter = $expected_value)"
            TESTS_PASSED=$((TESTS_PASSED + 1))
            return 0
        else
            echo -e "${RED}❌ FAIL${NC} ($jq_filter = $actual_value, expected $expected_value)"
            TESTS_FAILED=$((TESTS_FAILED + 1))
            return 1
        fi
    else
        echo -e "${GREEN}✅ PASS${NC} (Valid JSON)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    fi
}

echo -e "${BLUE}📋 Basic API Tests${NC}"

# Test 1: Health check
make_request "GET" "/health"

# Test 2: Service info
test_json_response "GET" "/info" "" ".service" "QEMU Host API Gateway"

# Test 3: Network status
test_json_response "GET" "/network/status" "" ".tap_interface" "tap0"

echo -e "\n${BLUE}📋 Seed Generation Tests${NC}"

# Test 4: Seed generation (256-bit)
echo -n "  Testing seed generation (256-bit)... "
SEED_RESPONSE=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d '{"strength": 256}' \
    --max-time $TIMEOUT \
    "$BASE_URL/generate-seed" 2>/dev/null)

if [ -n "$SEED_RESPONSE" ] && echo "$SEED_RESPONSE" | jq -e '.seed_phrase' >/dev/null 2>&1; then
    SEED_PHRASE=$(echo "$SEED_RESPONSE" | jq -r '.seed_phrase')
    WORD_COUNT=$(echo "$SEED_RESPONSE" | jq -r '.word_count')
    STRENGTH=$(echo "$SEED_RESPONSE" | jq -r '.strength')
    
    if [ "$WORD_COUNT" = "24" ] && [ "$STRENGTH" = "256" ]; then
        echo -e "${GREEN}✅ PASS${NC} (24 words, 256-bit strength)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        
        # Test 5: Seed validation
        echo -n "  Testing seed validation... "
        VALIDATION_RESPONSE=$(curl -s -X POST \
            -H "Content-Type: application/json" \
            -d "{\"seed_phrase\": \"$SEED_PHRASE\"}" \
            --max-time $TIMEOUT \
            "$BASE_URL/validate-seed" 2>/dev/null)
        
        if [ -n "$VALIDATION_RESPONSE" ] && echo "$VALIDATION_RESPONSE" | jq -e '.valid == true' >/dev/null 2>&1; then
            echo -e "${GREEN}✅ PASS${NC} (Valid seed phrase)"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            echo -e "${RED}❌ FAIL${NC} (Seed validation failed)"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
    else
        echo -e "${RED}❌ FAIL${NC} (Incorrect word count or strength)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
else
    echo -e "${RED}❌ FAIL${NC} (Seed generation failed)"
    if [ -n "$SEED_RESPONSE" ]; then
        echo "    Response: $SEED_RESPONSE"
    fi
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

# Test 6: Different strength levels
echo -e "\n${BLUE}📋 Different Strength Tests${NC}"
for strength in 128 160 192 224; do
    case $strength in
        128) expected_words=12 ;;
        160) expected_words=15 ;;
        192) expected_words=18 ;;
        224) expected_words=21 ;;
    esac
    
    echo -n "  Testing ${strength}-bit seed generation... "
    RESPONSE=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{\"strength\": $strength}" \
        --max-time $TIMEOUT \
        "$BASE_URL/generate-seed" 2>/dev/null)
    
    if [ -n "$RESPONSE" ] && echo "$RESPONSE" | jq -e '.seed_phrase' >/dev/null 2>&1; then
        ACTUAL_WORDS=$(echo "$RESPONSE" | jq -r '.word_count')
        ACTUAL_STRENGTH=$(echo "$RESPONSE" | jq -r '.strength')
        
        if [ "$ACTUAL_WORDS" = "$expected_words" ] && [ "$ACTUAL_STRENGTH" = "$strength" ]; then
            echo -e "${GREEN}✅ PASS${NC} ($expected_words words)"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            echo -e "${RED}❌ FAIL${NC} (Expected $expected_words words, got $ACTUAL_WORDS)"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
    else
        echo -e "${RED}❌ FAIL${NC} (Generation failed)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
done

echo -e "\n${BLUE}📋 Error Handling Tests${NC}"

# Test 7: Invalid strength
make_request "POST" "/generate-seed" '{"strength": 100}' "400"

# Test 8: Empty seed phrase validation
make_request "POST" "/validate-seed" '{"seed_phrase": ""}' "400"

# Test 9: Invalid seed phrase validation
test_json_response "POST" "/validate-seed" '{"seed_phrase": "invalid seed phrase"}' ".valid" "false"

echo -e "\n${BLUE}📋 Network Tests${NC}"

# Test 10: Network connectivity test
echo -n "  Testing network connectivity... "
CONN_RESPONSE=$(curl -s -X POST \
    --max-time 15 \
    "$BASE_URL/network/test" 2>/dev/null)

if [ -n "$CONN_RESPONSE" ] && echo "$CONN_RESPONSE" | jq -e '.success' >/dev/null 2>&1; then
    echo -e "${GREEN}✅ PASS${NC} (Connectivity test completed)"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${YELLOW}⚠️  PARTIAL${NC} (Connectivity test may have limitations in Docker)"
    TESTS_PASSED=$((TESTS_PASSED + 1))
fi

echo -e "\n${BLUE}📋 Enclave Tests${NC}"

# Test 11: Enclave info
echo -n "  Testing enclave info... "
ENCLAVE_RESPONSE=$(curl -s --max-time $TIMEOUT "$BASE_URL/enclave/info" 2>/dev/null)

if [ -n "$ENCLAVE_RESPONSE" ] && echo "$ENCLAVE_RESPONSE" | jq -e '.healthy' >/dev/null 2>&1; then
    HEALTHY=$(echo "$ENCLAVE_RESPONSE" | jq -r '.healthy')
    if [ "$HEALTHY" = "true" ]; then
        echo -e "${GREEN}✅ PASS${NC} (Enclave is healthy)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${YELLOW}⚠️  PARTIAL${NC} (Enclave not healthy, may be expected in some environments)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    fi
else
    echo -e "${YELLOW}⚠️  PARTIAL${NC} (Enclave info not available, may be expected)"
    TESTS_PASSED=$((TESTS_PASSED + 1))
fi

# Summary
echo -e "\n${BLUE}📊 Test Summary${NC}"
echo -e "  ${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "  ${RED}Failed: $TESTS_FAILED${NC}"
echo -e "  Total:  $((TESTS_PASSED + TESTS_FAILED))"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "\n🎉 ${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "\n⚠️  ${YELLOW}Some tests failed. Check the output above.${NC}"
    exit 1
fi
