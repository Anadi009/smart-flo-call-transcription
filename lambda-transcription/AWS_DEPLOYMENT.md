# AWS Lambda Deployment Guide

## 🚀 Quick Deployment

### 1. Upload the Lambda Function
- Upload `deployment.zip` to your AWS Lambda function
- Set runtime to **Go 1.x**
- Set handler to **main**

### 2. Configure Environment Variables
In AWS Lambda Console → Configuration → Environment variables, add:

```
DB_CONNECTION_STRING = postgres://postgres:Badho_1301@db.badho.in:5432/badho-app
GEMINI_API_KEY = AIzaSyATn1vcksF5BIJiBSn31CGfdslfysGtpOc
```

### 3. Configure Runtime Settings
- **Memory**: 512 MB (minimum)
- **Timeout**: 15 minutes (900 seconds)
- **Architecture**: x86_64

### 4. Test the Function
Use this test event:
```json
{
  "call_id": "c86d4b0d-5c9b-4edf-8b07-08a4833dcf50"
}
```

Expected response:
```json
{
  "statusCode": 200,
  "body": {
    "call_id": "c86d4b0d-5c9b-4edf-8b07-08a4833dcf50",
    "transcription": "transcribed text...",
    "answers": {
      "question-id-1": "answer 1",
      "question-id-2": "answer 2"
    },
    "processed_at": "2024-01-01T00:00:00Z"
  }
}
```

## 📦 Package Contents
- `main` - Compiled Go binary
- `deployment.zip` - Ready-to-upload package (6.4MB)

## 🔧 Features
- ✅ Single Gemini API call (transcription + Q&A)
- ✅ Optimized response (minimal output tokens)
- ✅ PostgreSQL database integration
- ✅ Error handling and logging
- ✅ Cost-effective execution

## 📊 Performance
- **Execution time**: ~30-60 seconds
- **Memory usage**: ~200-400 MB
- **API calls**: 1 (instead of 2)
- **Response size**: ~70% smaller than original

