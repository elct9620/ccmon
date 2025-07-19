package entity

// UsageVariable represents a predefined variable for quick query formatting
type UsageVariable struct {
	name string
	key  string
}

// Predefined variables for usage queries
var (
	DailyCostVariable        = UsageVariable{name: "Daily Cost", key: "@daily_cost"}
	MonthlyCostVariable      = UsageVariable{name: "Monthly Cost", key: "@monthly_cost"}
	DailyPlanUsageVariable   = UsageVariable{name: "Daily Plan Usage", key: "@daily_plan_usage"}
	MonthlyPlanUsageVariable = UsageVariable{name: "Monthly Plan Usage", key: "@monthly_plan_usage"}
)

// GetAllUsageVariables returns all available predefined variables
func GetAllUsageVariables() []UsageVariable {
	return []UsageVariable{
		DailyCostVariable,
		MonthlyCostVariable,
		DailyPlanUsageVariable,
		MonthlyPlanUsageVariable,
	}
}

// Key returns the variable key used in format strings (e.g., "@daily_cost")
func (v UsageVariable) Key() string {
	return v.key
}

// Name returns the human-readable name of the variable
func (v UsageVariable) Name() string {
	return v.name
}
