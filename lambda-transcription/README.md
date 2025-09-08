# SmartFlo Call Transcription Lambda Function

This is a Go-based AWS Lambda function that replicates the Python transcription pipeline functionality.

## Features

- Fetches call data from PostgreSQL database
- Downloads audio recordings from URLs
- **Single Gemini API call** for both transcription and question answering
- Optimized response with minimal output tokens (only question ID -> answer mapping)
- Saves results back to the database
- Returns simplified JSON response

## Dependencies

- Go 1.21+
- AWS Lambda Go runtime
- PostgreSQL driver
- Environment variable support

## Configuration

Set the following environment variables:

- `DB_CONNECTION_STRING`: PostgreSQL connection string
- `GEMINI_API_KEY`: Google Gemini API key

## Usage

### Local Testing

1. Install dependencies:
```bash
go mod tidy
```

2. Run locally:
```bash
go run main.go
```

### AWS Lambda Deployment

1. Build the binary:
```bash
GOOS=linux GOARCH=amd64 go build -o main main.go
```

2. Create a deployment package:
```bash
zip deployment.zip main
```

3. Upload to AWS Lambda

## Input Format

```json
{
  "call_id": "your-call-id-here"
}
```

## Output Format

```json
{
  "statusCode": 200,
  "body": {
    "call_id": "call-id",
    "transcription": "transcribed text...",
    "answers": {
      "question-id-1": "answer 1",
      "question-id-2": "answer 2"
    },
    "processed_at": "2024-01-01T00:00:00Z"
  }
}
```

## Optimizations

- **Single API Call**: Combines transcription and question answering in one Gemini request
- **Reduced Output**: Only returns question ID -> answer mapping instead of full question objects
- **Lower Costs**: Significantly reduced output tokens for cheaper API usage
- **Faster Execution**: Single network call instead of two separate requests

## Error Handling

The function returns appropriate HTTP status codes:
- 200: Success
- 500: Internal server error

Error details are included in the response body.
