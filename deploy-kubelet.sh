#!/bin/bash
# Quick deployment script for kubelet with allowed-numa-nodes feature
# Run on GB200 system with sudo

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Kubelet Deployment with allowed-numa-nodes ===${NC}"

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Please run as root or with sudo${NC}"
    exit 1
fi

# Backup current kubelet
BACKUP_FILE="/usr/bin/kubelet.backup.$(date +%Y%m%d-%H%M%S)"
echo -e "${YELLOW}Backing up current kubelet to ${BACKUP_FILE}${NC}"
cp /usr/bin/kubelet "$BACKUP_FILE"

# Check if new kubelet exists
NEW_KUBELET="./kubernetes/_output/local/bin/linux/arm64/kubelet"
if [ ! -f "$NEW_KUBELET" ]; then
    echo -e "${RED}New kubelet binary not found at ${NEW_KUBELET}${NC}"
    echo "Please build it first with: make WHAT=cmd/kubelet KUBE_BUILD_PLATFORMS=linux/arm64"
    exit 1
fi

# Stop kubelet
echo -e "${YELLOW}Stopping kubelet service...${NC}"
systemctl stop kubelet

# Clean state files
echo -e "${YELLOW}Cleaning state files...${NC}"
rm -f /var/lib/kubelet/cpu_manager_state
rm -f /var/lib/kubelet/memory_manager_state
echo -e "${GREEN}State files cleaned${NC}"

# Replace kubelet
echo -e "${YELLOW}Replacing kubelet binary...${NC}"
cp "$NEW_KUBELET" /usr/bin/kubelet
chmod +x /usr/bin/kubelet

# Verify version
echo -e "${GREEN}New kubelet version:${NC}"
/usr/bin/kubelet --version

# Check config file
CONFIG_FILE="/var/lib/kubelet/config.yaml"
if grep -q "allowed-numa-nodes" "$CONFIG_FILE"; then
    echo -e "${GREEN}✓ Config file contains allowed-numa-nodes option${NC}"
else
    echo -e "${YELLOW}⚠ Warning: allowed-numa-nodes not found in config${NC}"
    echo -e "${YELLOW}Please add to ${CONFIG_FILE}:${NC}"
    echo "topologyManagerPolicyOptions:"
    echo "  allowed-numa-nodes: \"0,1\""
    echo ""
    read -p "Edit config now? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        ${EDITOR:-vim} "$CONFIG_FILE"
    fi
fi

# Start kubelet
echo -e "${YELLOW}Starting kubelet service...${NC}"
systemctl start kubelet

# Wait a moment
sleep 3

# Check status
if systemctl is-active --quiet kubelet; then
    echo -e "${GREEN}✓ Kubelet started successfully${NC}"
    
    # Show recent logs
    echo -e "${GREEN}Recent kubelet logs:${NC}"
    journalctl -u kubelet -n 20 --no-pager | grep -i "numa\|topology" || echo "No NUMA/topology messages yet"
    
    echo ""
    echo -e "${GREEN}=== Deployment Complete ===${NC}"
    echo -e "Monitor logs with: ${YELLOW}sudo journalctl -u kubelet -f${NC}"
    echo -e "Check for NUMA filtering: ${YELLOW}sudo journalctl -u kubelet | grep allowed-numa-nodes${NC}"
else
    echo -e "${RED}✗ Kubelet failed to start${NC}"
    echo -e "${RED}Check logs: sudo journalctl -u kubelet -n 50 --no-pager${NC}"
    echo ""
    echo -e "${YELLOW}Rolling back...${NC}"
    cp "$BACKUP_FILE" /usr/bin/kubelet
    systemctl start kubelet
    exit 1
fi
