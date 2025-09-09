#!/bin/bash

API_ENDPOINT="https://wc4yvde65f.execute-api.ap-south-1.amazonaws.com/default/smartFloCallProcessingAPI"

echo "🚀 Testing Smart Flo Call Processing API"
echo "Endpoint: $API_ENDPOINT"
echo "=================================="

echo ""
echo "📋 Test 1: GET request (API Documentation)"
echo "-------------------------------------------"
curl -X GET "$API_ENDPOINT" \
  -H "Content-Type: application/json" \
  -w "\nStatus Code: %{http_code}\nResponse Time: %{time_total}s\n" \
  | jq '.' 2>/dev/null || echo "Response received (not JSON formatted)"

echo ""
echo "=================================="
echo ""

echo "📋 Test 2: POST request (Call Processing)"
echo "-------------------------------------------"
echo "Testing with call_logsId: ddf559f0-c076-471f-8824-9fde851bc70a"

curl -X POST "$API_ENDPOINT" \
  -H "Content-Type: application/json" \
  -d '{"call_logsId": "ddf559f0-c076-471f-8824-9fde851bc70a"}' \
  -w "\nStatus Code: %{http_code}\nResponse Time: %{time_total}s\n" \
  | jq '.' 2>/dev/null || echo "Response received (not JSON formatted)"

echo ""
echo "=================================="
echo ""

echo "📋 Test 3: POST request with invalid data"
echo "-------------------------------------------"
echo "Testing with missing call_logsId"

curl -X POST "$API_ENDPOINT" \
  -H "Content-Type: application/json" \
  -d '{}' \
  -w "\nStatus Code: %{http_code}\nResponse Time: %{time_total}s\n" \
  | jq '.' 2>/dev/null || echo "Response received (not JSON formatted)"

echo ""
echo "✅ API Testing Complete"
