package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// LambdaRequest represents the incoming Lambda event
type LambdaRequest struct {
	CallLogsID string `json:"call_logsId"`
}

// LambdaResponse represents the Lambda response
type LambdaResponse struct {
	StatusCode int         `json:"statusCode"`
	Body       interface{} `json:"body"`
	Error      string      `json:"error,omitempty"`
}

// CallData represents call information from the database
type CallData struct {
	ID              string    `json:"id"`
	RecordingURL    string    `json:"recording_url"`
	CallID          string    `json:"call_id"`
	CallerIDNumber  string    `json:"caller_id_number"`
	CallToNumber    string    `json:"call_to_number"`
	StartDate       string    `json:"start_date"`
	StartTime       string    `json:"start_time"`
	Duration        int       `json:"duration"`
	AgentName       string    `json:"agent_name"`
	CampaignName    string    `json:"campaign_name"`
	CampaignID      string    `json:"campaignId"`
}

// Question represents a question from the database
type Question struct {
	ID           string                 `json:"id"`
	Label        string                 `json:"label"`
	IsActive     bool                   `json:"isActive"`
	Details      map[string]interface{} `json:"details"`
	QuestionText string                 `json:"question_text"`
	AnswerType   string                 `json:"answer_type"`
	Instructions string                 `json:"instructions"`
	Answer       string                 `json:"answer,omitempty"`
	AnsweredAt   string                 `json:"answered_at,omitempty"`
}

// CallAnalysisData represents the data to be saved in callAnalysis column
type CallAnalysisData struct {
	Transcription string            `json:"transcription"`
	Answers       map[string]string `json:"answers"`
	ProcessedAt   string            `json:"processed_at"`
}

// GeminiRequest represents the request to Gemini API
type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text       string     `json:"text,omitempty"`
	InlineData *InlineData `json:"inline_data,omitempty"`
}

type InlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

// GeminiResponse represents the response from Gemini API
type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content Content `json:"content"`
}

// TranscriptionPipeline handles the transcription process
type TranscriptionPipeline struct {
	dbConnectionString string
	geminiAPIKey       string
	db                 *sql.DB
}

// NewTranscriptionPipeline creates a new pipeline instance
func NewTranscriptionPipeline(dbConnectionString, geminiAPIKey string) *TranscriptionPipeline {
	return &TranscriptionPipeline{
		dbConnectionString: dbConnectionString,
		geminiAPIKey:       geminiAPIKey,
	}
}

// ConnectToDatabase establishes connection to PostgreSQL
func (tp *TranscriptionPipeline) ConnectToDatabase() error {
	db, err := sql.Open("postgres", tp.dbConnectionString)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %v", err)
	}

	// Set connection timeouts
	db.SetConnMaxLifetime(30 * time.Second)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	tp.db = db
	return nil
}

// CloseDatabase closes the database connection
func (tp *TranscriptionPipeline) CloseDatabase() {
	if tp.db != nil {
		tp.db.Close()
	}
}

// GetCallData retrieves call data from the database
func (tp *TranscriptionPipeline) GetCallData(callLogsID string) (*CallData, error) {
	query := `
		SELECT id, recording_url, call_id, caller_id_number, call_to_number, 
		       start_date, start_time, duration, agent_name, campaign_name, "campaignId"
		FROM "smartFlo".call_logs 
		WHERE id = $1
	`

	var callData CallData
	err := tp.db.QueryRow(query, callLogsID).Scan(
		&callData.ID,
		&callData.RecordingURL,
		&callData.CallID,
		&callData.CallerIDNumber,
		&callData.CallToNumber,
		&callData.StartDate,
		&callData.StartTime,
		&callData.Duration,
		&callData.AgentName,
		&callData.CampaignName,
		&callData.CampaignID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no call found with ID: %s", callLogsID)
		}
		return nil, fmt.Errorf("error fetching call data: %v", err)
	}

	return &callData, nil
}

// GetQuestionsForCampaign retrieves questions specific to the campaign
func (tp *TranscriptionPipeline) GetQuestionsForCampaign(campaignID string) ([]Question, error) {
	query := `
		SELECT q.id, q.label, q."isActive", q.details
		FROM "smartFlo".question q
		INNER JOIN "smartFlo".campaign_question cq ON q.id = cq."questionId"
		WHERE q."isActive" = true AND cq."campaignId" = $1
		ORDER BY q.id
	`

	rows, err := tp.db.Query(query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("error fetching questions for campaign: %v", err)
	}
	defer rows.Close()

	var questions []Question
	for rows.Next() {
		var q Question
		var detailsJSON []byte

		err := rows.Scan(&q.ID, &q.Label, &q.IsActive, &detailsJSON)
		if err != nil {
			return nil, fmt.Errorf("error scanning question row: %v", err)
		}

		// Parse details JSON
		if err := json.Unmarshal(detailsJSON, &q.Details); err != nil {
			return nil, fmt.Errorf("error parsing question details: %v", err)
		}

		// Extract question text and other fields from details
		if questionText, ok := q.Details["questionText"].(string); ok {
			q.QuestionText = questionText
		}
		if answerType, ok := q.Details["answerType"].(string); ok {
			q.AnswerType = answerType
		} else {
			q.AnswerType = "text"
		}
		if instructions, ok := q.Details["instructions"].(string); ok {
			q.Instructions = instructions
		}

		questions = append(questions, q)
	}

	return questions, nil
}

// DownloadAudio downloads audio file from URL
func (tp *TranscriptionPipeline) DownloadAudio(recordingURL string) ([]byte, error) {
	resp, err := http.Get(recordingURL)
	if err != nil {
		return nil, fmt.Errorf("error downloading audio: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error downloading audio: status %d", resp.StatusCode)
	}

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading audio data: %v", err)
	}

	return audioData, nil
}

// TranscribeAudioOnly transcribes audio without answering questions
func (tp *TranscriptionPipeline) TranscribeAudioOnly(audioContent []byte) (string, error) {
	// Encode audio to base64
	audioBase64 := base64.StdEncoding.EncodeToString(audioContent)

	prompt := "Please transcribe the following audio file."

	// Prepare the request
	geminiURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-pro:generateContent"
	
	requestData := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{
						Text: prompt,
					},
					{
						InlineData: &InlineData{
							MimeType: "audio/mpeg",
							Data:     audioBase64,
						},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", geminiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	
	// Add API key as query parameter
	q := req.URL.Query()
	q.Add("key", tp.geminiAPIKey)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gemini API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	if len(geminiResp.Candidates) == 0 {
		return "", fmt.Errorf("no response generated from Gemini API")
	}

	if len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content parts in Gemini response")
	}

	transcription := geminiResp.Candidates[0].Content.Parts[0].Text
	if transcription == "" {
		return "", fmt.Errorf("empty transcription received from Gemini API")
	}

	return transcription, nil
}

// ProcessAudioWithGemini transcribes audio and answers questions in a single call
func (tp *TranscriptionPipeline) ProcessAudioWithGemini(audioContent []byte, questions []Question) (string, map[string]string, error) {
	// Encode audio to base64
	audioBase64 := base64.StdEncoding.EncodeToString(audioContent)

	// Prepare questions text for Gemini using details from database
	questionsText := ""
	var answerConstraints []string
	questionIDs := make([]string, len(questions))

	for i, q := range questions {
		questionIDs[i] = q.ID
		questionsText += fmt.Sprintf("%d. %s\n", i+1, q.QuestionText)

		// Use instructions from details column instead of hardcoded constraints
		if q.Instructions != "" {
			answerConstraints = append(answerConstraints, fmt.Sprintf("Question %d: %s", i+1, q.Instructions))
		} else {
			// Fallback to basic constraints if no instructions in details
			switch q.AnswerType {
			case "boolean":
				answerConstraints = append(answerConstraints, fmt.Sprintf("Question %d: Answer must be ONLY 'true' or 'false'", i+1))
			case "integer":
				answerConstraints = append(answerConstraints, fmt.Sprintf("Question %d: Answer must be ONLY a number (no units, no text)", i+1))
			case "description":
				answerConstraints = append(answerConstraints, fmt.Sprintf("Question %d: Answer must be a descriptive summary", i+1))
			default:
				answerConstraints = append(answerConstraints, fmt.Sprintf("Question %d: Answer should be clear and concise", i+1))
			}
		}
	}

	constraintsText := strings.Join(answerConstraints, "\n")

	prompt := fmt.Sprintf(`
Please transcribe the following audio file and then answer the questions based on the transcription.

QUESTIONS TO ANSWER:
%s

ANSWER CONSTRAINTS:
%s

IMPORTANT: Follow the answer constraints exactly as specified for each question.

Please provide your response in the following format:
TRANSCRIPTION:
[transcribed text here]

ANSWERS:
Answer 1: [your answer]
Answer 2: [your answer]
etc.
`, questionsText, constraintsText)

	// Prepare the request
	geminiURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-pro:generateContent"
	
	requestData := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{
						Text: prompt,
					},
					{
						InlineData: &InlineData{
							MimeType: "audio/mpeg",
							Data:     audioBase64,
						},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", nil, fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", geminiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	
	// Add API key as query parameter
	q := req.URL.Query()
	q.Add("key", tp.geminiAPIKey)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 45 * time.Second} // Reduced timeout for faster failure
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("gemini API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", nil, fmt.Errorf("error decoding response: %v", err)
	}

	if len(geminiResp.Candidates) == 0 {
		return "", nil, fmt.Errorf("no response generated from Gemini API")
	}

	if len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", nil, fmt.Errorf("no content parts in Gemini response")
	}

	responseText := geminiResp.Candidates[0].Content.Parts[0].Text
	if responseText == "" {
		return "", nil, fmt.Errorf("empty response received from Gemini API")
	}
	
	// Parse transcription and answers
	transcription, answers := tp.parseTranscriptionAndAnswers(responseText, questionIDs)
	
	return transcription, answers, nil
}

// parseTranscriptionAndAnswers parses the combined response from Gemini
func (tp *TranscriptionPipeline) parseTranscriptionAndAnswers(responseText string, questionIDs []string) (string, map[string]string) {
	transcription := ""
	answers := make(map[string]string)
	
	lines := strings.Split(responseText, "\n")
	inTranscription := false
	inAnswers := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "TRANSCRIPTION:") {
			inTranscription = true
			inAnswers = false
			// Get transcription content after the colon
			if parts := strings.SplitN(line, ":", 2); len(parts) > 1 {
				transcription = strings.TrimSpace(parts[1])
			}
			continue
		}
		
		if strings.HasPrefix(line, "ANSWERS:") {
			inTranscription = false
			inAnswers = true
			continue
		}
		
		if inTranscription && line != "" {
			if transcription != "" {
				transcription += "\n"
			}
			transcription += line
		}
		
		if inAnswers && strings.HasPrefix(line, "Answer ") {
			// Parse answer lines like "Answer 1: [answer]"
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				answerNum := strings.TrimSpace(parts[0])
				answer := strings.TrimSpace(parts[1])
				
				// Extract number from "Answer X"
				if strings.HasPrefix(answerNum, "Answer ") {
					numStr := strings.TrimSpace(strings.TrimPrefix(answerNum, "Answer"))
					if num, err := strconv.Atoi(numStr); err == nil && num > 0 && num <= len(questionIDs) {
						answers[questionIDs[num-1]] = answer
					}
				}
			}
		}
	}
	
	return transcription, answers
}

// SaveCallAnalysis saves the analysis data to the callAnalysis column
func (tp *TranscriptionPipeline) SaveCallAnalysis(callLogsID string, transcription string, answers map[string]string) error {
	// Prepare the analysis data
	analysisData := CallAnalysisData{
		Transcription: transcription,
		Answers:       answers,
		ProcessedAt:   time.Now().Format(time.RFC3339),
	}

	// Convert to JSON
	analysisJSON, err := json.Marshal(analysisData)
	if err != nil {
		return fmt.Errorf("error marshaling analysis data: %v", err)
	}

	// Update only the callAnalysis column for the specific ID
	updateQuery := `
		UPDATE "smartFlo".call_logs 
		SET "callAnalysis" = $1
		WHERE id = $2
	`

	_, err = tp.db.Exec(updateQuery, string(analysisJSON), callLogsID)
	if err != nil {
		return fmt.Errorf("error updating callAnalysis: %v", err)
	}

	return nil
}

// ProcessCall processes a call: transcribe audio and answer questions
func (tp *TranscriptionPipeline) ProcessCall(callLogsID string) (map[string]interface{}, error) {
	// Connect to database
	if err := tp.ConnectToDatabase(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer tp.CloseDatabase()

	// Get call data
	callData, err := tp.GetCallData(callLogsID)
	if err != nil {
		return nil, fmt.Errorf("failed to get call data: %v", err)
	}

	if callData.RecordingURL == "" {
		return nil, fmt.Errorf("no recording URL found for this call")
	}

	if callData.CampaignID == "" {
		return nil, fmt.Errorf("no campaign ID found for this call")
	}

	// Get questions specific to the campaign
	questions, err := tp.GetQuestionsForCampaign(callData.CampaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get questions for campaign: %v", err)
	}

	// Download audio
	audioContent, err := tp.DownloadAudio(callData.RecordingURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download audio: %v", err)
	}

	// Check if audio content is empty
	if len(audioContent) == 0 {
		return nil, fmt.Errorf("downloaded audio file is empty")
	}

	var transcription string
	var answers map[string]string

	if len(questions) == 0 {
		// No questions linked to campaign - only transcribe audio
		transcription, err = tp.TranscribeAudioOnly(audioContent)
		if err != nil {
			return nil, fmt.Errorf("failed to transcribe audio: %v", err)
		}
		answers = make(map[string]string)
	} else {
		// Process audio and answer questions in a single call
		transcription, answers, err = tp.ProcessAudioWithGemini(audioContent, questions)
		if err != nil {
			return nil, fmt.Errorf("failed to process audio: %v", err)
		}
	}

	// Save analysis data to callAnalysis column
	if err := tp.SaveCallAnalysis(callLogsID, transcription, answers); err != nil {
		return nil, fmt.Errorf("failed to save call analysis: %v", err)
	}

	// Create minimal response with only essential data
	result := map[string]interface{}{
		"call_logsId":  callLogsID,
		"campaignId":   callData.CampaignID,
		"transcription": transcription,
		"answers":       answers,
		"processed_at":  time.Now().Format(time.RFC3339),
	}

	return result, nil
}

// LambdaHandler handles Lambda events
func LambdaHandler(ctx context.Context, request LambdaRequest) (LambdaResponse, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		// If .env file doesn't exist, continue with environment variables
	}

	// Get configuration from environment variables
	dbConnectionString := os.Getenv("DB_CONNECTION_STRING")
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")

	if dbConnectionString == "" {
		dbConnectionString = "postgres://postgres:Badho_1301@db.badho.in:5432/badho-app?connect_timeout=10&statement_timeout=30000"
	}
	if geminiAPIKey == "" {
		geminiAPIKey = "AIzaSyATn1vcksF5BIJiBSn31CGfdslfysGtpOc"
	}

	// Create pipeline
	pipeline := NewTranscriptionPipeline(dbConnectionString, geminiAPIKey)

	// Process the call
	result, err := pipeline.ProcessCall(request.CallLogsID)
	if err != nil {
		return LambdaResponse{
			StatusCode: 500,
			Error:      err.Error(),
		}, nil
	}

	return LambdaResponse{
		StatusCode: 200,
		Body:       result,
	}, nil
}

func main() {
	lambda.Start(LambdaHandler)
}

