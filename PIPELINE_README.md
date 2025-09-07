# SmartFlo Call Transcription Pipeline

A complete pipeline that transcribes call recordings using Google Gemini AI and answers predefined questions from the database.

## 🚀 Features

- **Audio Transcription**: Downloads call recordings and transcribes them using Google Gemini AI
- **Question Answering**: Automatically answers questions from the `smartFlo.question` table
- **Database Integration**: Connects to PostgreSQL database to fetch call data and save results
- **Multi-language Support**: Handles Hindi and English call transcriptions
- **Result Storage**: Saves transcriptions to database and detailed results to JSON files

## 📋 Prerequisites

- Python 3.7+
- PostgreSQL database access
- Google Gemini API key
- Required Python packages (see requirements.txt)

## 🛠️ Installation

1. **Install dependencies:**
   ```bash
   pip3 install -r requirements.txt
   ```

2. **Configure API keys:**
   - The script is pre-configured with your Google Gemini API key
   - Database connection string is already set

## 📖 Usage

### Basic Usage

```bash
python3 transcription_pipeline.py <call_id>
```

### Example

```bash
python3 transcription_pipeline.py c86d4b0d-5c9b-4edf-8b07-08a4833dcf50
```

### Using the Helper Script

```bash
python3 run_pipeline.py c86d4b0d-5c9b-4edf-8b07-08a4833dcf50
```

## 🔄 Pipeline Process

1. **Database Connection**: Connects to PostgreSQL database
2. **Call Data Retrieval**: Fetches call information from `smartFlo.call_logs`
3. **Question Loading**: Loads active questions from `smartFlo.question` table
4. **Audio Download**: Downloads the call recording from the provided URL
5. **Transcription**: Sends audio to Google Gemini for transcription
6. **Question Answering**: Uses Gemini to answer questions based on transcription
7. **Result Storage**: Saves results to database and JSON file

## 📊 Output

### Console Output
- Real-time progress updates
- Transcription text display
- Questions and answers
- Success/error messages

### Database Updates
- Updates `transcriptionJSON` field in `smartFlo.call_logs`
- Increments `transcribeAttempt` counter

### JSON File
Creates a detailed JSON file with:
- Call metadata
- Full transcription
- Questions and AI-generated answers
- Processing timestamps

## 🗃️ Database Schema

### smartFlo.call_logs
- `id`: Call log ID (primary key)
- `recording_url`: URL to download audio recording
- `transcriptionJSON`: JSON field storing transcription data
- `transcribeAttempt`: Counter for transcription attempts

### smartFlo.question
- `id`: Question ID
- `label`: Question label/description
- `isActive`: Boolean flag for active questions
- `details`: JSONB field containing question text and instructions

## 🔧 Configuration

### API Keys
- **Google Gemini API**: `AIzaSyATn1vcksF5BIJiBSn31CGfdslfysGtpOc`
- **Database**: `postgres://postgres:Badho_1301@db.badho.in:5432/badho-app`

### Model Settings
- **Gemini Model**: `gemini-1.5-flash`
- **Audio Format**: MP3 (auto-detected)
- **Transcription Language**: Auto-detected (supports Hindi, English, etc.)

## 📁 File Structure

```
SmartFlo/
├── transcription_pipeline.py    # Main pipeline script
├── run_pipeline.py             # Helper script for easy execution
├── db_connection.py            # Database connection utilities
├── gemini_chat.py              # Gemini chat interface
├── requirements.txt            # Python dependencies
├── PIPELINE_README.md          # This documentation
└── transcription_results_*.json # Generated result files
```

## 🚨 Error Handling

- **Database Connection Errors**: Automatic retry and rollback
- **Audio Download Failures**: Timeout and error reporting
- **API Errors**: Detailed error messages and fallback handling
- **Transaction Issues**: Automatic rollback on failures

## 📈 Performance

- **Audio Processing**: Handles various audio formats
- **Token Optimization**: Uses single API call for question answering
- **Memory Efficient**: Streams audio data without storing large files
- **Concurrent Safe**: Database transactions ensure data integrity

## 🔍 Example Output

```
🚀 Starting transcription pipeline for call ID: c86d4b0d-5c9b-4edf-8b07-08a4833dcf50
================================================================================
✅ Connected to database successfully
✅ Found call data for ID: c86d4b0d-5c9b-4edf-8b07-08a4833dcf50
   Recording URL: https://cloudphone.tatateleservices.com/file/recording?...
   Caller: +919240211859
   Agent: Divya Nigam
✅ Found 3 active questions
🎵 Downloading audio from: https://cloudphone.tatateleservices.com/file/recording?...
✅ Audio downloaded successfully (88128 bytes)
🤖 Sending audio to Gemini for transcription...
✅ Transcription completed successfully

📝 TRANSCRIPTION:
--------------------------------------------------------------------------------
हेलो। हेलो। जी, नमस्कार सर। मेरी बात श्लोक और जीतू से हो रही है...
--------------------------------------------------------------------------------

🤖 Sending questions to Gemini for answering...
✅ Questions answered successfully

❓ QUESTIONS AND ANSWERS:
--------------------------------------------------------------------------------
1. Was there a network issue?
   Answer: Information not available in the call.
   Label: Was there a network issue?

2. How long was the call?
   Answer: Information not available in the call.
   Label: How long was the call?

3. What feedback related to app experience was given by the user to our calling executive?
   Answer: No feedback related to app experience was given by the user.
   Label: What feedback related to app experience was given by the user to our calling executive?

✅ Transcription saved to database
✅ Results saved to: transcription_results_c86d4b0d-5c9b-4edf-8b07-08a4833dcf50_20250907_144849.json
🎉 Pipeline completed successfully!
```

## 🛡️ Security Notes

- API keys are embedded in the script for convenience
- In production, consider using environment variables
- Database credentials are in the connection string
- Audio files are processed in memory and not permanently stored

## 📞 Support

For issues or questions about the pipeline, check:
1. Database connectivity
2. API key validity
3. Audio URL accessibility
4. Network connectivity

The pipeline provides detailed error messages to help diagnose issues.
