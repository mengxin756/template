package order

import (
	"context"
)

// OrderRepository order repository interface
type OrderRepository interface {
	// Create creates order
	Create(ctx context.Context, order *Order) error

	// GetByID gets order by ID
	GetByID(ctx context.Context, id int) (*Order, error)

	// GetByOrderNo gets order by order number
	GetByOrderNo(ctx context.Context, orderNo string) (*Order, error)

	// Update updates order
	Update(ctx context.Context, order *Order) error

	// Save saves order aggregate
	Save(ctx context.Context, aggregate *OrderAggregate) error

	// GetAggregateByID gets order aggregate by ID
	GetAggregateByID(ctx context.Context, id int) (*OrderAggregate, error)

	// GetAggregateByOrderNo gets order aggregate by order number
	GetAggregateByOrderNo(ctx context.Context, orderNo string) (*OrderAggregate, error)

	// ListByUserID lists orders by user ID
	ListByUserID(ctx context.Context, userID int, query *OrderQuery) ([]*Order, int64, error)

	// CreateOrderItem creates order item
	CreateOrderItem(ctx context.Context, item *OrderItem) error

	// GetOrderItems gets order items by order ID
	GetOrderItems(ctx context.Context, orderID int) ([]*OrderItem, error)
}

// OrderQuery order query conditions
type OrderQuery struct {
	ID        *int
	UserID    *int
	OrderNo   *string
	Status    *OrderStatus
	Page      int
	PageSize  int
	StartDate *string
	EndDate   *string
}
