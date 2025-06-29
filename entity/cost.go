package entity

// Cost represents a monetary cost value object
type Cost struct {
	amount float64
}

// NewCost creates a new Cost value object
func NewCost(amount float64) Cost {
	return Cost{amount: amount}
}

// Amount returns the cost amount in USD
func (c Cost) Amount() float64 {
	return c.amount
}

// Add returns a new Cost with the sum of two costs
func (c Cost) Add(other Cost) Cost {
	return Cost{amount: c.amount + other.amount}
}
