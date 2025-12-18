#!/bin/bash

# Archon Uninstallation Script
# This script removes the Archon TUI binary

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}   Archon TUI Uninstallation Script${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

# Check both possible install locations
SYSTEM_BIN="/usr/local/bin/archon"
USER_BIN="$HOME/.local/bin/archon"

FOUND=false

# Check system-wide installation
if [ -f "$SYSTEM_BIN" ]; then
    echo "Found system-wide installation at $SYSTEM_BIN"
    if [ "$EUID" -eq 0 ]; then
        rm -f "$SYSTEM_BIN"
        echo -e "${GREEN}✓ Removed $SYSTEM_BIN${NC}"
        FOUND=true
    else
        echo -e "${RED}✗ Need root privileges to remove system-wide installation${NC}"
        echo -e "${YELLOW}Run: sudo ./uninstall.sh${NC}"
        exit 1
    fi
fi

# Check user installation
if [ -f "$USER_BIN" ]; then
    echo "Found user installation at $USER_BIN"
    rm -f "$USER_BIN"
    echo -e "${GREEN}✓ Removed $USER_BIN${NC}"
    FOUND=true
fi

if [ "$FOUND" = false ]; then
    echo -e "${YELLOW}No Archon installation found${NC}"
    echo "Checked:"
    echo "  - $SYSTEM_BIN"
    echo "  - $USER_BIN"
    exit 0
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}   Uninstallation Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${YELLOW}Note: Configuration files in ~/.config/archon were NOT removed${NC}"
echo "To remove config files, run:"
echo -e "${RED}  rm -rf ~/.config/archon${NC}"
echo ""
