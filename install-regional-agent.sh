
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
    echo "❌ Error: Missing required parameters"
    echo ""
    show_usage
    exit 1
fi

echo "🚀 CheckCle Regional Monitoring Agent - Universal Installation"
echo "=============================================================="
echo ""

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo "❌ This script must be run as root (use sudo)" 
   echo "   Usage: sudo bash $0 [options]"
   exit 1
fi

# Detect operating system
echo "🔍 Detecting system information..."
OS_TYPE=$(uname -s | tr '[:upper:]' '[:lower:]')
echo "   Operating System: $OS_TYPE"

# Check if it's a supported OS
if [[ "$OS_TYPE" != "linux" ]]; then
    echo "❌ Unsupported operating system: $OS_TYPE"
    echo "   This installer only supports Linux systems"
    exit 1
fi

# Detect architecture
ARCH=$(uname -m)
echo "   Hardware Architecture: $ARCH"

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
        echo "⚠️  ARM 32-bit detected, using ARM64 package (may require compatibility layer)"
        ;;
    *)
        echo "❌ Unsupported architecture: $ARCH"
        echo "   Supported architectures: x86_64 (amd64), aarch64 (arm64)"
        echo "   Please contact support for your architecture: $ARCH"
        exit 1
        ;;
esac

echo "   Package Architecture: $PKG_ARCH"

# Construct package URLs and names
PACKAGE_URL="$BASE_PACKAGE_URL/distributed-regional-check-agent_${PACKAGE_VERSION}_${PKG_ARCH}.deb"
PACKAGE_NAME="distributed-regional-check-agent_${PACKAGE_VERSION}_${PKG_ARCH}.deb"

echo ""
echo "📋 Installation Configuration:"
echo "   Region Name: $REGION_NAME"
echo "   Agent ID: $AGENT_ID"
echo "   Agent IP: $AGENT_IP_ADDRESS"
echo "   Package Architecture: $PKG_ARCH"
echo "   Package URL: $PACKAGE_URL"
echo ""

# Check for required tools
echo "🔧 Checking system requirements..."
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
    echo "❌ Missing required tools: ${MISSING_TOOLS[*]}"
    echo "   Please install missing tools and try again"
    echo "   On Debian/Ubuntu: sudo apt-get update && sudo apt-get install wget curl"
    exit 1
fi

echo "✅ System requirements satisfied"

# Create temporary directory
TEMP_DIR=$(mktemp -d)
echo "📁 Created temporary directory: $TEMP_DIR"

# Download the .deb package
echo ""
echo "📥 Downloading Regional Monitoring Agent package for $PKG_ARCH..."
cd "$TEMP_DIR"

# Test if package exists first - Accept both 200 and 302 (redirect) as success
echo "🔍 Checking package availability..."
if command -v curl >/dev/null 2>&1; then
    HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -I "$PACKAGE_URL")
    if [ "$HTTP_STATUS" != "200" ] && [ "$HTTP_STATUS" != "302" ]; then
        echo "❌ Package not found at $PACKAGE_URL (HTTP $HTTP_STATUS)"
        echo "   Available packages should be:"
        echo "   - distributed-regional-check-agent_${PACKAGE_VERSION}_amd64.deb"
        echo "   - distributed-regional-check-agent_${PACKAGE_VERSION}_arm64.deb"
        echo ""
        echo "   Please check the GitHub releases page:"
        echo "   https://github.com/operacle/Distributed-Regional-Monitoring/releases"
        rm -rf "$TEMP_DIR"
        exit 1
    fi
    echo "✅ Package found (HTTP $HTTP_STATUS), proceeding with download..."
fi

# Try wget first, then curl as fallback
DOWNLOAD_SUCCESS=false

if command -v wget >/dev/null 2>&1; then
    echo "📥 Downloading with wget..."
    if wget -q --show-progress --timeout=30 --tries=3 "$PACKAGE_URL" -O "$PACKAGE_NAME"; then
        DOWNLOAD_SUCCESS=true
        echo "✅ Package downloaded successfully using wget"
    fi
fi

if [ "$DOWNLOAD_SUCCESS" = false ] && command -v curl >/dev/null 2>&1; then
    echo "📥 Downloading with curl..."
    if curl -L --connect-timeout 30 --max-time 300 --retry 3 --retry-delay 2 -o "$PACKAGE_NAME" "$PACKAGE_URL" --progress-bar; then
        DOWNLOAD_SUCCESS=true
        echo "✅ Package downloaded successfully using curl"
    fi
fi

if [ "$DOWNLOAD_SUCCESS" = false ]; then
    echo "❌ Failed to download package from $PACKAGE_URL"
    echo "   Please check:"
    echo "   - Internet connection"
    echo "   - Package availability for $PKG_ARCH architecture"
    echo "   - GitHub repository access: https://github.com/operacle/Distributed-Regional-Monitoring/releases"
    echo "   - Firewall/proxy settings"
    echo ""
    echo "   Available packages should be:"
    echo "   - distributed-regional-check-agent_${PACKAGE_VERSION}_amd64.deb"
    echo "   - distributed-regional-check-agent_${PACKAGE_VERSION}_arm64.deb"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Verify download was successful
if [ ! -f "$PACKAGE_NAME" ] || [ ! -s "$PACKAGE_NAME" ]; then
    echo "❌ Downloaded package is empty or missing"
    echo "   File size: $(ls -lh "$PACKAGE_NAME" 2>/dev/null | awk '{print $5}' || echo 'file not found')"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Verify package integrity
echo ""
echo "🔍 Verifying package..."
if dpkg-deb --info "$PACKAGE_NAME" > /dev/null 2>&1; then
    echo "✅ Package verification successful"
    
    # Show package info
    echo "📦 Package Information:"
    dpkg-deb --field "$PACKAGE_NAME" Package Version Architecture Description | head -4
else
    echo "❌ Package verification failed - corrupted download"
    echo "   File size: $(ls -lh "$PACKAGE_NAME" | awk '{print $5}')"
    echo "   Try downloading manually from: $PACKAGE_URL"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Install the package
echo ""
echo "📦 Installing Regional Monitoring Agent package..."
if dpkg -i "$PACKAGE_NAME" 2>/dev/null; then
    echo "✅ Package installed successfully"
else
    echo "⚠️  Package installation had dependency issues, attempting to fix..."
    if apt-get update && apt-get install -f -y; then
        echo "✅ Dependencies fixed and package installed successfully"
    else
        echo "❌ Failed to install package and fix dependencies"
        echo "   This might be due to:"
        echo "   - Missing system dependencies"
        echo "   - Architecture compatibility issues"
        echo "   - Package conflicts"
        echo "   - Insufficient disk space"
        echo ""
        echo "   Manual resolution:"
        echo "   1. Run: sudo apt-get update"
        echo "   2. Run: sudo apt-get install -f"
        echo "   3. Retry installation: sudo dpkg -i $PACKAGE_NAME"
        rm -rf "$TEMP_DIR"
        exit 1
    fi
fi

# Configure the agent
echo ""
echo "⚙️  Configuring Regional Monitoring Agent..."

# Ensure configuration directory exists
mkdir -p /etc/regional-check-agent

# Create the environment configuration file
cat > /etc/regional-check-agent/regional-check-agent.env << EOF
# Distributed Regional Check Agent Configuration
# Auto-generated on $(date)
# Architecture: $PKG_ARCH

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

# Regional Agent Configuration - Auto-configured
REGION_NAME=$REGION_NAME
AGENT_ID=$AGENT_ID
AGENT_IP_ADDRESS=$AGENT_IP_ADDRESS
AGENT_TOKEN=$AGENT_TOKEN

# Monitoring configuration
CHECK_INTERVAL=30s
MAX_RETRIES=3
REQUEST_TIMEOUT=10s
EOF

echo "✅ Configuration file created at /etc/regional-check-agent/regional-check-agent.env"

# Set proper permissions
if id "regional-check-agent" &>/dev/null; then
    chown root:regional-check-agent /etc/regional-check-agent/regional-check-agent.env
    chmod 640 /etc/regional-check-agent/regional-check-agent.env
    echo "✅ Configuration file permissions set"
else
    echo "⚠️  regional-check-agent user not found, using root permissions"
    chown root:root /etc/regional-check-agent/regional-check-agent.env
    chmod 600 /etc/regional-check-agent/regional-check-agent.env
fi

# Enable and start the service
echo ""
echo "🔧 Starting Regional Monitoring Agent service..."

# Reload systemd daemon
systemctl daemon-reload

# Enable the service for auto-start
if systemctl enable $SERVICE_NAME; then
    echo "✅ Service enabled for auto-start"
else
    echo "❌ Failed to enable service"
    echo "   Check systemd configuration"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Start the service
if systemctl start $SERVICE_NAME; then
    echo "✅ Service started successfully"
else
    echo "❌ Failed to start service"
    echo "   Common issues:"
    echo "   - Configuration errors"
    echo "   - Port 8091 already in use"
    echo "   - Permission issues"
    echo ""
    echo "   Troubleshooting:"
    echo "   - Check logs: sudo journalctl -u $SERVICE_NAME -f"
    echo "   - Check config: sudo nano /etc/regional-check-agent/regional-check-agent.env"
    echo "   - Manual start: sudo systemctl start $SERVICE_NAME"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Wait a moment for service to initialize
echo "⏳ Waiting for service to initialize..."
sleep 5

# Check service status
echo ""
echo "📊 Service Status:"
systemctl status $SERVICE_NAME --no-pager -l --lines=5

# Test health endpoint
echo ""
echo "🩺 Testing agent health endpoint..."
HEALTH_CHECK_ATTEMPTS=3
HEALTH_CHECK_DELAY=2

for i in $(seq 1 $HEALTH_CHECK_ATTEMPTS); do
    if curl -s -f --connect-timeout 5 http://localhost:8091/health > /dev/null; then
        echo "✅ Agent health endpoint is responding"
        HEALTH_OK=true
        break
    else
        if [ $i -lt $HEALTH_CHECK_ATTEMPTS ]; then
            echo "⏳ Health check attempt $i/$HEALTH_CHECK_ATTEMPTS failed, retrying in ${HEALTH_CHECK_DELAY}s..."
            sleep $HEALTH_CHECK_DELAY
        else
            echo "⚠️  Agent health endpoint not responding after $HEALTH_CHECK_ATTEMPTS attempts"
            echo "   This may be normal if the service is still starting up"
            echo "   Check status later with: curl http://localhost:8091/health"
        fi
    fi
done

# Cleanup
rm -rf "$TEMP_DIR"
echo ""
echo "🎉 Regional Monitoring Agent Installation Complete!"
echo ""
echo "📋 Installation Summary:"
echo "   Agent ID: $AGENT_ID"
echo "   Region: $REGION_NAME"
echo "   Architecture: $PKG_ARCH ($ARCH)"
echo "   Status: $(systemctl is-active $SERVICE_NAME 2>/dev/null || echo 'unknown')"
echo "   Health URL: http://localhost:8091/health"
echo "   Service endpoint: http://localhost:8091/operation"
echo "   Config file: /etc/regional-check-agent/regional-check-agent.env"
echo ""
echo "📝 Useful commands:"
echo "   Check status: sudo systemctl status $SERVICE_NAME"
echo "   View logs: sudo journalctl -u $SERVICE_NAME -f"
echo "   Restart: sudo systemctl restart $SERVICE_NAME"
echo "   Stop: sudo systemctl stop $SERVICE_NAME"
echo "   Health check: curl http://localhost:8091/health"
echo ""
echo "🔧 Troubleshooting:"
echo "   If the service fails to start:"
echo "   1. Check logs: sudo journalctl -u $SERVICE_NAME -n 50"
echo "   2. Verify config: cat /etc/regional-check-agent/regional-check-agent.env"
echo "   3. Test connectivity: ping $(echo $POCKETBASE_URL | sed 's|https\?://||' | sed 's|/.*||')"
echo "   4. Check port availability: sudo netstat -tlnp | grep 8091"
echo ""
echo "✨ The agent is now monitoring and reporting to your CheckCle dashboard!"