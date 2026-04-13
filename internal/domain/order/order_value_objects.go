package order

import (
	"fmt"
	"time"
)

// OrderStatus order status enum
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"   // pending payment
	OrderStatusPaid      OrderStatus = "paid"      // paid, waiting for processing
	OrderStatusProcessing OrderStatus = "processing" // processing
	OrderStatusShipped   OrderStatus = "shipped"   // shipped
	OrderStatusCompleted OrderStatus = "completed" // completed
	OrderStatusCancelled OrderStatus = "cancelled" // cancelled
	OrderStatusRefunded  OrderStatus = "refunded"  // refunded
)

// status transitions: defines valid status transitions
var orderStatusTransitions = map[OrderStatus][]OrderStatus{
	OrderStatusPending: {
		OrderStatusPaid,
		OrderStatusCancelled,
	},
	OrderStatusPaid: {
		OrderStatusProcessing,
		OrderStatusCancelled,
		OrderStatusRefunded,
	},
	OrderStatusProcessing: {
		OrderStatusShipped,
		OrderStatusCancelled,
		OrderStatusRefunded,
	},
	OrderStatusShipped: {
		OrderStatusCompleted,
		OrderStatusRefunded,
	},
	OrderStatusCompleted: {
		OrderStatusRefunded,
	},
	OrderStatusCancelled: {},
	OrderStatusRefunded:  {},
}

// IsValid validates order status
func (s OrderStatus) IsValid() bool {
	_, ok := orderStatusTransitions[s]
	return ok
}

// CanTransitionTo checks if can transition to target status
func (s OrderStatus) CanTransitionTo(target OrderStatus) bool {
	allowed, ok := orderStatusTransitions[s]
	if !ok {
		return false
	}
	for _, status := range allowed {
		if status == target {
			return true
		}
	}
	return false
}

// ValidTransitions returns all valid transitions
func (s OrderStatus) ValidTransitions() []OrderStatus {
	return orderStatusTransitions[s]
}

// String returns string representation
func (s OrderStatus) String() string {
	return string(s)
}

// ParseOrderStatus parses string to OrderStatus
func ParseOrderStatus(str string) (OrderStatus, error) {
	status := OrderStatus(str)
	if !status.IsValid() {
		return "", fmt.Errorf("invalid order status: %s", str)
	}
	return status, nil
}

// PaymentMethod payment method enum
type PaymentMethod string

const (
	PaymentMethodWallet  PaymentMethod = "wallet"  // wallet balance
	PaymentMethodPoints  PaymentMethod = "points"  // points
	PaymentMethodMixed   PaymentMethod = "mixed"   // mixed payment
	PaymentMethodExternal PaymentMethod = "external" // external payment (alipay, wechat)
)

// IsValid validates payment method
func (m PaymentMethod) IsValid() bool {
	switch m {
	case PaymentMethodWallet, PaymentMethodPoints, PaymentMethodMixed, PaymentMethodExternal:
		return true
	default:
		return false
	}
}

// String returns string representation
func (m PaymentMethod) String() string {
	return string(m)
}

// OrderItem order item entity
type OrderItem struct {
	productID   int
	productName string
	quantity    int
	unitPrice   float64
	discount    float64
	subtotal    float64
}

// NewOrderItem creates order item
func NewOrderItem(productID int, productName string, quantity int, unitPrice, discount float64) (*OrderItem, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}
	if unitPrice < 0 {
		return nil, fmt.Errorf("unit price cannot be negative")
	}
	if discount < 0 || discount > 1 {
		return nil, fmt.Errorf("discount must be between 0 and 1")
	}

	subtotal := float64(quantity) * unitPrice * (1 - discount)
	return &OrderItem{
		productID:   productID,
		productName: productName,
		quantity:    quantity,
		unitPrice:   unitPrice,
		discount:    discount,
		subtotal:    subtotal,
	}, nil
}

// Getters
func (i *OrderItem) ProductID() int       { return i.productID }
func (i *OrderItem) ProductName() string  { return i.productName }
func (i *OrderItem) Quantity() int        { return i.quantity }
func (i *OrderItem) UnitPrice() float64  { return i.unitPrice }
func (i *OrderItem) Discount() float64   { return i.discount }
func (i *OrderItem) Subtotal() float64   { return i.subtotal }

// PaymentInfo payment info value object
type PaymentInfo struct {
	method         PaymentMethod
	walletAmount   float64 // amount paid from wallet
	pointsAmount   float64 // points value used
	externalAmount float64 // external payment amount
	totalAmount    float64
	currency       string
	paidAt         time.Time
}

// NewPaymentInfo creates payment info
func NewPaymentInfo(method PaymentMethod, walletAmount, pointsAmount, externalAmount float64, currency string) *PaymentInfo {
	return &PaymentInfo{
		method:         method,
		walletAmount:   walletAmount,
		pointsAmount:   pointsAmount,
		externalAmount: externalAmount,
		totalAmount:    walletAmount + externalAmount,
		currency:       currency,
		paidAt:         time.Now(),
	}
}

// Getters
func (p *PaymentInfo) Method() PaymentMethod       { return p.method }
func (p *PaymentInfo) WalletAmount() float64       { return p.walletAmount }
func (p *PaymentInfo) PointsAmount() float64       { return p.pointsAmount }
func (p *PaymentInfo) ExternalAmount() float64     { return p.externalAmount }
func (p *PaymentInfo) TotalAmount() float64        { return p.totalAmount }
func (p *PaymentInfo) Currency() string            { return p.currency }
func (p *PaymentInfo) PaidAt() time.Time           { return p.paidAt }

// RefundInfo refund info value object
type RefundInfo struct {
	refundID       string
	amount         float64
	reason         string
	refundedAt     time.Time
	refundMethod   PaymentMethod
}

// NewRefundInfo creates refund info
func NewRefundInfo(refundID string, amount float64, reason string, method PaymentMethod) *RefundInfo {
	return &RefundInfo{
		refundID:     refundID,
		amount:       amount,
		reason:       reason,
		refundedAt:   time.Now(),
		refundMethod: method,
	}
}

// Getters
func (r *RefundInfo) RefundID() string         { return r.refundID }
func (r *RefundInfo) Amount() float64          { return r.amount }
func (r *RefundInfo) Reason() string           { return r.reason }
func (r *RefundInfo) RefundedAt() time.Time    { return r.refundedAt }
func (r *RefundInfo) RefundMethod() PaymentMethod { return r.refundMethod }
