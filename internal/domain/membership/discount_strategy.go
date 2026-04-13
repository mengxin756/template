package membership

import (
	"fmt"
)

// Money money value object
type Money struct {
	amount   float64
	currency string
}

// NewMoney creates money value object
func NewMoney(amount float64, currency string) (*Money, error) {
	if amount < 0 {
		return nil, fmt.Errorf("money amount cannot be negative")
	}
	if currency == "" {
		currency = "CNY"
	}
	return &Money{amount: amount, currency: currency}, nil
}

// Amount returns amount
func (m Money) Amount() float64 {
	return m.amount
}

// Currency returns currency
func (m Money) Currency() string {
	return m.currency
}

// Add adds money
func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("cannot add different currencies")
	}
	return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}

// Subtract subtracts money
func (m Money) Subtract(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("cannot subtract different currencies")
	}
	if m.amount < other.amount {
		return Money{}, fmt.Errorf("insufficient amount")
	}
	return Money{amount: m.amount - other.amount, currency: m.currency}, nil
}

// Multiply multiplies money by factor
func (m Money) Multiply(factor float64) Money {
	return Money{amount: m.amount * factor, currency: m.currency}
}

// DiscountStrategy discount strategy interface (Strategy Pattern)
type DiscountStrategy interface {
	// CalculateDiscount calculates discount amount
	CalculateDiscount(originalPrice Money) (Money, error)
	// StrategyName returns strategy name
	StrategyName() string
}

// LevelDiscountStrategy level-based discount strategy
type LevelDiscountStrategy struct {
	level MemberLevel
}

// NewLevelDiscountStrategy creates level discount strategy
func NewLevelDiscountStrategy(level MemberLevel) *LevelDiscountStrategy {
	return &LevelDiscountStrategy{level: level}
}

// CalculateDiscount calculates discount based on level
func (s *LevelDiscountStrategy) CalculateDiscount(originalPrice Money) (Money, error) {
	config := s.level.Config()
	discountAmount := originalPrice.Amount() * (1 - config.DiscountRate)
	return Money{amount: discountAmount, currency: originalPrice.Currency()}, nil
}

// StrategyName returns strategy name
func (s *LevelDiscountStrategy) StrategyName() string {
	return fmt.Sprintf("level_discount_%s", s.level)
}

// CouponDiscountStrategy coupon-based discount strategy
type CouponDiscountStrategy struct {
	discountRate float64
	maxDiscount  float64
	couponCode   string
}

// NewCouponDiscountStrategy creates coupon discount strategy
func NewCouponDiscountStrategy(couponCode string, discountRate, maxDiscount float64) *CouponDiscountStrategy {
	return &CouponDiscountStrategy{
		discountRate: discountRate,
		maxDiscount:  maxDiscount,
		couponCode:   couponCode,
	}
}

// CalculateDiscount calculates discount based on coupon
func (s *CouponDiscountStrategy) CalculateDiscount(originalPrice Money) (Money, error) {
	discountAmount := originalPrice.Amount() * (1 - s.discountRate)
	if discountAmount > s.maxDiscount {
		discountAmount = s.maxDiscount
	}
	return Money{amount: discountAmount, currency: originalPrice.Currency()}, nil
}

// StrategyName returns strategy name
func (s *CouponDiscountStrategy) StrategyName() string {
	return fmt.Sprintf("coupon_discount_%s", s.couponCode)
}

// FixedAmountDiscountStrategy fixed amount discount strategy
type FixedAmountDiscountStrategy struct {
	discountAmount float64
}

// NewFixedAmountDiscountStrategy creates fixed amount discount strategy
func NewFixedAmountDiscountStrategy(discountAmount float64) *FixedAmountDiscountStrategy {
	return &FixedAmountDiscountStrategy{discountAmount: discountAmount}
}

// CalculateDiscount returns fixed discount amount
func (s *FixedAmountDiscountStrategy) CalculateDiscount(originalPrice Money) (Money, error) {
	if s.discountAmount > originalPrice.Amount() {
		return Money{amount: originalPrice.Amount(), currency: originalPrice.Currency()}, nil
	}
	return Money{amount: s.discountAmount, currency: originalPrice.Currency()}, nil
}

// StrategyName returns strategy name
func (s *FixedAmountDiscountStrategy) StrategyName() string {
	return "fixed_amount_discount"
}

// CompositeDiscountStrategy composite strategy (stack multiple discounts)
type CompositeDiscountStrategy struct {
	strategies []DiscountStrategy
	maxTotalDiscount float64
}

// NewCompositeDiscountStrategy creates composite discount strategy
func NewCompositeDiscountStrategy(maxTotalDiscount float64, strategies ...DiscountStrategy) *CompositeDiscountStrategy {
	return &CompositeDiscountStrategy{
		strategies:      strategies,
		maxTotalDiscount: maxTotalDiscount,
	}
}

// CalculateDiscount calculates total discount from all strategies
func (s *CompositeDiscountStrategy) CalculateDiscount(originalPrice Money) (Money, error) {
	totalDiscount := Money{amount: 0, currency: originalPrice.Currency()}

	for _, strategy := range s.strategies {
		discount, err := strategy.CalculateDiscount(originalPrice)
		if err != nil {
			continue // skip failed strategy
		}
		totalDiscount, _ = totalDiscount.Add(discount)
	}

	// cap at max total discount
	if totalDiscount.Amount() > s.maxTotalDiscount {
		totalDiscount = Money{amount: s.maxTotalDiscount, currency: originalPrice.Currency()}
	}

	// cannot exceed original price
	if totalDiscount.Amount() > originalPrice.Amount() {
		totalDiscount = originalPrice
	}

	return totalDiscount, nil
}

// StrategyName returns strategy name
func (s *CompositeDiscountStrategy) StrategyName() string {
	return "composite_discount"
}

// PricingService pricing domain service
type PricingService interface {
	// CalculateFinalPrice calculates final price after all discounts
	CalculateFinalPrice(originalPrice Money, strategies ...DiscountStrategy) (Money, Money, error)
}

// pricingServiceImpl pricing service implementation
type pricingServiceImpl struct{}

// NewPricingService creates pricing service
func NewPricingService() PricingService {
	return &pricingServiceImpl{}
}

// CalculateFinalPrice calculates final price
func (s *pricingServiceImpl) CalculateFinalPrice(originalPrice Money, strategies ...DiscountStrategy) (Money, Money, error) {
	if len(strategies) == 0 {
		return originalPrice, Money{amount: 0, currency: originalPrice.Currency()}, nil
	}

	composite := NewCompositeDiscountStrategy(originalPrice.Amount(), strategies...)
	totalDiscount, err := composite.CalculateDiscount(originalPrice)
	if err != nil {
		return Money{}, Money{}, err
	}

	finalPrice, err := originalPrice.Subtract(totalDiscount)
	if err != nil {
		return Money{}, Money{}, err
	}

	return finalPrice, totalDiscount, nil
}
