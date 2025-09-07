# Google Gemini Chat Script

A simple Python script to interact with Google's Gemini AI model using the API.

## Features

- Interactive chat interface
- Conversation history tracking
- Automatic conversation saving
- Error handling and connection testing
- Clean, user-friendly interface

## Setup

1. Install the required dependencies:
   ```bash
   pip install -r requirements.txt
   ```

2. Run the script:
   ```bash
   python gemini_chat.py
   ```

## Usage

- Type your questions or messages to chat with Gemini
- Type `quit`, `exit`, or `bye` to end the conversation
- Type `clear` to clear the conversation history
- Press `Ctrl+C` to exit at any time
- Conversations are automatically saved as JSON files

## API Key

The script is pre-configured with your Google Gemini API key. The API key is embedded in the script for convenience, but in production environments, consider using environment variables for better security.

## Example

```
🤖 Google Gemini Chat
==================================================
Type 'quit', 'exit', or 'bye' to end the conversation
Type 'clear' to clear the conversation history
==================================================

👤 You: What is artificial intelligence?
🤖 Gemini: Artificial intelligence (AI) is a broad field of computer science...

👤 You: quit
👋 Goodbye!
💾 Conversation saved to: conversation_20241201_143022.json
```

