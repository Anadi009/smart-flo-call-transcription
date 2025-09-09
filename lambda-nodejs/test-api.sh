#!/bin/bash

# Test script for API Gateway endpoints
# Replace YOUR_API_GATEWAY_URL with your actual API Gateway URL

API_URL="https://YOUR_API_GATEWAY_URL"
TEST_CALL_ID="ddf559f0-c076-471f-8824-9fde851bc70a"

echo "🧪 Testing Smart Flo Call Transcription API"
echo "=========================================="
echo ""

# Test 1: GET request for API info
echo "📋 Test 1: GET request for API information"
echo "curl -X GET '$API_URL'"
echo ""
curl -X GET "$API_URL" \
  -H "Accept: application/json" \
  -w "\nStatus: %{http_code}\n" \
  | jq '.' 2>/dev/null || echo "Response received"

echo ""
echo "----------------------------------------"
echo ""

# Test 2: POST request with call_logsId
echo "📋 Test 2: POST request to process call"
echo "curl -X POST '$API_URL' -d '{\"call_logsId\": \"$TEST_CALL_ID\"}'"
echo ""
curl -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d "{\"call_logsId\": \"$TEST_CALL_ID\"}" \
  -w "\nStatus: %{http_code}\n" \
  | jq '.' 2>/dev/null || echo "Response received"

echo ""
echo "----------------------------------------"
echo ""

# Test 3: OPTIONS request (CORS preflight)
echo "📋 Test 3: OPTIONS request (CORS preflight)"
echo "curl -X OPTIONS '$API_URL'"
echo ""
curl -X OPTIONS "$API_URL" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type" \
  -w "\nStatus: %{http_code}\n" \
  | jq '.' 2>/dev/null || echo "Response received"

echo ""
echo "----------------------------------------"
echo ""

# Test 4: Invalid POST request (missing call_logsId)
echo "📋 Test 4: Invalid POST request (should return error)"
echo "curl -X POST '$API_URL' -d '{}'"
echo ""
curl -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{}' \
  -w "\nStatus: %{http_code}\n" \
  | jq '.' 2>/dev/null || echo "Response received"

echo ""
echo "🎉 API testing completed!"
echo ""
echo "📝 Usage Instructions:"
echo "1. Replace YOUR_API_GATEWAY_URL with your actual API Gateway URL"
echo "2. Replace the test call ID with a valid call_logsId from your database"
echo "3. Run this script to test all API endpoints"
