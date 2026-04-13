package service

import (
	"context"
	"fmt"
	"time"

	"example.com/classic/internal/domain"
	"example.com/classic/internal/domain/membership"
	"example.com/classic/internal/domain/order"
	"example.com/classic/internal/domain/points"
	"example.com/classic/internal/domain/wallet"
	"example.com/classic/pkg/errors"
	"example.com/classic/pkg/logger"
)

// OrderService order application service
type OrderService interface {
	// CreateOrder creates order
	CreateOrder(ctx context.Context, req *CreateOrderRequest) (*order.Order, error)

	// PayOrder pays order with wallet
	PayOrder(ctx context.Context, req *PayOrderRequest) error

	// PayOrderWithMemberDiscount pays order with member discount
	PayOrderWithMemberDiscount(ctx context.Context, req *PayOrderRequest) error

	// PayOrderWithMixed pays order with mixed payment
	PayOrderWithMixed(ctx context.Context, req *MixedPaymentRequest) error

	// CancelOrder cancels order
	CancelOrder(ctx context.Context, orderID int, reason string) error

	// RefundOrder refunds order
	RefundOrder(ctx context.Context, orderID int, reason string) error

	// GetOrder gets order by ID
	GetOrder(ctx context.Context, orderID int) (*order.Order, error)

	// ListOrders lists orders by user ID
	ListOrders(ctx context.Context, userID int, query *order.OrderQuery) ([]*order.Order, int64, error)

	// ProcessOrder starts processing order
	ProcessOrder(ctx context.Context, orderID int) error

	// ShipOrder ships order
	ShipOrder(ctx context.Context, orderID int, trackingNo string) error

	// CompleteOrder completes order
	CompleteOrder(ctx context.Context, orderID int) error
}

// CreateOrderRequest create order request
type CreateOrderRequest struct {
	UserID      int
	Items       []OrderItemInput
	Remark      string
}

// OrderItemInput order item input
type OrderItemInput struct {
	ProductID   int
	ProductName string
	Quantity    int
	UnitPrice   float64
	Discount    float64
}

// PayOrderRequest pay order request
type PayOrderRequest struct {
	OrderID int
	UserID  int
}

// MixedPaymentRequest mixed payment request
type MixedPaymentRequest struct {
	OrderID      int
	UserID       int
	WalletAmount float64
	PointsAmount int
}

// orderServiceImpl order service implementation
type orderServiceImpl struct {
	orderRepo       order.OrderRepository
	memberRepo      membership.MemberRepository
	walletRepo      wallet.WalletRepository
	pointsRepo      points.PointsAccountRepository
	orderPaymentSvc order.OrderPaymentService
	orderPointsSvc  order.OrderPointsService
	eventPublisher  domain.EventPublisher
	log             logger.Logger
}

// NewOrderService creates order service
func NewOrderService(
	orderRepo order.OrderRepository,
	memberRepo membership.MemberRepository,
	walletRepo wallet.WalletRepository,
	pointsRepo points.PointsAccountRepository,
	eventPublisher domain.EventPublisher,
	log logger.Logger,
) OrderService {
	return &orderServiceImpl{
		orderRepo:       orderRepo,
		memberRepo:      memberRepo,
		walletRepo:      walletRepo,
		pointsRepo:      pointsRepo,
		orderPaymentSvc: order.NewOrderPaymentService(),
		orderPointsSvc:  order.NewOrderPointsService(),
		eventPublisher:  eventPublisher,
		log:             log,
	}
}

// CreateOrder creates order
func (s *orderServiceImpl) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*order.Order, error) {
	s.log.Info(ctx, "creating order", logger.F("user_id", req.UserID))

	// 1. create order items
	items := make([]*order.OrderItem, 0, len(req.Items))
	for _, input := range req.Items {
		item, err := order.NewOrderItem(input.ProductID, input.ProductName, input.Quantity, input.UnitPrice, input.Discount)
		if err != nil {
			return nil, errors.WrapInvalidParam(err, "invalid order item")
		}
		items = append(items, item)
	}

	// 2. generate order number
	orderNo := s.generateOrderNo()

	// 3. create order aggregate
	orderAggregate, err := order.NewOrderAggregate(orderNo, req.UserID, items, req.Remark)
	if err != nil {
		return nil, errors.WrapInternalError(err, "failed to create order")
	}

	// 4. save order
	if err := s.orderRepo.Save(ctx, orderAggregate); err != nil {
		return nil, errors.WrapInternalError(err, "failed to save order")
	}

	// 5. publish events
	s.publishEvents(ctx, orderAggregate)

	s.log.Info(ctx, "order created", logger.F("order_id", orderAggregate.ID()), logger.F("order_no", orderNo))

	return orderAggregate.Order(), nil
}

// PayOrder pays order with wallet
func (s *orderServiceImpl) PayOrder(ctx context.Context, req *PayOrderRequest) error {
	s.log.Info(ctx, "paying order", logger.F("order_id", req.OrderID), logger.F("user_id", req.UserID))

	// 1. get order aggregate
	orderAggregate, err := s.orderRepo.GetAggregateByID(ctx, req.OrderID)
	if err != nil {
		return err
	}

	// 2. validate user
	if orderAggregate.Order().UserID() != req.UserID {
		return errors.ErrForbidden
	}

	// 3. get wallet aggregate
	walletAggregate, err := s.walletRepo.GetAggregateByUserID(ctx, req.UserID)
	if err != nil {
		return err
	}

	// 4. pay using domain service
	if err := s.orderPaymentSvc.PayWithWallet(ctx, orderAggregate, walletAggregate); err != nil {
		return errors.WrapInternalError(err, "payment failed")
	}

	// 5. save both aggregates (should be in transaction)
	if err := s.orderRepo.Save(ctx, orderAggregate); err != nil {
		return errors.WrapInternalError(err, "failed to save order")
	}
	if err := s.walletRepo.Save(ctx, walletAggregate); err != nil {
		return errors.WrapInternalError(err, "failed to save wallet")
	}

	// 6. publish events
	s.publishEvents(ctx, orderAggregate)
	s.publishEvents(ctx, walletAggregate)

	// 7. earn points after payment
	if err := s.earnPointsAfterPayment(ctx, orderAggregate, req.UserID); err != nil {
		s.log.Warn(ctx, "failed to earn points", logger.F("error", err))
		// don't fail the payment, just log
	}

	s.log.Info(ctx, "order paid successfully", logger.F("order_id", req.OrderID))

	return nil
}

// PayOrderWithMemberDiscount pays order with member discount
func (s *orderServiceImpl) PayOrderWithMemberDiscount(ctx context.Context, req *PayOrderRequest) error {
	s.log.Info(ctx, "paying order with member discount", logger.F("order_id", req.OrderID))

	// 1. get order aggregate
	orderAggregate, err := s.orderRepo.GetAggregateByID(ctx, req.OrderID)
	if err != nil {
		return err
	}

	// 2. validate user
	if orderAggregate.Order().UserID() != req.UserID {
		return errors.ErrForbidden
	}

	// 3. get member aggregate
	memberAggregate, err := s.memberRepo.GetAggregateByUserID(ctx, req.UserID)
	if err != nil {
		return err
	}

	// 4. get wallet aggregate
	walletAggregate, err := s.walletRepo.GetAggregateByUserID(ctx, req.UserID)
	if err != nil {
		return err
	}

	// 5. pay using domain service
	if err := s.orderPaymentSvc.PayWithMemberDiscount(ctx, orderAggregate, memberAggregate, walletAggregate); err != nil {
		return errors.WrapInternalError(err, "payment failed")
	}

	// 6. save aggregates
	if err := s.orderRepo.Save(ctx, orderAggregate); err != nil {
		return errors.WrapInternalError(err, "failed to save order")
	}
	if err := s.walletRepo.Save(ctx, walletAggregate); err != nil {
		return errors.WrapInternalError(err, "failed to save wallet")
	}

	// 7. publish events
	s.publishEvents(ctx, orderAggregate)
	s.publishEvents(ctx, walletAggregate)

	// 8. earn points
	if err := s.earnPointsAfterPayment(ctx, orderAggregate, req.UserID); err != nil {
		s.log.Warn(ctx, "failed to earn points", logger.F("error", err))
	}

	return nil
}

// PayOrderWithMixed pays order with mixed payment
func (s *orderServiceImpl) PayOrderWithMixed(ctx context.Context, req *MixedPaymentRequest) error {
	s.log.Info(ctx, "paying order with mixed payment", logger.F("order_id", req.OrderID))

	// 1. get order aggregate
	orderAggregate, err := s.orderRepo.GetAggregateByID(ctx, req.OrderID)
	if err != nil {
		return err
	}

	// 2. validate user
	if orderAggregate.Order().UserID() != req.UserID {
		return errors.ErrForbidden
	}

	// 3. get wallet and points aggregates
	walletAggregate, err := s.walletRepo.GetAggregateByUserID(ctx, req.UserID)
	if err != nil {
		return err
	}

	pointsAggregate, err := s.pointsRepo.GetAggregateByUserID(ctx, req.UserID)
	if err != nil {
		return err
	}

	// 4. pay using domain service
	if err := s.orderPaymentSvc.PayWithMixed(ctx, orderAggregate, walletAggregate, pointsAggregate, req.WalletAmount, req.PointsAmount); err != nil {
		return errors.WrapInternalError(err, "mixed payment failed")
	}

	// 5. save aggregates
	if err := s.orderRepo.Save(ctx, orderAggregate); err != nil {
		return errors.WrapInternalError(err, "failed to save order")
	}
	if err := s.walletRepo.Save(ctx, walletAggregate); err != nil {
		return errors.WrapInternalError(err, "failed to save wallet")
	}
	if err := s.pointsRepo.Save(ctx, pointsAggregate); err != nil {
		return errors.WrapInternalError(err, "failed to save points account")
	}

	// 6. publish events
	s.publishEvents(ctx, orderAggregate)
	s.publishEvents(ctx, walletAggregate)
	s.publishEvents(ctx, pointsAggregate)

	return nil
}

// CancelOrder cancels order
func (s *orderServiceImpl) CancelOrder(ctx context.Context, orderID int, reason string) error {
	s.log.Info(ctx, "cancelling order", logger.F("order_id", orderID))

	// 1. get order aggregate
	orderAggregate, err := s.orderRepo.GetAggregateByID(ctx, orderID)
	if err != nil {
		return err
	}

	// 2. cancel using domain service
	if err := s.orderPaymentSvc.CancelUnpaidOrder(ctx, orderAggregate, reason); err != nil {
		return errors.WrapInternalError(err, "cancel failed")
	}

	// 3. save
	if err := s.orderRepo.Save(ctx, orderAggregate); err != nil {
		return errors.WrapInternalError(err, "failed to save order")
	}

	// 4. publish events
	s.publishEvents(ctx, orderAggregate)

	return nil
}

// RefundOrder refunds order
func (s *orderServiceImpl) RefundOrder(ctx context.Context, orderID int, reason string) error {
	s.log.Info(ctx, "refunding order", logger.F("order_id", orderID))

	// 1. get order aggregate
	orderAggregate, err := s.orderRepo.GetAggregateByID(ctx, orderID)
	if err != nil {
		return err
	}

	// 2. get wallet aggregate
	walletAggregate, err := s.walletRepo.GetAggregateByUserID(ctx, orderAggregate.Order().UserID())
	if err != nil {
		return err
	}

	// 3. refund using domain service
	if err := s.orderPaymentSvc.RefundOrder(ctx, orderAggregate, walletAggregate, reason); err != nil {
		return errors.WrapInternalError(err, "refund failed")
	}

	// 4. save aggregates
	if err := s.orderRepo.Save(ctx, orderAggregate); err != nil {
		return errors.WrapInternalError(err, "failed to save order")
	}
	if err := s.walletRepo.Save(ctx, walletAggregate); err != nil {
		return errors.WrapInternalError(err, "failed to save wallet")
	}

	// 5. publish events
	s.publishEvents(ctx, orderAggregate)
	s.publishEvents(ctx, walletAggregate)

	return nil
}

// GetOrder gets order by ID
func (s *orderServiceImpl) GetOrder(ctx context.Context, orderID int) (*order.Order, error) {
	aggregate, err := s.orderRepo.GetAggregateByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return aggregate.Order(), nil
}

// ListOrders lists orders by user ID
func (s *orderServiceImpl) ListOrders(ctx context.Context, userID int, query *order.OrderQuery) ([]*order.Order, int64, error) {
	return s.orderRepo.ListByUserID(ctx, userID, query)
}

// ProcessOrder starts processing order
func (s *orderServiceImpl) ProcessOrder(ctx context.Context, orderID int) error {
	aggregate, err := s.orderRepo.GetAggregateByID(ctx, orderID)
	if err != nil {
		return err
	}

	if err := aggregate.StartProcessing(); err != nil {
		return errors.WrapInvalidParam(err, "cannot process order")
	}

	if err := s.orderRepo.Save(ctx, aggregate); err != nil {
		return errors.WrapInternalError(err, "failed to save order")
	}

	s.publishEvents(ctx, aggregate)
	return nil
}

// ShipOrder ships order
func (s *orderServiceImpl) ShipOrder(ctx context.Context, orderID int, trackingNo string) error {
	aggregate, err := s.orderRepo.GetAggregateByID(ctx, orderID)
	if err != nil {
		return err
	}

	if err := aggregate.Ship(trackingNo); err != nil {
		return errors.WrapInvalidParam(err, "cannot ship order")
	}

	if err := s.orderRepo.Save(ctx, aggregate); err != nil {
		return errors.WrapInternalError(err, "failed to save order")
	}

	s.publishEvents(ctx, aggregate)
	return nil
}

// CompleteOrder completes order
func (s *orderServiceImpl) CompleteOrder(ctx context.Context, orderID int) error {
	aggregate, err := s.orderRepo.GetAggregateByID(ctx, orderID)
	if err != nil {
		return err
	}

	if err := aggregate.Complete(); err != nil {
		return errors.WrapInvalidParam(err, "cannot complete order")
	}

	if err := s.orderRepo.Save(ctx, aggregate); err != nil {
		return errors.WrapInternalError(err, "failed to save order")
	}

	s.publishEvents(ctx, aggregate)
	return nil
}

// Helper methods

func (s *orderServiceImpl) generateOrderNo() string {
	return fmt.Sprintf("ORD%s%06d", time.Now().Format("20060102150405"), time.Now().Nanosecond()/1000)
}

func (s *orderServiceImpl) publishEvents(ctx context.Context, aggregate interface{ Events() []domain.DomainEvent }) {
	for _, event := range aggregate.Events() {
		if err := s.eventPublisher.Publish(ctx, event); err != nil {
			s.log.Error(ctx, "failed to publish event", logger.F("event_type", event.EventType()), logger.F("error", err))
		}
	}
}

func (s *orderServiceImpl) earnPointsAfterPayment(ctx context.Context, orderAggregate *order.OrderAggregate, userID int) error {
	// get member and points account
	memberAggregate, err := s.memberRepo.GetAggregateByUserID(ctx, userID)
	if err != nil {
		return err
	}

	pointsAggregate, err := s.pointsRepo.GetAggregateByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// earn points
	if err := s.orderPointsSvc.EarnPointsAfterOrder(ctx, orderAggregate, memberAggregate, pointsAggregate); err != nil {
		return err
	}

	// save
	if err := s.memberRepo.Save(ctx, memberAggregate); err != nil {
		return err
	}
	if err := s.pointsRepo.Save(ctx, pointsAggregate); err != nil {
		return err
	}

	s.publishEvents(ctx, memberAggregate)
	s.publishEvents(ctx, pointsAggregate)

	return nil
}
