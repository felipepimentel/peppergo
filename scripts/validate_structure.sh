#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Required directories
REQUIRED_DIRS=(
    "cmd/peppergo"
    "internal/agent"
    "internal/capability"
    "internal/provider"
    "internal/tool"
    "internal/config"
    "pkg/types"
    "pkg/log"
    "pkg/errors"
    "api/proto"
    "assets/agents"
    "assets/prompts"
    "scripts"
    "test"
    "docs"
    "docs/agents"
)

# Required files
REQUIRED_FILES=(
    "go.mod"
    "go.sum"
    "README.md"
    "docs/status.md"
    ".golangci.yml"
    "Makefile"
)

# Counter for issues
ISSUES=0

# Function to check directory
check_directory() {
    if [ ! -d "$1" ]; then
        echo -e "${RED}Missing directory: $1${NC}"
        ISSUES=$((ISSUES + 1))
    else
        echo -e "${GREEN}✓ Directory exists: $1${NC}"
    fi
}

# Function to check file
check_file() {
    if [ ! -f "$1" ]; then
        echo -e "${RED}Missing file: $1${NC}"
        ISSUES=$((ISSUES + 1))
    else
        echo -e "${GREEN}✓ File exists: $1${NC}"
    fi
}

# Print header
echo -e "${YELLOW}Validating project structure...${NC}\n"

# Check directories
echo -e "${YELLOW}Checking required directories...${NC}"
for dir in "${REQUIRED_DIRS[@]}"; do
    check_directory "$dir"
done

echo -e "\n${YELLOW}Checking required files...${NC}"
for file in "${REQUIRED_FILES[@]}"; do
    check_file "$file"
done

# Check Go module
if [ -f "go.mod" ]; then
    echo -e "\n${YELLOW}Checking Go module...${NC}"
    if ! go mod verify > /dev/null 2>&1; then
        echo -e "${RED}Go module verification failed${NC}"
        ISSUES=$((ISSUES + 1))
    else
        echo -e "${GREEN}✓ Go module verified${NC}"
    fi
fi

# Check if golangci-lint is installed
echo -e "\n${YELLOW}Checking golangci-lint...${NC}"
if ! command -v golangci-lint > /dev/null 2>&1; then
    echo -e "${RED}golangci-lint is not installed${NC}"
    ISSUES=$((ISSUES + 1))
else
    echo -e "${GREEN}✓ golangci-lint is installed${NC}"
fi

# Print summary
echo -e "\n${YELLOW}Validation complete${NC}"
if [ $ISSUES -eq 0 ]; then
    echo -e "${GREEN}✓ All checks passed${NC}"
    exit 0
else
    echo -e "${RED}✗ Found $ISSUES issue(s)${NC}"
    exit 1
fi 