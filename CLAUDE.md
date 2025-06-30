# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ccmon is a TUI (Terminal User Interface) application that monitors Claude Code API usage by receiving OpenTelemetry (OTLP) telemetry data. It displays real-time statistics for token usage, costs, and request counts, with separate tracking for base (Haiku) and premium (Sonnet/Opus) models.

## Operating Modes

ccmon has two distinct operating modes:

1. **Monitor Mode** (default): TUI dashboard that displays usage statistics via gRPC queries
   ```bash
   ./ccmon
   ```

2. **Server Mode**: Headless OTLP collector + gRPC query service that receives and stores telemetry data
   ```bash
   ./ccmon -s
   # or
   ./ccmon --server
   ```

3. **Block Tracking Mode**: Monitor with Claude token limit progress bars for 5-hour blocks
   ```bash
   ./ccmon -b 5am      # Track usage from 5am start blocks
   ./ccmon --block 11pm # Track usage from 11pm start blocks
   ```

## Development Requirements

### Protocol Buffers Toolchain
- **Required protoc version**: v28.0 or higher
- **Required protoc-gen-go**: v1.28.1 (pinned for consistency)
- **Required protoc-gen-go-grpc**: v1.2.0 (pinned for consistency)
- **Installation**: 
  - protoc: Download from [GitHub Releases](https://github.com/protocolbuffers/protobuf/releases)
  - Go plugins: `go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1`
  - gRPC plugin: `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0`
- **Version check**: Run `make check-protoc` to verify complete toolchain

**Important**: Using different protoc/plugin versions between development and CI can cause inconsistent generated files that may break Homebrew formula updates. The release workflow now validates proto files and prevents problematic commits.

## Build and Development Commands

```bash
# Check protoc version compatibility
make check-protoc

# Build the application (includes protobuf generation)
make build

# Generate protobuf code
make generate

# Format code
gofmt -w .

# Run monitor (TUI mode)
./ccmon

# Run server (OTLP collector + Query service)
./ccmon -s

# Clean build artifacts
make clean

# Docker development
docker build -t ccmon:dev .
docker run --rm -p 4317:4317 ccmon:dev
```

## Verification Workflow

After making code changes and before committing, always run this verification workflow to ensure code quality:

```bash
# 1. Format code
gofmt -w .

# 2. Lint code (fix all issues)
golangci-lint run

# 3. Test code with coverage
go test -cover ./...
```

### Coverage Guidelines
- Aim for **>80% test coverage** for new code
- Focus on testing business logic in `usecase/` and `entity/` packages
- Repository and handler layers should have integration tests
- Use `go test -coverprofile=coverage.out ./...` for detailed coverage analysis

## Docker Deployment

ccmon provides production-ready Docker images with multi-architecture support:

```bash
# Run production server
docker run -d \
  --name ccmon-server \
  -p 4317:4317 \
  -v ccmon-data:/data \
  ghcr.io/elct9620/ccmon:latest

# Connect monitor to server
docker run --rm -it \
  --network host \
  ghcr.io/elct9620/ccmon:latest
```

### Docker Configuration
- **Base Image**: Alpine Linux for minimal size and security
- **User**: Non-root user for security best practices
- **Volumes**: `/data` for database persistence
- **Network**: Binds to `0.0.0.0:4317` for external access
- **Multi-arch**: Supports both amd64 and arm64 platforms

## Release Process

ccmon uses automated release management with conventional commits:

### Conventional Commits
- `feat:` - New features (minor version bump)
- `fix:` - Bug fixes (patch version bump) 
- `feat!:` or `fix!:` - Breaking changes (major version bump)
- `docs:`, `chore:`, `refactor:` - No version bump

### Automated Release Flow
1. **Push to main** - Triggers release-please workflow
2. **Release PR Creation** - release-please creates/updates release PR with changelog
3. **Merge Release PR** - Creates GitHub release with semantic version tag
4. **Asset Building** - GoReleaser builds cross-platform binaries and Docker images
5. **Publication** - Binaries attached to release, Docker images pushed to ghcr.io

### Manual Release Commands
```bash
# Check what would be released (dry run)
goreleaser release --snapshot --clean

# Test Docker build locally
docker build -t ccmon:test .

# Set specific version (e.g., 0.1.0) using Release-As footer
git commit --allow-empty -m "chore: release 0.1.0" -m "Release-As: 0.1.0"
```

## Architecture

The application follows a modular architecture with clear separation of concerns:

### Handler Architecture
1. **main.go** - Entry point with command-line flag parsing and mode routing
2. **config.go** - Configuration system using Viper with TOML/YAML/JSON support

#### TUI Handler (`handler/tui/`)
1. **program.go** - Sets up the TUI monitor mode with config handling and initialization
2. **view_model.go** - Bubble Tea model that queries data via gRPC and refreshes periodically  
3. **renderer.go** - Rendering logic using Lipgloss for styled terminal output
4. **block.go** - Block time parsing and calculation logic (private functions)

#### gRPC Handler (`handler/grpc/`)
1. **server.go** - gRPC server lifecycle management and service registration
2. **receiver/receiver.go** - OTLP message processing and data extraction
3. **query/service.go** - Query service implementation for retrieving statistics and requests

### Shared Components
1. **db.go** - Database factory functions for BoltDB initialization
2. **repository/** - Data access layer with entity conversion and database operations
3. **usecase/** - Business logic layer implementing CQRS commands and queries
4. **entity/** - Domain entities following DDD principles with encapsulation
5. **proto/** - Protocol Buffers definitions and generated gRPC code

### Key Design Patterns

- **Handler Separation**: Clear separation between TUI and gRPC handlers with distinct responsibilities
- **Server Lifecycle Management**: gRPC server setup and lifecycle managed in dedicated server layer
- **Message Processing**: OTLP receiver focused solely on protocol message parsing and data extraction
- **Query Service**: Dedicated gRPC service for read-only data access via protobuf interface
- **Dependency Injection**: All dependencies initialized in main.go and injected into handlers
- **gRPC Communication**: Monitor mode communicates with server via gRPC instead of direct database access
- **Periodic Refresh**: Monitor mode refreshes every 5 seconds via gRPC queries
- **Dual Statistics System**: Separate `stats` (filtered display) and `blockStats` (progress tracking) for independent concerns
- **Complete Usecase Integration**: All statistics go through proper usecase layer, no direct entity calculations in TUI
- **BlockStatsRepository Pattern**: Dedicated interface for block-specific statistics with interface segregation
- **Model Tier Separation**: Distinguishes between base models (Haiku) and premium models (Sonnet/Opus)
- **Domain-Driven Design**: Entities with private fields, getter methods, and encapsulated business logic
- **Entity-Based Architecture**: Handlers depend only on domain entities, not database implementation types
- **UI-Owned Time Filtering**: Time range calculations handled in presentation layer using entity.Period
- **Repository Pattern**: Repository layer handles entity conversion and database operations
- **Database Factory Functions**: Simple factory functions for database initialization at root level
- **Clean Architecture Compliance**: Proper dependency inversion with usecase and repository layers
- **Handler Ownership**: Each handler owns its specific business logic (e.g., TUI owns block calculations)

### Data Flow

**Server Mode:**
1. main.go creates BoltDB instance using factory functions and injects into repository
2. Repository and usecase layers are initialized with proper dependency injection
3. Claude Code sends OTLP telemetry data to port 4317
4. gRPC server (`handler/grpc/server.go`) handles connection and service registration
5. OTLP receiver (`handler/grpc/receiver/`) parses log records with body "claude_code.api_request"
6. Data is saved via usecase commands that coordinate with repository layer
7. Query service (`handler/grpc/query/`) provides gRPC API using usecase queries with efficient limiting

**Monitor Mode:**
1. main.go creates MonitorConfig struct with timezone, refresh interval, token limit, and block time
2. TUI handler receives config and usecases, then handles all initialization internally:
   - Loads timezone and validates refresh interval
   - Parses block time and calculates current block if block tracking enabled
   - Initializes view model with computed values
3. Queries data via gRPC calls through repository abstraction with limit parameters
4. Refreshes statistics every 5 seconds via dual stats system:
   - **Display statistics**: Uses `GetStatsQuery` with current filter period
   - **Block statistics**: Uses `GetBlockStatsQuery` with current block period (when enabled)
5. Uses dual query strategy: limit=100 for display, limit=0 for accurate statistics
6. Separate statistics tracking: `stats` for display table, `blockStats` for progress bar
7. Allows time-based filtering and sorting with keyboard shortcuts using entity.Period
8. Displays data in a sortable TUI table with latest-first default ordering
9. Progress bars remain stable regardless of display filter changes

### Environment Variables Required

For Claude Code to send telemetry to ccmon:
```bash
export CLAUDE_CODE_ENABLE_TELEMETRY=1
export OTEL_METRICS_EXPORTER=otlp
export OTEL_LOGS_EXPORTER=otlp
export OTEL_EXPORTER_OTLP_PROTOCOL=grpc
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
```

## Configuration

ccmon supports configuration files in TOML, YAML, or JSON format. The application searches for configuration files in the following locations (first found wins):

1. Current directory: `./config.{toml,yaml,json}` (highest priority)
2. User config directory: `~/.ccmon/config.{toml,yaml,json}`

If no configuration file is found, default values are used.

### Configuration Options

```toml
[database]
# Path to the BoltDB database file
# Default: ~/.ccmon/ccmon.db
path = "~/.ccmon/ccmon.db"

[server]
# gRPC server address for OTLP receiver + Query service
# Default: 127.0.0.1:4317
address = "127.0.0.1:4317"

[monitor]
# gRPC server address for query service
# Default: 127.0.0.1:4317
# Monitor connects to this address to query data from server
server = "127.0.0.1:4317"

# Timezone for time filtering and display in monitor mode
# Default: "UTC"
# Examples: "UTC", "America/New_York", "Europe/London", "Asia/Tokyo"
timezone = "UTC"

# Monitor refresh interval for updating data in TUI
# Default: "5s"
# Examples: "1s", "5s", "10s", "30s", "1m"
# Minimum: 1s, Maximum: 5m
refresh_interval = "5s"

[claude]
# Claude subscription plan
# Default: "unset"
# Valid values: "unset", "pro", "max", "max20"
# Used for automatic token limit detection when using block tracking (-b flag)
plan = "unset"

# Custom token limit override
# Default: 0 (use plan defaults)
# Set to override default limits: pro=7000, max=35000, max20=140000
# Use with block tracking (-b flag) to monitor token usage within 5-hour blocks
max_tokens = 0
```

See `config.toml.example` for a complete example configuration file.

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

The `entity/` package contains domain entities following Domain-Driven Design (DDD) principles. All entities must adhere to these patterns:

### Core Principles

1. **Private Fields**: All struct fields must be private (lowercase) to ensure encapsulation
2. **Getter Methods**: Provide public getter methods for accessing field values
3. **Immutability**: Entities are immutable after creation - no setter methods
4. **Factory Functions**: Use `NewXxx()` functions for entity creation
5. **Business Logic**: Encapsulate domain behavior within entities

### Implementation Pattern

```go
// Entity with private fields
type APIRequest struct {
    sessionID string
    timestamp time.Time
    model     Model
    tokens    Token
    cost      Cost
    duration  time.Duration
}

// Factory function for creation
func NewAPIRequest(sessionID string, timestamp time.Time, ...) APIRequest {
    return APIRequest{
        sessionID: sessionID,
        timestamp: timestamp,
        // ... initialize all fields
    }
}

// Getter methods for field access
func (a APIRequest) SessionID() string {
    return a.sessionID
}

func (a APIRequest) Timestamp() time.Time {
    return a.timestamp
}

// Business logic methods
func (a APIRequest) ID() string {
    return fmt.Sprintf("%s_%s", a.timestamp.Format(time.RFC3339Nano), a.sessionID)
}
```

### Value Objects

Value objects like `Token`, `Cost`, `Model`, and `Period` follow the same pattern:

```go
type Cost struct {
    amount float64
}

func NewCost(amount float64) Cost {
    return Cost{amount: amount}
}

func (c Cost) Amount() float64 {
    return c.amount
}

// Immutable operations return new instances
func (c Cost) Add(other Cost) Cost {
    return Cost{amount: c.amount + other.amount}
}
```

### Period Value Object

The `Period` value object represents time ranges for filtering operations:

```go
type Period struct {
    startAt time.Time
    endAt   time.Time
}

func NewPeriod(startAt, endAt time.Time) Period {
    return Period{startAt: startAt, endAt: endAt}
}

func NewPeriodFromDuration(now time.Time, duration time.Duration) Period {
    return Period{startAt: now.Add(-duration), endAt: now}
}

func NewAllTimePeriod(now time.Time) Period {
    return Period{startAt: time.Time{}, endAt: now}
}

func (p Period) StartAt() time.Time { return p.startAt }
func (p Period) EndAt() time.Time { return p.endAt }
func (p Period) IsAllTime() bool { return p.startAt.IsZero() }
```

### Benefits

- **Encapsulation**: Internal representation is hidden from external packages
- **Immutability**: Prevents accidental state mutations
- **Testability**: Business logic can be tested without infrastructure dependencies
- **Maintainability**: Changes to internal structure don't affect external code
- **Domain Focus**: Entities contain only domain logic, no infrastructure concerns

## Testing Conventions

- **ALWAYS use table-driven tests** for comprehensive test coverage and readability

## Design Principles and Guidelines

- **Always apply YAGNI principle**: Only implement features and complexity that are immediately necessary, avoiding premature optimization and unnecessary abstractions
