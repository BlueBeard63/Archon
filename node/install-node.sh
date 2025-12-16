#!/bin/bash
# Archon Node Installation Script
# This script installs the archon-node server on a VPS

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}╔════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║   Archon Node Server Installation         ║${NC}"
echo -e "${GREEN}╔════════════════════════════════════════════╗${NC}"
echo

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run as root${NC}"
    echo "Please run: sudo $0"
    exit 1
fi

# Check if archon-node binary exists
if [ ! -f "./archon-node" ]; then
    echo -e "${RED}Error: archon-node binary not found in current directory${NC}"
    echo "Please build or download the archon-node binary first"
    exit 1
fi

# Check if config file exists
if [ ! -f "./node-config.toml" ]; then
    echo -e "${YELLOW}Warning: node-config.toml not found in current directory${NC}"
    echo "You'll need to create it manually at /etc/archon/node-config.toml"
    echo
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

echo -e "${GREEN}[1/6] Creating directories...${NC}"
mkdir -p /etc/archon
mkdir -p /var/lib/archon
mkdir -p /etc/archon/ssl
echo "  ✓ Created /etc/archon"
echo "  ✓ Created /var/lib/archon"
echo "  ✓ Created /etc/archon/ssl"
echo

echo -e "${GREEN}[2/6] Installing archon-node binary...${NC}"
cp ./archon-node /usr/local/bin/archon-node
chmod +x /usr/local/bin/archon-node
echo "  ✓ Installed to /usr/local/bin/archon-node"
echo

echo -e "${GREEN}[3/6] Installing configuration...${NC}"
if [ -f "./node-config.toml" ]; then
    cp ./node-config.toml /etc/archon/node-config.toml
    chmod 600 /etc/archon/node-config.toml
    echo "  ✓ Installed config to /etc/archon/node-config.toml"
else
    echo "  ⚠ Skipping config installation (file not found)"
fi
echo

echo -e "${GREEN}[4/6] Creating systemd service...${NC}"
if [ -f "./archon-node.service" ]; then
    cp ./archon-node.service /etc/systemd/system/archon-node.service
else
    cat > /etc/systemd/system/archon-node.service <<'EOF'
[Unit]
Description=Archon Node Server - Docker Site Management
Documentation=https://github.com/BlueBeard63/archon
After=network-online.target docker.service
Wants=network-online.target
Requires=docker.service

[Service]
Type=simple
User=root
Group=root
ExecStart=/usr/local/bin/archon-node --config /etc/archon/node-config.toml
Restart=on-failure
RestartSec=5s
LimitNOFILE=65536
LimitNPROC=4096
Environment="PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
StandardOutput=journal
StandardError=journal
SyslogIdentifier=archon-node
WorkingDirectory=/var/lib/archon

[Install]
WantedBy=multi-user.target
EOF
fi
chmod 644 /etc/systemd/system/archon-node.service
echo "  ✓ Created systemd service"
echo

echo -e "${GREEN}[5/6] Enabling service...${NC}"
systemctl daemon-reload
systemctl enable archon-node
echo "  ✓ Service enabled"
echo

echo -e "${GREEN}[6/6] Starting service...${NC}"
if systemctl start archon-node; then
    echo "  ✓ Service started successfully"
else
    echo -e "${RED}  ✗ Failed to start service${NC}"
    echo "  Check logs with: journalctl -u archon-node -f"
    exit 1
fi
echo

echo -e "${GREEN}╔════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║   Installation Complete!                   ║${NC}"
echo -e "${GREEN}╔════════════════════════════════════════════╗${NC}"
echo

systemctl status archon-node --no-pager

echo
echo -e "${YELLOW}Useful commands:${NC}"
echo "  View logs:       journalctl -u archon-node -f"
echo "  Stop service:    systemctl stop archon-node"
echo "  Restart service: systemctl restart archon-node"
echo "  Check status:    systemctl status archon-node"
echo "  Disable service: systemctl disable archon-node"
echo
echo -e "${GREEN}The archon-node server is now running!${NC}"
