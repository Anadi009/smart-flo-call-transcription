#!/bin/bash

echo "🚀 Building Node.js Lambda function..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}$1${NC}"
}

print_success() {
    echo -e "${GREEN}$1${NC}"
}

print_warning() {
    echo -e "${YELLOW}$1${NC}"
}

print_error() {
    echo -e "${RED}$1${NC}"
}

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    print_error "❌ npm is not installed. Please install Node.js and npm first."
    exit 1
fi

# Check if node is installed
if ! command -v node &> /dev/null; then
    print_error "❌ Node.js is not installed. Please install Node.js first."
    exit 1
fi

print_status "📋 Node.js version: $(node --version)"
print_status "📋 npm version: $(npm --version)"

# Install dependencies if node_modules doesn't exist
if [ ! -d "node_modules" ]; then
    print_status "📦 Installing dependencies..."
    npm install
    if [ $? -ne 0 ]; then
        print_error "❌ Failed to install dependencies!"
        exit 1
    fi
    print_success "✅ Dependencies installed successfully"
else
    print_status "📦 Dependencies already installed"
fi

# Clean previous build
print_status "🧹 Cleaning previous build..."
rm -rf deployment.zip lambda-package/

# Create deployment package directory
print_status "📦 Creating deployment package..."
mkdir -p lambda-package

# Copy source files
cp index.js lambda-package/
cp package.json lambda-package/

# Copy node_modules (only production dependencies)
print_status "📦 Installing production dependencies in package..."
cd lambda-package
npm install --only=production --silent

if [ $? -ne 0 ]; then
    print_error "❌ Failed to install production dependencies!"
    exit 1
fi

# Remove unnecessary files to reduce package size
print_status "🗑️  Removing unnecessary files..."
find node_modules -name "*.md" -delete 2>/dev/null || true
find node_modules -name "*.txt" -delete 2>/dev/null || true
find node_modules -name "LICENSE*" -delete 2>/dev/null || true
find node_modules -name "CHANGELOG*" -delete 2>/dev/null || true
find node_modules -name "*.map" -delete 2>/dev/null || true
find node_modules -name "test" -type d -exec rm -rf {} + 2>/dev/null || true
find node_modules -name "tests" -type d -exec rm -rf {} + 2>/dev/null || true
find node_modules -name "example" -type d -exec rm -rf {} + 2>/dev/null || true
find node_modules -name "examples" -type d -exec rm -rf {} + 2>/dev/null || true

cd ..

# Create ZIP file
print_status "🗜️  Creating deployment ZIP..."
cd lambda-package
zip -r ../deployment.zip . -q

if [ $? -ne 0 ]; then
    print_error "❌ Failed to create deployment package!"
    exit 1
fi

cd ..

# Clean up temporary directory
rm -rf lambda-package/

# Get package size
PACKAGE_SIZE=$(du -h deployment.zip | cut -f1)

print_success "✅ Build successful!"
print_success "📦 Deployment package created: deployment.zip"
print_success "📦 Package size: $PACKAGE_SIZE"

echo ""
print_status "🚀 AWS Lambda Configuration:"
print_status "   Runtime: Node.js 18.x (or later)"
print_status "   Handler: index.handler"
print_status "   Architecture: x86_64"
print_status "   Timeout: 15 minutes (recommended)"
print_status "   Memory: 512 MB (recommended)"

echo ""
print_success "🚀 Ready for AWS Lambda deployment!"
print_success "Upload deployment.zip to your Lambda function"
print_status "Set environment variables: DB_CONNECTION_STRING, GEMINI_API_KEY"
