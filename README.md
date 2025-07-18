# ccmon

A TUI (Terminal User Interface) application for monitoring Claude Code API usage through OpenTelemetry (OTLP) telemetry data. ccmon displays real-time statistics for token usage, costs, and request counts with separate tracking for base (Haiku) and premium (Sonnet/Opus) models.

Inspired by [ccusage](https://github.com/ryoppippi/ccusage), but uses OTLP to receive Claude Code usage data.

## Features

- **Real-time Monitoring**: Live TUI dashboard showing Claude Code API usage statistics
- **Token Tracking**: Separate monitoring for base (Haiku) and premium (Sonnet/Opus) models
- **Cost Analysis**: Track API costs and usage patterns
- **Block Progress**: Monitor Claude token limit progress with 5-hour block tracking and beautiful gradient progress bars
- **Time Filtering**: Filter data by various time periods (last hour, day, week, etc.)
- **Configurable Refresh**: Customizable monitor refresh intervals (1s to 5m)
- **Data Retention**: Automatic cleanup of old telemetry data with configurable retention periods
- **OTLP Integration**: Receives telemetry data via OpenTelemetry protocol
- **Dual Operating Modes**: Monitor mode (TUI) and server mode (headless collector)

## Installation

### Homebrew (macOS and Linux)

```bash
# Add the tap from the main repository
brew tap elct9620/ccmon https://github.com/elct9620/ccmon

# Install stable release from pre-built binaries
brew install ccmon

# Or install latest development version from source (requires Go and protobuf)
brew install --head ccmon
```

### Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/elct9620/ccmon/releases).

### Docker

```bash
# Pull the latest image
docker pull ghcr.io/elct9620/ccmon:latest

# Check version
docker run --rm ghcr.io/elct9620/ccmon:latest --version

# Run in server mode (recommended for security)
# Note: Binding to 127.0.0.1:4317 restricts access to localhost only
docker run -d \
  --name ccmon \
  -p 127.0.0.1:4317:4317 \
  -v ccmon-data:/data \
  ghcr.io/elct9620/ccmon:latest
```

### Build from Source

```bash
git clone https://github.com/elct9620/ccmon.git
cd ccmon
make build
```

## Usage

### Operating Modes

ccmon has two distinct operating modes that work together:

#### 1. Server Mode (Required First)
Headless OTLP collector + gRPC query service that receives telemetry data from Claude Code:
```bash
./ccmon -s
# or
./ccmon --server

# With data retention (automatically delete old records)
./ccmon -s --server-retention 7d   # Keep 7 days of data
./ccmon -s --server-retention 30d  # Keep 30 days of data
```

**Important**: You must run the server mode first to collect telemetry data before using the monitor.

#### 2. Monitor Mode
TUI dashboard that connects to the server and displays usage statistics:
```bash
./ccmon                    # Connect to default server (localhost:4317)
./ccmon --monitor-server host:port # Connect to specific server
```

#### 3. Block Tracking Mode
Monitor with Claude token limit progress bars for 5-hour blocks:
```bash
./ccmon -b 5am      # Track usage from 5am start blocks
./ccmon --block 11pm # Track usage from 11pm start blocks
```

#### 4. Format Query Mode
Quick query mode that outputs formatted usage data directly to stdout:
```bash
./ccmon --format "@daily_cost"              # Today's cost (e.g., $1.2)
./ccmon --format "@monthly_cost"            # This month's cost
./ccmon --format "Today: @daily_cost"       # Custom format with text
./ccmon --format "@daily_plan_usage"        # Daily plan usage percentage
./ccmon --format "@monthly_plan_usage"      # Monthly plan usage percentage
```

**Available Variables:**
- `@daily_cost` - Today's total cost (e.g., "$1.2")
- `@monthly_cost` - This month's total cost
- `@daily_plan_usage` - Daily usage as percentage of plan limit (e.g., "15%")
- `@monthly_plan_usage` - Monthly usage as percentage of plan limit

**Example Usage:**
```bash
# Simple cost query
./ccmon --format "@daily_cost"
# Output: $1.2

# Custom format with multiple variables
./ccmon --format "Daily: @daily_cost (@daily_plan_usage of plan)"
# Output: Daily: $1.2 (15% of plan)

# Use in scripts
DAILY_COST=$(./ccmon --format "@daily_cost")
echo "Today's Claude usage cost: $DAILY_COST"
```

### Version Information

Check the installed version of ccmon:
```bash
./ccmon --version
# or
./ccmon -v
```

This will display:
- Version number (e.g., v0.4.0 for releases or "dev" for development builds)
- Git commit hash
- Build date

### Quick Start

1. **Start the server** (receives telemetry data):
```bash
# Using Docker (recommended)
docker run -d \
  --name ccmon \
  -p 127.0.0.1:4317:4317 \
  -v ccmon-data:/data \
  ghcr.io/elct9620/ccmon:latest

# Or using binary
./ccmon --server
```

2. **Configure Claude Code** to send telemetry:
```bash
export CLAUDE_CODE_ENABLE_TELEMETRY=1
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
# ... other OTEL variables (see configuration section)
```

3. **Use Claude Code** to generate some API requests

4. **Start the monitor** to view data:
```bash
# Using Docker
docker run --rm -it --network host ghcr.io/elct9620/ccmon:latest --monitor-server localhost:4317

# Or using binary
./ccmon
```

### Docker Usage

#### Server Mode (Step 1 - Required)
```bash
# Run server with persistent data (bind to localhost only for security)
docker run -d \
  --name ccmon \
  -p 127.0.0.1:4317:4317 \
  -v ccmon-data:/data \
  -e TZ=UTC \
  ghcr.io/elct9620/ccmon:latest

# With data retention
docker run -d \
  --name ccmon \
  -p 127.0.0.1:4317:4317 \
  -v ccmon-data:/data \
  -e TZ=UTC \
  ghcr.io/elct9620/ccmon:latest \
  --server --server-retention 7d

# Check server logs
docker logs ccmon
```

#### Monitor Mode (Step 4 - After server is running)
```bash
# Connect to existing server on same host
docker run --rm -it \
  --network host \
  ghcr.io/elct9620/ccmon:latest

# Connect to server on different host
docker run --rm -it \
  ghcr.io/elct9620/ccmon:latest \
  --monitor-server your-server:4317
```

### Docker Compose

Create a `docker-compose.yml` file:

```yaml
version: '3.8'

services:
  ccmon:
    image: ghcr.io/elct9620/ccmon:latest
    container_name: ccmon
    ports:
      - "127.0.0.1:4317:4317"  # Bind to localhost only for security
    volumes:
      - ccmon-data:/data
      - ./config.toml:/app/config.toml:ro  # Optional: custom config
    environment:
      - TZ=UTC
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "./ccmon", "--help"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  ccmon-data:
```

Run with Docker Compose:
```bash
# Start the server
docker-compose up -d

# View logs
docker-compose logs -f

# Connect with monitor mode
docker run --rm -it \
  --network host \
  ghcr.io/elct9620/ccmon:latest \
  --monitor-server localhost:4317

# Stop the server
docker-compose down
```

## Configuration

ccmon supports configuration files in TOML, YAML, or JSON format. The application searches for configuration files in:

1. Current directory: `./config.{toml,yaml,json}`
2. User config directory: `~/.ccmon/config.{toml,yaml,json}`

### Example Configuration

```toml
[database]
# Path to the BoltDB database file
path = "~/.ccmon/ccmon.db"

[server]
# gRPC server address for OTLP receiver + Query service
address = "127.0.0.1:4317"
# Data retention period (optional)
# retention = "7d"  # Keep 7 days of data
# retention = "30d" # Keep 30 days of data
# retention = "never" # Keep all data (default)

[monitor]
# gRPC server address for query service
server = "127.0.0.1:4317"
# Timezone for time filtering and display
timezone = "UTC"
# Monitor refresh interval (how often the TUI updates)
refresh_interval = "5s"  # Options: "1s", "5s", "10s", "30s", "1m", etc.

[claude]
# Claude subscription plan for automatic token limit detection
plan = "pro"  # Options: "unset", "pro", "max", "max20"
# Custom token limit override (optional)
max_tokens = 7000
```

See `config.toml.example` for a complete configuration example.

### Monitor Customization

The monitor mode can be customized to fit different usage patterns and system capabilities:

#### Refresh Interval
Control how frequently the TUI updates its data display:

```toml
[monitor]
refresh_interval = "5s"    # Default: matches Claude Code telemetry frequency
# refresh_interval = "1s"  # Fast refresh for active development
# refresh_interval = "10s" # Balanced refresh for normal usage
# refresh_interval = "30s" # Slower refresh to save resources
# refresh_interval = "1m"  # Minimal overhead for background monitoring
```

**Guidelines:**
- **1-2 seconds**: Best for active development and real-time monitoring
- **5 seconds**: Default rate, aligns with Claude Code's telemetry frequency
- **10-30 seconds**: Good balance between responsiveness and resource usage
- **1-5 minutes**: Minimal overhead for background monitoring on slower systems

**Note:** Claude Code sends telemetry approximately every 5 seconds, so refresh intervals shorter than 5s may not show new data more frequently.

### Data Retention

ccmon supports automatic cleanup of old telemetry data to manage storage space. When enabled, the server will automatically delete records older than the specified period.

#### Configuration Options

1. **Via Configuration File** (config.toml):
```toml
[server]
retention = "7d"    # Keep 7 days of data
# retention = "30d"   # Keep 30 days of data  
# retention = "never" # Keep all data (default)
```

2. **Via Command Line Flag**:
```bash
./ccmon -s --server-retention 7d
```

3. **Via Docker**:
```bash
docker run -d \
  --name ccmon \
  -p 127.0.0.1:4317:4317 \
  -v ccmon-data:/data \
  ghcr.io/elct9620/ccmon:latest \
  --server --server-retention 30d
```

#### Retention Period Format
- Supported formats: `"1d"`, `"7d"`, `"30d"`, `"24h"`, `"168h"`, `"720h"`
- Minimum retention: 24 hours (prevents accidental data loss)
- Default: `"never"` (no automatic cleanup)

#### How It Works
- Cleanup runs automatically every 6 hours when retention is enabled
- Only deletes records older than the specified period
- Runs in the background without affecting server performance

## Claude Code Integration

To send telemetry data to ccmon, configure Claude Code with these environment variables:

```bash
export CLAUDE_CODE_ENABLE_TELEMETRY=1
export OTEL_METRICS_EXPORTER=otlp
export OTEL_LOGS_EXPORTER=otlp
export OTEL_EXPORTER_OTLP_PROTOCOL=grpc
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
```

If running ccmon server on a different host:
```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=http://your-server:4317
```

## Development

### Prerequisites
- Go 1.24.3+
- Make
- Protocol Buffers compiler (protoc) v30.2+

### Build Commands
```bash
# Generate protobuf code
make generate

# Build the application
make build

# Format code
gofmt -w .

# Clean build artifacts
make clean
```

### Development with Docker
```bash
# Build local image
docker build -t ccmon:dev .

# Run development server (bind to localhost for security)
docker run --rm -p 127.0.0.1:4317:4317 ccmon:dev
```

## Architecture

ccmon follows Clean Architecture and Domain-Driven Design (DDD) principles:

- **Handler Layer**: Separate TUI and gRPC handlers
- **Usecase Layer**: Business logic with CQRS commands and queries
- **Repository Layer**: Data access with entity conversion
- **Entity Layer**: Domain entities with encapsulated business logic
- **gRPC Communication**: Monitor mode communicates via gRPC queries

For detailed architecture documentation, see [CLAUDE.md](./CLAUDE.md).

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes using [conventional commits](https://conventionalcommits.org/)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by [ccusage](https://github.com/example/ccusage)
- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the TUI
- Uses [OpenTelemetry](https://opentelemetry.io/) for telemetry collection
