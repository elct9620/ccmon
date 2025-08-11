# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ccmon is a TUI (Terminal User Interface) application that monitors Claude Code API usage by receiving OpenTelemetry (OTLP) telemetry data. It displays real-time statistics for token usage, costs, and request counts, with separate tracking for base (Haiku) and premium (Sonnet/Opus) models.

## Documentation

- The `docs/glossary.md` is defined our domain concepts

## Architecture

ccmon follows Clean Architecture and Domain-Driven Design (DDD) principles with clear separation of concerns:

### Directory Structure
- `entity/` - Domain entities and business rules (DDD principles)
- `usecase/` - Business logic layer implementing CQRS commands and queries
- `repository/` - Data access implementations with entity conversion
- `handler/` - External interfaces (TUI, gRPC, CLI)
- `service/` - Infrastructure services (time handling, external adapters)

## Development Requirements

### Protocol Buffers Toolchain
- **Required protoc version**: v30.2 or higher
- **Required protoc-gen-go**: v1.28.1 (pinned for consistency)
- **Required protoc-gen-go-grpc**: v1.2.0 (pinned for consistency)
- **Installation**: 
  - protoc: Download from [GitHub Releases](https://github.com/protocolbuffers/protobuf/releases)
  - Go plugins: `go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1`
  - gRPC plugin: `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0`
- **Version check**: Run `make check-protoc` to verify complete toolchain

**Important**: Using different protoc/plugin versions between development and CI can cause inconsistent generated files that may break Homebrew formula updates.

## Operating Modes

ccmon has four distinct operating modes:

1. **Monitor Mode** (default): TUI dashboard that displays usage statistics via gRPC queries
   ```bash
   ./ccmon
   ```

2. **Server Mode**: Headless OTLP collector + gRPC query service that receives and stores telemetry data with optional data retention
   ```bash
   ./ccmon -s
   # or
   ./ccmon --server
   # Configure retention via config file or flag:
   ./ccmon -s --server-retention 30d
   ```

3. **Block Tracking Mode**: Monitor with Claude token limit progress bars for 5-hour blocks
   ```bash
   ./ccmon -b 5am      # Track usage from 5am start blocks
   ./ccmon --block 11pm # Track usage from 11pm start blocks
   ```

4. **Format Query Mode**: Non-interactive command-line output for scripting
   ```bash
   ./ccmon --format "@daily_cost"              # Today's cost
   ./ccmon --format "Today: @daily_cost"       # Custom format
   ./ccmon --format "@daily_plan_usage"        # Plan usage percentage
   ```

## Development Commands

### Build Commands
```bash
# Check protoc version compatibility
make check-protoc

# Build the application (includes protobuf generation)
make build

# Generate protobuf code
make generate

# Install/update dependencies
make deps

# Clean build artifacts
make clean

# Development shortcuts
make dev-server    # Clean, build, and run server mode
make dev-monitor   # Clean, build, and run monitor mode
```

### Testing and Verification
```bash
# Quick verification workflow (before committing)
make fmt && make vet && make test

# Format code
make fmt

# Vet code
make vet

# Run all tests
make test

# Run tests with coverage
go test -cover ./...

# Detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific tests
go test ./entity/api_request_test.go -v
go test -run TestAPIRequest_ID ./entity/ -v
go test -run Integration ./...
go test ./handler/tui/ -v  # TUI tests with teatest

# Lint code (if available)
golangci-lint run
```

### Coverage Guidelines
- Aim for **>80% test coverage** for new code
- Focus on testing business logic in `usecase/` and `entity/` packages
- Repository and handler layers should have integration tests

## Configuration

ccmon supports configuration files in TOML, YAML, or JSON format in these locations (first found wins):

1. Current directory: `./config.{toml,yaml,json}` (highest priority)
2. User config directory: `~/.ccmon/config.{toml,yaml,json}`

### Key Configuration Options

```toml
[database]
path = "~/.ccmon/ccmon.db"  # Database file path

[server]
address = "127.0.0.1:4317"  # gRPC server address

[monitor]
server = "127.0.0.1:4317"   # Query service address
timezone = "UTC"            # Timezone for display
refresh_interval = "5s"     # TUI refresh rate

[claude]
plan = "unset"              # Subscription plan: unset/pro/max/max20
max_tokens = 0              # Custom token limit override
```

### Environment Variables for Claude Code Telemetry

```bash
export CLAUDE_CODE_ENABLE_TELEMETRY=1
export OTEL_METRICS_EXPORTER=otlp
export OTEL_LOGS_EXPORTER=otlp
export OTEL_EXPORTER_OTLP_PROTOCOL=grpc
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
```

### Key Design Patterns
- **Handler Separation**: TUI and gRPC handlers with distinct responsibilities
- **gRPC Communication**: Monitor communicates with server via gRPC (no direct DB access)
- **Dual Statistics System**: Separate `stats` (display) and `blockStats` (progress tracking)
- **Entity-Based Architecture**: Handlers depend only on domain entities
- **Domain-Driven Design**: Entities with private fields and encapsulated business logic

## Authentication

ccmon supports optional authentication using Authorization headers, ideal for securing servers behind Cloudflare Zero Trust or other proxy services.

### Usage Scenarios

1. **Local Development**: No authentication required (default)
2. **Public Deployment**: Use Bearer tokens or service tokens for security 
3. **Proxy Services**: Use service tokens or custom authentication formats
4. **Custom Authentication**: Use any Authorization header format

### Bearer Token Example

For standard Bearer authentication:

```toml
[server.auth]
token = "Bearer your-secret-api-key"

[monitor.auth]
token = "Bearer your-secret-api-key"
```

The token value is sent as-is in the `Authorization` header via gRPC metadata (equivalent to HTTP headers).

### Data Flow
**Server Mode**: Claude Code → OTLP (port 4317) → gRPC receiver → usecase → repository → BoltDB

**Monitor Mode**: TUI → gRPC queries → server → statistics display (refreshes every 5s)

## Block Tracking Feature

ccmon supports Claude's 5-hour token limit blocks for monitoring API usage against subscription plan limits.

### Usage
```bash
./ccmon -b 5am      # Track from 5am blocks (5am-10am, 10am-3pm, etc.)
./ccmon --block 11pm # Track from 11pm blocks (11pm-4am, 4am-9am, etc.)
```

### Features
- **Visual Progress Bar**: Color-coded bars (green → orange → red) showing usage percentage
- **Token Counting**: Only premium tokens (Sonnet/Opus) count toward limits; Haiku tokens are free
- **Time Remaining**: Shows countdown until next 5-hour block starts
- **Automatic Block Advancement**: Blocks automatically advance every 5 hours without requiring restart
- **Plan Integration**: Auto-detects limits based on `claude.plan` config (pro=7K, max=35K, max20=140K)
- **Custom Limits**: Override with `claude.max_tokens` config for custom token limits
- **Timezone Support**: Uses `monitor.timezone` config for accurate block calculations
- **Keyboard Filter**: Press 'b' key to filter requests by current block timeframe
- **Always Valid Blocks**: Shows current or upcoming block - no "too early" errors

### Display Format
```
Block Progress (5am - 10am):
[████████░░] 80% (5,600/7,000 tokens)
Time remaining: 2h 15m
```

## Model Identification Logic

- Base models (not counted against limits): Identified by checking if model name contains "haiku" (case-insensitive)
- Premium models (counted against limits): All other models (Sonnet, Opus, etc.)

## Important Implementation Details

- All numeric values from Claude Code telemetry are sent as strings and must be parsed using `fmt.Sscanf()`
- Monitor mode refreshes via gRPC queries every 5 seconds
- Statistics are tracked separately for base/premium tiers and combined totals
- The gRPC server runs on port 4317 (standard OTLP port) providing both OTLP and Query services
- Table height and column widths are dynamically adjusted based on terminal size in monitor mode
- TUI supports tab navigation with Tab key to switch between "Current" and "Daily Usage" views
- TUI supports sortable request list with 'o' key to toggle between latest-first and oldest-first
- Daily usage tab shows last 30 days with detailed premium token breakdown (Input/Output/Cache)
- Block tracking mode shows progress bars for 5-hour token limit periods with 'b' key filtering and automatic advancement
- Multiple monitors can connect to the same server via gRPC (no database conflicts)
- Database limits stored requests to last 10,000 entries with efficient limiting support
- Monitor and server can run on different machines by configuring monitor.server address
- Network traffic optimized: TUI requests only 100 records for display, separate unlimited query for statistics

## Development Conventions

- Always write devlog in `docs/devlog/` use Markdown format, group by day. e.g. 20250628.md

## Entity Design Patterns

All entities in `entity/` package follow Domain-Driven Design (DDD) principles:

### Core Principles
1. **Private Fields**: All struct fields must be private (lowercase) for encapsulation
2. **Getter Methods**: Provide public getter methods for accessing field values
3. **Immutability**: Entities are immutable after creation - no setter methods
4. **Factory Functions**: Use `NewXxx()` functions for entity creation
5. **Business Logic**: Encapsulate domain behavior within entities

### Implementation Example
```go
// Entity with private fields
type APIRequest struct {
    sessionID string
    timestamp time.Time
    model     Model
    tokens    Token
    cost      Cost
}

// Factory function for creation
func NewAPIRequest(sessionID string, timestamp time.Time, ...) APIRequest {
    return APIRequest{
        sessionID: sessionID,
        timestamp: timestamp,
        // ... initialize fields
    }
}

// Getter methods for field access
func (a APIRequest) SessionID() string { return a.sessionID }
func (a APIRequest) Timestamp() time.Time { return a.timestamp }

// Business logic methods
func (a APIRequest) ID() string {
    return fmt.Sprintf("%s_%s", a.timestamp.Format(time.RFC3339Nano), a.sessionID)
}
```

## Testing Conventions

### TUI Testing with teatest
- **ALWAYS use `teatest` for TUI integration tests** instead of unit testing individual methods
- **Focus on real user interactions**: Test keyboard navigation, rendering, and workflows
- **Use consistent patterns**: All TUI tests follow `teatest.NewTestModel()` approach
- **Parallel execution**: Use `t.Parallel()` in all table-driven tests for performance
- **Optimized timeouts**: Keep `WaitFor` timeouts ≤500ms for responsive testing

### General Testing
- **ALWAYS use table-driven tests** for comprehensive test coverage
- **Test files naming**: Use `*_test.go` for unit tests, separate files per component

## Important Implementation Details

### Model Classification
- **Base models** (free, not counted): Contains "haiku" (case-insensitive)
- **Premium models** (counted against limits): All other models (Sonnet, Opus, etc.)

### Key Technical Notes
- All numeric values from Claude Code telemetry are sent as strings - parse using `fmt.Sscanf()`
- Monitor mode refreshes via gRPC queries every 5 seconds
- gRPC server runs on port 4317 (standard OTLP port)
- Database limits stored requests to last 10,000 entries
- Network traffic optimized: TUI requests 100 records for display, unlimited for statistics
- Multiple monitors can connect to same server via gRPC (no database conflicts)

### Development Conventions
- Write devlog in `docs/devlog/` using Markdown format, grouped by day (e.g., `20250628.md`)
- Apply YAGNI principle: Only implement immediately necessary features
- Avoid premature optimization and unnecessary abstractions

