# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ccmon is a TUI (Terminal User Interface) application that monitors Claude Code API usage by receiving OpenTelemetry (OTLP) telemetry data. It displays real-time statistics for token usage, costs, and request counts, with separate tracking for base (Haiku) and premium (Sonnet/Opus) models.

## Build and Development Commands

```bash
# Build the application
go build -o ccmon .

# Format code
gofmt -w .

# Run the application
./ccmon

# Clean build artifacts
rm ccmon
```

## Architecture

The application follows a modular architecture with clear separation of concerns:

1. **main.go** - Entry point that initializes the Bubble Tea program, creates channels for communication, and starts the OTLP receiver
2. **receiver.go** - OTLP gRPC server that receives telemetry data from Claude Code and parses API request information
3. **model.go** - Bubble Tea model implementing the Elm architecture pattern, manages application state and handles updates
4. **ui.go** - Rendering logic using Lipgloss for styled terminal output

### Key Design Patterns

- **Channel-based Communication**: Uses Go channels to pass `APIRequest` structs from the OTLP receiver to the TUI
- **Elm Architecture**: Bubble Tea's Update/View pattern for reactive UI updates
- **Model Tier Separation**: Distinguishes between base models (Haiku) and premium models (Sonnet/Opus) using the `isBaseModel()` function

### Data Flow

1. Claude Code sends OTLP telemetry data to port 4317
2. The receiver parses log records with body "claude_code.api_request"
3. Extracted data is sent through a channel to the TUI model
4. The model updates statistics and the table view
5. The UI renders the updated state

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
- The application maintains a rolling buffer of the last 100 requests
- Statistics are tracked separately for base/premium tiers and combined totals
- The gRPC server runs on port 4317 (standard OTLP port)
- Table height is dynamically adjusted based on terminal size

## Development Conventions

- Always write devlog in `docs/devlog/` use Markdown format, group by day. e.g. 20250628.md