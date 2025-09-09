# Smart Flo Call Transcription Lambda

Node.js AWS Lambda function for call transcription and analysis using Google Gemini AI.

## 🔧 Required Environment Variables

Set these in your AWS Lambda console:

### Option 1: Connection String (Recommended)
```
DB_CONNECTION_STRING=postgres://username:password@host:port/database
GEMINI_API_KEY=your_gemini_api_key
```

### Option 2: Individual Variables (Fallback)
```
DB_HOST=your_database_host
DB_PORT=5432
DB_NAME=your_database_name
DB_USER=your_database_user
DB_PASSWORD=your_database_password
GEMINI_API_KEY=your_gemini_api_key
```

## 🚀 Deployment

1. Upload `deployment.zip` to AWS Lambda
2. Set handler to `index.handler`
3. Configure environment variables above
4. Set timeout to 15 minutes and memory to 512MB

## 📡 API Usage

**Note: This API processes calls asynchronously. It returns immediately with a 202 Accepted status while processing continues in the background.**

### Direct Lambda Invocation
```json
{
  "call_logsId": "uuid-of-call-record"
}
```

### API Gateway (GET) - API Information
```bash
curl -X GET "https://your-api-gateway-url"
```

### API Gateway (POST) - Start Processing
```bash
curl -X POST "https://your-api-gateway-url" \
  -H "Content-Type: application/json" \
  -d '{"call_logsId": "uuid-of-call-record"}'
```

**Response (202 Accepted):**
```json
{
  "statusCode": 202,
  "message": "Call processing started",
  "call_logsId": "uuid-of-call-record",
  "timestamp": "2025-01-01T00:00:00.000Z"
}
```

## 🔒 Security

- No credentials are hardcoded in the source code
- All sensitive data must be provided via environment variables
- Function will fail with clear error messages if required variables are missing
