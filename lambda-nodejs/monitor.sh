#!/bin/bash

# Monitoring script for Smart Flo Call Transcription Lambda
# This script helps you monitor the function execution and check results

echo "🔍 Smart Flo Call Transcription Monitor"
echo "======================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Function to check CloudWatch logs
check_logs() {
    local call_logs_id=$1
    local minutes=${2:-5}
    
    print_info "Checking CloudWatch logs for call_logsId: $call_logs_id"
    print_info "Looking for logs from the last $minutes minutes..."
    echo ""
    
    # AWS CLI command to get logs (requires AWS CLI configured)
    if command -v aws &> /dev/null; then
        local log_group="/aws/lambda/your-lambda-function-name"
        local start_time=$(date -d "$minutes minutes ago" -u +%s)000
        
        echo "📋 Recent log entries:"
        aws logs filter-log-events \
            --log-group-name "$log_group" \
            --start-time $start_time \
            --filter-pattern "$call_logs_id" \
            --query 'events[*].[timestamp,message]' \
            --output table 2>/dev/null || print_warning "Could not fetch logs. Make sure AWS CLI is configured and log group exists."
    else
        print_warning "AWS CLI not found. Install AWS CLI to check CloudWatch logs."
        print_info "Manual steps:"
        print_info "1. Go to AWS CloudWatch Console"
        print_info "2. Navigate to Log Groups"
        print_info "3. Find your Lambda function log group"
        print_info "4. Search for: $call_logs_id"
    fi
}

# Function to check database results
check_database() {
    local call_logs_id=$1
    
    print_info "Checking database for results..."
    print_info "Query: SELECT callAnalysis FROM smartFlo.call_logs WHERE id = '$call_logs_id'"
    echo ""
    
    print_warning "To check database results manually:"
    print_info "1. Connect to your PostgreSQL database"
    print_info "2. Run: SELECT callAnalysis FROM \"smartFlo\".call_logs WHERE id = '$call_logs_id';"
    print_info "3. Look for a JSON object with transcription and answers"
}

# Function to test the API
test_api() {
    local api_url=$1
    local call_logs_id=$2
    
    print_info "Testing API endpoint..."
    echo ""
    
    # Test GET request
    print_info "1. Testing GET request (API info):"
    curl -s -X GET "$api_url" | jq '.' 2>/dev/null || echo "Response received"
    echo ""
    
    # Test POST request
    print_info "2. Testing POST request (start processing):"
    local response=$(curl -s -X POST "$api_url" \
        -H "Content-Type: application/json" \
        -d "{\"call_logsId\": \"$call_logs_id\"}")
    
    echo "$response" | jq '.' 2>/dev/null || echo "$response"
    echo ""
    
    # Check response status
    local status_code=$(echo "$response" | jq -r '.statusCode // "unknown"' 2>/dev/null)
    if [ "$status_code" = "202" ]; then
        print_success "API call successful - Processing started"
    else
        print_error "API call failed - Status: $status_code"
    fi
}

# Function to show monitoring checklist
show_checklist() {
    print_info "Monitoring Checklist:"
    echo ""
    echo "1. 📊 Check API Response (should be 202 Accepted)"
    echo "2. 📋 Check CloudWatch Logs for processing steps"
    echo "3. 🗄️  Check Database for saved results"
    echo "4. ⏱️  Monitor processing time (typically 30-120 seconds)"
    echo "5. 🔍 Look for error messages in logs"
    echo ""
    print_info "Expected log sequence:"
    echo "   🚀 Starting async processing for call_logsId: xxx"
    echo "   🔌 Connecting to database..."
    echo "   ✅ Database connected successfully"
    echo "   📋 Fetching call data..."
    echo "   ✅ Call data retrieved - Campaign: xxx, Recording URL: Present"
    echo "   ❓ Fetching questions for campaign: xxx"
    echo "   ✅ Questions fetched: x, Audio downloaded: xxxx bytes"
    echo "   🤖 Processing with Gemini AI..."
    echo "   ✅ Gemini processing completed - Transcription: xxx chars, Answers: x"
    echo "   💾 Saving results to database..."
    echo "   ✅ Results saved successfully"
    echo "   🎉 Processing completed in xxxxms"
}

# Main execution
if [ $# -eq 0 ]; then
    echo "Usage: $0 <command> [options]"
    echo ""
    echo "Commands:"
    echo "  test <api_url> <call_logs_id>     - Test API endpoint"
    echo "  logs <call_logs_id> [minutes]     - Check CloudWatch logs"
    echo "  db <call_logs_id>                 - Check database results"
    echo "  checklist                         - Show monitoring checklist"
    echo ""
    echo "Examples:"
    echo "  $0 test https://api.example.com/endpoint abc123-def456"
    echo "  $0 logs abc123-def456 10"
    echo "  $0 db abc123-def456"
    echo "  $0 checklist"
    exit 1
fi

case $1 in
    "test")
        if [ $# -lt 3 ]; then
            print_error "Usage: $0 test <api_url> <call_logs_id>"
            exit 1
        fi
        test_api "$2" "$3"
        ;;
    "logs")
        if [ $# -lt 2 ]; then
            print_error "Usage: $0 logs <call_logs_id> [minutes]"
            exit 1
        fi
        check_logs "$2" "${3:-5}"
        ;;
    "db")
        if [ $# -lt 2 ]; then
            print_error "Usage: $0 db <call_logs_id>"
            exit 1
        fi
        check_database "$2"
        ;;
    "checklist")
        show_checklist
        ;;
    *)
        print_error "Unknown command: $1"
        echo "Run '$0' without arguments to see usage."
        exit 1
        ;;
esac
