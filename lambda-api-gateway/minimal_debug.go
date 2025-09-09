package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("🚀 HANDLER STARTED - Method: '%s'", request.HTTPMethod)
	
	// Log everything about the request
	log.Printf("📋 Request Details:")
	log.Printf("   Method: '%s'", request.HTTPMethod)
	log.Printf("   Path: '%s'", request.Path)
	log.Printf("   Body: '%s'", request.Body)
	log.Printf("   Headers: %+v", request.Headers)
	
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ PANIC RECOVERED: %v", r)
		}
	}()

	// Test environment variables
	log.Printf("🔑 Environment Variables:")
	log.Printf("   DB_CONNECTION_STRING exists: %v", os.Getenv("DB_CONNECTION_STRING") != "")
	log.Printf("   GEMINI_API_KEY exists: %v", os.Getenv("GEMINI_API_KEY") != "")

	// Simple JSON parsing test
	log.Printf("🔄 Testing JSON parsing...")
	var req map[string]interface{}
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		log.Printf("❌ JSON Parse Error: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: fmt.Sprintf(`{"error": "JSON parse failed: %s"}`, err.Error()),
		}, nil
	}
	log.Printf("✅ JSON parsed successfully: %+v", req)

	// Check for call_logsId
	callLogsId, exists := req["call_logsId"]
	if !exists {
		log.Printf("❌ Missing call_logsId in request")
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: `{"error": "call_logsId is required"}`,
		}, nil
	}
	log.Printf("✅ call_logsId found: %v", callLogsId)

	// Return success with minimal processing
	response := map[string]interface{}{
		"status": "minimal_debug_success",
		"call_logsId": callLogsId,
		"message": "Basic request processing successful",
		"timestamp": "2025-09-08T21:00:00Z",
	}

	log.Printf("✅ Creating response...")
	jsonBody, err := json.Marshal(response)
	if err != nil {
		log.Printf("❌ Response marshal error: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: `{"error": "Response marshal failed"}`,
		}, nil
	}

	log.Printf("🎉 SUCCESS - Returning response")
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(jsonBody),
	}, nil
}

func main() {
	log.Printf("🌟 Lambda starting up...")
	lambda.Start(HandleRequest)
}
