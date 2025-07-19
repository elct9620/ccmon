# Unsupported Event Logging Feature Requirements

## 1. Log Non-API Request Events

As a ccmon developer
I want to log non-API request log events from OTLP data
So that I can analyze which other event types might be valuable for future features

```gherkin
Scenario: Log with unsupported body value
  Given ccmon server is running in server mode
  And OTLP receiver receives a log with string body value other than "claude_code.api_request"
  When the receiver processes the log
  Then it should log "Unsupported log event: [actual_body_value]"
```

## Technical Notes

- **Log Level**: Use INFO level for unsupported event logging
- **Log Format**: "Unsupported log event: [event_type]"
- **Skip Conditions**: Empty body values, non-string body values, parsing errors
- **Purpose**: Analysis only - to identify potential new event types to support
- **Current Supported Event**: Only "claude_code.api_request" body values are processed