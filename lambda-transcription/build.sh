#!/bin/bash

# Build script for AWS Lambda deployment

echo "Building Go Lambda function for AWS..."

# Clean previous builds
rm -f main bootstrap deployment.zip

# Build for Linux (required for AWS Lambda)
echo "🔨 Compiling Go binary..."
GOOS=linux GOARCH=amd64 go build -o bootstrap main.go

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
    
    # Make sure the binary is executable
    chmod +x bootstrap
    
    # Create deployment package
    echo "📦 Creating deployment package..."
    zip deployment.zip bootstrap
    
    if [ $? -eq 0 ]; then
        echo "✅ Deployment package created: deployment.zip"
        echo "📦 Package size: $(du -h deployment.zip | cut -f1)"
        echo ""
        echo "�� Lambda Configuration:"
        echo "   Runtime: Go 1.x"
        echo "   Handler: bootstrap"
        echo "   Architecture: x86_64"
    else
        echo "❌ Failed to create deployment package"
        exit 1
    fi
else
    echo "❌ Build failed"
    exit 1
fi

echo ""
echo "🚀 Ready for AWS Lambda deployment!"
echo "Upload deployment.zip to your Lambda function"

