# AWS Lambda Deployment Guide

## Prerequisites

1. AWS CLI configured with appropriate permissions
2. Go 1.21+ installed
3. AWS Lambda function created in the AWS Console

## Deployment Steps

### 1. Build the Lambda Function

```bash
./build.sh
```

This will create a `deployment.zip` file ready for upload.

### 2. Upload to AWS Lambda

#### Option A: Using AWS CLI
```bash
aws lambda update-function-code \
    --function-name your-function-name \
    --zip-file fileb://deployment.zip
```

#### Option B: Using AWS Console
1. Go to AWS Lambda Console
2. Select your function
3. Go to "Code" tab
4. Click "Upload from" → ".zip file"
5. Upload `deployment.zip`

### 3. Configure Environment Variables

In the AWS Lambda Console:
1. Go to "Configuration" → "Environment variables"
2. Add the following variables:
   - `DB_CONNECTION_STRING`: `postgres://postgres:Badho_1301@db.badho.in:5432/badho-app`
   - `GEMINI_API_KEY`: `AIzaSyATn1vcksF5BIJiBSn31CGfdslfysGtpOc`

### 4. Configure Runtime Settings

- **Runtime**: Go 1.x
- **Handler**: main
- **Timeout**: 15 minutes (900 seconds)
- **Memory**: 512 MB (minimum recommended)

### 5. Test the Function

#### Test Event Format:
```json
{
  "call_id": "c86d4b0d-5c9b-4edf-8b07-08a4833dcf50"
}
```

#### Expected Response:
```json
{
  "statusCode": 200,
  "body": {
    "call_id": "c86d4b0d-5c9b-4edf-8b07-08a4833dcf50",
    "call_data": { ... },
    "transcription": "transcribed text...",
    "questions_and_answers": [ ... ],
    "processed_at": "2024-01-01T00:00:00Z"
  }
}
```

## Local Testing

To test locally before deployment:

```bash
go run main.go test
```

## Monitoring

- Check CloudWatch logs for execution details
- Monitor function duration and memory usage
- Set up CloudWatch alarms for errors

## Troubleshooting

1. **Database Connection Issues**: Verify VPC configuration and security groups
2. **Timeout Errors**: Increase function timeout
3. **Memory Issues**: Increase memory allocation
4. **API Key Issues**: Verify environment variables are set correctly

