# ccmon Domain Glossary

This glossary defines terms and concepts used in the ccmon project to ensure consistent terminology and make domain concepts clear to all contributors and users.

## Core Entities

| Term | Definition |
|------|------------|
| APIRequest | A domain entity representing a single Claude Code API request with session ID, timestamp, model, token usage, cost, and duration. Forms the core of the telemetry data collected from Claude Code interactions. |
| Token | A value object representing token usage for an API request, containing input tokens, output tokens, cache read tokens, and cache creation tokens. Provides methods for calculating totals and limited tokens. |
| Cost | A value object representing monetary cost in USD for API requests. Immutable with operations returning new instances. |
| Model | A value object identifying the AI model used (e.g., "claude-3-haiku", "claude-3-sonnet"). Distinguishes between base models (Haiku) and premium models (Sonnet/Opus) for billing and limit purposes. |
| Period | A value object representing a time range with start and end times. Used for filtering API requests and calculating statistics over specific timeframes. |
| Stats | An aggregate entity containing usage statistics for both base and premium models within a specific period, including request counts, token usage, costs, and burn rate calculations. |
| Usage | An entity representing usage statistics grouped by periods, typically used for daily usage displays and historical analysis. |
| Block | A value object representing Claude's 5-hour token limit blocks with start time and optional token limit. Used for monitoring usage against subscription plan limits. |

## Business Concepts

| Term | Definition |
|------|------------|
| Base Models | Free AI models (currently Haiku) that don't count against Claude's token limits. Identified by checking if the model name contains "haiku" (case-insensitive). |
| Premium Models | Paid AI models (Sonnet, Opus, etc.) that count against Claude's token limits and subscription quotas. All models except Haiku are considered premium. |
| Token Limits | Claude subscription plans impose limits on premium token usage within 5-hour blocks: Pro (7,000), Max (35,000), Max20 (140,000) tokens per block. |
| Limited Tokens | Only input and output tokens count against Claude's rate limits. Cache tokens (read/creation) are excluded from limit calculations. |
| Burn Rate | The rate of premium token consumption per minute, calculated as limited tokens divided by period duration. Used to monitor API usage intensity. |
| Block Tracking | Monitoring API usage within Claude's 5-hour token limit blocks, showing progress bars and time remaining until the next block starts. |
| Session ID | A unique identifier for a Claude Code session, used to group related API requests and track usage patterns. |

## Technical Concepts

| Term | Definition |
|------|------------|
| OTLP | OpenTelemetry Protocol used by Claude Code to send usage data to ccmon via log records with body "claude_code.api_request". |
| gRPC Query Service | Service providing read-only access to stored telemetry data via Protocol Buffers interface, enabling multiple monitors to connect to a single server. |
| TUI | Terminal User Interface - the interactive terminal-based dashboard that displays real-time statistics with keyboard navigation and multiple views. |
| Server Mode | Headless operation mode running the OTLP collector and gRPC query service to receive and store telemetry data. |
| Monitor Mode | Interactive TUI mode that connects to a server via gRPC to display usage statistics and request details. |

## Cache Concepts

| Term | Definition |
|------|------------|
| Cache Read Tokens | Tokens consumed when reading from Claude's prompt cache, typically at reduced cost compared to input tokens. |
| Cache Creation Tokens | Tokens used when creating entries in Claude's prompt cache, enabling faster subsequent requests with similar context. |
| Cache Tokens | Combined cache read and creation tokens that don't count against rate limits but contribute to total usage statistics. |

## Display and UI Concepts

| Term | Definition |
|------|------------|
| Daily Usage | Historical view showing usage statistics aggregated by day, typically covering the last 30 days with detailed token breakdowns. |
| Current View | Real-time dashboard showing recent API requests, current statistics, and optional block progress tracking. |
| Display Modes | Different table layouts: FullMode (9 columns), GroupedMode (4 main columns), CompactMode (4 simplified columns for narrow terminals). |
| Time Filtering | Ability to filter displayed requests by time periods (1 hour, 24 hours, all time) while maintaining separate statistics for display and block tracking. |

## Repository and Data Concepts

| Term | Definition |
|------|------------|
| BoltDB | Embedded key-value database used for storing API request data locally, providing persistence without external dependencies. |
| Entity Conversion | Process of converting between database storage types and domain entities, handled by the repository layer to maintain clean architecture. |
| Request Limiting | Database queries support limit and offset parameters to manage memory usage and optimize network traffic between monitor and server. |