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

- [x] 3. Implement format renderer with variable substitution
  - Create `handler/cli/format_renderer.go` with FormatRenderer struct
  - Add Render method that processes format strings and substitutes variables
  - Support @daily_cost, @monthly_cost, @daily_plan_usage, @monthly_plan_usage variables
  - Use hardcoded values initially, format currency as USD with one decimal place
  - _Requirements: R1, R2, R3

- [x] 4. Connect to existing gRPC service for real cost data
  - Wire up FormatRenderer with existing CalculateStatsQuery
  - Replace hardcoded @daily_cost and @monthly_cost with real data from gRPC
  - Add dependency injection for stats query in QueryHandler
  - Ensure time zone consistency with existing TUI calculations
  - _Requirements: R1, R2, R5

- [x] 5. Create embedded plans data file with go:embed setup
  - Create `data/plans.json` with plan definitions (unset=0, pro=20, max=100, max20=200)
  - Add go:embed directive in main.go to embed data directory into binary
  - Ensure JSON structure matches design specification with name and price fields
  - Test embedded data loading during application startup
  - _Requirements: R4

- [x] 6. Create Plan entity with business rules
  - Implement `entity/plan.go` with Plan struct, NewPlan constructor, and core methods
  - Add Name(), Price(), IsValid(), and CalculateUsagePercentage() methods
  - Include validation logic for plan names and percentage calculations
  - Support plan types: unset, pro, max, max20 with respective USD prices
  - _Requirements: R4

- [x] 7. Implement plan repository with embedded data access
  - Create `repository/embedded_plan_repository.go` with EmbeddedPlanRepository struct
  - Implement GetConfiguredPlan method reading from config and embedded JSON
  - Add PlanRepository interface and implement JSON parsing logic
  - Include dependency injection for Config struct and embed.FS
  - _Requirements: R4

- [ ] 8. Extend config system for plan configuration
  - Update config.toml structure to include `[claude]` section with `plan` field
  - Modify Config struct to support Claude.Plan string field
  - Ensure graceful handling when plan is not configured (default to "unset")
  - Validate config parsing works with existing configuration system
  - _Requirements: R4

- [ ] 9. Implement Get Plan Query usecase
  - Create `usecase/get_plan_query.go` with GetPlanQuery struct and Execute method
  - Add PlanRepository interface for dependency injection pattern
  - Connect GetPlanQuery with EmbeddedPlanRepository for plan retrieval
  - Ensure error handling for missing/invalid plan configurations
  - _Requirements: R4

- [ ] 10. Implement Calculate Percentage Query usecase
  - Create `usecase/calculate_percentage_query.go` with CalculatePercentageQuery struct
  - Add ExecuteDaily and ExecuteMonthly methods for percentage calculations
  - Integrate with GetPlanQuery and existing CalculateStatsQuery
  - Return integer percentages (e.g., "50%" not "50.0%") as per specification
  - _Requirements: R2, R4

- [ ] 11. Wire up plan functionality in format renderer
  - Replace hardcoded percentage values in FormatRenderer with CalculatePercentageQuery
  - Add dependency injection for CalculatePercentageQuery in FormatRenderer constructor
  - Update variable substitution to use real plan calculations
  - Ensure 0% return for unset/invalid plans as specified
  - _Requirements: R2, R4

- [ ] 12. Add comprehensive error handling and timeout management
  - Enhance QueryHandler to handle all connection and data retrieval errors
  - Implement timeout handling for gRPC queries (prevent hanging)
  - Ensure graceful degradation with "❌ ERROR" output for all failure scenarios
  - Test error scenarios: server down, invalid configuration, network timeouts
  - _Requirements: R1, R2, R3, R5

- [ ] 13. Create integration tests for end-to-end functionality
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