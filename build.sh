#!/bin/bash

# build.sh - Script to build the Plannet application
# This script builds the Plannet CLI application for various platforms

# Set colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Function to print section headers
print_header() {
  echo -e "\n${YELLOW}=== $1 ===${NC}"
}

# Function to run a command and report result
run_command() {
  echo -e "\n${YELLOW}Running: $1${NC}"
  $2
  if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Command successful: $1${NC}"
    return 0
  else
    echo -e "${RED}✗ Command failed: $1${NC}"
    return 1
  fi
}

# Check if Go is installed
print_header "Checking Go Installation"
if ! command -v go &> /dev/null; then
  echo -e "${RED}Go is not installed or not in PATH${NC}"
  echo "Please install Go first: https://golang.org/doc/install"
  exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo -e "Go version: ${GREEN}$GO_VERSION${NC}"

# Create build directory
print_header "Creating Build Directory"
BUILD_DIR="build"
mkdir -p $BUILD_DIR
run_command "Create build directory" "mkdir -p $BUILD_DIR"

# Run tests
print_header "Running Tests"
run_command "Run tests" "go test ./..."

# Build for current platform
print_header "Building for Current Platform"
PLATFORM=$(go env GOOS)
ARCH=$(go env GOARCH)
OUTPUT="$BUILD_DIR/plannet"
if [ "$PLATFORM" = "windows" ]; then
  OUTPUT="$OUTPUT.exe"
fi
run_command "Build for $PLATFORM/$ARCH" "go build -o $OUTPUT ."

# Build for other platforms
print_header "Building for Other Platforms"

# Build for Linux
run_command "Build for Linux/amd64" "GOOS=linux GOARCH=amd64 go build -o $BUILD_DIR/plannet-linux-amd64 ."
run_command "Build for Linux/arm64" "GOOS=linux GOARCH=arm64 go build -o $BUILD_DIR/plannet-linux-arm64 ."

# Build for macOS
run_command "Build for macOS/amd64" "GOOS=darwin GOARCH=amd64 go build -o $BUILD_DIR/plannet-darwin-amd64 ."
run_command "Build for macOS/arm64" "GOOS=darwin GOARCH=arm64 go build -o $BUILD_DIR/plannet-darwin-arm64 ."

# Build for Windows
run_command "Build for Windows/amd64" "GOOS=windows GOARCH=amd64 go build -o $BUILD_DIR/plannet-windows-amd64.exe ."
run_command "Build for Windows/arm64" "GOOS=windows GOARCH=arm64 go build -o $BUILD_DIR/plannet-windows-arm64.exe ."

# Create release archive
print_header "Creating Release Archives"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
run_command "Create Linux archives" "cd $BUILD_DIR && tar -czf plannet-linux-amd64-$VERSION.tar.gz plannet-linux-amd64 && tar -czf plannet-linux-arm64-$VERSION.tar.gz plannet-linux-arm64"
run_command "Create macOS archives" "cd $BUILD_DIR && tar -czf plannet-darwin-amd64-$VERSION.tar.gz plannet-darwin-amd64 && tar -czf plannet-darwin-arm64-$VERSION.tar.gz plannet-darwin-arm64"
run_command "Create Windows archives" "cd $BUILD_DIR && zip plannet-windows-amd64-$VERSION.zip plannet-windows-amd64.exe && zip plannet-windows-arm64-$VERSION.zip plannet-windows-arm64.exe"

# Install locally
print_header "Installing Locally"
run_command "Install locally" "go install"

print_header "Build Complete"
echo -e "${GREEN}Plannet has been built successfully!${NC}"
echo -e "Binaries are available in the ${YELLOW}$BUILD_DIR${NC} directory"
echo -e "The application has been installed locally and can be run with ${YELLOW}plannet${NC}"
echo -e "To test the track feature, run ${YELLOW}./test_track.sh${NC}" 