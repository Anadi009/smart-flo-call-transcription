#!/usr/bin/env python3
"""
SmartFlo Call Transcription and Question Answering Pipeline
A pipeline that takes a call_logs ID, transcribes the audio using Gemini, and answers questions.
"""

import psycopg2
import requests
import json
import sys
import os
from urllib.parse import urlparse
from typing import Optional, Dict, Any, List
from datetime import datetime
import base64

class TranscriptionPipeline:
    def __init__(self, db_connection_string: str, gemini_api_key: str):
        """
        Initialize the transcription pipeline.
        
        Args:
            db_connection_string (str): PostgreSQL connection string
            gemini_api_key (str): Google Gemini API key
        """
        self.db_connection_string = db_connection_string
        self.gemini_api_key = gemini_api_key
        self.db_connection = None
        self.cursor = None
        
    def connect_to_database(self) -> bool:
        """Connect to the PostgreSQL database."""
        try:
            parsed_url = urlparse(self.db_connection_string)
            self.db_connection = psycopg2.connect(
                host=parsed_url.hostname,
                port=parsed_url.port,
                database=parsed_url.path[1:],
                user=parsed_url.username,
                password=parsed_url.password
            )
            self.cursor = self.db_connection.cursor()
            print("✅ Connected to database successfully")
            return True
        except Exception as e:
            print(f"❌ Database connection error: {e}")
            return False
    
    def disconnect_from_database(self):
        """Disconnect from the database."""
        if self.cursor:
            self.cursor.close()
        if self.db_connection:
            self.db_connection.close()
        print("🔌 Database connection closed")
    
    def get_call_data(self, call_id: str) -> Optional[Dict[str, Any]]:
        """
        Get call data from smartFlo.call_logs table.
        
        Args:
            call_id (str): The call log ID
            
        Returns:
            Optional[Dict[str, Any]]: Call data including recording URL
        """
        try:
            query = """
                SELECT id, recording_url, call_id, caller_id_number, call_to_number, 
                       start_date, start_time, duration, agent_name, campaign_name
                FROM "smartFlo".call_logs 
                WHERE id = %s;
            """
            self.cursor.execute(query, (call_id,))
            result = self.cursor.fetchone()
            
            if result:
                columns = [desc[0] for desc in self.cursor.description]
                call_data = dict(zip(columns, result))
                print(f"✅ Found call data for ID: {call_id}")
                print(f"   Recording URL: {call_data.get('recording_url', 'N/A')}")
                print(f"   Caller: {call_data.get('caller_id_number', 'N/A')}")
                print(f"   Agent: {call_data.get('agent_name', 'N/A')}")
                return call_data
            else:
                print(f"❌ No call found with ID: {call_id}")
                return None
                
        except Exception as e:
            print(f"❌ Error fetching call data: {e}")
            return None
    
    def get_questions(self) -> Optional[List[Dict[str, Any]]]:
        """
        Get questions from smartFlo.question table.
        
        Returns:
            Optional[List[Dict[str, Any]]]: List of questions
        """
        try:
            query = """
                SELECT id, label, "isActive", details
                FROM "smartFlo".question 
                WHERE "isActive" = true
                ORDER BY id;
            """
            self.cursor.execute(query)
            results = self.cursor.fetchall()
            
            if results:
                columns = [desc[0] for desc in self.cursor.description]
                questions = []
                for row in results:
                    question_data = dict(zip(columns, row))
                    # Extract question text from details JSONB
                    if question_data.get('details') and 'questionText' in question_data['details']:
                        question_data['question_text'] = question_data['details']['questionText']
                        question_data['answer_type'] = question_data['details'].get('answerType', 'text')
                        question_data['instructions'] = question_data['details'].get('instructions', '')
                    questions.append(question_data)
                print(f"✅ Found {len(questions)} active questions")
                return questions
            else:
                print("❌ No active questions found")
                return None
                
        except Exception as e:
            print(f"❌ Error fetching questions: {e}")
            return None
    
    def download_audio(self, recording_url: str) -> Optional[bytes]:
        """
        Download audio file from the recording URL.
        
        Args:
            recording_url (str): URL of the audio recording
            
        Returns:
            Optional[bytes]: Audio file content
        """
        try:
            print(f"🎵 Downloading audio from: {recording_url}")
            response = requests.get(recording_url, timeout=30)
            response.raise_for_status()
            
            print(f"✅ Audio downloaded successfully ({len(response.content)} bytes)")
            return response.content
            
        except Exception as e:
            print(f"❌ Error downloading audio: {e}")
            return None
    
    def transcribe_audio_with_gemini(self, audio_content: bytes) -> Optional[str]:
        """
        Transcribe audio using Google Gemini API.
        
        Args:
            audio_content (bytes): Audio file content
            
        Returns:
            Optional[str]: Transcription text
        """
        try:
            # Encode audio to base64
            audio_base64 = base64.b64encode(audio_content).decode('utf-8')
            
            # Prepare the request
            url = f"https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-pro:generateContent"
            
            headers = {
                "Content-Type": "application/json",
            }
            
            data = {
                "contents": [
                    {
                        "parts": [
                            {
                                "text": "Please transcribe the following audio file. Provide a clear, accurate transcription of the conversation."
                            },
                            {
                                "inline_data": {
                                    "mime_type": "audio/mpeg",  # Assuming MP3 format
                                    "data": audio_base64
                                }
                            }
                        ]
                    }
                ]
            }
            
            params = {
                "key": self.gemini_api_key
            }
            
            print("🤖 Sending audio to Gemini for transcription...")
            response = requests.post(url, headers=headers, json=data, params=params)
            response.raise_for_status()
            
            result = response.json()
            
            if "candidates" in result and len(result["candidates"]) > 0:
                transcription = result["candidates"][0]["content"]["parts"][0]["text"]
                print("✅ Transcription completed successfully")
                return transcription
            else:
                print("❌ No transcription generated")
                return None
                
        except Exception as e:
            print(f"❌ Error transcribing audio: {e}")
            return None
    
    def answer_questions_with_gemini(self, transcription: str, questions: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """
        Answer questions using the transcription and Gemini.
        
        Args:
            transcription (str): Transcribed text
            questions (List[Dict[str, Any]]): List of questions to answer
            
        Returns:
            List[Dict[str, Any]]: Questions with answers
        """
        try:
            # Prepare questions text for Gemini with specific instructions and answer type constraints
            questions_text = ""
            answer_constraints = []
            
            for i, q in enumerate(questions, 1):
                question_text = q['question_text']
                answer_type = q.get('answer_type', 'text')
                instructions = q.get('instructions', '')
                
                questions_text += f"{i}. {question_text}\n"
                
                # Add answer type constraints
                if answer_type == 'boolean':
                    answer_constraints.append(f"Question {i}: Answer must be ONLY 'true' or 'false'")
                elif answer_type == 'integer':
                    answer_constraints.append(f"Question {i}: Answer must be ONLY a number (no units, no text)")
                elif answer_type == 'description':
                    answer_constraints.append(f"Question {i}: Answer must be a descriptive summary")
                else:
                    answer_constraints.append(f"Question {i}: Answer should be clear and concise")
                
                if instructions:
                    answer_constraints.append(f"Question {i}: {instructions}")
            
            constraints_text = "\n".join(answer_constraints)
            
            prompt = f"""
            Based on the following call transcription, please answer the questions below. 
            Provide clear, concise answers based on the information available in the transcription.
            If information is not available in the transcription, please state "Information not available in the call."

            TRANSCRIPTION:
            {transcription}

            QUESTIONS TO ANSWER:
            {questions_text}

            ANSWER CONSTRAINTS:
            {constraints_text}

            IMPORTANT: Follow the answer type constraints exactly. For boolean questions, answer only 'true' or 'false'. For integer questions, answer only the number. For description questions, provide a summary.

            Please provide your answers in the following format:
            Answer 1: [your answer]
            Answer 2: [your answer]
            etc.
            """
            
            url = f"https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-pro:generateContent"
            
            headers = {
                "Content-Type": "application/json",
            }
            
            data = {
                "contents": [
                    {
                        "parts": [
                            {
                                "text": prompt
                            }
                        ]
                    }
                ]
            }
            
            params = {
                "key": self.gemini_api_key
            }
            
            print("🤖 Sending questions to Gemini for answering...")
            print(f"\n📝 PROMPT BEING SENT TO GEMINI:")
            print("=" * 80)
            print(prompt)
            print("=" * 80)
            
            response = requests.post(url, headers=headers, json=data, params=params)
            response.raise_for_status()
            
            result = response.json()
            
            if "candidates" in result and len(result["candidates"]) > 0:
                answers_text = result["candidates"][0]["content"]["parts"][0]["text"]
                print("✅ Questions answered successfully")
                
                # Parse answers and add to questions
                answers = self.parse_answers(answers_text, len(questions))
                
                for i, question in enumerate(questions):
                    # Update the answer in the details JSONB field
                    if 'details' in question and question['details']:
                        question['details']['answer'] = answers[i] if i < len(answers) else "Answer not available"
                    
                    # Also add to the main question object for easy access
                    question['answer'] = answers[i] if i < len(answers) else "Answer not available"
                    question['answered_at'] = datetime.now().isoformat()
                
                return questions
            else:
                print("❌ No answers generated")
                return questions
                
        except Exception as e:
            print(f"❌ Error answering questions: {e}")
            return questions
    
    def parse_answers(self, answers_text: str, num_questions: int) -> List[str]:
        """
        Parse answers from Gemini response.
        
        Args:
            answers_text (str): Raw answers text from Gemini
            num_questions (int): Number of questions
            
        Returns:
            List[str]: List of parsed answers
        """
        answers = []
        lines = answers_text.split('\n')
        
        print(f"\n🔍 Parsing answers from Gemini response:")
        print(f"Raw response: {answers_text[:200]}...")
        
        for i in range(1, num_questions + 1):
            answer_found = False
            for line in lines:
                line = line.strip()
                # Try different patterns
                if (line.startswith(f"Answer {i}:") or 
                    line.startswith(f"{i}.") or 
                    line.startswith(f"{i}:") or
                    line.startswith(f"Answer {i}") or
                    line.startswith(f"Q{i}") or
                    line.startswith(f"Question {i}")):
                    
                    # Extract answer after colon or period
                    if ':' in line:
                        answer = line.split(':', 1)[1].strip()
                    elif '.' in line and not line.startswith(f"{i}."):
                        answer = line.split('.', 1)[1].strip()
                    else:
                        answer = line
                    
                    # Clean up the answer
                    answer = answer.strip('.,-').strip()
                    answers.append(answer)
                    answer_found = True
                    print(f"   Question {i}: {answer}")
                    break
            
            if not answer_found:
                answers.append("Answer not found in response")
                print(f"   Question {i}: Answer not found in response")
        
        return answers
    
    def save_results(self, call_data: Dict[str, Any], transcription: str, questions_with_answers: List[Dict[str, Any]]):
        """
        Save results to database and files.
        
        Args:
            call_data (Dict[str, Any]): Call data
            transcription (str): Transcription text
            questions_with_answers (List[Dict[str, Any]]): Questions with answers
        """
        try:
            # Save transcription to call_logs table
            update_query = """
                UPDATE "smartFlo".call_logs 
                SET "transcriptionJSON" = %s, "transcribeAttempt" = "transcribeAttempt" + 1
                WHERE id = %s;
            """
            
            transcription_data = {
                "transcription": transcription,
                "transcribed_at": datetime.now().isoformat(),
                "questions_answered": len(questions_with_answers)
            }
            
            self.cursor.execute(update_query, (json.dumps(transcription_data), call_data['id']))
            self.db_connection.commit()
            print("✅ Transcription saved to database")
            
            # Save results to JSON file
            results = {
                "call_id": call_data['id'],
                "call_data": call_data,
                "transcription": transcription,
                "questions_and_answers": questions_with_answers,
                "processed_at": datetime.now().isoformat()
            }
            
            timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
            filename = f"transcription_results_{call_data['id']}_{timestamp}.json"
            
            with open(filename, 'w', encoding='utf-8') as f:
                json.dump(results, f, indent=2, ensure_ascii=False)
            
            print(f"✅ Results saved to: {filename}")
            
        except Exception as e:
            print(f"❌ Error saving results: {e}")
            try:
                self.db_connection.rollback()
                print("🔄 Database transaction rolled back")
            except:
                pass
    
    def process_call(self, call_id: str):
        """
        Process a call: transcribe audio and answer questions.
        
        Args:
            call_id (str): The call log ID to process
        """
        print(f"🚀 Starting transcription pipeline for call ID: {call_id}")
        print("="*80)
        
        # Step 1: Connect to database
        if not self.connect_to_database():
            return
        
        try:
            # Step 2: Get call data
            call_data = self.get_call_data(call_id)
            if not call_data:
                return
            
            if not call_data.get('recording_url'):
                print("❌ No recording URL found for this call")
                return
            
            # Step 3: Get questions
            questions = self.get_questions()
            if not questions:
                print("⚠️  No questions found, proceeding with transcription only")
                questions = []
            
            # Step 4: Download audio
            audio_content = self.download_audio(call_data['recording_url'])
            if not audio_content:
                return
            
            # Step 5: Transcribe audio
            transcription = self.transcribe_audio_with_gemini(audio_content)
            if not transcription:
                return
            
            print(f"\n📝 TRANSCRIPTION:")
            print("-" * 80)
            print(transcription)
            print("-" * 80)
            
            # Step 6: Answer questions
            questions_with_answers = questions
            if questions:
                questions_with_answers = self.answer_questions_with_gemini(transcription, questions)
                
                print(f"\n❓ QUESTIONS AND ANSWERS:")
                print("-" * 80)
                for i, qa in enumerate(questions_with_answers, 1):
                    print(f"{i}. {qa['question_text']}")
                    print(f"   Answer: {qa.get('answer', 'No answer available')}")
                    print(f"   Label: {qa.get('label', 'N/A')}")
                    print()
            
            # Step 7: Save results
            self.save_results(call_data, transcription, questions_with_answers)
            
            print("🎉 Pipeline completed successfully!")
            
        except Exception as e:
            print(f"❌ Pipeline error: {e}")
        finally:
            self.disconnect_from_database()

def main():
    """Main function to run the transcription pipeline."""
    # Configuration
    DB_CONNECTION_STRING = "postgres://postgres:Badho_1301@db.badho.in:5432/badho-app"
    GEMINI_API_KEY = "AIzaSyATn1vcksF5BIJiBSn31CGfdslfysGtpOc"
    
    # Get call ID from command line argument
    if len(sys.argv) != 2:
        print("Usage: python3 transcription_pipeline.py <call_id>")
        print("Example: python3 transcription_pipeline.py c86d4b0d-5c9b-4edf-8b07-08a4833dcf50")
        sys.exit(1)
    
    call_id = sys.argv[1]
    
    # Create and run pipeline
    pipeline = TranscriptionPipeline(DB_CONNECTION_STRING, GEMINI_API_KEY)
    pipeline.process_call(call_id)

if __name__ == "__main__":
    main()
