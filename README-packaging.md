
# Distributed Regional Check Agent - Packaging Guide

This directory contains the packaging infrastructure for building the Distributed Regional Check Agent as .deb packages for Linux distributions.

## Quick Start

```bash
# Build for all architectures (AMD64 and ARM64)
./build.sh

# Build for specific architecture
./build.sh amd64
./build.sh arm64

# Install the package
sudo dpkg -i build/distributed-regional-check-agent_1.0.0_amd64.deb
```

## Package Structure

```
packaging/
├── control              # Debian package metadata
├── postinst            # Post-installation script
├── prerm               # Pre-removal script
├── postrm              # Post-removal script
├── regional-check-agent.service  # Systemd service file
└── regional-check-agent.conf     # Default configuration template
```

## Build System

### Make Targets

- `make build` - Build binary for current architecture
- `make build-amd64` - Build binary for AMD64
- `make build-arm64` - Build binary for ARM64
- `make deb` - Create .deb package for current architecture
- `make deb-amd64` - Create .deb package for AMD64
- `make deb-arm64` - Create .deb package for ARM64
- `make deb-all` - Create .deb packages for both architectures
- `make clean` - Clean build artifacts
- `make install` - Install the package

### Build Script

The `build.sh` script provides a convenient wrapper:

```bash
./build.sh           # Build for all architectures
./build.sh amd64     # Build only for AMD64
./build.sh arm64     # Build only for ARM64
```

## Installation

### System Requirements

- Linux distribution with systemd
- libc6 (>= 2.17)
- ca-certificates

### Installation Steps

1. **Install the package:**
   ```bash
   sudo dpkg -i distributed-regional-check-agent_1.0.0_amd64.deb
   ```

2. **Configure the agent (REQUIRED):**
   ```bash
   sudo nano /etc/regional-check-agent/regional-check-agent.env
   ```

   Set these required variables:
   ```bash
   REGION_NAME=your-region-name
   AGENT_ID=your-unique-agent-id
   AGENT_IP_ADDRESS=your-agent-ip
   POCKETBASE_URL=http://your-pocketbase-server:8090
   ```

3. **Enable and start the service:**
   ```bash
   sudo systemctl enable regional-check-agent
   sudo systemctl start regional-check-agent
   ```

4. **Verify the installation:**
   ```bash
   sudo systemctl status regional-check-agent
   curl http://localhost:8091/health
   ```

## Configuration

### Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `REGION_NAME` | Name of the monitoring region | Yes | - |
| `AGENT_ID` | Unique identifier for this agent | Yes | - |
| `AGENT_IP_ADDRESS` | IP address of this agent | Yes | - |
| `POCKETBASE_URL` | PocketBase server URL | Yes | `http://localhost:8090` |
| `PORT` | Port for the agent API | No | `8091` |
| `CHECK_INTERVAL` | Monitoring check interval | No | `30s` |

### Configuration Files

- **Service config**: `/etc/regional-check-agent/regional-check-agent.env`
- **Systemd service**: `/etc/systemd/system/regional-check-agent.service`
- **Working directory**: `/var/lib/regional-check-agent`
- **Log directory**: `/var/log/regional-check-agent`

## Service Management

### Common Commands

```bash
# Start the service
sudo systemctl start regional-check-agent

# Stop the service
sudo systemctl stop regional-check-agent

# Restart the service
sudo systemctl restart regional-check-agent

# Check service status
sudo systemctl status regional-check-agent

# View logs
sudo journalctl -u regional-check-agent -f

# Enable auto-start
sudo systemctl enable regional-check-agent

# Disable auto-start
sudo systemctl disable regional-check-agent
```

### Health Check

The agent provides a health check endpoint:

```bash
curl http://localhost:8091/health
```

### API Endpoints

- `GET /health` - Health check
- `POST /operation` - Perform monitoring operations
- `GET /operation/quick` - Quick operation test

## Security

The package implements several security measures:

- **Dedicated user**: Runs as `regional-check-agent` user
- **Systemd hardening**: Multiple security restrictions enabled
- **File permissions**: Proper ownership and permissions
- **Network isolation**: Only necessary network access
- **No privilege escalation**: NoNewPrivileges enabled

## Troubleshooting

### Common Issues

1. **Service won't start**
   - Check configuration: `sudo systemctl status regional-check-agent`
   - Verify environment variables are set correctly
   - Check logs: `sudo journalctl -u regional-check-agent`

2. **Permission denied errors**
   - Ensure proper file ownership: `sudo chown -R regional-check-agent:regional-check-agent /var/lib/regional-check-agent`

3. **Network connectivity issues**
   - Verify PocketBase URL is accessible
   - Check firewall settings
   - Test with curl: `curl http://your-pocketbase-url/api/health`

### Log Files

- **Systemd logs**: `sudo journalctl -u regional-check-agent`
- **Application logs**: `/var/log/regional-check-agent/` (if file logging is enabled)

## Uninstallation

```bash
# Remove package but keep configuration
sudo dpkg -r distributed-regional-check-agent

# Remove package and all configuration
sudo dpkg --purge distributed-regional-check-agent
```

## Development

### Building from Source

```bash
# Clone and build
git clone <repository>
cd micro-services/distributed-regional-check-agent
./build.sh
```

### Testing

```bash
# Run tests
make test

# Test installation
make install
```

## Package Information

- **Package name**: `distributed-regional-check-agent`
- **Version**: `1.0.0`
- **Architectures**: `amd64`, `arm64`
- **Dependencies**: `libc6 (>= 2.17)`, `ca-certificates`
- **Service name**: `regional-check-agent`
- **Binary location**: `/usr/bin/distributed-regional-check-agent`