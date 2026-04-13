package order

import (
	"time"

	"example.com/classic/internal/domain"
)

// OrderCreatedEvent order created event
type OrderCreatedEvent struct {
	orderID      int
	orderNo      string
	userID       int
	totalAmount  float64
	finalAmount  float64
	occurredAt   time.Time
}

func NewOrderCreatedEvent(orderID int, orderNo string, userID int, totalAmount, finalAmount float64) *OrderCreatedEvent {
	return &OrderCreatedEvent{
		orderID:     orderID,
		orderNo:     orderNo,
		userID:      userID,
		totalAmount: totalAmount,
		finalAmount: finalAmount,
		occurredAt:  time.Now(),
	}
}

func (e *OrderCreatedEvent) EventType() string     { return "order.created" }
func (e *OrderCreatedEvent) OccurredAt() time.Time { return e.occurredAt }
func (e *OrderCreatedEvent) AggregateID() string   { return domain.AggregateID("order", e.orderID) }
func (e *OrderCreatedEvent) OrderID() int          { return e.orderID }
func (e *OrderCreatedEvent) OrderNo() string       { return e.orderNo }
func (e *OrderCreatedEvent) UserID() int           { return e.userID }
func (e *OrderCreatedEvent) TotalAmount() float64  { return e.totalAmount }
func (e *OrderCreatedEvent) FinalAmount() float64  { return e.finalAmount }

// OrderPaidEvent order paid event
type OrderPaidEvent struct {
	orderID       int
	orderNo       string
	userID        int
	paymentMethod string
	paidAmount    float64
	orderAmount   float64
	occurredAt    time.Time
}

func NewOrderPaidEvent(orderID int, orderNo string, userID int, paymentMethod string, paidAmount, orderAmount float64) *OrderPaidEvent {
	return &OrderPaidEvent{
		orderID:       orderID,
		orderNo:       orderNo,
		userID:        userID,
		paymentMethod: paymentMethod,
		paidAmount:    paidAmount,
		orderAmount:   orderAmount,
		occurredAt:    time.Now(),
	}
}

func (e *OrderPaidEvent) EventType() string     { return "order.paid" }
func (e *OrderPaidEvent) OccurredAt() time.Time { return e.occurredAt }
func (e *OrderPaidEvent) AggregateID() string   { return domain.AggregateID("order", e.orderID) }
func (e *OrderPaidEvent) OrderID() int          { return e.orderID }
func (e *OrderPaidEvent) OrderNo() string       { return e.orderNo }
func (e *OrderPaidEvent) UserID() int           { return e.userID }
func (e *OrderPaidEvent) PaymentMethod() string { return e.paymentMethod }
func (e *OrderPaidEvent) PaidAmount() float64   { return e.paidAmount }
func (e *OrderPaidEvent) OrderAmount() float64  { return e.orderAmount }

// OrderProcessingEvent order processing event
type OrderProcessingEvent struct {
	orderID    int
	orderNo    string
	occurredAt time.Time
}

func NewOrderProcessingEvent(orderID int, orderNo string) *OrderProcessingEvent {
	return &OrderProcessingEvent{
		orderID:    orderID,
		orderNo:    orderNo,
		occurredAt: time.Now(),
	}
}

func (e *OrderProcessingEvent) EventType() string     { return "order.processing" }
func (e *OrderProcessingEvent) OccurredAt() time.Time { return e.occurredAt }
func (e *OrderProcessingEvent) AggregateID() string   { return domain.AggregateID("order", e.orderID) }
func (e *OrderProcessingEvent) OrderID() int          { return e.orderID }
func (e *OrderProcessingEvent) OrderNo() string       { return e.orderNo }

// OrderShippedEvent order shipped event
type OrderShippedEvent struct {
	orderID     int
	orderNo     string
	trackingNo  string
	occurredAt  time.Time
}

func NewOrderShippedEvent(orderID int, orderNo, trackingNo string) *OrderShippedEvent {
	return &OrderShippedEvent{
		orderID:    orderID,
		orderNo:    orderNo,
		trackingNo: trackingNo,
		occurredAt: time.Now(),
	}
}

func (e *OrderShippedEvent) EventType() string     { return "order.shipped" }
func (e *OrderShippedEvent) OccurredAt() time.Time { return e.occurredAt }
func (e *OrderShippedEvent) AggregateID() string   { return domain.AggregateID("order", e.orderID) }
func (e *OrderShippedEvent) OrderID() int          { return e.orderID }
func (e *OrderShippedEvent) OrderNo() string       { return e.orderNo }
func (e *OrderShippedEvent) TrackingNo() string    { return e.trackingNo }

// OrderCompletedEvent order completed event
type OrderCompletedEvent struct {
	orderID    int
	orderNo    string
	occurredAt time.Time
}

func NewOrderCompletedEvent(orderID int, orderNo string) *OrderCompletedEvent {
	return &OrderCompletedEvent{
		orderID:    orderID,
		orderNo:    orderNo,
		occurredAt: time.Now(),
	}
}

func (e *OrderCompletedEvent) EventType() string     { return "order.completed" }
func (e *OrderCompletedEvent) OccurredAt() time.Time { return e.occurredAt }
func (e *OrderCompletedEvent) AggregateID() string   { return domain.AggregateID("order", e.orderID) }
func (e *OrderCompletedEvent) OrderID() int          { return e.orderID }
func (e *OrderCompletedEvent) OrderNo() string       { return e.orderNo }

// OrderCancelledEvent order cancelled event
type OrderCancelledEvent struct {
	orderID      int
	orderNo      string
	previousStatus string
	reason       string
	occurredAt   time.Time
}

func NewOrderCancelledEvent(orderID int, orderNo, previousStatus, reason string) *OrderCancelledEvent {
	return &OrderCancelledEvent{
		orderID:       orderID,
		orderNo:       orderNo,
		previousStatus: previousStatus,
		reason:        reason,
		occurredAt:    time.Now(),
	}
}

func (e *OrderCancelledEvent) EventType() string     { return "order.cancelled" }
func (e *OrderCancelledEvent) OccurredAt() time.Time { return e.occurredAt }
func (e *OrderCancelledEvent) AggregateID() string   { return domain.AggregateID("order", e.orderID) }
func (e *OrderCancelledEvent) OrderID() int          { return e.orderID }
func (e *OrderCancelledEvent) OrderNo() string       { return e.orderNo }
func (e *OrderCancelledEvent) PreviousStatus() string { return e.previousStatus }
func (e *OrderCancelledEvent) Reason() string        { return e.reason }

// OrderRefundedEvent order refunded event
type OrderRefundedEvent struct {
	orderID     int
	orderNo     string
	refundAmount float64
	reason      string
	occurredAt  time.Time
}

func NewOrderRefundedEvent(orderID int, orderNo string, refundAmount float64, reason string) *OrderRefundedEvent {
	return &OrderRefundedEvent{
		orderID:      orderID,
		orderNo:      orderNo,
		refundAmount: refundAmount,
		reason:       reason,
		occurredAt:   time.Now(),
	}
}

func (e *OrderRefundedEvent) EventType() string     { return "order.refunded" }
func (e *OrderRefundedEvent) OccurredAt() time.Time { return e.occurredAt }
func (e *OrderRefundedEvent) AggregateID() string   { return domain.AggregateID("order", e.orderID) }
func (e *OrderRefundedEvent) OrderID() int          { return e.orderID }
func (e *OrderRefundedEvent) OrderNo() string       { return e.orderNo }
func (e *OrderRefundedEvent) RefundAmount() float64 { return e.refundAmount }
func (e *OrderRefundedEvent) Reason() string        { return e.reason }
