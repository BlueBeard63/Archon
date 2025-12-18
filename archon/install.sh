#!/bin/bash

# Archon Installation Script
# This script builds and installs the Archon TUI to make it accessible from anywhere

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}   Archon TUI Installation Script${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Detect install location
if [ "$EUID" -eq 0 ]; then
    # Running as root, install system-wide
    INSTALL_DIR="/usr/local/bin"
    echo -e "${YELLOW}Running as root. Installing system-wide to ${INSTALL_DIR}${NC}"
else
    # Running as user, install to user bin
    INSTALL_DIR="$HOME/.local/bin"
    echo -e "${YELLOW}Running as user. Installing to ${INSTALL_DIR}${NC}"

    # Create directory if it doesn't exist
    mkdir -p "$INSTALL_DIR"

    # Check if ~/.local/bin is in PATH
    if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
        echo -e "${YELLOW}Warning: $HOME/.local/bin is not in your PATH${NC}"
        echo -e "${YELLOW}Add this line to your ~/.bashrc or ~/.zshrc:${NC}"
        echo -e "${GREEN}export PATH=\"\$HOME/.local/bin:\$PATH\"${NC}"
        echo ""
    fi
fi

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "Building Archon..."
cd "$SCRIPT_DIR"

# Build the binary
if go build -o archon; then
    echo -e "${GREEN}✓ Build successful${NC}"
else
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi

echo ""
echo "Installing to ${INSTALL_DIR}..."

# Install the binary
if cp archon "$INSTALL_DIR/archon"; then
    echo -e "${GREEN}✓ Installed to ${INSTALL_DIR}/archon${NC}"
else
    echo -e "${RED}✗ Installation failed${NC}"
    exit 1
fi

# Make sure it's executable
chmod +x "$INSTALL_DIR/archon"

# Create config directory if it doesn't exist
CONFIG_DIR="$HOME/.config/archon"
if [ ! -d "$CONFIG_DIR" ]; then
    echo ""
    echo "Creating config directory..."
    mkdir -p "$CONFIG_DIR"
    echo -e "${GREEN}✓ Created ${CONFIG_DIR}${NC}"
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}   Installation Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "You can now run Archon from anywhere by typing:"
echo -e "${GREEN}  archon${NC}"
echo ""

# Check if command is immediately available
if command -v archon &> /dev/null; then
    echo -e "${GREEN}✓ 'archon' command is ready to use${NC}"
else
    if [ "$INSTALL_DIR" = "$HOME/.local/bin" ]; then
        echo -e "${YELLOW}Note: You may need to restart your shell or run:${NC}"
        echo -e "${GREEN}  source ~/.bashrc${NC}"
        echo -e "${GREEN}or${NC}"
        echo -e "${GREEN}  source ~/.zshrc${NC}"
    fi
fi

echo ""
