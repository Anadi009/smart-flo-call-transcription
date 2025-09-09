#!/bin/bash

echo "Building Go Lambda function for API Gateway..."

# Clean up any previous builds
rm -f main bootstrap deployment.zip

echo "🔨 Compiling Go binary..."
# Build for Linux (AWS Lambda environment) - AWS Lambda Go runtime expects 'bootstrap'
GOOS=linux GOARCH=amd64 go build -o bootstrap main.go

# Check if build was successful
if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
else
    echo "❌ Build failed!"
    exit 1
fi

# Make executable
chmod +x bootstrap

echo "📦 Creating deployment package..."
# Create zip file for AWS Lambda
zip deployment.zip bootstrap

# Get file size
SIZE=$(du -h deployment.zip | cut -f1)
echo "✅ Deployment package created: deployment.zip"
echo "📦 Package size: $SIZE"

echo ""
echo "🚀 Lambda Configuration:"
echo "   Runtime: Go 1.x"
echo "   Handler: bootstrap"
echo "   Architecture: x86_64"
echo ""
echo "🚀 Ready for AWS Lambda deployment!"
echo "Upload deployment.zip to your Lambda function"
