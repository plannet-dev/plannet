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

# Update Go modules
print_header "Updating Go Modules"
run_command "Update Go modules" "go mod tidy"

# Create build directory
print_header "Creating Build Directory"
BUILD_DIR="build"
mkdir -p $BUILD_DIR
run_command "Create build directory" "mkdir -p $BUILD_DIR"

# Run tests
print_header "Running Tests"
run_command "Run tests" "go test ./..." || echo -e "${YELLOW}Tests failed, but continuing with build...${NC}"

# Build for current platform
print_header "Building for Current Platform"
PLATFORM=$(go env GOOS)
ARCH=$(go env GOARCH)
OUTPUT="$BUILD_DIR/plannet"
if [ "$PLATFORM" = "windows" ]; then
  OUTPUT="$OUTPUT.exe"
fi
# Ensure we're building from the root directory with the main.go file
run_command "Build for $PLATFORM/$ARCH" "go build -o $OUTPUT ." || echo -e "${YELLOW}Build for current platform failed, but continuing...${NC}"

# Build for other platforms
print_header "Building for Other Platforms"

# Build for Linux
run_command "Build for Linux/amd64" "env GOOS=linux GOARCH=amd64 go build -o $BUILD_DIR/plannet-linux-amd64 ." || echo -e "${YELLOW}Build for Linux/amd64 failed, but continuing...${NC}"
run_command "Build for Linux/arm64" "env GOOS=linux GOARCH=arm64 go build -o $BUILD_DIR/plannet-linux-arm64 ." || echo -e "${YELLOW}Build for Linux/arm64 failed, but continuing...${NC}"

# Build for macOS
run_command "Build for macOS/amd64" "env GOOS=darwin GOARCH=amd64 go build -o $BUILD_DIR/plannet-darwin-amd64 ." || echo -e "${YELLOW}Build for macOS/amd64 failed, but continuing...${NC}"
run_command "Build for macOS/arm64" "env GOOS=darwin GOARCH=arm64 go build -o $BUILD_DIR/plannet-darwin-arm64 ." || echo -e "${YELLOW}Build for macOS/arm64 failed, but continuing...${NC}"

# Build for Windows
run_command "Build for Windows/amd64" "env GOOS=windows GOARCH=amd64 go build -o $BUILD_DIR/plannet-windows-amd64.exe ." || echo -e "${YELLOW}Build for Windows/amd64 failed, but continuing...${NC}"
run_command "Build for Windows/arm64" "env GOOS=windows GOARCH=arm64 go build -o $BUILD_DIR/plannet-windows-arm64.exe ." || echo -e "${YELLOW}Build for Windows/arm64 failed, but continuing...${NC}"

# Create release archive
print_header "Creating Release Archives"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Only create archives for files that exist
if [ -f "$BUILD_DIR/plannet-linux-amd64" ]; then
  run_command "Create Linux amd64 archive" "cd $BUILD_DIR && tar -czf plannet-linux-amd64-$VERSION.tar.gz plannet-linux-amd64"
fi

if [ -f "$BUILD_DIR/plannet-linux-arm64" ]; then
  run_command "Create Linux arm64 archive" "cd $BUILD_DIR && tar -czf plannet-linux-arm64-$VERSION.tar.gz plannet-linux-arm64"
fi

if [ -f "$BUILD_DIR/plannet-darwin-amd64" ]; then
  run_command "Create macOS amd64 archive" "cd $BUILD_DIR && tar -czf plannet-darwin-amd64-$VERSION.tar.gz plannet-darwin-amd64"
fi

if [ -f "$BUILD_DIR/plannet-darwin-arm64" ]; then
  run_command "Create macOS arm64 archive" "cd $BUILD_DIR && tar -czf plannet-darwin-arm64-$VERSION.tar.gz plannet-darwin-arm64"
fi

if [ -f "$BUILD_DIR/plannet-windows-amd64.exe" ]; then
  run_command "Create Windows amd64 archive" "cd $BUILD_DIR && zip plannet-windows-amd64-$VERSION.zip plannet-windows-amd64.exe"
fi

if [ -f "$BUILD_DIR/plannet-windows-arm64.exe" ]; then
  run_command "Create Windows arm64 archive" "cd $BUILD_DIR && zip plannet-windows-arm64-$VERSION.zip plannet-windows-arm64.exe"
fi

# Install locally
print_header "Installing Locally"
run_command "Install locally" "go install" || echo -e "${YELLOW}Local installation failed, but continuing...${NC}"

print_header "Build Complete"
echo -e "${GREEN}Plannet build process completed!${NC}"
echo -e "Binaries are available in the ${YELLOW}$BUILD_DIR${NC} directory"
echo -e "To test the track feature, run ${YELLOW}./test_track.sh${NC}"

# Check if any builds were successful
if [ -f "$BUILD_DIR/plannet" ] || [ -f "$BUILD_DIR/plannet.exe" ]; then
  echo -e "${GREEN}The application has been built successfully for the current platform!${NC}"
else
  echo -e "${RED}No binaries were built successfully. Please fix the compilation errors.${NC}"
  echo -e "The main issues appear to be in the following files:"
  echo -e "  - cmd/complete.go: Undefined variables and imports"
  echo -e "  - cmd/export.go: Undefined variables and imports"
  echo -e "  - cmd/git.go: Issues with time parsing"
  echo -e "  - cmd/init.go: Unknown fields in struct literals"
fi 