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

## Build and Development Commands

```bash
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
```

## Architecture

The application follows a modular architecture with clear separation of concerns:

### Handler Architecture
1. **main.go** - Entry point with command-line flag parsing and mode routing
2. **config.go** - Configuration system using Viper with TOML/YAML/JSON support

#### TUI Handler (`handler/tui/`)
1. **monitor.go** - Sets up the TUI monitor mode with gRPC client
2. **model.go** - Bubble Tea model that queries data via gRPC and refreshes periodically  
3. **ui.go** - Rendering logic using Lipgloss for styled terminal output

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
- **Model Tier Separation**: Distinguishes between base models (Haiku) and premium models (Sonnet/Opus)
- **Domain-Driven Design**: Entities with private fields, getter methods, and encapsulated business logic
- **Entity-Based Architecture**: Handlers depend only on domain entities, not database implementation types
- **UI-Owned Time Filtering**: Time range calculations handled in presentation layer using entity.Period
- **Repository Pattern**: Repository layer handles entity conversion and database operations
- **Database Factory Functions**: Simple factory functions for database initialization at root level
- **Clean Architecture Compliance**: Proper dependency inversion with usecase and repository layers

### Data Flow

**Server Mode:**
1. main.go creates BoltDB instance using factory functions and injects into repository
2. Repository and usecase layers are initialized with proper dependency injection
3. Claude Code sends OTLP telemetry data to port 4317
4. gRPC server (`handler/grpc/server.go`) handles connection and service registration
5. OTLP receiver (`handler/grpc/receiver/`) parses log records with body "claude_code.api_request"
6. Data is saved via usecase commands that coordinate with repository layer
7. Query service (`handler/grpc/query/`) provides gRPC API using usecase queries
8. Requests are logged to console

**Monitor Mode:**
1. main.go initializes gRPC repository and usecase layer for monitor mode
2. TUI handler receives usecase queries for data access
3. Queries data via gRPC calls through repository abstraction
4. Refreshes statistics every 5 seconds via usecase layer
5. Allows time-based filtering with keyboard shortcuts using entity.Period
6. Displays data in a TUI table

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

[claude]
# Claude subscription plan
# Default: "unset"
# Valid values: "unset", "pro", "max", "max20"
plan = "unset"
```

See `config.toml.example` for a complete example configuration file.

## Model Identification Logic

- Base models (not counted against limits): Identified by checking if model name contains "haiku" (case-insensitive)
- Premium models (counted against limits): All other models (Sonnet, Opus, etc.)

## Important Implementation Details

- All numeric values from Claude Code telemetry are sent as strings and must be parsed using `fmt.Sscanf()`
- Monitor mode refreshes via gRPC queries every 5 seconds
- Server mode logs each request to console for visibility
- Statistics are tracked separately for base/premium tiers and combined totals
- The gRPC server runs on port 4317 (standard OTLP port) providing both OTLP and Query services
- Table height is dynamically adjusted based on terminal size in monitor mode
- Multiple monitors can connect to the same server via gRPC (no database conflicts)
- Database limits stored requests to last 10,000 entries
- Monitor and server can run on different machines by configuring monitor.server address

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

func NewPeriodFromDuration(duration time.Duration) Period {
    now := time.Now()
    return Period{startAt: now.Add(-duration), endAt: now}
}

func NewAllTimePeriod() Period {
    return Period{startAt: time.Time{}, endAt: time.Now()}
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