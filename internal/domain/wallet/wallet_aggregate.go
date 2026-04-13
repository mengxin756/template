package wallet

import (
	"fmt"
	"time"

	"example.com/classic/internal/domain"
)

// Transaction transaction entity
type Transaction struct {
	id              int
	walletID        int
	transactionType TransactionType
	amount          Money
	balance         Money    // balance after transaction
	status          TransactionStatus
	referenceType   string   // order, refund, transfer, etc.
	referenceID     string   // reference ID
	description     string
	operatedBy      string   // operator (user or system)
	operatedAt      time.Time
	createdAt       time.Time
}

// NewTransaction creates new transaction
func NewTransaction(
	walletID int,
	transactionType TransactionType,
	amount Money,
	balance Money,
	referenceType string,
	referenceID string,
	description string,
	operatedBy string,
) *Transaction {
	now := time.Now()
	return &Transaction{
		walletID:        walletID,
		transactionType: transactionType,
		amount:          amount,
		balance:         balance,
		status:          TransactionStatusPending,
		referenceType:   referenceType,
		referenceID:     referenceID,
		description:     description,
		operatedBy:      operatedBy,
		operatedAt:      now,
		createdAt:       now,
	}
}

// Getters
func (t *Transaction) ID() int                       { return t.id }
func (t *Transaction) WalletID() int                 { return t.walletID }
func (t *Transaction) TransactionType() TransactionType { return t.transactionType }
func (t *Transaction) Amount() Money                  { return t.amount }
func (t *Transaction) Balance() Money                 { return t.balance }
func (t *Transaction) Status() TransactionStatus     { return t.status }
func (t *Transaction) ReferenceType() string         { return t.referenceType }
func (t *Transaction) ReferenceID() string           { return t.referenceID }
func (t *Transaction) Description() string           { return t.description }
func (t *Transaction) OperatedBy() string            { return t.operatedBy }
func (t *Transaction) OperatedAt() time.Time         { return t.operatedAt }
func (t *Transaction) CreatedAt() time.Time          { return t.createdAt }

// SetID sets ID (called by repository)
func (t *Transaction) SetID(id int) {
	t.id = id
}

// Complete marks transaction as completed
func (t *Transaction) Complete() error {
	if t.status != TransactionStatusPending {
		return fmt.Errorf("can only complete pending transaction")
	}
	t.status = TransactionStatusCompleted
	return nil
}

// Fail marks transaction as failed
func (t *Transaction) Fail(reason string) error {
	if t.status != TransactionStatusPending {
		return fmt.Errorf("can only fail pending transaction")
	}
	t.status = TransactionStatusFailed
	t.description = fmt.Sprintf("%s (failed: %s)", t.description, reason)
	return nil
}

// Cancel cancels transaction
func (t *Transaction) Cancel(reason string) error {
	if t.status != TransactionStatusPending {
		return fmt.Errorf("can only cancel pending transaction")
	}
	t.status = TransactionStatusCancelled
	t.description = fmt.Sprintf("%s (cancelled: %s)", t.description, reason)
	return nil
}

// Wallet wallet entity
type Wallet struct {
	id            int
	userID        int
	balance       Money
	frozenAmount  Money // frozen amount (cannot use)
	currency      Currency
	version       int    // optimistic lock version
	createdAt     time.Time
	updatedAt     time.Time
}

// NewWallet creates new wallet
func NewWallet(userID int, currency Currency) (*Wallet, error) {
	if !currency.IsValid() {
		return nil, fmt.Errorf("invalid currency: %s", currency)
	}
	now := time.Now()
	return &Wallet{
		userID:       userID,
		balance:      Money{amount: 0, currency: currency},
		frozenAmount: Money{amount: 0, currency: currency},
		currency:     currency,
		version:      1,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

// RebuildWallet rebuilds wallet from persistence
func RebuildWallet(
	id int,
	userID int,
	balance float64,
	frozenAmount float64,
	currency Currency,
	version int,
	createdAt time.Time,
	updatedAt time.Time,
) *Wallet {
	return &Wallet{
		id:           id,
		userID:       userID,
		balance:      Money{amount: balance, currency: currency},
		frozenAmount: Money{amount: frozenAmount, currency: currency},
		currency:     currency,
		version:      version,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

// Getters
func (w *Wallet) ID() int               { return w.id }
func (w *Wallet) UserID() int           { return w.userID }
func (w *Wallet) Balance() Money        { return w.balance }
func (w *Wallet) FrozenAmount() Money   { return w.frozenAmount }
func (w *Wallet) AvailableBalance() Money {
	available, _ := w.balance.Subtract(w.frozenAmount)
	return available
}
func (w *Wallet) Currency() Currency    { return w.currency }
func (w *Wallet) Version() int          { return w.version }
func (w *Wallet) CreatedAt() time.Time  { return w.createdAt }
func (w *Wallet) UpdatedAt() time.Time  { return w.updatedAt }

// SetID sets ID (called by repository)
func (w *Wallet) SetID(id int) {
	w.id = id
}

// incrementVersion increments version for optimistic lock
func (w *Wallet) incrementVersion() {
	w.version++
	w.updatedAt = time.Now()
}

// WalletAggregate wallet aggregate root
type WalletAggregate struct {
	wallet          *Wallet
	pendingTransactions []*Transaction
	transactions    []*Transaction
	events          *domain.EventRecorder
}

// NewWalletAggregate creates new wallet aggregate
func NewWalletAggregate(userID int, currency Currency) (*WalletAggregate, error) {
	wallet, err := NewWallet(userID, currency)
	if err != nil {
		return nil, err
	}
	return &WalletAggregate{
		wallet:              wallet,
		pendingTransactions: make([]*Transaction, 0),
		transactions:        make([]*Transaction, 0),
		events:              domain.NewEventRecorder(),
	}, nil
}

// RebuildWalletAggregate rebuilds wallet aggregate
func RebuildWalletAggregate(wallet *Wallet, transactions []*Transaction) *WalletAggregate {
	return &WalletAggregate{
		wallet:              wallet,
		pendingTransactions: make([]*Transaction, 0),
		transactions:        transactions,
		events:              domain.NewEventRecorder(),
	}
}

// Wallet returns wallet entity
func (a *WalletAggregate) Wallet() *Wallet {
	return a.wallet
}

// Transactions returns transactions
func (a *WalletAggregate) Transactions() []*Transaction {
	return a.transactions
}

// PendingTransactions returns pending transactions
func (a *WalletAggregate) PendingTransactions() []*Transaction {
	return a.pendingTransactions
}

// ClearPendingTransactions clears pending transactions
func (a *WalletAggregate) ClearPendingTransactions() {
	a.pendingTransactions = make([]*Transaction, 0)
}

// ID returns aggregate ID
func (a *WalletAggregate) ID() int {
	return a.wallet.ID()
}

// Deposit deposits money
func (a *WalletAggregate) Deposit(amount Money, referenceType, referenceID, description, operatedBy string) error {
	if amount.Currency() != a.wallet.currency {
		return fmt.Errorf("currency mismatch: wallet is %s, deposit is %s", a.wallet.currency, amount.Currency())
	}
	if amount.IsZero() {
		return fmt.Errorf("deposit amount cannot be zero")
	}

	// update balance
	newBalance, err := a.wallet.balance.Add(amount)
	if err != nil {
		return err
	}
	a.wallet.balance = newBalance
	a.wallet.incrementVersion()

	// create transaction
	transaction := NewTransaction(
		a.wallet.id,
		TransactionTypeDeposit,
		amount,
		a.wallet.balance,
		referenceType,
		referenceID,
		description,
		operatedBy,
	)
	transaction.Complete() // deposit is immediate
	a.pendingTransactions = append(a.pendingTransactions, transaction)

	// emit event
	a.events.AddEvent(NewWalletDepositedEvent(
		a.wallet.id,
		a.wallet.userID,
		amount.Amount(),
		a.wallet.balance.Amount(),
		string(a.wallet.currency),
		referenceType,
		referenceID,
	))

	return nil
}

// Withdraw withdraws money
func (a *WalletAggregate) Withdraw(amount Money, referenceType, referenceID, description, operatedBy string) error {
	if amount.Currency() != a.wallet.currency {
		return fmt.Errorf("currency mismatch: wallet is %s, withdraw is %s", a.wallet.currency, amount.Currency())
	}
	if amount.IsZero() {
		return fmt.Errorf("withdraw amount cannot be zero")
	}

	available := a.wallet.AvailableBalance()
	if available.Amount() < amount.Amount() {
		return fmt.Errorf("insufficient balance: available %.2f, need %.2f", available.Amount(), amount.Amount())
	}

	// update balance
	newBalance, err := a.wallet.balance.Subtract(amount)
	if err != nil {
		return err
	}
	a.wallet.balance = newBalance
	a.wallet.incrementVersion()

	// create transaction
	transaction := NewTransaction(
		a.wallet.id,
		TransactionTypeWithdraw,
		amount,
		a.wallet.balance,
		referenceType,
		referenceID,
		description,
		operatedBy,
	)
	transaction.Complete()
	a.pendingTransactions = append(a.pendingTransactions, transaction)

	// emit event
	a.events.AddEvent(NewWalletWithdrawnEvent(
		a.wallet.id,
		a.wallet.userID,
		amount.Amount(),
		a.wallet.balance.Amount(),
		string(a.wallet.currency),
		referenceType,
		referenceID,
	))

	return nil
}

// Pay pays for order
func (a *WalletAggregate) Pay(amount Money, orderID string, description string) error {
	if amount.Currency() != a.wallet.currency {
		return fmt.Errorf("currency mismatch")
	}
	if amount.IsZero() {
		return fmt.Errorf("payment amount cannot be zero")
	}

	available := a.wallet.AvailableBalance()
	if available.Amount() < amount.Amount() {
		return fmt.Errorf("insufficient balance: available %.2f, need %.2f", available.Amount(), amount.Amount())
	}

	// update balance
	newBalance, err := a.wallet.balance.Subtract(amount)
	if err != nil {
		return err
	}
	a.wallet.balance = newBalance
	a.wallet.incrementVersion()

	// create transaction
	transaction := NewTransaction(
		a.wallet.id,
		TransactionTypePayment,
		amount,
		a.wallet.balance,
		"order",
		orderID,
		description,
		"user",
	)
	transaction.Complete()
	a.pendingTransactions = append(a.pendingTransactions, transaction)

	// emit event
	a.events.AddEvent(NewWalletPaymentEvent(
		a.wallet.id,
		a.wallet.userID,
		amount.Amount(),
		a.wallet.balance.Amount(),
		string(a.wallet.currency),
		orderID,
	))

	return nil
}

// Refund refunds money
func (a *WalletAggregate) Refund(amount Money, orderID string, reason string) error {
	if amount.Currency() != a.wallet.currency {
		return fmt.Errorf("currency mismatch")
	}
	if amount.IsZero() {
		return fmt.Errorf("refund amount cannot be zero")
	}

	// update balance
	newBalance, err := a.wallet.balance.Add(amount)
	if err != nil {
		return err
	}
	a.wallet.balance = newBalance
	a.wallet.incrementVersion()

	// create transaction
	transaction := NewTransaction(
		a.wallet.id,
		TransactionTypeRefund,
		amount,
		a.wallet.balance,
		"order",
		orderID,
		fmt.Sprintf("refund: %s", reason),
		"system",
	)
	transaction.Complete()
	a.pendingTransactions = append(a.pendingTransactions, transaction)

	// emit event
	a.events.AddEvent(NewWalletRefundedEvent(
		a.wallet.id,
		a.wallet.userID,
		amount.Amount(),
		a.wallet.balance.Amount(),
		string(a.wallet.currency),
		orderID,
		reason,
	))

	return nil
}

// Freeze freezes amount
func (a *WalletAggregate) Freeze(amount Money, reason string) error {
	if amount.Currency() != a.wallet.currency {
		return fmt.Errorf("currency mismatch")
	}
	if amount.IsZero() {
		return fmt.Errorf("freeze amount cannot be zero")
	}

	available := a.wallet.AvailableBalance()
	if available.Amount() < amount.Amount() {
		return fmt.Errorf("insufficient balance to freeze")
	}

	// update frozen amount
	newFrozen, err := a.wallet.frozenAmount.Add(amount)
	if err != nil {
		return err
	}
	a.wallet.frozenAmount = newFrozen
	a.wallet.incrementVersion()

	// create transaction
	transaction := NewTransaction(
		a.wallet.id,
		TransactionTypeFreeze,
		amount,
		a.wallet.balance,
		"freeze",
		"",
		reason,
		"system",
	)
	transaction.Complete()
	a.pendingTransactions = append(a.pendingTransactions, transaction)

	// emit event
	a.events.AddEvent(NewWalletFrozenEvent(
		a.wallet.id,
		a.wallet.userID,
		amount.Amount(),
		a.wallet.frozenAmount.Amount(),
		string(a.wallet.currency),
		reason,
	))

	return nil
}

// Unfreeze unfreezes amount
func (a *WalletAggregate) Unfreeze(amount Money, reason string) error {
	if amount.Currency() != a.wallet.currency {
		return fmt.Errorf("currency mismatch")
	}
	if amount.IsZero() {
		return fmt.Errorf("unfreeze amount cannot be zero")
	}

	if a.wallet.frozenAmount.Amount() < amount.Amount() {
		return fmt.Errorf("insufficient frozen amount")
	}

	// update frozen amount
	newFrozen, err := a.wallet.frozenAmount.Subtract(amount)
	if err != nil {
		return err
	}
	a.wallet.frozenAmount = newFrozen
	a.wallet.incrementVersion()

	// create transaction
	transaction := NewTransaction(
		a.wallet.id,
		TransactionTypeUnfreeze,
		amount,
		a.wallet.balance,
		"unfreeze",
		"",
		reason,
		"system",
	)
	transaction.Complete()
	a.pendingTransactions = append(a.pendingTransactions, transaction)

	// emit event
	a.events.AddEvent(NewWalletUnfrozenEvent(
		a.wallet.id,
		a.wallet.userID,
		amount.Amount(),
		a.wallet.frozenAmount.Amount(),
		string(a.wallet.currency),
		reason,
	))

	return nil
}

// CanPay checks if can pay specified amount
func (a *WalletAggregate) CanPay(amount Money) bool {
	if amount.Currency() != a.wallet.currency {
		return false
	}
	return a.wallet.AvailableBalance().Amount() >= amount.Amount()
}

// Events returns domain events
func (a *WalletAggregate) Events() []domain.DomainEvent {
	return a.events.Events()
}

// ClearEvents clears domain events
func (a *WalletAggregate) ClearEvents() {
	a.events.ClearEvents()
}

// HasEvents checks if has events
func (a *WalletAggregate) HasEvents() bool {
	return a.events.HasEvents()
}
