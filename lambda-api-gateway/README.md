# Smart Flo Call Processing - API Gateway Lambda

This is an AWS Lambda function designed to work with API Gateway for processing call transcriptions and question answering.

## Features

- **API Gateway Integration**: Handles HTTP POST requests via API Gateway
- **Campaign-specific Questions**: Only processes questions linked to the call's campaign
- **Dynamic Instructions**: Uses instructions from database `details` column
- **Single Gemini API Call**: Efficient processing combining transcription and Q&A
- **Database Integration**: Updates `callAnalysis` column with results
- **Error Handling**: Comprehensive error handling with proper HTTP status codes
- **CORS Support**: Includes CORS headers for web applications

## API Endpoint

```
POST https://your-api-gateway-url/smartFloCallProcessing
Content-Type: application/json

{
  "call_logsId": "ddf559f0-c076-471f-8824-9fde851bc70a"
}
```

## Response Format

```json
{
  "call_logsId": "ddf559f0-c076-471f-8824-9fde851bc70a",
  "campaignId": "a3b8afd2-07fe-4c2d-b025-a88aa1ee5580",
  "transcription": "Full transcription of the call...",
  "answers": {
    "question-id-1": "answer 1",
    "question-id-2": "answer 2"
  },
  "processed_at": "2025-09-08T19:28:29Z"
}
```

## Build and Deploy

1. **Build the function:**
   ```bash
   chmod +x build.sh
   ./build.sh
   ```

2. **Deploy to AWS Lambda:**
   - Upload `deployment.zip` to your Lambda function
   - Set Handler to: `main`
   - Set Runtime to: `Go 1.x`
   - Set Architecture to: `x86_64`

3. **Configure Environment Variables:**
   - `DB_CONNECTION_STRING`: PostgreSQL connection string
   - `GEMINI_API_KEY`: Google Gemini API key

4. **Set up API Gateway:**
   - Create API Gateway HTTP API
   - Create route: `POST /smartFloCallProcessing`
   - Connect to your Lambda function
   - Enable CORS if needed

## Testing

### Using curl:
```bash
curl -X POST "https://your-api-gateway-url/smartFloCallProcessing" \
  -H "Content-Type: application/json" \
  -d '{"call_logsId": "ddf559f0-c076-471f-8824-9fde851bc70a"}'
```

### Using PostgreSQL (via HTTP):
```sql
-- Using http extension
SELECT content FROM http((
    'POST',
    'https://your-api-gateway-url/smartFloCallProcessing',
    ARRAY[http_header('Content-Type', 'application/json')],
    'application/json',
    '{"call_logsId": "ddf559f0-c076-471f-8824-9fde851bc70a"}'
));
```

## Error Responses

- `400 Bad Request`: Invalid JSON or missing call_logsId
- `405 Method Not Allowed`: Non-POST requests
- `500 Internal Server Error`: Processing errors

## Architecture

- **Runtime**: Go 1.x
- **Handler**: main
- **Timeout**: Recommended 5-10 minutes
- **Memory**: Recommended 1024 MB
- **Architecture**: x86_64
