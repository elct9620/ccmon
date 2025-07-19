# Quick Query Feature Tasks

## Overview
Implementation of command-line quick query functionality for ccmon, enabling users to retrieve usage metrics via format strings with predefined variables (e.g., `ccmon -format '@daily_cost'`).

## Task List

- [x] 1. Add format flag to main CLI interface
  - Extend main.go flag parsing to include `-format` flag
  - Create basic command structure that accepts format string and exits without TUI
  - Output hardcoded "Format: {input}" to verify flag parsing works
  - Ensure format query bypasses TUI initialization and outputs directly to stdout
  - _Requirements: R1

- [x] 2. Create CLI query handler with hardcoded responses
  - Implement `handler/cli/query_handler.go` with QueryHandler struct
  - Add HandleFormatQuery method that outputs hardcoded values for testing
  - Return "$10.0" for @daily_cost, "$150.0" for @monthly_cost, "50%" for percentages
  - Implement basic error output "❌ ERROR" for connection simulation
  - _Requirements: R1, R2, R3

- [ ] 3. Implement format renderer with variable substitution
  - Create `handler/cli/format_renderer.go` with FormatRenderer struct
  - Add Render method that processes format strings and substitutes variables
  - Support @daily_cost, @monthly_cost, @daily_plan_usage, @monthly_plan_usage variables
  - Use hardcoded values initially, format currency as USD with one decimal place
  - _Requirements: R1, R2, R3

- [ ] 4. Connect to existing gRPC service for real cost data
  - Wire up FormatRenderer with existing CalculateStatsQuery
  - Replace hardcoded @daily_cost and @monthly_cost with real data from gRPC
  - Add dependency injection for stats query in QueryHandler
  - Ensure time zone consistency with existing TUI calculations
  - _Requirements: R1, R2, R5

- [ ] 5. Add plan configuration to config system
  - Extend existing config.toml structure to include `[claude]` section with `plan` field
  - Update configuration reading logic to support plan setting
  - Create ConfigReader interface for accessing plan configuration
  - Ensure default behavior when plan is not configured (return "0%" for percentages)
  - _Requirements: R4

- [ ] 6. Create Plan entity with business rules
  - Implement `entity/plan.go` with Plan struct, NewPlan constructor, and core methods
  - Add validation logic for plan names and price calculations
  - Include CalculateUsagePercentage method for percentage calculations
  - Support plan types: unset, pro, max, max20 with respective prices
  - _Requirements: R4

- [ ] 7. Create embedded plans data file
  - Create `data/plans.json` with plan definitions (unset, pro, max, max20)
  - Configure go:embed directive in main.go to embed data directory
  - Ensure JSON structure matches design specification with name and price fields
  - Test embedded data loading during application startup
  - _Requirements: R4

- [ ] 8. Implement plan repository with embedded data
  - Create `repository/embedded_plan_repository.go` with EmbeddedPlanRepository struct
  - Implement GetConfiguredPlan method reading from config and embedded JSON
  - Add error handling for missing/invalid plan configurations
  - Include dependency injection for ConfigReader interface
  - _Requirements: R4

- [ ] 9. Implement Calculate Percentage Query usecase
  - Create `usecase/calculate_percentage_query.go` with CalculatePercentageQuery struct
  - Add ExecuteDaily and ExecuteMonthly methods for percentage calculations
  - Integrate with existing CalculateStatsQuery and new GetPlanQuery
  - Replace hardcoded percentage values in FormatRenderer with real calculations
  - _Requirements: R2, R4

- [ ] 10. Implement Get Plan Query usecase
  - Create `usecase/get_plan_query.go` with GetPlanQuery struct
  - Add Execute method that retrieves configured plan via repository
  - Define PlanRepository interface for dependency injection
  - Wire up with CalculatePercentageQuery for complete percentage functionality
  - _Requirements: R4

- [ ] 11. Add comprehensive error handling
  - Enhance QueryHandler to handle all connection and data retrieval errors
  - Ensure graceful degradation with "❌ ERROR" output for all failure scenarios
  - Add timeout handling for gRPC queries
  - Test error scenarios: server down, invalid configuration, network issues
  - _Requirements: R1, R2, R3, R5

- [ ] 12. Create integration tests for end-to-end functionality
  - Test complete format query execution with real gRPC responses
  - Verify all supported variables work correctly in various combinations
  - Test time zone consistency with existing TUI calculations
  - Validate output format matches specification requirements exactly
  - _Requirements: R1, R2, R3, R5

## Requirements Reference
- R1: Basic usage information query with simple command
- R2: Multiple variables in single query
- R3: Predefined variable support (@daily_cost, @monthly_cost, etc.)
- R4: Plan configuration support (Pro, Max, Max20)
- R5: Time zone consistency with TUI interface