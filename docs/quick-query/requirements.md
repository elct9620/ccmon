# Quick Query Feature Requirements

## 1. Basic Usage Information Query

As a Claude Code user
I want to quickly query my usage information with a simple command
So that I can monitor my API costs without opening the full TUI interface

```gherkin
Feature: Quick usage information query

  Scenario: Query daily cost with simple format
    Given ccmon server is running with usage data
    When I run "ccmon --format '@daily_cost'"
    Then I should see output like "$10.0"

  Scenario: Query with custom text and emoji
    Given ccmon server is running with usage data
    When I run "ccmon --format '💰 @daily_cost'"
    Then I should see output like "💰 $10.0"

  Scenario: Server connection error
    Given ccmon server is not accessible
    When I run "ccmon --format '@daily_cost'"
    Then I should see "❌ ERROR"
```

## 2. Multiple Variables in Single Query

As a Claude Code user
I want to combine multiple usage metrics in one query
So that I can see comprehensive information in my status bar or tmux panel

```gherkin
Feature: Multiple variables in format string

  Scenario: Combine cost and percentage
    Given ccmon server is running with usage data
    And user has Pro plan configured ($20)
    When I run "ccmon --format '💰 @daily_cost / @daily_plan_usage'"
    Then I should see output like "💰 $10.0 / 50%"

  Scenario: Multiple daily metrics
    Given ccmon server is running with usage data
    When I run "ccmon --format 'Daily: @daily_cost Monthly: @monthly_cost'"
    Then I should see output like "Daily: $10.0 Monthly: $150.0"
```

## 3. Predefined Variable Support

As a Claude Code user
I want to use predefined variables for common metrics
So that I can easily access frequently needed information

```gherkin
Feature: Predefined variables

  Scenario: Daily cost variable
    Given ccmon server has daily usage data
    When I use "@daily_cost" variable
    Then it should return today's total cost in USD format

  Scenario: Monthly cost variable
    Given ccmon server has monthly usage data
    When I use "@monthly_cost" variable
    Then it should return current month's total cost in USD format

  Scenario: Daily plan usage percentage
    Given ccmon server has daily usage data
    And user has plan configured in settings
    When I use "@daily_plan_usage" variable
    Then it should return daily cost as percentage of plan price

  Scenario: Monthly plan usage percentage
    Given ccmon server has monthly usage data
    And user has plan configured in settings
    When I use "@monthly_plan_usage" variable
    Then it should return monthly cost as percentage of plan price

  Scenario: No plan configured
    Given user has no plan configured in settings
    When I use "@daily_plan_usage" or "@monthly_plan_usage"
    Then it should return "0%"
```

## 4. Plan Configuration Support

As a Claude Code user
I want ccmon to read my plan configuration from settings
So that percentage calculations are accurate for my subscription

```gherkin
Feature: Plan configuration

  Scenario: Pro plan configuration
    Given user has "Pro" plan configured in settings
    When calculating plan usage percentages
    Then it should use $20 as the plan price

  Scenario: Max plan configuration
    Given user has "Max" plan configured in settings
    When calculating plan usage percentages
    Then it should use $100 as the plan price

  Scenario: Max20 plan configuration
    Given user has "Max20" plan configured in settings
    When calculating plan usage percentages
    Then it should use $200 as the plan price

  Scenario: Invalid or missing plan
    Given user has no plan or invalid plan configured
    When calculating plan usage percentages
    Then percentage should always be "0%"
```

## 5. Time Zone Consistency

As a Claude Code user
I want the quick query to use the same time zone logic as the TUI
So that the data is consistent between interfaces

```gherkin
Feature: Time zone consistency

  Scenario: Daily calculation consistency
    Given TUI shows specific daily cost
    When I run quick query for "@daily_cost"
    Then it should show the same value as TUI daily cost

  Scenario: Monthly calculation consistency
    Given TUI shows specific monthly cost
    When I run quick query for "@monthly_cost"
    Then it should show the same value as TUI monthly cost

  Scenario: User time zone respect
    Given user is in a specific time zone
    When calculating daily metrics
    Then it should use user's local 00:00-23:59 for "today"

  Scenario: Month boundary respect
    Given user is in a specific time zone
    When calculating monthly metrics
    Then it should use 1st to last day of current month in user's time zone
```

## Technical Notes

- **Data Source**: Use existing gRPC query service
- **Error Handling**: Display "❌ ERROR" for any connection or data retrieval failures
- **No Caching**: Real-time queries without local caching
- **Command Flag**: Consider `-format` vs `-print` for final implementation
- **Plan Prices**: Pro=$20, Max=$100, Max20=$200
- **Percentage Format**: Show as integer percentage (e.g., "50%" not "50.0%")
- **Currency Format**: USD with one decimal place (e.g., "$10.0")
