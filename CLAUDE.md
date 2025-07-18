# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ccmon is a TUI (Terminal User Interface) application that monitors Claude Code API usage by receiving OpenTelemetry (OTLP) telemetry data. It displays real-time statistics for token usage, costs, and request counts, with separate tracking for base (Haiku) and premium (Sonnet/Opus) models.

## Documentation

- The `docs/glossary.md` is defined our domain concepts

## Architecture Conventions

### Directory Structure

- `entity/` - Domain entities and business rules
- `usecase/` - Business logic and application services with interface definitions
- `repository/` - Data access implementations  
- `handler/` - External interfaces (TUI, gRPC, CLI)
- `service/` - Infrastructure services and non-business logic implementations (e.g., time handling, external service adapters)

## Operating Modes

ccmon has two distinct operating modes:

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

[... rest of the file remains unchanged ...]