#!/bin/bash

set -e

echo "=========================================="
echo "Installing Linting Tools"
echo "=========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go first."
    exit 1
fi

# Install golangci-lint
echo -e "${YELLOW}Installing golangci-lint...${NC}"
if command -v golangci-lint &> /dev/null; then
    echo "golangci-lint is already installed: $(golangci-lint --version)"
else
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.61.0
    echo -e "${GREEN}✓ golangci-lint installed${NC}"
fi
echo ""

# Install markdownlint-cli
echo -e "${YELLOW}Installing markdownlint-cli...${NC}"
if command -v markdownlint &> /dev/null; then
    echo "markdownlint is already installed: $(markdownlint --version)"
else
    if command -v npm &> /dev/null; then
        npm install -g markdownlint-cli
        echo -e "${GREEN}✓ markdownlint-cli installed${NC}"
    else
        echo "Warning: npm not found. Skipping markdownlint installation."
        echo "Install Node.js and npm to use markdownlint."
    fi
fi
echo ""

# Install yamllint
echo -e "${YELLOW}Installing yamllint...${NC}"
if command -v yamllint &> /dev/null; then
    echo "yamllint is already installed: $(yamllint --version)"
else
    if command -v pip3 &> /dev/null; then
        pip3 install --user yamllint
        echo -e "${GREEN}✓ yamllint installed${NC}"
    elif command -v pip &> /dev/null; then
        pip install --user yamllint
        echo -e "${GREEN}✓ yamllint installed${NC}"
    else
        echo "Warning: pip/pip3 not found. Skipping yamllint installation."
        echo "Install Python and pip to use yamllint."
    fi
fi
echo ""

echo "=========================================="
echo -e "${GREEN}Installation Complete!${NC}"
echo "=========================================="
echo ""
echo "Verify installations:"
echo "  golangci-lint version: $(golangci-lint --version 2>/dev/null || echo 'not installed')"
echo "  markdownlint version:  $(markdownlint --version 2>/dev/null || echo 'not installed')"
echo "  yamllint version:      $(yamllint --version 2>/dev/null || echo 'not installed')"
echo ""
echo "Run 'make lint' to lint all code"
echo "Run 'make help' to see all available commands"
