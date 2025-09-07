#!/usr/bin/env python3
"""
Simple script to run the transcription pipeline
Usage: python3 run_pipeline.py <call_id>
"""

import sys
import subprocess
import os

def main():
    if len(sys.argv) != 2:
        print("🚀 SmartFlo Call Transcription Pipeline")
        print("=" * 50)
        print("Usage: python3 run_pipeline.py <call_id>")
        print()
        print("Example:")
        print("  python3 run_pipeline.py c86d4b0d-5c9b-4edf-8b07-08a4833dcf50")
        print()
        print("This will:")
        print("  1. Fetch call data from smartFlo.call_logs")
        print("  2. Download the audio recording")
        print("  3. Transcribe using Google Gemini AI")
        print("  4. Answer questions from smartFlo.question table")
        print("  5. Save results to database and JSON file")
        sys.exit(1)
    
    call_id = sys.argv[1]
    
    print(f"🚀 Starting transcription pipeline for call ID: {call_id}")
    print("=" * 80)
    
    try:
        # Run the transcription pipeline
        result = subprocess.run([
            sys.executable, 
            "transcription_pipeline.py", 
            call_id
        ], check=True, capture_output=False)
        
        print("\n🎉 Pipeline completed successfully!")
        
    except subprocess.CalledProcessError as e:
        print(f"\n❌ Pipeline failed with error code: {e.returncode}")
        sys.exit(1)
    except KeyboardInterrupt:
        print("\n\n⏹️  Pipeline interrupted by user")
        sys.exit(1)
    except Exception as e:
        print(f"\n❌ Unexpected error: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
