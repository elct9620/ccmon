# Unsupported Event Logging Feature Tasks

- [ ] 1. Add unsupported event logging to OTLP receiver Export method
  - Modify `handler/grpc/receiver/receiver.go` logsReceiver.Export() method
  - Add logging for non-API request string body values that are not empty
  - Use INFO level with format: "Unsupported log event: [event_type]"
  - Skip empty, nil, or non-string body values silently (existing behavior)
  - Ensure existing "claude_code.api_request" processing remains unchanged
  - _Requirements: R1

- [ ] 2. Add test cases for unsupported event logging scenarios
  - Modify `handler/grpc/receiver/receiver_test.go` to include unsupported event tests
  - Test case: Log with string body "custom.event" should log "Unsupported log event: custom.event"
  - Test case: Log with string body "system.monitoring" should log "Unsupported log event: system.monitoring"
  - Test case: Log with empty string body should be skipped silently
  - Test case: Log with nil body should be skipped silently
  - Test case: Log with non-string body (int/bool) should be skipped silently
  - Test case: Log with "claude_code.api_request" should process normally (regression test)
  - Verify correct log output format and INFO level
  - _Requirements: R1

- [ ] 3. Verify integration and validate logging behavior
  - Run existing receiver tests to ensure no regressions in API request processing
  - Manually test unsupported event logging by sending OTLP data with different body types
  - Confirm INFO level logs appear in output with correct format
  - Validate that empty/nil/non-string bodies are handled silently
  - Ensure zero performance impact on existing claude_code.api_request processing
  - _Requirements: R1