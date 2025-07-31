
#!/bin/bash

# CheckCle Regional Monitoring Agent - Universal Installation Script
# This script automatically detects system architecture and installs the appropriate package
# Usage: curl -fsSL https://raw.githubusercontent.com/operacle/checkcle/main/scripts/install-regional-agent.sh | sudo bash -s -- [options]

set -e

# Default values
REGION_NAME=""
AGENT_ID=""
AGENT_IP_ADDRESS=""
AGENT_TOKEN=""
POCKETBASE_URL=""
BASE_PACKAGE_URL="https://github.com/operacle/Distributed-Regional-Monitoring/releases/download/V1.0.0"
PACKAGE_VERSION="1.0.0"
SERVICE_NAME="regional-check-agent"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --region-name=NAME        Set the region name (required)"
    echo "  --agent-id=ID            Set the agent ID (required)"
    echo "  --agent-ip=IP            Set the agent IP address (required)"
    echo "  --token=TOKEN            Set the authentication token (required)"
    echo "  --pocketbase-url=URL     Set the PocketBase API URL (required)"
    echo "  --package-version=VER    Set package version (default: $PACKAGE_VERSION)"
    echo "  --help                   Show this help message"
    echo ""
    echo "Example:"
    echo "  $0 --region-name=\"us-east-1\" --agent-id=\"agent_abc123\" \\"
    echo "     --agent-ip=\"192.168.1.100\" --token=\"your-token\" \\"
    echo "     --pocketbase-url=\"https://your-pb.com\""
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --region-name=*)
            REGION_NAME="${1#*=}"
            shift
            ;;
        --agent-id=*)
            AGENT_ID="${1#*=}"
            shift
            ;;
        --agent-ip=*)
            AGENT_IP_ADDRESS="${1#*=}"
            shift
            ;;
        --token=*)
            AGENT_TOKEN="${1#*=}"
            shift
            ;;
        --pocketbase-url=*)
            POCKETBASE_URL="${1#*=}"
            shift
            ;;
        --package-version=*)
            PACKAGE_VERSION="${1#*=}"
            shift
            ;;
        --help)
            show_usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Validate required parameters
if [[ -z "$REGION_NAME" || -z "$AGENT_ID" || -z "$AGENT_IP_ADDRESS" || -z "$AGENT_TOKEN" || -z "$POCKETBASE_URL" ]]; then
    log_error "Missing required parameters"
    echo ""
    show_usage
    exit 1
fi

echo "============================================="
echo "  CheckCle Regional Monitoring Agent"
echo "  Universal Installation"
echo "============================================="

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   log_error "This script must be run as root (use sudo)" 
   echo "   Usage: sudo bash $0 [options]"
   exit 1
fi

# Detect operating system
OS_TYPE=$(uname -s | tr '[:upper:]' '[:lower:]')

# Check if it's a supported OS
if [[ "$OS_TYPE" != "linux" ]]; then
    log_error "Unsupported operating system: $OS_TYPE"
    log_info "This installer only supports Linux systems"
    exit 1
fi

# Detect architecture
ARCH=$(uname -m)

# Map architecture to package architecture
case $ARCH in
    x86_64|amd64)
        PKG_ARCH="amd64"
        ;;
    aarch64|arm64)
        PKG_ARCH="arm64"
        ;;
    armv7l|armv6l)
        PKG_ARCH="arm64"
        log_warning "ARM 32-bit detected, using ARM64 package (may require compatibility layer)"
        ;;
    *)
        log_error "Unsupported architecture: $ARCH"
        log_info "Supported architectures: x86_64 (amd64), aarch64 (arm64)"
        exit 1
        ;;
esac

log_success "System: $OS_TYPE ($ARCH -> $PKG_ARCH)"

# Construct package URLs and names
PACKAGE_URL="$BASE_PACKAGE_URL/distributed-regional-check-agent_${PACKAGE_VERSION}_${PKG_ARCH}.deb"
PACKAGE_NAME="distributed-regional-check-agent_${PACKAGE_VERSION}_${PKG_ARCH}.deb"

# Check for required tools
MISSING_TOOLS=()

if ! command -v wget >/dev/null 2>&1 && ! command -v curl >/dev/null 2>&1; then
    MISSING_TOOLS+=("wget or curl")
fi

if ! command -v dpkg >/dev/null 2>&1; then
    MISSING_TOOLS+=("dpkg")
fi

if ! command -v systemctl >/dev/null 2>&1; then
    MISSING_TOOLS+=("systemd")
fi

if [ ${#MISSING_TOOLS[@]} -ne 0 ]; then
    log_error "Missing required tools: ${MISSING_TOOLS[*]}"
    log_info "On Debian/Ubuntu: sudo apt-get update && sudo apt-get install wget curl"
    exit 1
fi

# Create temporary directory
TEMP_DIR=$(mktemp -d)

# Download the .deb package
log_info "Downloading package for $PKG_ARCH..."
cd "$TEMP_DIR"

# Test if package exists first
if command -v curl >/dev/null 2>&1; then
    HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -I "$PACKAGE_URL")
    if [ "$HTTP_STATUS" != "200" ] && [ "$HTTP_STATUS" != "302" ]; then
        log_error "Package not found at $PACKAGE_URL (HTTP $HTTP_STATUS)"
        log_info "Check: https://github.com/operacle/Distributed-Regional-Monitoring/releases"
        rm -rf "$TEMP_DIR"
        exit 1
    fi
fi

# Try wget first, then curl as fallback
DOWNLOAD_SUCCESS=false

if command -v wget >/dev/null 2>&1; then
    if wget -q --show-progress --timeout=30 --tries=3 "$PACKAGE_URL" -O "$PACKAGE_NAME"; then
        DOWNLOAD_SUCCESS=true
    fi
fi

if [ "$DOWNLOAD_SUCCESS" = false ] && command -v curl >/dev/null 2>&1; then
    if curl -L --connect-timeout 30 --max-time 300 --retry 3 --retry-delay 2 -o "$PACKAGE_NAME" "$PACKAGE_URL" --progress-bar; then
        DOWNLOAD_SUCCESS=true
    fi
fi

if [ "$DOWNLOAD_SUCCESS" = false ]; then
    log_error "Failed to download package from $PACKAGE_URL"
    log_info "Check internet connection and package availability"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Verify download was successful
if [ ! -f "$PACKAGE_NAME" ] || [ ! -s "$PACKAGE_NAME" ]; then
    log_error "Downloaded package is empty or missing"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Verify package integrity
if dpkg-deb --info "$PACKAGE_NAME" > /dev/null 2>&1; then
    log_success "Package verified"
else
    log_error "Package verification failed - corrupted download"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Install the package
log_info "Installing package..."
if dpkg -i "$PACKAGE_NAME" 2>/dev/null; then
    log_success "Package installed"
else
    log_warning "Fixing dependencies..."
    if apt-get update && apt-get install -f -y; then
        log_success "Package installed with dependencies"
    else
        log_error "Failed to install package"
        log_info "Try: sudo apt-get update && sudo apt-get install -f"
        rm -rf "$TEMP_DIR"
        exit 1
    fi
fi

# Configure the agent
log_info "Configuring agent..."

# Ensure configuration directory exists
mkdir -p /etc/regional-check-agent

# Create the environment configuration file
cat > /etc/regional-check-agent/regional-check-agent.env << EOF
# Distributed Regional Check Agent Configuration
# Auto-generated on $(date)

# Server Configuration
PORT=8091

# Operation defaults
DEFAULT_COUNT=4
DEFAULT_TIMEOUT=10s
MAX_COUNT=20
MAX_TIMEOUT=30s

# Logging
ENABLE_LOGGING=true

# PocketBase integration
POCKETBASE_ENABLED=true
POCKETBASE_URL=$POCKETBASE_URL

# Regional Agent Configuration
REGION_NAME=$REGION_NAME
AGENT_ID=$AGENT_ID
AGENT_IP_ADDRESS=$AGENT_IP_ADDRESS
AGENT_TOKEN=$AGENT_TOKEN

# Monitoring configuration
CHECK_INTERVAL=30s
MAX_RETRIES=3
REQUEST_TIMEOUT=10s
EOF

# Set proper permissions
if id "regional-check-agent" &>/dev/null; then
    chown root:regional-check-agent /etc/regional-check-agent/regional-check-agent.env
    chmod 640 /etc/regional-check-agent/regional-check-agent.env
else
    chown root:root /etc/regional-check-agent/regional-check-agent.env
    chmod 600 /etc/regional-check-agent/regional-check-agent.env
fi

log_success "Configuration complete"

# Enable and start the service
log_info "Starting service..."

# Reload systemd daemon
systemctl daemon-reload

# Enable the service for auto-start
if systemctl enable $SERVICE_NAME; then
    log_success "Service enabled"
else
    log_error "Failed to enable service"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Start the service
if systemctl start $SERVICE_NAME; then
    log_success "Service started"
else
    log_error "Failed to start service"
    log_info "Check logs: sudo journalctl -u $SERVICE_NAME -f"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Wait a moment for service to initialize
sleep 5

# Test health endpoint
HEALTH_CHECK_ATTEMPTS=3
for i in $(seq 1 $HEALTH_CHECK_ATTEMPTS); do
    if curl -s -f --connect-timeout 5 http://localhost:8091/health > /dev/null; then
        log_success "Health endpoint responding"
        break
    else
        if [ $i -lt $HEALTH_CHECK_ATTEMPTS ]; then
            sleep 2
        else
            log_warning "Health endpoint not responding (service may still be starting)"
        fi
    fi
done

# Cleanup
rm -rf "$TEMP_DIR"

echo ""
echo "============================================="
echo "  Installation Complete!"
echo "============================================="
echo ""
log_success "CheckCle Regional Monitoring Agent installed successfully"
echo ""
echo "Agent Details:"
echo "  Region: $REGION_NAME"
echo "  Agent ID: $AGENT_ID"
echo "  Status: $(systemctl is-active $SERVICE_NAME 2>/dev/null || echo 'unknown')"
echo "  Health: http://localhost:8091/health"
echo ""
echo "Management Commands:"
echo "  Status: sudo systemctl status $SERVICE_NAME"
echo "  Logs: sudo journalctl -u $SERVICE_NAME -f"
echo "  Restart: sudo systemctl restart $SERVICE_NAME"
echo ""
log_success "Agent is now monitoring and reporting to your dashboard!"