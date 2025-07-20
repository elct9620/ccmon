package entity

import "time"

type Plan struct {
	name  string
	price Cost
}

func NewPlan(name string, price Cost) Plan {
	return Plan{
		name:  name,
		price: price,
	}
}

func (p Plan) Name() string {
	return p.name
}

func (p Plan) Price() Cost {
	return p.price
}

func (p Plan) IsValid() bool {
	validPlans := map[string]bool{
		"unset": true,
		"pro":   true,
		"max":   true,
		"max20": true,
	}
	return validPlans[p.name]
}

func (p Plan) CalculateUsagePercentage(actualCost Cost) int {
	if !p.IsValid() || p.price.Amount() == 0 {
		return 0
	}

	percentage := (actualCost.Amount() / p.price.Amount()) * 100
	return int(percentage)
}

// CalculateUsagePercentageInPeriod calculates the percentage of period budget used
// based on the actual cost for a specific period and the plan's period budget
func (p Plan) CalculateUsagePercentageInPeriod(actualCost Cost, period Period) int {
	if !p.IsValid() || p.price.Amount() == 0 {
		return 0
	}

	// Calculate days in the month that contains the period start time
	periodStart := period.StartAt()
	daysInMonth := time.Date(periodStart.Year(), periodStart.Month()+1, 0, 0, 0, 0, 0, periodStart.Location()).Day()

	// Calculate period budget (plan price / days in month)
	periodBudget := p.price.Amount() / float64(daysInMonth)

	// Calculate percentage: (actual cost / period budget) * 100
	percentage := (actualCost.Amount() / periodBudget) * 100
	return int(percentage)
}
