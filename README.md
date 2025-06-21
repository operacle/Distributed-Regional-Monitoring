
# CheckCle Distributed Regional Monitoring Agent

A Go-based microservice for Uptime Regional Monitoring Agent including ICMP ping, DNS resolution, and TCP connectivity testing.

## Features

- **ICMP Ping**: Full ping functionality with packet statistics
- **DNS Resolution**: A, AAAA, MX, and TXT record lookups
- **TCP Connectivity**: Port connectivity testing
- **SSL Certificate**: SSL Certificate Check
- REST API endpoints
- Health check endpoint
- Configurable via environment variables
- Comprehensive operation statistics

## API Endpoints

### POST /operation
Perform various network operations (ping, dns, tcp).

**Ping Request:**
```json
{
  "type": "ping",
  "host": "google.com",
  "count": 4,
  "timeout": 3
}
```

**DNS Request:**
```json
{
  "type": "dns",
  "host": "google.com",
  "query": "A",
  "timeout": 3
}
```

**TCP Request:**
```json
{
  "type": "tcp",
  "host": "google.com",
  "port": 443,
  "timeout": 3
}
```

**Response:**
```json
{
  "type": "ping",
  "host": "google.com",
  "success": true,
  "response_time": "20ms",
  "packets_sent": 4,
  "packets_recv": 4,
  "packet_loss": 0,
  "min_rtt": "15ms",
  "max_rtt": "25ms",
  "avg_rtt": "20ms",
  "rtts": ["15ms", "20ms", "25ms", "18ms"],
  "start_time": "2023-12-01T10:00:00Z",
  "end_time": "2023-12-01T10:00:03Z"
}
```

### GET /operation/quick
Quick operation test with query parameters.

**Examples:**
- `/operation/quick?type=ping&host=google.com&count=1`
- `/operation/quick?type=dns&host=google.com&query=A`
- `/operation/quick?type=tcp&host=google.com&port=443`

### GET /health
Health check endpoint.

### Legacy Endpoints
- `POST /ping` - Legacy ping endpoint (backward compatibility)
- `GET /ping/quick` - Legacy quick ping endpoint

## Operation Types

### Ping (ICMP)
- **Type**: `ping`
- **Parameters**: `host`, `count`, `timeout`
- **Features**: Packet loss calculation, RTT statistics, multiple packets

### DNS Resolution
- **Type**: `dns`
- **Parameters**: `host`, `query` (A, AAAA, MX, TXT), `timeout`
- **Features**: Multiple record types, resolution time tracking

### TCP Connectivity
- **Type**: `tcp`
- **Parameters**: `host`, `port`, `timeout`
- **Features**: Connection testing, response time measurement

## Configuration

Environment variables:

- `PORT` - Service port (default: 8080)
- `DEFAULT_COUNT` - Default ping count (default: 4)
- `DEFAULT_TIMEOUT` - Default timeout (default: 3s)
- `MAX_COUNT` - Maximum ping count (default: 20)
- `MAX_TIMEOUT` - Maximum timeout (default: 30s)
- `ENABLE_LOGGING` - Enable logging (default: true)

## Running

### Local Development
```bash
go run main.go
```

### Docker
```bash
docker build -t service-operation .
docker run -p 8091:8091 service-operation
```

## Requirements

- Go 1.21+
- Root privileges for ICMP (on Linux)

## Note

This service requires elevated privileges to send ICMP packets. Run with sudo on Linux systems or use capabilities:

```bash
sudo setcap cap_net_raw=+ep ./service-operation
```
