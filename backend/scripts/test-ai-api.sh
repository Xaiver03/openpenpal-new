#!/bin/bash

# Test script for the Multi-Provider AI API system
# Usage: ./test-ai-api.sh [BASE_URL]

BASE_URL=${1:-"http://localhost:8080"}
API_BASE="$BASE_URL/api"

echo "🤖 Testing Multi-Provider AI API System"
echo "Base URL: $BASE_URL"
echo "========================================"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test function
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    local expected_status=${5:-200}
    
    echo -e "${BLUE}Testing:${NC} $description"
    echo -e "${YELLOW}$method${NC} $endpoint"
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$endpoint")
    fi
    
    # Extract HTTP status code (last line)
    http_code=$(echo "$response" | tail -n1)
    # Extract response body (all but last line)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "$expected_status" ]; then
        echo -e "${GREEN}✓ Success${NC} (HTTP $http_code)"
        echo "$body" | jq -r '.message // .data.provider // "Response received"' 2>/dev/null || echo "Response: OK"
    else
        echo -e "${RED}✗ Failed${NC} (HTTP $http_code, expected $expected_status)"
        echo "$body" | jq -r '.error // .message // .' 2>/dev/null || echo "$body"
    fi
    echo ""
}

# 1. Test Provider Status (Public endpoint)
echo -e "${BLUE}=== Testing Provider Management ===${NC}"
test_endpoint "GET" "$API_BASE/ai/providers/status" "" "Get AI provider status"

# 2. Test Text Generation
echo -e "${BLUE}=== Testing Text Generation ===${NC}"
test_endpoint "POST" "$API_BASE/ai/generate" '{
    "prompt": "Write a short greeting for a pen pal letter",
    "max_tokens": 100,
    "temperature": 0.7,
    "preferred_provider": "local"
}' "Generate text with local provider"

# 3. Test Chat API
echo -e "${BLUE}=== Testing Chat API ===${NC}"
test_endpoint "POST" "$API_BASE/ai/chat" '{
    "messages": [
        {"role": "user", "content": "Hello, how are you?"}
    ],
    "max_tokens": 150,
    "temperature": 0.7,
    "preferred_provider": "local"
}' "Chat conversation"

# 4. Test Text Summarization
echo -e "${BLUE}=== Testing Summarization ===${NC}"
test_endpoint "POST" "$API_BASE/ai/summarize" '{
    "text": "This is a long text that needs to be summarized. It contains multiple sentences and covers various topics. The purpose is to test the summarization functionality of our AI system. We want to see if it can extract the key points and present them in a concise manner.",
    "max_tokens": 50,
    "temperature": 0.3,
    "preferred_provider": "local"
}' "Summarize text"

# 5. Test Translation
echo -e "${BLUE}=== Testing Translation ===${NC}"
test_endpoint "POST" "$API_BASE/ai/translate" '{
    "text": "Hello, how are you today?",
    "target_language": "中文",
    "preferred_provider": "local"
}' "Translate text to Chinese"

# 6. Test Sentiment Analysis
echo -e "${BLUE}=== Testing Sentiment Analysis ===${NC}"
test_endpoint "POST" "$API_BASE/ai/sentiment" '{
    "text": "I am very happy and excited about this new AI system!",
    "preferred_provider": "local"
}' "Analyze sentiment"

# 7. Test Content Moderation
echo -e "${BLUE}=== Testing Content Moderation ===${NC}"
test_endpoint "POST" "$API_BASE/ai/moderate" '{
    "text": "This is a test message for content moderation.",
    "preferred_provider": "local"
}' "Moderate content"

# 8. Test Letter Writing Assistance
echo -e "${BLUE}=== Testing Letter Writing Assistance ===${NC}"
test_endpoint "POST" "$API_BASE/ai/letter/assist" '{
    "topic": "友情",
    "style": "friendly",
    "tone": "warm",
    "length": "medium",
    "preferred_provider": "local"
}' "Letter writing assistance"

# 9. Test Batch Translation
echo -e "${BLUE}=== Testing Batch Translation ===${NC}"
test_endpoint "POST" "$API_BASE/ai/translate/batch" '{
    "texts": ["Hello", "Good morning", "Thank you"],
    "target_language": "中文",
    "preferred_provider": "local"
}' "Batch translate texts"

# 10. Test Usage Statistics (will need authentication in real scenario)
echo -e "${BLUE}=== Testing Usage Statistics ===${NC}"
test_endpoint "GET" "$API_BASE/ai/usage/stats?days=7" "" "Get usage statistics" 401

# 11. Test Admin Endpoints (will need authentication and admin role)
echo -e "${BLUE}=== Testing Admin Endpoints ===${NC}"
test_endpoint "POST" "$API_BASE/admin/ai/providers/reload" "" "Reload AI providers" 401

test_endpoint "GET" "$API_BASE/admin/ai/config" "" "Get AI configuration" 401

test_endpoint "GET" "$API_BASE/admin/ai/monitoring" "" "Get AI monitoring data" 401

test_endpoint "GET" "$API_BASE/admin/ai/analytics" "" "Get AI analytics" 401

test_endpoint "GET" "$API_BASE/admin/ai/logs" "" "Get AI logs" 401

test_endpoint "POST" "$API_BASE/admin/ai/test-provider" '{
    "provider": "local",
    "test_type": "connection"
}' "Test AI provider connection" 401

echo -e "${GREEN}=== AI API Testing Complete ===${NC}"
echo ""
echo "Note: Authentication-required endpoints (marked with 401) are expected to fail"
echo "without proper JWT tokens. This is normal behavior for security."
echo ""
echo "To test authenticated endpoints, you need to:"
echo "1. Register/login to get a JWT token"
echo "2. Include the token in the Authorization header"
echo "3. Have appropriate permissions for admin endpoints"