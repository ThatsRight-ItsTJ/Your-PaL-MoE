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
