package entity

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
