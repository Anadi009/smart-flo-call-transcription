#!/usr/bin/env python3
"""
Google Gemini API Chat Script
A simple script to interact with Google's Gemini AI model using the API.
"""

import os
import sys
import json
from typing import Optional
import requests
from datetime import datetime

class GeminiChat:
    def __init__(self, api_key: str):
        """
        Initialize the Gemini chat client.
        
        Args:
            api_key (str): Your Google Gemini API key
        """
        self.api_key = api_key
        self.base_url = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"
        self.session = requests.Session()
        
    def send_message(self, message: str, model: str = "gemini-2.0-flash") -> Optional[str]:
        """
        Send a message to the Gemini API and return the response.
        
        Args:
            message (str): The message to send to the AI
            model (str): The model to use (default: gemini-pro)
            
        Returns:
            Optional[str]: The AI's response or None if error occurred
        """
        url = f"https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent"
        
        headers = {
            "Content-Type": "application/json",
        }
        
        data = {
            "contents": [
                {
                    "parts": [
                        {
                            "text": message
                        }
                    ]
                }
            ]
        }
        
        params = {
            "key": self.api_key
        }
        
        try:
            response = self.session.post(url, headers=headers, json=data, params=params)
            response.raise_for_status()
            
            result = response.json()
            
            if "candidates" in result and len(result["candidates"]) > 0:
                return result["candidates"][0]["content"]["parts"][0]["text"]
            else:
                print("Error: No response generated")
                return None
                
        except requests.exceptions.RequestException as e:
            print(f"Error making request: {e}")
            return None
        except KeyError as e:
            print(f"Error parsing response: {e}")
            print(f"Response: {response.text}")
            return None
        except Exception as e:
            print(f"Unexpected error: {e}")
            return None
    
    def chat_loop(self):
        """
        Start an interactive chat loop with the user.
        """
        print("🤖 Google Gemini Chat")
        print("=" * 50)
        print("Type 'quit', 'exit', or 'bye' to end the conversation")
        print("Type 'clear' to clear the conversation history")
        print("=" * 50)
        
        conversation_history = []
        
        while True:
            try:
                user_input = input("\n👤 You: ").strip()
                
                if user_input.lower() in ['quit', 'exit', 'bye']:
                    print("\n👋 Goodbye!")
                    break
                    
                if user_input.lower() == 'clear':
                    conversation_history = []
                    print("🧹 Conversation history cleared!")
                    continue
                    
                if not user_input:
                    continue
                
                # Add timestamp and user message to history
                conversation_history.append({
                    "timestamp": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
                    "user": user_input
                })
                
                print("\n🤖 Gemini: ", end="", flush=True)
                
                # Send message to Gemini
                response = self.send_message(user_input)
                
                if response:
                    print(response)
                    # Add AI response to history
                    conversation_history.append({
                        "timestamp": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
                        "ai": response
                    })
                else:
                    print("Sorry, I couldn't process your request. Please try again.")
                    
            except KeyboardInterrupt:
                print("\n\n👋 Goodbye!")
                break
            except Exception as e:
                print(f"\n❌ An error occurred: {e}")
        
        # Save conversation history
        self.save_conversation(conversation_history)
    
    def save_conversation(self, history: list):
        """
        Save conversation history to a JSON file.
        
        Args:
            history (list): List of conversation messages
        """
        if not history:
            return
            
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        filename = f"conversation_{timestamp}.json"
        
        try:
            with open(filename, 'w', encoding='utf-8') as f:
                json.dump(history, f, indent=2, ensure_ascii=False)
            print(f"\n💾 Conversation saved to: {filename}")
        except Exception as e:
            print(f"\n❌ Error saving conversation: {e}")

def main():
    """
    Main function to run the Gemini chat script.
    """
    # Your API key
    from dotenv import load_dotenv
    load_dotenv()
    API_KEY = os.environ.get("GEMINI_API_KEY")
    # Check if API key is provided
    if not API_KEY:
        print("❌ Error: API key not found!")
        print("Please set your Google Gemini API key in the script.")
        sys.exit(1)
    
    # Create and start the chat client
    try:
        chat_client = GeminiChat(API_KEY)
        
        # Test the connection first
        print("🔍 Testing connection to Gemini API...")
        test_response = chat_client.send_message("Hello, can you hear me?")
        
        if test_response:
            print("✅ Connection successful!")
            chat_client.chat_loop()
        else:
            print("❌ Failed to connect to Gemini API. Please check your API key and internet connection.")
            
    except Exception as e:
        print(f"❌ Error initializing chat client: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()

