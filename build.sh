
#!/bin/bash

# Build script for distributed-regional-check-agent .deb package

set -e

echo "Building CheckCle Distributed Regional Check Agent .deb package..."

# Check if required tools are installed
command -v dpkg-deb >/dev/null 2>&1 || {
    echo "Error: dpkg-deb is required but not installed."
    echo "Install with: sudo apt-get install dpkg-dev"
    exit 1
}

command -v go >/dev/null 2>&1 || {
    echo "Error: Go is required but not installed."
    exit 1
}

# Show current directory for context
echo "Building from: $(pwd)"

# Clean previous builds
echo "Cleaning previous builds..."
make clean

# Check for architecture argument
ARCH=${1:-"all"}

case $ARCH in
    amd64)
        echo "Building for AMD64 architecture..."
        make deb-amd64
        ;;
    arm64)
        echo "Building for ARM64 architecture..."
        make deb-arm64
        ;;
    all)
        echo "Building for both AMD64 and ARM64 architectures..."
        make deb-all
        ;;
    *)
        echo "Usage: $0 [amd64|arm64|all]"
        echo "  amd64 - Build only for AMD64"
        echo "  arm64 - Build only for ARM64"
        echo "  all   - Build for both architectures (default)"
        exit 1
        ;;
esac

echo ""
echo "✅ Build complete!"
echo ""
echo "📦 Generated packages:"
ls -la build/*.deb 2>/dev/null || echo "No packages found"
echo ""
echo "🚀 To install a package:"
echo "  sudo dpkg -i build/distributed-regional-check-agent_1.0.0_amd64.deb"
echo "  # or"
echo "  sudo dpkg -i build/distributed-regional-check-agent_1.0.0_arm64.deb"
echo ""
echo "📋 To install dependencies if needed:"
echo "  sudo apt-get install -f"
echo ""
echo "⚙️  To configure (REQUIRED before starting):"
echo "  sudo nano /etc/regional-check-agent/regional-check-agent.env"
echo ""
echo "🔧 To start after configuration:"
echo "  sudo systemctl enable regional-check-agent"
echo "  sudo systemctl start regional-check-agent"
echo ""
echo "📊 To check status:"
echo "  sudo systemctl status regional-check-agent"
echo ""
echo "🩺 Health check endpoint:"
echo "  curl http://localhost:8091/health"