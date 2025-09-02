#!/bin/bash

# Test script for Dynamic Model Discovery System
# Usage: ./test_dynamic_models.sh

set -e

echo "🚀 Dynamic Model Discovery Test Script"
echo "======================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVER_URL="http://localhost:8080"
PROVIDERS_CSV="providers.csv"
ENHANCED_SERVER="./enhanced-server-v3"

echo -e "${BLUE}📋 Test Configuration:${NC}"
echo "  Server URL: $SERVER_URL"
echo "  Providers CSV: $PROVIDERS_CSV"
echo "  Enhanced Server: $ENHANCED_SERVER"
echo ""

# Function to check if server is running
check_server() {
    echo -e "${BLUE}🔍 Checking if server is running...${NC}"
    if curl -s "$SERVER_URL/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Server is running${NC}"
        return 0
    else
        echo -e "${RED}❌ Server is not running${NC}"
        return 1
    fi
}

# Function to start server if not running
start_server() {
    if ! check_server; then
        echo -e "${YELLOW}🚀 Starting enhanced server...${NC}"
        
        # Check if binary exists
        if [ ! -f "$ENHANCED_SERVER" ]; then
            echo -e "${RED}❌ Enhanced server binary not found: $ENHANCED_SERVER${NC}"
            echo -e "${YELLOW}💡 Please build the server first with: go build -o enhanced-server-v3 ./cmd/enhanced-server/main_v3.go${NC}"
            exit 1
        fi
        
        # Start server in background
        nohup $ENHANCED_SERVER > server.log 2>&1 &
        SERVER_PID=$!
        echo "Server PID: $SERVER_PID"
        
        # Wait for server to start
        echo -e "${BLUE}⏳ Waiting for server to start...${NC}"
        for i in {1..30}; do
            if check_server; then
                echo -e "${GREEN}✅ Server started successfully${NC}"
                break
            fi
            sleep 1
            echo -n "."
        done
        
        if ! check_server; then
            echo -e "${RED}❌ Failed to start server after 30 seconds${NC}"
            echo -e "${YELLOW}📋 Server log:${NC}"
            tail -20 server.log
            exit 1
        fi
    fi
}

# Function to test API endpoint
test_endpoint() {
    local endpoint=$1
    local method=${2:-GET}
    local data=${3:-}
    local description=$4
    
    echo -e "${BLUE}🔍 Testing: $description${NC}"
    echo "  Endpoint: $method $endpoint"
    
    if [ -n "$data" ]; then
        response=$(curl -s -X "$method" -H "Content-Type: application/json" -d "$data" "$SERVER_URL$endpoint" 2>/dev/null || echo "ERROR")
    else
        response=$(curl -s -X "$method" "$SERVER_URL$endpoint" 2>/dev/null || echo "ERROR")
    fi
    
    if [ "$response" = "ERROR" ]; then
        echo -e "${RED}❌ Failed to connect to endpoint${NC}"
        return 1
    elif echo "$response" | jq . > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Success - Valid JSON response${NC}"
        echo "$response" | jq . | head -10
        return 0
    else
        echo -e "${YELLOW}⚠️  Response received but not valid JSON${NC}"
        echo "$response" | head -5
        return 1
    fi
    echo ""
}

# Function to test dynamic model loading
test_dynamic_models() {
    echo -e "${BLUE}🌐 Testing Dynamic Model Discovery${NC}"
    echo "=================================="
    
    # Test 1: Get current providers
    echo -e "${BLUE}Test 1: Get Current Providers${NC}"
    test_endpoint "/api/v1/providers" "GET" "" "Get all providers with model counts"
    echo ""
    
    # Test 2: Get system info with dynamic stats
    echo -e "${BLUE}Test 2: System Information${NC}"
    test_endpoint "/api/v1/system/info" "GET" "" "Get system info with dynamic loading stats"
    echo ""
    
    # Test 3: Get dynamic model stats
    echo -e "${BLUE}Test 3: Dynamic Model Statistics${NC}"
    test_endpoint "/api/v1/models/stats" "GET" "" "Get dynamic model loading statistics"
    echo ""
    
    # Test 4: Refresh all models
    echo -e "${BLUE}Test 4: Refresh All Models${NC}"
    test_endpoint "/api/v1/models/refresh" "POST" "" "Refresh all provider models"
    echo ""
    
    # Test 5: Process a simple request
    echo -e "${BLUE}Test 5: Process Request${NC}"
    test_endpoint "/api/v1/process" "POST" '{"request":"Write a simple Python function"}' "Process a coding request"
    echo ""
    
    # Test 6: Analyze request
    echo -e "${BLUE}Test 6: Analyze Request${NC}"
    test_endpoint "/api/v1/analyze" "POST" '{"request":"Generate an image of a sunset"}' "Analyze an image generation request"
    echo ""
    
    # Test 7: Get provider capabilities
    echo -e "${BLUE}Test 7: Provider Capabilities${NC}"
    test_endpoint "/api/v1/providers/capabilities" "GET" "" "Get provider capabilities"
    echo ""
}

# Function to test specific provider endpoints
test_provider_endpoints() {
    echo -e "${BLUE}🔗 Testing Provider Endpoints${NC}"
    echo "============================="
    
    # Check if providers.csv exists
    if [ ! -f "$PROVIDERS_CSV" ]; then
        echo -e "${RED}❌ Providers CSV not found: $PROVIDERS_CSV${NC}"
        return 1
    fi
    
    echo -e "${BLUE}📋 Providers in CSV:${NC}"
    head -5 "$PROVIDERS_CSV"
    echo ""
    
    # Test specific endpoints that should have dynamic models
    echo -e "${BLUE}🌐 Testing Known Dynamic Endpoints:${NC}"
    
    # Test Pollinations (if accessible)
    echo -e "${YELLOW}Testing Pollinations models endpoint...${NC}"
    if curl -s "https://text.pollinations.ai/models" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Pollinations endpoint accessible${NC}"
        curl -s "https://text.pollinations.ai/models" | head -5
    else
        echo -e "${YELLOW}⚠️  Pollinations endpoint not accessible (may be rate limited)${NC}"
    fi
    echo ""
    
    # Test Local Ollama (if running)
    echo -e "${YELLOW}Testing Local Ollama endpoint...${NC}"
    if curl -s "http://localhost:11434/api/tags" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Local Ollama accessible${NC}"
        curl -s "http://localhost:11434/api/tags" | head -5
    else
        echo -e "${YELLOW}⚠️  Local Ollama not running or not accessible${NC}"
    fi
    echo ""
}

# Function to run comprehensive tests
run_comprehensive_tests() {
    echo -e "${BLUE}🧪 Running Comprehensive Tests${NC}"
    echo "=============================="
    
    local passed=0
    local total=0
    
    # Test different request types
    test_requests=(
        '{"request":"Hello, how are you?"}:Text Generation'
        '{"request":"Write a Python function to sort a list"}:Code Generation'
        '{"request":"Generate an image of a mountain landscape"}:Image Generation'
        '{"request":"Analyze this complex machine learning problem"}:Complex Analysis'
        '{"request":"Create a simple web page with HTML and CSS"}:Web Development'
    )
    
    for test_case in "${test_requests[@]}"; do
        IFS=':' read -r request_data description <<< "$test_case"
        total=$((total + 1))
        
        echo -e "${BLUE}Test Case $total: $description${NC}"
        if test_endpoint "/api/v1/process" "POST" "$request_data" "$description"; then
            passed=$((passed + 1))
        fi
        echo "---"
    done
    
    echo -e "${BLUE}📊 Test Results Summary${NC}"
    echo "======================"
    echo -e "Passed: ${GREEN}$passed${NC}/$total"
    echo -e "Success Rate: ${GREEN}$(( passed * 100 / total ))%${NC}"
    
    if [ $passed -eq $total ]; then
        echo -e "${GREEN}🎉 All tests passed!${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠️  Some tests failed${NC}"
        return 1
    fi
}

# Function to cleanup
cleanup() {
    if [ -n "$SERVER_PID" ]; then
        echo -e "${YELLOW}🛑 Stopping server (PID: $SERVER_PID)...${NC}"
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
        echo -e "${GREEN}✅ Server stopped${NC}"
    fi
}

# Main execution
main() {
    echo -e "${BLUE}🚀 Starting Dynamic Model Discovery Tests${NC}"
    echo ""
    
    # Trap to cleanup on exit
    trap cleanup EXIT
    
    # Start server if needed
    start_server
    
    # Run tests
    test_dynamic_models
    test_provider_endpoints
    run_comprehensive_tests
    
    echo ""
    echo -e "${GREEN}🎉 Dynamic Model Discovery Testing Complete!${NC}"
    echo ""
    echo -e "${BLUE}📋 Next Steps:${NC}"
    echo "1. Check server.log for detailed server output"
    echo "2. Verify that dynamic models are being fetched from URLs"
    echo "3. Test with your own provider endpoints"
    echo "4. Monitor model refresh functionality"
    echo ""
    echo -e "${YELLOW}💡 Tips:${NC}"
    echo "- Use 'curl -X POST $SERVER_URL/api/v1/models/refresh' to refresh models"
    echo "- Check '$SERVER_URL/api/v1/models/stats' for dynamic loading statistics"
    echo "- View '$SERVER_URL/api/v1/system/info' for comprehensive system information"
}

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}⚠️  jq is not installed. Installing for better JSON parsing...${NC}"
    if command -v apt-get &> /dev/null; then
        sudo apt-get update && sudo apt-get install -y jq
    elif command -v yum &> /dev/null; then
        sudo yum install -y jq
    elif command -v brew &> /dev/null; then
        brew install jq
    else
        echo -e "${RED}❌ Could not install jq. JSON responses may not be formatted.${NC}"
    fi
fi

# Run main function
main "$@"