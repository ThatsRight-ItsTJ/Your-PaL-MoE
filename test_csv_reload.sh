#!/bin/bash

echo "🚀 Testing CSV Hot Reload and Manual Reload"
echo "==========================================="

# Create initial test CSV
echo
echo "=== Creating Initial CSV ==="
cat > test_providers.csv << 'EOF'
ID,Name,Tier,Endpoint,APIKey,Model,CostPerToken,MaxTokens,Capabilities,AdditionalInfo
pollinations,Pollinations,community,https://text.pollinations.ai,,openai,0.000001,2048,chat;creative,no_auth:true,free_tier:true
test_provider,Test Provider,community,https://example.com/api,,test-model,0.001,2048,chat,rate_limit:100/min
EOF

echo "✅ Initial CSV created with 2 providers:"
echo "  - pollinations (Pollinations)"
echo "  - test_provider (Test Provider)"

# Display CSV content
echo
echo "=== Initial CSV Content ==="
cat test_providers.csv
echo

# Test CSV modification (simulating hot reload)
echo "=== Testing CSV Modification ==="
echo "Adding a new provider to CSV..."

# Add new provider to CSV
cat >> test_providers.csv << 'EOF'
local_llama,Local Llama,unofficial,http://localhost:8080,none,llama-2-7b,0,4096,chat,local:true,gpu_required:false
EOF

echo "✅ Added new provider: local_llama (Local Llama)"

# Display updated CSV content
echo
echo "=== Updated CSV Content ==="
cat test_providers.csv
echo

# Show file modification details
echo "=== File Modification Details ==="
ls -la test_providers.csv
echo "File modification time: $(stat -c %y test_providers.csv 2>/dev/null || stat -f "%Sm" test_providers.csv)"

# Create example providers.csv for the server
echo
echo "=== Creating Production CSV ==="
cp test_providers.csv providers.csv
echo "✅ Copied test CSV to providers.csv for server use"

# Test manual reload using curl (when server is running)
echo
echo "=== Manual Reload Test Instructions ==="
echo "To test manual reload when server is running:"
echo "1. Start the enhanced server: go run cmd/enhanced-server/main.go"
echo "2. Test manual reload endpoint:"
echo "   curl -X POST http://localhost:8080/api/v1/providers/reload"
echo "3. Check providers list:"
echo "   curl http://localhost:8080/api/v1/providers"

# Create a comprehensive test script for when server is available
cat > test_server_endpoints.sh << 'EOF'
#!/bin/bash

echo "Testing Enhanced Server Endpoints"
echo "================================="

BASE_URL="http://localhost:8080"

# Test health endpoint
echo
echo "=== Testing Health Endpoint ==="
curl -s "$BASE_URL/health" && echo

# Test providers list
echo
echo "=== Testing Providers List ==="
curl -s "$BASE_URL/api/v1/providers" | head -20 && echo

# Test manual reload
echo
echo "=== Testing Manual Reload ==="
response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST "$BASE_URL/api/v1/providers/reload")
http_code=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
content=$(echo $response | sed -e 's/HTTPSTATUS\:.*//g')

if [ "$http_code" -eq 200 ]; then
    echo "✅ Manual reload successful (Status: $http_code)"
    echo "Response: $content"
else
    echo "❌ Manual reload failed (Status: $http_code)"
fi

# Test YAML generation
echo
echo "=== Testing YAML Generation ==="
response=$(curl -s -w "HTTPSTATUS:%{http_code}" "$BASE_URL/api/v1/providers/pollinations/yaml")
http_code=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
content=$(echo $response | sed -e 's/HTTPSTATUS\:.*//g')

if [ "$http_code" -eq 200 ]; then
    echo "✅ YAML generation successful (Status: $http_code)"
    echo "Generated YAML:"
    echo "================"
    echo "$content"
    echo "================"
else
    echo "❌ YAML generation failed (Status: $http_code)"
    echo "Error: $content"
fi

# Test batch YAML generation
echo
echo "=== Testing Batch YAML Generation ==="
response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST "$BASE_URL/api/v1/providers/yaml/generate-all")
http_code=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
content=$(echo $response | sed -e 's/HTTPSTATUS\:.*//g')

if [ "$http_code" -eq 200 ]; then
    echo "✅ Batch YAML generation successful (Status: $http_code)"
    echo "Response summary: $content" | head -10
else
    echo "❌ Batch YAML generation failed (Status: $http_code)"
fi

echo
echo "🏁 Server endpoint testing completed!"
EOF

chmod +x test_server_endpoints.sh
echo "✅ Created test_server_endpoints.sh for server testing"

echo
echo "🏁 CSV Hot Reload test setup completed!"
echo
echo "Files created:"
echo "- test_providers.csv (test data)"
echo "- providers.csv (server data)" 
echo "- test_server_endpoints.sh (server tests)"
echo
echo "Manual testing steps:"
echo "1. The CSV files demonstrate hot reload capability"
echo "2. File watcher will detect changes to providers.csv when server runs"
echo "3. Manual reload endpoint: POST /api/v1/providers/reload"
echo "4. Run test_server_endpoints.sh when server is running"