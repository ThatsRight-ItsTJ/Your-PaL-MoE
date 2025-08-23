#!/bin/bash

echo "🚀 Testing Pollinations API Connectivity"
echo "========================================"

# Test 1: Simple connectivity test
echo
echo "=== Test 1: Basic Connectivity ==="
echo "Testing: https://text.pollinations.ai/Hello"

response=$(curl -s -w "HTTPSTATUS:%{http_code}" "https://text.pollinations.ai/Hello")
http_code=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
content=$(echo $response | sed -e 's/HTTPSTATUS\:.*//g')

if [ "$http_code" -eq 200 ]; then
    echo "✅ Pollinations API is reachable (Status: $http_code)"
    echo "Response: $content"
else
    echo "❌ API request failed (Status: $http_code)"
fi

# Test 2: Simple generation test
echo
echo "=== Test 2: Simple Text Generation ==="
prompt="Generate%20a%20simple%20greeting"
echo "Testing prompt: Generate a simple greeting"
echo "URL: https://text.pollinations.ai/$prompt"

response=$(curl -s -w "HTTPSTATUS:%{http_code}" "https://text.pollinations.ai/$prompt")
http_code=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
content=$(echo $response | sed -e 's/HTTPSTATUS\:.*//g')

if [ "$http_code" -eq 200 ]; then
    echo "✅ Simple generation successful (Status: $http_code)"
    echo "Response: $content"
else
    echo "❌ Generation failed (Status: $http_code)"
    echo "Error: $content"
fi

# Test 3: YAML generation test
echo
echo "=== Test 3: YAML Configuration Generation ==="
yaml_prompt="Generate%20a%20YAML%20configuration%20for%20an%20AI%20provider%3A%0AProvider%3A%20test_provider%0AEndpoint%3A%20https%3A//example.com/api%0AModel%3A%20test-model%0ACost%3A%200.001%0AFormat%20as%20valid%20YAML"

echo "Testing YAML generation..."
echo "URL: https://text.pollinations.ai/$yaml_prompt"

response=$(curl -s -w "HTTPSTATUS:%{http_code}" "https://text.pollinations.ai/$yaml_prompt")
http_code=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
content=$(echo $response | sed -e 's/HTTPSTATUS\:.*//g')

if [ "$http_code" -eq 200 ]; then
    echo "✅ YAML generation successful (Status: $http_code)"
    echo "Generated YAML:"
    echo "================"
    echo "$content"
    echo "================"
    
    # Save to file
    echo "$content" > test_generated.yaml
    echo "✅ YAML saved to test_generated.yaml"
else
    echo "❌ YAML generation failed (Status: $http_code)"
    echo "Error: $content"
fi

echo
echo "🏁 Pollinations API test completed!"
echo
echo "Summary:"
echo "- Basic connectivity: $([ "$http_code" -eq 200 ] && echo "✅ PASS" || echo "❌ FAIL")"
echo "- Text generation: Test completed"
echo "- YAML generation: Test completed"
echo
echo "Next steps:"
echo "1. Check the generated YAML file: test_generated.yaml"
echo "2. Test the enhanced server with manual reload endpoint"
echo "3. Verify CSV hot reload functionality"