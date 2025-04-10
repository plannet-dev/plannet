#!/bin/bash

# test_track.sh - Script to test the Plannet track feature
# This script runs a series of tests on the track command

# Set colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Function to print section headers
print_header() {
  echo -e "\n${YELLOW}=== $1 ===${NC}"
}

# Function to run a test and report result
run_test() {
  echo -e "\n${YELLOW}Running: $1${NC}"
  $2
  if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Test passed: $1${NC}"
    return 0
  else
    echo -e "${RED}✗ Test failed: $1${NC}"
    return 1
  fi
}

# Initialize test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Check if plannet is installed
print_header "Checking Plannet Installation"
if ! command -v plannet &> /dev/null; then
  echo -e "${RED}Plannet is not installed or not in PATH${NC}"
  echo "Please build and install Plannet first using build.sh"
  exit 1
fi

# Check if plannet is initialized
print_header "Checking Plannet Initialization"
if ! plannet init --check &> /dev/null; then
  echo -e "${YELLOW}Plannet is not initialized. Running init...${NC}"
  plannet init --non-interactive
  if [ $? -ne 0 ]; then
    echo -e "${RED}Failed to initialize Plannet${NC}"
    echo "Please run 'plannet init' manually to set up the configuration"
    exit 1
  fi
fi

# Test 1: Track work with description
print_header "Test 1: Track work with description"
run_test "Track work with description" "plannet track 'Implement user authentication'"
if [ $? -eq 0 ]; then
  TESTS_PASSED=$((TESTS_PASSED+1))
else
  TESTS_FAILED=$((TESTS_FAILED+1))
fi

# Test 2: Track work with ticket ID
print_header "Test 2: Track work with ticket ID"
run_test "Track work with ticket ID" "plannet track 'Fix login bug' --ticket DEV-123"
if [ $? -eq 0 ]; then
  TESTS_PASSED=$((TESTS_PASSED+1))
else
  TESTS_FAILED=$((TESTS_FAILED+1))
fi

# Test 3: Track work with tags
print_header "Test 3: Track work with tags"
run_test "Track work with tags" "plannet track 'Update documentation' --tags docs,api"
if [ $? -eq 0 ]; then
  TESTS_PASSED=$((TESTS_PASSED+1))
else
  TESTS_FAILED=$((TESTS_FAILED+1))
fi

# Test 4: Track work with all options
print_header "Test 4: Track work with all options"
run_test "Track work with all options" "plannet track 'Refactor database layer' --ticket DEV-456 --tags refactor,backend"
if [ $? -eq 0 ]; then
  TESTS_PASSED=$((TESTS_PASSED+1))
else
  TESTS_FAILED=$((TESTS_FAILED+1))
fi

# Test 5: Track work with git context
print_header "Test 5: Track work with git context"
# Create a temporary git repo for testing
TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR
git init
echo "# Test repo" > README.md
git add README.md
git commit -m "Initial commit"
git checkout -b feature/DEV-789
echo "# Updated" > README.md
git add README.md
git commit -m "Update README"
run_test "Track work with git context" "plannet track 'Update project documentation'"
if [ $? -eq 0 ]; then
  TESTS_PASSED=$((TESTS_PASSED+1))
else
  TESTS_FAILED=$((TESTS_FAILED+1))
fi
cd - > /dev/null
rm -rf $TEMP_DIR

# Test 6: List tracked work
print_header "Test 6: List tracked work"
run_test "List tracked work" "plannet list"
if [ $? -eq 0 ]; then
  TESTS_PASSED=$((TESTS_PASSED+1))
else
  TESTS_FAILED=$((TESTS_FAILED+1))
fi

# Test 7: View specific tracked work
print_header "Test 7: View specific tracked work"
# Get the ID of the first tracked work
WORK_ID=$(plannet list --json 2>/dev/null | jq -r '.[0].id' 2>/dev/null)
if [ -n "$WORK_ID" ]; then
  run_test "View specific tracked work" "plannet view $WORK_ID"
  if [ $? -eq 0 ]; then
    TESTS_PASSED=$((TESTS_PASSED+1))
  else
    TESTS_FAILED=$((TESTS_FAILED+1))
  fi
else
  echo -e "${YELLOW}No tracked work found to view${NC}"
  TESTS_FAILED=$((TESTS_FAILED+1))
fi

# Test 8: Complete tracked work
print_header "Test 8: Complete tracked work"
if [ -n "$WORK_ID" ]; then
  run_test "Complete tracked work" "plannet complete $WORK_ID"
  if [ $? -eq 0 ]; then
    TESTS_PASSED=$((TESTS_PASSED+1))
  else
    TESTS_FAILED=$((TESTS_FAILED+1))
  fi
else
  echo -e "${YELLOW}No tracked work found to complete${NC}"
  TESTS_FAILED=$((TESTS_FAILED+1))
fi

# Test 9: Export tracked work
print_header "Test 9: Export tracked work"
run_test "Export tracked work to CSV" "plannet export --format csv > tracked_work.csv 2>/dev/null"
if [ $? -eq 0 ]; then
  TESTS_PASSED=$((TESTS_PASSED+1))
else
  TESTS_FAILED=$((TESTS_FAILED+1))
fi

run_test "Export tracked work to JSON" "plannet export --format json > tracked_work.json 2>/dev/null"
if [ $? -eq 0 ]; then
  TESTS_PASSED=$((TESTS_PASSED+1))
else
  TESTS_FAILED=$((TESTS_FAILED+1))
fi

# Test 10: Track work with LLM assistance
print_header "Test 10: Track work with LLM assistance"
run_test "Track work with LLM assistance" "plannet track --llm 'I just fixed a bug in the authentication system'"
if [ $? -eq 0 ]; then
  TESTS_PASSED=$((TESTS_PASSED+1))
else
  TESTS_FAILED=$((TESTS_FAILED+1))
fi

print_header "All Tests Completed"
echo -e "${GREEN}Tests passed: $TESTS_PASSED${NC}"
echo -e "${RED}Tests failed: $TESTS_FAILED${NC}"
echo -e "Total tests: $((TESTS_PASSED+TESTS_FAILED))"

if [ $TESTS_FAILED -eq 0 ]; then
  echo -e "${GREEN}All tests passed successfully!${NC}"
  echo -e "Check the exported files (tracked_work.csv and tracked_work.json) for results."
else
  echo -e "${YELLOW}Some tests failed. Please check the error messages above.${NC}"
  echo -e "You may need to fix issues in the Plannet codebase before all tests will pass." 