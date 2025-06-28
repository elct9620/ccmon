# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ccmon is a TUI (Terminal User Interface) application that monitors Claude Code API usage by receiving OpenTelemetry (OTLP) telemetry data. It displays real-time statistics for token usage, costs, and request counts, with separate tracking for base (Haiku) and premium (Sonnet/Opus) models.

## Operating Modes

ccmon has two distinct operating modes:

1. **Monitor Mode** (default): TUI dashboard that displays usage statistics from the database
   ```bash
   ./ccmon
   ```

2. **Server Mode**: Headless OTLP collector that receives and stores telemetry data
   ```bash
   ./ccmon -s
   # or
   ./ccmon --server
   ```

## Build and Development Commands

```bash
# Build the application
go build -o ccmon .

# Format code
gofmt -w .

# Run monitor (TUI mode)
./ccmon

# Run server (OTLP collector)
./ccmon -s

# Clean build artifacts
rm ccmon
```

## Architecture

The application follows a modular architecture with clear separation of concerns:

### Monitor Mode Files
1. **main.go** - Entry point with command-line flag parsing to determine mode
2. **monitor.go** - Sets up the TUI monitor mode
3. **model.go** - Bubble Tea model that reads from database and refreshes periodically
4. **ui.go** - Rendering logic using Lipgloss for styled terminal output

### Server Mode Files
1. **server.go** - Headless OTLP server with console logging
2. **receiver.go** - OTLP gRPC server that receives telemetry data and saves to database

### Shared Components
1. **db.go** - BoltDB database operations for persistent storage

### Key Design Patterns

- **Mode Separation**: Monitor mode (TUI) and server mode (headless) run independently
- **Database-Centric**: Both modes interact through the BoltDB database
- **Periodic Refresh**: Monitor mode refreshes every 5 seconds from database
- **Model Tier Separation**: Distinguishes between base models (Haiku) and premium models (Sonnet/Opus)

### Data Flow

**Server Mode:**
1. Claude Code sends OTLP telemetry data to port 4317
2. The receiver parses log records with body "claude_code.api_request"
3. Extracted data is saved to BoltDB database
4. Requests are logged to console

**Monitor Mode:**
1. Reads existing data from BoltDB database
2. Refreshes statistics every 5 seconds
3. Allows time-based filtering with keyboard shortcuts
4. Displays data in a TUI table

### Environment Variables Required

For Claude Code to send telemetry to ccmon:
```bash
export CLAUDE_CODE_ENABLE_TELEMETRY=1
export OTEL_METRICS_EXPORTER=otlp
export OTEL_LOGS_EXPORTER=otlp
export OTEL_EXPORTER_OTLP_PROTOCOL=grpc
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
```

## Model Identification Logic

- Base models (not counted against limits): Identified by checking if model name contains "haiku" (case-insensitive)
- Premium models (counted against limits): All other models (Sonnet, Opus, etc.)

## Important Implementation Details

- All numeric values from Claude Code telemetry are sent as strings and must be parsed using `fmt.Sscanf()`
- Monitor mode refreshes from database every 5 seconds
- Server mode logs each request to console for visibility
- Statistics are tracked separately for base/premium tiers and combined totals
- The gRPC server runs on port 4317 (standard OTLP port) in server mode only
- Table height is dynamically adjusted based on terminal size in monitor mode
- Multiple monitors can connect to the same database file
- Database limits stored requests to last 10,000 entries

## Development Conventions

- Always write devlog in `docs/devlog/` use Markdown format, group by day. e.g. 20250628.md