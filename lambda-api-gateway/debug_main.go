package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Request represents the incoming request body
type Request struct {
	CallLogsID string `json:"call_logsId"`
}

// HandleRequest handles API Gateway proxy integration requests
func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Printf("🚀 Debug Lambda started\n")
	fmt.Printf("Request Method: '%s'\n", request.HTTPMethod)
	fmt.Printf("Request Body: '%s'\n", request.Body)

	// Check if it's a GET request
	if request.Body == "" || request.Body == "{}" {
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type": "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: `{"status": "debug mode", "message": "Lambda is working"}`,
		}, nil
	}

	// Parse request
	var req Request
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		fmt.Printf("❌ JSON parsing failed: %v\n", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: fmt.Sprintf(`{"error": "JSON parsing failed: %s"}`, err.Error()),
		}, nil
	}

	fmt.Printf("✅ JSON parsed. CallLogsID: %s\n", req.CallLogsID)

	// Test environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Printf("⚠️  godotenv.Load() failed: %v\n", err)
	}

	dbConnectionString := os.Getenv("DB_CONNECTION_STRING")
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")

	if dbConnectionString == "" {
		dbConnectionString = "postgres://postgres:Badho_1301@db.badho.in:5432/badho-app?connect_timeout=10&statement_timeout=30000"
		fmt.Printf("🔧 Using fallback DB connection\n")
	} else {
		fmt.Printf("✅ DB connection string from env\n")
	}

	if geminiAPIKey == "" {
		geminiAPIKey = "AIzaSyATn1vcksF5BIJiBSn31CGfdslfysGtpOc"
		fmt.Printf("🔧 Using fallback Gemini API key\n")
	} else {
		fmt.Printf("✅ Gemini API key from env\n")
	}

	// Test database connection
	fmt.Printf("🔌 Testing database connection...\n")
	db, err := sql.Open("postgres", dbConnectionString)
	if err != nil {
		fmt.Printf("❌ Database open failed: %v\n", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: fmt.Sprintf(`{"error": "Database open failed: %s"}`, err.Error()),
		}, nil
	}
	defer db.Close()

	// Set connection timeouts
	db.SetConnMaxLifetime(30 * time.Second)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	fmt.Printf("🏓 Pinging database...\n")
	if err := db.Ping(); err != nil {
		fmt.Printf("❌ Database ping failed: %v\n", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: fmt.Sprintf(`{"error": "Database ping failed: %s"}`, err.Error()),
		}, nil
	}

	fmt.Printf("✅ Database connection successful\n")

	// Test simple query
	fmt.Printf("🔍 Testing simple query...\n")
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM \"smartFlo\".call_logs").Scan(&count)
	if err != nil {
		fmt.Printf("❌ Query failed: %v\n", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: fmt.Sprintf(`{"error": "Query failed: %s"}`, err.Error()),
		}, nil
	}

	fmt.Printf("✅ Query successful. Total call_logs: %d\n", count)

	// Test specific call lookup
	fmt.Printf("🔍 Looking up specific call: %s\n", req.CallLogsID)
	var callExists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM \"smartFlo\".call_logs WHERE id = $1)", req.CallLogsID).Scan(&callExists)
	if err != nil {
		fmt.Printf("❌ Call lookup failed: %v\n", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: fmt.Sprintf(`{"error": "Call lookup failed: %s"}`, err.Error()),
		}, nil
	}

	if !callExists {
		fmt.Printf("❌ Call not found: %s\n", req.CallLogsID)
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Headers: map[string]string{
				"Content-Type": "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: `{"error": "Call not found"}`,
		}, nil
	}

	fmt.Printf("✅ Call found: %s\n", req.CallLogsID)

	// Return success
	result := map[string]interface{}{
		"status":          "debug_success",
		"call_logsId":     req.CallLogsID,
		"call_exists":     callExists,
		"total_calls":     count,
		"db_connected":    true,
		"gemini_key_set":  geminiAPIKey != "",
		"processed_at":    time.Now().Format(time.RFC3339),
	}

	jsonBody, _ := json.Marshal(result)
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
	lambda.Start(HandleRequest)
}
