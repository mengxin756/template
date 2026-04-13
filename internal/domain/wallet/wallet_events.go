package wallet

import (
	"time"

	"example.com/classic/internal/domain"
)

// WalletDepositedEvent wallet deposited event
type WalletDepositedEvent struct {
	walletID      int
	userID        int
	amount        float64
	balance       float64
	currency      string
	referenceType string
	referenceID   string
	occurredAt    time.Time
}

func NewWalletDepositedEvent(walletID, userID int, amount, balance float64, currency, referenceType, referenceID string) *WalletDepositedEvent {
	return &WalletDepositedEvent{
		walletID:      walletID,
		userID:        userID,
		amount:        amount,
		balance:       balance,
		currency:      currency,
		referenceType: referenceType,
		referenceID:   referenceID,
		occurredAt:    time.Now(),
	}
}

func (e *WalletDepositedEvent) EventType() string     { return "wallet.deposited" }
func (e *WalletDepositedEvent) OccurredAt() time.Time { return e.occurredAt }
func (e *WalletDepositedEvent) AggregateID() string   { return domain.AggregateID("wallet", e.walletID) }
func (e *WalletDepositedEvent) WalletID() int          { return e.walletID }
func (e *WalletDepositedEvent) UserID() int            { return e.userID }
func (e *WalletDepositedEvent) Amount() float64       { return e.amount }
func (e *WalletDepositedEvent) Balance() float64      { return e.balance }
func (e *WalletDepositedEvent) Currency() string      { return e.currency }
func (e *WalletDepositedEvent) ReferenceType() string { return e.referenceType }
func (e *WalletDepositedEvent) ReferenceID() string   { return e.referenceID }

// WalletWithdrawnEvent wallet withdrawn event
type WalletWithdrawnEvent struct {
	walletID      int
	userID        int
	amount        float64
	balance       float64
	currency      string
	referenceType string
	referenceID   string
	occurredAt    time.Time
}

func NewWalletWithdrawnEvent(walletID, userID int, amount, balance float64, currency, referenceType, referenceID string) *WalletWithdrawnEvent {
	return &WalletWithdrawnEvent{
		walletID:      walletID,
		userID:        userID,
		amount:        amount,
		balance:       balance,
		currency:      currency,
		referenceType: referenceType,
		referenceID:   referenceID,
		occurredAt:    time.Now(),
	}
}

func (e *WalletWithdrawnEvent) EventType() string     { return "wallet.withdrawn" }
func (e *WalletWithdrawnEvent) OccurredAt() time.Time { return e.occurredAt }
func (e *WalletWithdrawnEvent) AggregateID() string   { return domain.AggregateID("wallet", e.walletID) }
func (e *WalletWithdrawnEvent) WalletID() int          { return e.walletID }
func (e *WalletWithdrawnEvent) UserID() int            { return e.userID }
func (e *WalletWithdrawnEvent) Amount() float64       { return e.amount }
func (e *WalletWithdrawnEvent) Balance() float64      { return e.balance }
func (e *WalletWithdrawnEvent) Currency() string      { return e.currency }
func (e *WalletWithdrawnEvent) ReferenceType() string { return e.referenceType }
func (e *WalletWithdrawnEvent) ReferenceID() string   { return e.referenceID }

// WalletPaymentEvent wallet payment event
type WalletPaymentEvent struct {
	walletID   int
	userID     int
	amount     float64
	balance    float64
	currency   string
	orderID    string
	occurredAt time.Time
}

func NewWalletPaymentEvent(walletID, userID int, amount, balance float64, currency, orderID string) *WalletPaymentEvent {
	return &WalletPaymentEvent{
		walletID:   walletID,
		userID:     userID,
		amount:     amount,
		balance:    balance,
		currency:   currency,
		orderID:    orderID,
		occurredAt: time.Now(),
	}
}

func (e *WalletPaymentEvent) EventType() string     { return "wallet.payment" }
func (e *WalletPaymentEvent) OccurredAt() time.Time { return e.occurredAt }
func (e *WalletPaymentEvent) AggregateID() string   { return domain.AggregateID("wallet", e.walletID) }
func (e *WalletPaymentEvent) WalletID() int         { return e.walletID }
func (e *WalletPaymentEvent) UserID() int           { return e.userID }
func (e *WalletPaymentEvent) Amount() float64       { return e.amount }
func (e *WalletPaymentEvent) Balance() float64      { return e.balance }
func (e *WalletPaymentEvent) Currency() string      { return e.currency }
func (e *WalletPaymentEvent) OrderID() string       { return e.orderID }

// WalletRefundedEvent wallet refunded event
type WalletRefundedEvent struct {
	walletID   int
	userID     int
	amount     float64
	balance    float64
	currency   string
	orderID    string
	reason     string
	occurredAt time.Time
}

func NewWalletRefundedEvent(walletID, userID int, amount, balance float64, currency, orderID, reason string) *WalletRefundedEvent {
	return &WalletRefundedEvent{
		walletID:   walletID,
		userID:     userID,
		amount:     amount,
		balance:    balance,
		currency:   currency,
		orderID:    orderID,
		reason:     reason,
		occurredAt: time.Now(),
	}
}

func (e *WalletRefundedEvent) EventType() string     { return "wallet.refunded" }
func (e *WalletRefundedEvent) OccurredAt() time.Time { return e.occurredAt }
func (e *WalletRefundedEvent) AggregateID() string   { return domain.AggregateID("wallet", e.walletID) }
func (e *WalletRefundedEvent) WalletID() int         { return e.walletID }
func (e *WalletRefundedEvent) UserID() int           { return e.userID }
func (e *WalletRefundedEvent) Amount() float64       { return e.amount }
func (e *WalletRefundedEvent) Balance() float64      { return e.balance }
func (e *WalletRefundedEvent) Currency() string      { return e.currency }
func (e *WalletRefundedEvent) OrderID() string       { return e.orderID }
func (e *WalletRefundedEvent) Reason() string        { return e.reason }

// WalletFrozenEvent wallet frozen event
type WalletFrozenEvent struct {
	walletID       int
	userID         int
	amount         float64
	frozenAmount   float64
	currency       string
	reason         string
	occurredAt     time.Time
}

func NewWalletFrozenEvent(walletID, userID int, amount, frozenAmount float64, currency, reason string) *WalletFrozenEvent {
	return &WalletFrozenEvent{
		walletID:     walletID,
		userID:       userID,
		amount:       amount,
		frozenAmount: frozenAmount,
		currency:     currency,
		reason:       reason,
		occurredAt:   time.Now(),
	}
}

func (e *WalletFrozenEvent) EventType() string     { return "wallet.frozen" }
func (e *WalletFrozenEvent) OccurredAt() time.Time { return e.occurredAt }
func (e *WalletFrozenEvent) AggregateID() string   { return domain.AggregateID("wallet", e.walletID) }
func (e *WalletFrozenEvent) WalletID() int         { return e.walletID }
func (e *WalletFrozenEvent) UserID() int           { return e.userID }
func (e *WalletFrozenEvent) Amount() float64       { return e.amount }
func (e *WalletFrozenEvent) FrozenAmount() float64 { return e.frozenAmount }
func (e *WalletFrozenEvent) Currency() string      { return e.currency }
func (e *WalletFrozenEvent) Reason() string        { return e.reason }

// WalletUnfrozenEvent wallet unfrozen event
type WalletUnfrozenEvent struct {
	walletID       int
	userID         int
	amount         float64
	frozenAmount   float64
	currency       string
	reason         string
	occurredAt     time.Time
}

func NewWalletUnfrozenEvent(walletID, userID int, amount, frozenAmount float64, currency, reason string) *WalletUnfrozenEvent {
	return &WalletUnfrozenEvent{
		walletID:     walletID,
		userID:       userID,
		amount:       amount,
		frozenAmount: frozenAmount,
		currency:     currency,
		reason:       reason,
		occurredAt:   time.Now(),
	}
}

func (e *WalletUnfrozenEvent) EventType() string     { return "wallet.unfrozen" }
func (e *WalletUnfrozenEvent) OccurredAt() time.Time { return e.occurredAt }
func (e *WalletUnfrozenEvent) AggregateID() string   { return domain.AggregateID("wallet", e.walletID) }
func (e *WalletUnfrozenEvent) WalletID() int         { return e.walletID }
func (e *WalletUnfrozenEvent) UserID() int           { return e.userID }
func (e *WalletUnfrozenEvent) Amount() float64       { return e.amount }
func (e *WalletUnfrozenEvent) FrozenAmount() float64 { return e.frozenAmount }
func (e *WalletUnfrozenEvent) Currency() string      { return e.currency }
func (e *WalletUnfrozenEvent) Reason() string        { return e.reason }
