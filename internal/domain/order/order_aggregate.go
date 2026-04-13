package order

import (
	"fmt"
	"time"

	"example.com/classic/internal/domain"
	"example.com/classic/internal/domain/membership"
)

// Order order entity
type Order struct {
	id             int
	orderNo        string
	userID         int
	items          []*OrderItem
	status         OrderStatus
	totalAmount    float64
	discountAmount float64
	finalAmount    float64
	paymentInfo    *PaymentInfo
	refundInfo     *RefundInfo
	remark         string
	createdAt      time.Time
	updatedAt      time.Time
}

// NewOrder creates new order
func NewOrder(orderNo string, userID int, items []*OrderItem, remark string) (*Order, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("order must have at least one item")
	}

	var totalAmount float64
	for _, item := range items {
		totalAmount += item.Subtotal()
	}

	now := time.Now()
	return &Order{
		orderNo:     orderNo,
		userID:      userID,
		items:       items,
		status:      OrderStatusPending,
		totalAmount: totalAmount,
		finalAmount: totalAmount,
		remark:      remark,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// RebuildOrder rebuilds order from persistence
func RebuildOrder(
	id int,
	orderNo string,
	userID int,
	items []*OrderItem,
	status OrderStatus,
	totalAmount float64,
	discountAmount float64,
	finalAmount float64,
	paymentInfo *PaymentInfo,
	refundInfo *RefundInfo,
	remark string,
	createdAt time.Time,
	updatedAt time.Time,
) *Order {
	return &Order{
		id:             id,
		orderNo:        orderNo,
		userID:         userID,
		items:          items,
		status:         status,
		totalAmount:    totalAmount,
		discountAmount: discountAmount,
		finalAmount:    finalAmount,
		paymentInfo:    paymentInfo,
		refundInfo:     refundInfo,
		remark:         remark,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}

// Getters
func (o *Order) ID() int                    { return o.id }
func (o *Order) OrderNo() string            { return o.orderNo }
func (o *Order) UserID() int                { return o.userID }
func (o *Order) Items() []*OrderItem        { return o.items }
func (o *Order) Status() OrderStatus        { return o.status }
func (o *Order) TotalAmount() float64       { return o.totalAmount }
func (o *Order) DiscountAmount() float64    { return o.discountAmount }
func (o *Order) FinalAmount() float64       { return o.finalAmount }
func (o *Order) PaymentInfo() *PaymentInfo  { return o.paymentInfo }
func (o *Order) RefundInfo() *RefundInfo    { return o.refundInfo }
func (o *Order) Remark() string             { return o.remark }
func (o *Order) CreatedAt() time.Time       { return o.createdAt }
func (o *Order) UpdatedAt() time.Time       { return o.updatedAt }

// SetID sets ID (called by repository)
func (o *Order) SetID(id int) {
	o.id = id
}

// ApplyDiscount applies discount
func (o *Order) ApplyDiscount(discount float64) error {
	if o.status != OrderStatusPending {
		return fmt.Errorf("can only apply discount to pending order")
	}
	if discount < 0 || discount > o.totalAmount {
		return fmt.Errorf("invalid discount amount")
	}

	o.discountAmount = discount
	o.finalAmount = o.totalAmount - discount
	o.updatedAt = time.Now()
	return nil
}

// ApplyMemberDiscount applies member level discount
func (o *Order) ApplyMemberDiscount(level membership.MemberLevel) error {
	if o.status != OrderStatusPending {
		return fmt.Errorf("can only apply discount to pending order")
	}

	// calculate discount using member level
	config := level.Config()
	discount := o.totalAmount * (1 - config.DiscountRate)
	return o.ApplyDiscount(discount)
}

// OrderAggregate order aggregate root
type OrderAggregate struct {
	order   *Order
	events  *domain.EventRecorder
}

// NewOrderAggregate creates new order aggregate
func NewOrderAggregate(orderNo string, userID int, items []*OrderItem, remark string) (*OrderAggregate, error) {
	order, err := NewOrder(orderNo, userID, items, remark)
	if err != nil {
		return nil, err
	}

	aggregate := &OrderAggregate{
		order:  order,
		events: domain.NewEventRecorder(),
	}

	// emit order created event
	aggregate.events.AddEvent(NewOrderCreatedEvent(
		order.id,
		order.orderNo,
		order.userID,
		order.totalAmount,
		order.finalAmount,
	))

	return aggregate, nil
}

// RebuildOrderAggregate rebuilds order aggregate
func RebuildOrderAggregate(order *Order) *OrderAggregate {
	return &OrderAggregate{
		order:  order,
		events: domain.NewEventRecorder(),
	}
}

// Order returns order entity
func (a *OrderAggregate) Order() *Order {
	return a.order
}

// ID returns aggregate ID
func (a *OrderAggregate) ID() int {
	return a.order.ID()
}

// MarkAsPaid marks order as paid
func (a *OrderAggregate) MarkAsPaid(paymentInfo *PaymentInfo) error {
	if a.order.status != OrderStatusPending {
		return fmt.Errorf("can only pay pending order")
	}

	if paymentInfo.TotalAmount() < a.order.finalAmount {
		return fmt.Errorf("payment amount %.2f less than order amount %.2f", paymentInfo.TotalAmount(), a.order.finalAmount)
	}

	a.order.status = OrderStatusPaid
	a.order.paymentInfo = paymentInfo
	a.order.updatedAt = time.Now()

	// emit event
	a.events.AddEvent(NewOrderPaidEvent(
		a.order.id,
		a.order.orderNo,
		a.order.userID,
		string(paymentInfo.Method()),
		paymentInfo.TotalAmount(),
		a.order.finalAmount,
	))

	return nil
}

// StartProcessing starts processing order
func (a *OrderAggregate) StartProcessing() error {
	if a.order.status != OrderStatusPaid {
		return fmt.Errorf("can only process paid order")
	}

	a.order.status = OrderStatusProcessing
	a.order.updatedAt = time.Now()

	a.events.AddEvent(NewOrderProcessingEvent(a.order.id, a.order.orderNo))
	return nil
}

// Ship ships order
func (a *OrderAggregate) Ship(trackingNo string) error {
	if a.order.status != OrderStatusProcessing {
		return fmt.Errorf("can only ship processing order")
	}

	a.order.status = OrderStatusShipped
	a.order.updatedAt = time.Now()

	a.events.AddEvent(NewOrderShippedEvent(a.order.id, a.order.orderNo, trackingNo))
	return nil
}

// Complete completes order
func (a *OrderAggregate) Complete() error {
	if a.order.status != OrderStatusShipped {
		return fmt.Errorf("can only complete shipped order")
	}

	a.order.status = OrderStatusCompleted
	a.order.updatedAt = time.Now()

	a.events.AddEvent(NewOrderCompletedEvent(a.order.id, a.order.orderNo))
	return nil
}

// Cancel cancels order
func (a *OrderAggregate) Cancel(reason string) error {
	if !a.order.status.CanTransitionTo(OrderStatusCancelled) {
		return fmt.Errorf("cannot cancel order in status %s", a.order.status)
	}

	oldStatus := a.order.status
	a.order.status = OrderStatusCancelled
	a.order.remark = fmt.Sprintf("%s (cancelled: %s)", a.order.remark, reason)
	a.order.updatedAt = time.Now()

	a.events.AddEvent(NewOrderCancelledEvent(
		a.order.id,
		a.order.orderNo,
		string(oldStatus),
		reason,
	))

	return nil
}

// Refund refunds order
func (a *OrderAggregate) Refund(refundInfo *RefundInfo) error {
	if !a.order.status.CanTransitionTo(OrderStatusRefunded) {
		return fmt.Errorf("cannot refund order in status %s", a.order.status)
	}

	a.order.status = OrderStatusRefunded
	a.order.refundInfo = refundInfo
	a.order.updatedAt = time.Now()

	a.events.AddEvent(NewOrderRefundedEvent(
		a.order.id,
		a.order.orderNo,
		refundInfo.Amount(),
		refundInfo.Reason(),
	))

	return nil
}

// CanPay checks if order can be paid
func (a *OrderAggregate) CanPay() bool {
	return a.order.status == OrderStatusPending
}

// CanCancel checks if order can be cancelled
func (a *OrderAggregate) CanCancel() bool {
	return a.order.status.CanTransitionTo(OrderStatusCancelled)
}

// CanRefund checks if order can be refunded
func (a *OrderAggregate) CanRefund() bool {
	return a.order.status.CanTransitionTo(OrderStatusRefunded)
}

// Events returns domain events
func (a *OrderAggregate) Events() []domain.DomainEvent {
	return a.events.Events()
}

// ClearEvents clears domain events
func (a *OrderAggregate) ClearEvents() {
	a.events.ClearEvents()
}

// HasEvents checks if has events
func (a *OrderAggregate) HasEvents() bool {
	return a.events.HasEvents()
}
