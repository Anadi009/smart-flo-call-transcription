const { Client } = require('pg');
const axios = require('axios');

class TranscriptionPipeline {
    constructor(dbConnectionString, geminiApiKey) {
        this.dbConnectionString = dbConnectionString;
        this.geminiApiKey = geminiApiKey;
        this.client = null;
    }

    async connectToDatabase() {
        // Parse connection string or use individual environment variables
        let config;
        
        if (this.dbConnectionString) {
            // Parse connection string: postgres://user:password@host:port/database
            const url = new URL(this.dbConnectionString);
            config = {
                host: url.hostname,
                port: parseInt(url.port) || 5432,
                database: url.pathname.substring(1), // Remove leading slash
                user: url.username,
                password: url.password,
                ssl: { rejectUnauthorized: false },
                connectionTimeoutMillis: 10000
            };
        } else {
            // Use individual environment variables as fallback
            const requiredEnvVars = ['DB_HOST', 'DB_NAME', 'DB_USER', 'DB_PASSWORD'];
            const missingVars = requiredEnvVars.filter(varName => !process.env[varName]);
            
            if (missingVars.length > 0) {
                throw new Error(`Missing required environment variables: ${missingVars.join(', ')}`);
            }

            config = {
                host: process.env.DB_HOST,
                port: parseInt(process.env.DB_PORT) || 5432,
                database: process.env.DB_NAME,
                user: process.env.DB_USER,
                password: process.env.DB_PASSWORD,
                ssl: { rejectUnauthorized: false },
                connectionTimeoutMillis: 10000
            };
        }
        
        this.client = new Client(config);
        await this.client.connect();
    }

    async closeDatabase() {
        await this.client?.end().catch(() => {});
    }

    async getCallData(callLogsId) {
        const { rows } = await this.client.query(
            'SELECT recording_url, "campaignId" FROM "smartFlo".call_logs WHERE id = $1',
            [callLogsId]
        );
        
        if (!rows.length) throw new Error(`No call found with ID: ${callLogsId}`);
        return { recordingUrl: rows[0].recording_url, campaignId: rows[0].campaignId };
    }

    async getQuestionsForCampaign(campaignId) {
        const { rows } = await this.client.query(`
            SELECT q.id, q.details FROM "smartFlo".question q
            INNER JOIN "smartFlo".campaign_question cq ON q.id = cq."questionId"
            WHERE q."isActive" = true AND cq."campaignId" = $1 ORDER BY q.id
        `, [campaignId]);
        
        return rows.map(({ id, details }) => {
            const d = typeof details === 'string' ? JSON.parse(details) : details;
            return {
                id,
                questionText: d.questionText,
                answerType: d.answerType || 'text',
                instructions: d.instructions || ''
            };
        });
    }

    async downloadAudio(recordingUrl) {
        const { data } = await axios({ url: recordingUrl, responseType: 'arraybuffer', timeout: 30000 });
        const audioData = Buffer.from(data);
        if (!audioData.length) throw new Error('Downloaded audio file is empty');
        return audioData;
    }

    async callGeminiAPI(audioContent, prompt) {
        const { data } = await axios({
            method: 'POST',
            url: `https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-pro:generateContent?key=${this.geminiApiKey}`,
            data: {
                contents: [{
                    parts: [
                        { text: prompt },
                        { inline_data: { mime_type: "audio/mpeg", data: audioContent.toString('base64') } }
                    ]
                }]
            },
            timeout: 90000
        });

        const text = data.candidates?.[0]?.content?.parts?.[0]?.text;
        if (!text) throw new Error('Invalid Gemini API response');
        return text;
    }

    async processAudioWithGemini(audioContent, questions) {
        const questionIds = questions.map(q => q.id);
        const questionsText = questions.map((q, i) => `${i + 1}. ${q.questionText}`).join('\n');
        const constraintsText = questions.map((q, i) => 
            `Question ${i + 1}: ${q.instructions || this.getDefaultConstraint(q.answerType)}`
        ).join('\n');

        const prompt = `Please transcribe the following audio file and then answer the questions based on the transcription.

QUESTIONS TO ANSWER:
${questionsText}

ANSWER CONSTRAINTS:
${constraintsText}

IMPORTANT: Follow the answer constraints exactly as specified for each question.

Please provide your response in the following format:
TRANSCRIPTION:
[transcribed text here]

ANSWERS:
Answer 1: [your answer]
Answer 2: [your answer]
etc.`;

        const responseText = await this.callGeminiAPI(audioContent, prompt);
        return this.parseTranscriptionAndAnswers(responseText, questionIds);
    }

    getDefaultConstraint(answerType) {
        switch (answerType) {
            case 'boolean': return "Answer must be ONLY 'true' or 'false'";
            case 'integer': return "Answer must be ONLY a number (no units, no text)";
            case 'description': return "Answer must be a descriptive summary";
            default: return "Answer should be clear and concise";
        }
    }

    parseTranscriptionAndAnswers(responseText, questionIds) {
        let transcription = "";
        const answers = {};
        let inTranscription = false, inAnswers = false;
        
        for (const line of responseText.split('\n')) {
            const trimmed = line.trim();
            
            if (trimmed.startsWith('TRANSCRIPTION:')) {
                inTranscription = true; inAnswers = false;
                const parts = trimmed.split(':');
                if (parts.length > 1) transcription = parts.slice(1).join(':').trim();
            } else if (trimmed.startsWith('ANSWERS:')) {
                inTranscription = false; inAnswers = true;
            } else if (inTranscription && trimmed) {
                transcription += (transcription ? '\n' : '') + trimmed;
            } else if (inAnswers && trimmed.startsWith('Answer ')) {
                const colonIndex = trimmed.indexOf(':');
                if (colonIndex > 0) {
                    const num = parseInt(trimmed.substring(7, colonIndex).trim());
                    if (num > 0 && num <= questionIds.length) {
                        answers[questionIds[num - 1]] = trimmed.substring(colonIndex + 1).trim();
                    }
                }
            }
        }
        
        return { transcription, answers };
    }

    async saveCallAnalysis(callLogsId, transcription, answers) {
        await this.client.query(
            'UPDATE "smartFlo".call_logs SET "callAnalysis" = $1 WHERE id = $2',
            [JSON.stringify({ transcription, answers, processed_at: new Date().toISOString() }), callLogsId]
        );
    }

    async processCall(callLogsId) {
        const startTime = Date.now();
        console.log(`🔄 Processing call_logsId: ${callLogsId}`);
        
        try {
            console.log(`🔌 Connecting to database...`);
            await this.connectToDatabase();
            console.log(`✅ Database connected successfully`);

            console.log(`📋 Fetching call data...`);
            const { recordingUrl, campaignId } = await this.getCallData(callLogsId);
            console.log(`✅ Call data retrieved - Campaign: ${campaignId}, Recording URL: ${recordingUrl ? 'Present' : 'Missing'}`);
            
            if (!recordingUrl || !campaignId) throw new Error('Missing recording URL or campaign ID');

            console.log(`❓ Fetching questions for campaign: ${campaignId}`);
            const [questions, audioContent] = await Promise.all([
                this.getQuestionsForCampaign(campaignId),
                this.downloadAudio(recordingUrl)
            ]);
            console.log(`✅ Questions fetched: ${questions.length}, Audio downloaded: ${audioContent.length} bytes`);

            console.log(`🤖 Processing with Gemini AI...`);
            const { transcription, answers } = questions.length === 0 
                ? { transcription: await this.callGeminiAPI(audioContent, "Please transcribe the following audio file."), answers: {} }
                : await this.processAudioWithGemini(audioContent, questions);
            console.log(`✅ Gemini processing completed - Transcription: ${transcription?.length || 0} chars, Answers: ${Object.keys(answers || {}).length}`);

            console.log(`💾 Saving results to database...`);
            await this.saveCallAnalysis(callLogsId, transcription, answers);
            console.log(`✅ Results saved successfully`);

            const duration = Date.now() - startTime;
            console.log(`🎉 Processing completed in ${duration}ms`);

            return {
                call_logsId: callLogsId,
                campaignId,
                transcription,
                answers,
                processed_at: new Date().toISOString()
            };
        } catch (error) {
            const duration = Date.now() - startTime;
            console.error(`❌ Processing failed after ${duration}ms:`, error.message);
            throw error;
        } finally {
            console.log(`🔌 Closing database connection...`);
            await this.closeDatabase();
        }
    }
}

exports.handler = async (event) => {
    const isApiGateway = event.httpMethod || event.requestContext;
    const headers = {
        'Content-Type': 'application/json',
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Headers': 'Content-Type',
        'Access-Control-Allow-Methods': 'GET, POST, OPTIONS'
    };

    try {
        // Handle preflight CORS requests
        if (event.httpMethod === 'OPTIONS') {
            return { statusCode: 200, headers, body: JSON.stringify({ message: 'CORS preflight' }) };
        }

        // Handle GET requests - return API info
        if (event.httpMethod === 'GET') {
            const apiInfo = {
                message: 'Smart Flo Call Transcription API',
                version: '1.0.0',
                status: 'online',
                processing: 'asynchronous',
                usage: 'Send POST request with call_logsId in JSON body',
                response: 'Returns 202 Accepted immediately, processing continues in background',
                example: { call_logsId: 'uuid-here' }
            };
            return {
                statusCode: 200,
                headers,
                body: JSON.stringify(apiInfo)
            };
        }

        // Extract call_logsId from different event types
        let callLogsId;
        if (event.call_logsId) {
            // Direct Lambda invocation
            callLogsId = event.call_logsId;
        } else if (event.body) {
            // API Gateway POST request
            try {
                const body = typeof event.body === 'string' ? JSON.parse(event.body) : event.body;
                callLogsId = body.call_logsId;
            } catch (parseError) {
                throw new Error('Invalid JSON in request body');
            }
        } else {
            throw new Error('call_logsId is required');
        }

        if (!callLogsId) throw new Error('call_logsId is required');

        // Validate required environment variables
        const dbConnectionString = process.env.DB_CONNECTION_STRING;
        const geminiApiKey = process.env.GEMINI_API_KEY;
        
        if (!geminiApiKey) {
            throw new Error('GEMINI_API_KEY environment variable is required');
        }

        const pipeline = new TranscriptionPipeline(dbConnectionString, geminiApiKey);
        
        // Fire-and-forget processing - start and return immediately
        console.log(`🚀 Starting fire-and-forget processing for call_logsId: ${callLogsId}`);
        console.log(`⏰ Processing started at: ${new Date().toISOString()}`);
        
        // Start processing in background using process.nextTick to ensure it runs
        process.nextTick(async () => {
            try {
                const result = await pipeline.processCall(callLogsId);
                console.log(`✅ Processing completed successfully for call_logsId: ${callLogsId}`);
                console.log(`📊 Result summary:`, {
                    call_logsId: result.call_logsId,
                    campaignId: result.campaignId,
                    transcriptionLength: result.transcription?.length || 0,
                    answersCount: Object.keys(result.answers || {}).length,
                    processed_at: result.processed_at
                });
            } catch (error) {
                console.error(`❌ Background processing failed for call_logsId: ${callLogsId}`);
                console.error(`🔍 Error details:`, {
                    message: error.message,
                    stack: error.stack,
                    timestamp: new Date().toISOString()
                });
            }
        });
        
        // Return immediately - don't wait for processing
        const response = {
            statusCode: 202,
            message: 'Call processing started in background',
            call_logsId: callLogsId,
            timestamp: new Date().toISOString(),
            note: 'Processing continues in background. Check CloudWatch logs for progress.'
        };
        
        return {
            statusCode: 202,
            ...(isApiGateway && { headers, body: JSON.stringify(response) }),
            ...(!isApiGateway && { body: response })
        };
    } catch (error) {
        return {
            statusCode: 500,
            ...(isApiGateway && { 
                headers, 
                body: JSON.stringify({ 
                    error: error.message,
                    timestamp: new Date().toISOString()
                }) 
            }),
            ...(!isApiGateway && { error: error.message })
        };
    }
};
