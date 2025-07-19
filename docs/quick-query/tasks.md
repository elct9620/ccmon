# Quick Query Feature Tasks

## Overview
Implementation of command-line quick query functionality for ccmon, enabling users to retrieve usage metrics via format strings with predefined variables (e.g., `ccmon --format '@daily_cost'`).

## Task List

- [x] 1. Add format flag to main CLI interface
  - Extend main.go flag parsing to include `--format` flag
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

- [x] 8. Extend config system for plan configuration
  - Update config.toml structure to include `[claude]` section with `plan` field
  - Modify Config struct to support Claude.Plan string field
  - Ensure graceful handling when plan is not configured (default to "unset")
  - Validate config parsing works with existing configuration system
  - _Requirements: R4

- [x] 9. Implement Get Plan Query usecase
  - Create `usecase/get_plan_query.go` with GetPlanQuery struct and Execute method
  - Add PlanRepository interface for dependency injection pattern
  - Connect GetPlanQuery with EmbeddedPlanRepository for plan retrieval
  - Ensure error handling for missing/invalid plan configurations
  - _Requirements: R4

- [x] 10. Create usage variable entity with predefined variables
  - Implement `entity/usage_variable.go` with UsageVariable struct
  - Define all predefined variables as constants (DailyCostVariable, etc.)
  - Add GetAllUsageVariables() function to list all available variables
  - Include Key() and Name() methods for accessing variable properties
  - _Requirements: R3

- [x] 11. Implement GetUsageVariablesQuery usecase
  - Create `usecase/get_usage_variables_query.go` with GetUsageVariablesQuery struct
  - Add Execute method that returns map[string]string for variable substitution
  - Wire dependencies: CalculateStatsQuery, PlanRepository, PeriodFactory
  - Implement generateVariableMap helper to format all variable values
  - _Requirements: R2, R3, R4

- [x] 12. Wire up percentage calculations in GetUsageVariablesQuery
  - Update GetUsageVariablesQuery to calculate plan usage percentages
  - Use Plan entity's CalculateUsagePercentage method for calculations
  - Format percentages as integers (e.g., "50%" not "50.0%")
  - Return "0%" for unset/invalid plans as specified
  - _Requirements: R2, R4

- [x] 13. Update format renderer to use GetUsageVariablesQuery
  - Replace hardcoded values and stats query in FormatRenderer with GetUsageVariablesQuery
  - Update constructor to accept usageVariablesQuery dependency
  - Modify Render method to call query.Execute() for variable map
  - Ensure substituteVariables uses the real variable map from query
  - _Requirements: R1, R2, R3

- [x] 14. Wire up complete dependency graph in main.go
  - Create PeriodFactory instance for daily/monthly period generation
  - Initialize GetUsageVariablesQuery with all required dependencies
  - Update FormatRenderer and QueryHandler initialization
  - Ensure proper error handling flow from query to output
  - _Requirements: R1, R4, R5

- [x] 15. Add comprehensive error handling and timeout management
  - Enhance QueryHandler to handle all connection and data retrieval errors
  - Implement timeout handling for gRPC queries (prevent hanging)
  - Ensure graceful degradation with "❌ ERROR" output for all failure scenarios
  - Test error scenarios: server down, invalid configuration, network timeouts
  - _Requirements: R1, R2, R3, R5

- [x] 16. Create integration tests for end-to-end functionality
  - Test complete format query execution with real gRPC responses
  - Verify all supported variables work correctly in various combinations
  - Test time zone consistency with existing TUI calculations
  - Validate output format matches specification requirements exactly
  - _Requirements: R1, R2, R3, R5

## Formula Update Tasks

- [x] 17. Update daily plan usage calculation formula in GetUsageVariablesQuery
  - Modify `generateVariableMap` method in `usecase/get_usage_variables_query.go`
  - Change daily plan usage calculation from `(dailyCost / planPrice) * 100` to `(dailyCost / (planPrice / daysInMonth)) * 100`
  - Get current month days count using `time.Now().AddDate(0, 1, -time.Now().Day()).Day()`
  - Keep monthly plan usage calculation unchanged (continue using existing `CalculateUsagePercentage`)
  - Ensure "0%" is returned for unset/invalid plans
  - _Requirements: R3 (Updated daily plan usage formula)

- [x] 18. Update tests to verify new daily plan usage calculation
  - Update existing tests for `GetUsageVariablesQuery` to validate new formula
  - Test Pro plan examples: $1.0 daily cost with 31 days = 155%, $2.0 with 28 days = 280%
  - Test different month lengths (28, 29, 30, 31 days)
  - Verify monthly calculation remains unchanged
  - Ensure edge cases work: unset plan returns 0%, zero costs, etc.
  - _Requirements: R3

## Requirements Reference
- R1: Basic usage information query with simple command
- R2: Multiple variables in single query
- R3: Predefined variable support (@daily_cost, @monthly_cost, etc.) - **Updated with new daily plan usage formula: Daily Cost / (Plan Price / Days in Current Month)**
- R4: Plan configuration support (Pro, Max, Max20)
- R5: Time zone consistency with TUI interface