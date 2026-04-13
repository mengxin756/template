package wallet

import (
	"fmt"
)

// Currency currency enum
type Currency string

const (
	CurrencyCNY  Currency = "CNY"
	CurrencyUSD  Currency = "USD"
	CurrencyEUR  Currency = "EUR"
)

// IsValid validates currency
func (c Currency) IsValid() bool {
	switch c {
	case CurrencyCNY, CurrencyUSD, CurrencyEUR:
		return true
	default:
		return false
	}
}

// String returns string representation
func (c Currency) String() string {
	return string(c)
}

// Money money value object (immutable)
type Money struct {
	amount   float64
	currency Currency
}

// NewMoney creates money value object
func NewMoney(amount float64, currency Currency) (*Money, error) {
	if amount < 0 {
		return nil, fmt.Errorf("money amount cannot be negative")
	}
	if !currency.IsValid() {
		return nil, fmt.Errorf("invalid currency: %s", currency)
	}
	return &Money{amount: amount, currency: currency}, nil
}

// Amount returns amount
func (m Money) Amount() float64 {
	return m.amount
}

// Currency returns currency
func (m Money) Currency() Currency {
	return m.currency
}

// IsZero checks if money is zero
func (m Money) IsZero() bool {
	return m.amount == 0
}

// Add adds money (same currency only)
func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("cannot add different currencies: %s and %s", m.currency, other.currency)
	}
	return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}

// Subtract subtracts money
func (m Money) Subtract(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("cannot subtract different currencies: %s and %s", m.currency, other.currency)
	}
	if m.amount < other.amount {
		return Money{}, fmt.Errorf("insufficient balance: have %.2f, need %.2f", m.amount, other.amount)
	}
	return Money{amount: m.amount - other.amount, currency: m.currency}, nil
}

// Multiply multiplies money by factor
func (m Money) Multiply(factor float64) Money {
	return Money{amount: m.amount * factor, currency: m.currency}
}

// Equals checks equality
func (m Money) Equals(other Money) bool {
	return m.amount == other.amount && m.currency == other.currency
}

// String returns string representation
func (m Money) String() string {
	return fmt.Sprintf("%.2f %s", m.amount, m.currency)
}

// TransactionType transaction type enum
type TransactionType string

const (
	TransactionTypeDeposit    TransactionType = "deposit"    // deposit
	TransactionTypeWithdraw   TransactionType = "withdraw"   // withdraw
	TransactionTypePayment    TransactionType = "payment"    // payment for order
	TransactionTypeRefund     TransactionType = "refund"     // refund
	TransactionTypeTransferIn TransactionType = "transfer_in"  // transfer in
	TransactionTypeTransferOut TransactionType = "transfer_out" // transfer out
	TransactionTypeBonus      TransactionType = "bonus"      // bonus/reward
	TransactionTypeAdjust     TransactionType = "adjust"    // manual adjustment
	TransactionTypeFreeze     TransactionType = "freeze"    // freeze amount
	TransactionTypeUnfreeze   TransactionType = "unfreeze"  // unfreeze amount
)

// IsValid validates transaction type
func (t TransactionType) IsValid() bool {
	switch t {
	case TransactionTypeDeposit, TransactionTypeWithdraw, TransactionTypePayment,
		TransactionTypeRefund, TransactionTypeTransferIn, TransactionTypeTransferOut,
		TransactionTypeBonus, TransactionTypeAdjust, TransactionTypeFreeze, TransactionTypeUnfreeze:
		return true
	default:
		return false
	}
}

// String returns string representation
func (t TransactionType) String() string {
	return string(t)
}

// IsInflow checks if transaction is inflow (positive)
func (t TransactionType) IsInflow() bool {
	switch t {
	case TransactionTypeDeposit, TransactionTypeRefund, TransactionTypeTransferIn,
		TransactionTypeBonus, TransactionTypeUnfreeze:
		return true
	default:
		return false
	}
}

// IsOutflow checks if transaction is outflow (negative)
func (t TransactionType) IsOutflow() bool {
	switch t {
	case TransactionTypeWithdraw, TransactionTypePayment, TransactionTypeTransferOut,
		TransactionTypeFreeze:
		return true
	default:
		return false
	}
}

// TransactionStatus transaction status enum
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"   // pending
	TransactionStatusCompleted TransactionStatus = "completed" // completed
	TransactionStatusFailed    TransactionStatus = "failed"    // failed
	TransactionStatusCancelled TransactionStatus = "cancelled" // cancelled
)

// IsValid validates transaction status
func (s TransactionStatus) IsValid() bool {
	switch s {
	case TransactionStatusPending, TransactionStatusCompleted, TransactionStatusFailed, TransactionStatusCancelled:
		return true
	default:
		return false
	}
}

// String returns string representation
func (s TransactionStatus) String() string {
	return string(s)
}
