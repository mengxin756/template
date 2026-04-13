package order

import (
	"context"
	"fmt"

	"example.com/classic/internal/domain/membership"
	"example.com/classic/internal/domain/points"
	"example.com/classic/internal/domain/wallet"
)

// OrderPaymentService order payment domain service
// This service coordinates multiple aggregates: Order, Wallet, PointsAccount, Member
type OrderPaymentService interface {
	// PayWithWallet pays order with wallet balance
	PayWithWallet(ctx context.Context, orderAgg *OrderAggregate, walletAgg *wallet.WalletAggregate) error

	// PayWithMixed pays order with mixed payment (wallet + points)
	PayWithMixed(ctx context.Context, orderAgg *OrderAggregate, walletAgg *wallet.WalletAggregate, pointsAccount *points.PointsAccountAggregate, walletAmount float64, pointsAmount int) error

	// PayWithMemberDiscount pays order with member discount
	PayWithMemberDiscount(ctx context.Context, orderAgg *OrderAggregate, memberAgg *membership.MemberAggregate, walletAgg *wallet.WalletAggregate) error

	// RefundOrder refunds order
	RefundOrder(ctx context.Context, orderAgg *OrderAggregate, walletAgg *wallet.WalletAggregate, reason string) error

	// CancelUnpaidOrder cancels unpaid order
	CancelUnpaidOrder(ctx context.Context, orderAgg *OrderAggregate, reason string) error
}

// orderPaymentServiceImpl order payment service implementation
type orderPaymentServiceImpl struct{}

// NewOrderPaymentService creates order payment service
func NewOrderPaymentService() OrderPaymentService {
	return &orderPaymentServiceImpl{}
}

// PayWithWallet pays order with wallet balance
func (s *orderPaymentServiceImpl) PayWithWallet(ctx context.Context, orderAgg *OrderAggregate, walletAgg *wallet.WalletAggregate) error {
	// 1. validate order can be paid
	if !orderAgg.CanPay() {
		return fmt.Errorf("order cannot be paid in current status")
	}

	// 2. get payment amount
	amount := orderAgg.Order().FinalAmount()

	// 3. create money value object
	money, err := wallet.NewMoney(amount, walletAgg.Wallet().Currency())
	if err != nil {
		return err
	}

	// 4. check wallet balance
	if !walletAgg.CanPay(*money) {
		return fmt.Errorf("insufficient wallet balance")
	}

	// 5. pay from wallet
	if err := walletAgg.Pay(*money, fmt.Sprintf("ORDER-%d", orderAgg.ID()), fmt.Sprintf("order payment: %s", orderAgg.Order().OrderNo())); err != nil {
		return fmt.Errorf("wallet payment failed: %w", err)
	}

	// 6. mark order as paid
	paymentInfo := NewPaymentInfo(PaymentMethodWallet, amount, 0, 0, string(walletAgg.Wallet().Currency()))
	if err := orderAgg.MarkAsPaid(paymentInfo); err != nil {
		// rollback wallet payment
		walletAgg.Refund(*money, fmt.Sprintf("ORDER-%d", orderAgg.ID()), "order payment failed, rollback")
		return fmt.Errorf("order payment failed: %w", err)
	}

	return nil
}

// PayWithMixed pays order with mixed payment
func (s *orderPaymentServiceImpl) PayWithMixed(
	ctx context.Context,
	orderAgg *OrderAggregate,
	walletAgg *wallet.WalletAggregate,
	pointsAccount *points.PointsAccountAggregate,
	walletAmount float64,
	pointsAmount int,
) error {
	// 1. validate order
	if !orderAgg.CanPay() {
		return fmt.Errorf("order cannot be paid in current status")
	}

	// 2. validate amounts
	totalPayment := walletAmount + float64(pointsAmount)/100.0 // assume 100 points = 1 currency unit
	if totalPayment < orderAgg.Order().FinalAmount() {
		return fmt.Errorf("payment amount %.2f less than order amount %.2f", totalPayment, orderAgg.Order().FinalAmount())
	}

	// 3. use points if specified
	if pointsAmount > 0 {
		pointsValue, err := points.NewPoints(pointsAmount)
		if err != nil {
			return fmt.Errorf("invalid points amount: %w", err)
		}
		if !pointsAccount.CanUse(*pointsValue) {
			return fmt.Errorf("insufficient points")
		}
		if err := pointsAccount.Use(*pointsValue, "order payment", orderAgg.Order().OrderNo()); err != nil {
			return fmt.Errorf("points usage failed: %w", err)
		}
	}

	// 4. pay from wallet
	if walletAmount > 0 {
		money, err := wallet.NewMoney(walletAmount, walletAgg.Wallet().Currency())
		if err != nil {
			return err
		}
		if !walletAgg.CanPay(*money) {
			// rollback points
			if pointsValue, err := points.NewPoints(pointsAmount); err == nil {
				pointsAccount.Earn(*pointsValue, points.PointsTypeRefund, "order payment rollback", orderAgg.Order().OrderNo())
			}
			return fmt.Errorf("insufficient wallet balance")
		}
		if err := walletAgg.Pay(*money, fmt.Sprintf("ORDER-%d", orderAgg.ID()), "mixed payment"); err != nil {
			// rollback points
			if pointsValue, err := points.NewPoints(pointsAmount); err == nil {
				pointsAccount.Earn(*pointsValue, points.PointsTypeRefund, "order payment rollback", orderAgg.Order().OrderNo())
			}
			return fmt.Errorf("wallet payment failed: %w", err)
		}
	}

	// 5. mark order as paid
	paymentInfo := NewPaymentInfo(PaymentMethodMixed, walletAmount, float64(pointsAmount), 0, string(walletAgg.Wallet().Currency()))
	if err := orderAgg.MarkAsPaid(paymentInfo); err != nil {
		// rollback all
		if walletAmount > 0 {
			money, _ := wallet.NewMoney(walletAmount, walletAgg.Wallet().Currency())
			walletAgg.Refund(*money, fmt.Sprintf("ORDER-%d", orderAgg.ID()), "order payment rollback")
		}
		if pointsAmount > 0 {
			if pointsValue, err := points.NewPoints(pointsAmount); err == nil {
				pointsAccount.Earn(*pointsValue, points.PointsTypeRefund, "order payment rollback", orderAgg.Order().OrderNo())
			}
		}
		return fmt.Errorf("order payment failed: %w", err)
	}

	return nil
}

// PayWithMemberDiscount pays order with member discount
func (s *orderPaymentServiceImpl) PayWithMemberDiscount(
	ctx context.Context,
	orderAgg *OrderAggregate,
	memberAgg *membership.MemberAggregate,
	walletAgg *wallet.WalletAggregate,
) error {
	// 1. validate order
	if !orderAgg.CanPay() {
		return fmt.Errorf("order cannot be paid in current status")
	}

	// 2. apply member discount
	if err := orderAgg.Order().ApplyMemberDiscount(memberAgg.Member().Level()); err != nil {
		return fmt.Errorf("failed to apply member discount: %w", err)
	}

	// 3. pay with wallet
	return s.PayWithWallet(ctx, orderAgg, walletAgg)
}

// RefundOrder refunds order
func (s *orderPaymentServiceImpl) RefundOrder(
	ctx context.Context,
	orderAgg *OrderAggregate,
	walletAgg *wallet.WalletAggregate,
	reason string,
) error {
	// 1. validate order can be refunded
	if !orderAgg.CanRefund() {
		return fmt.Errorf("order cannot be refunded in current status")
	}

	// 2. get refund amount
	refundAmount := orderAgg.Order().FinalAmount()
	if orderAgg.Order().PaymentInfo() != nil {
		refundAmount = orderAgg.Order().PaymentInfo().WalletAmount()
	}

	// 3. refund to wallet
	if refundAmount > 0 {
		money, err := wallet.NewMoney(refundAmount, walletAgg.Wallet().Currency())
		if err != nil {
			return err
		}
		if err := walletAgg.Refund(*money, fmt.Sprintf("ORDER-%d", orderAgg.ID()), reason); err != nil {
			return fmt.Errorf("wallet refund failed: %w", err)
		}
	}

	// 4. mark order as refunded
	refundInfo := NewRefundInfo(fmt.Sprintf("REFUND-%d", orderAgg.ID()), refundAmount, reason, PaymentMethodWallet)
	if err := orderAgg.Refund(refundInfo); err != nil {
		// rollback wallet refund (this is tricky, but we assume refund always succeeds)
		return fmt.Errorf("order refund failed: %w", err)
	}

	return nil
}

// CancelUnpaidOrder cancels unpaid order
func (s *orderPaymentServiceImpl) CancelUnpaidOrder(
	ctx context.Context,
	orderAgg *OrderAggregate,
	reason string,
) error {
	// 1. validate order can be cancelled
	if !orderAgg.CanCancel() {
		return fmt.Errorf("order cannot be cancelled in current status")
	}

	// 2. only unpaid orders can be cancelled without refund
	if orderAgg.Order().Status() != OrderStatusPending {
		return fmt.Errorf("only pending orders can be cancelled without refund")
	}

	// 3. cancel order
	return orderAgg.Cancel(reason)
}

// OrderPointsService order points domain service
// Handles points earning after order completion
type OrderPointsService interface {
	// EarnPointsAfterOrder earns points after order is paid
	EarnPointsAfterOrder(ctx context.Context, orderAgg *OrderAggregate, memberAgg *membership.MemberAggregate, pointsAccount *points.PointsAccountAggregate) error
}

// orderPointsServiceImpl order points service implementation
type orderPointsServiceImpl struct{}

// NewOrderPointsService creates order points service
func NewOrderPointsService() OrderPointsService {
	return &orderPointsServiceImpl{}
}

// EarnPointsAfterOrder earns points after order is paid
func (s *orderPointsServiceImpl) EarnPointsAfterOrder(
	ctx context.Context,
	orderAgg *OrderAggregate,
	memberAgg *membership.MemberAggregate,
	pointsAccount *points.PointsAccountAggregate,
) error {
	// 1. validate order is paid
	if orderAgg.Order().Status() != OrderStatusPaid && orderAgg.Order().Status() != OrderStatusCompleted {
		return fmt.Errorf("order must be paid or completed to earn points")
	}

	// 2. calculate points to earn (1 currency unit = 1 point, multiplied by member level)
	orderAmount := orderAgg.Order().FinalAmount()
	levelConfig := memberAgg.Member().Level().Config()
	pointsToEarn := int(orderAmount * levelConfig.PointMultiplier)

	// 3. earn points
	pointsValue, err := points.NewPoints(pointsToEarn)
	if err != nil {
		return fmt.Errorf("invalid points amount: %w", err)
	}
	if err := pointsAccount.Earn(*pointsValue, points.PointsTypeEarned, "order purchase", orderAgg.Order().OrderNo()); err != nil {
		return fmt.Errorf("failed to earn points: %w", err)
	}

	// 4. add points to member (this triggers level upgrade check)
	if err := memberAgg.AddPoints(pointsToEarn, "order purchase"); err != nil {
		// rollback points account
		pointsAccount.Use(*pointsValue, "member points rollback", orderAgg.Order().OrderNo())
		return fmt.Errorf("failed to add points to member: %w", err)
	}

	return nil
}
