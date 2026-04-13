package points

import (
	"fmt"
	"time"

	"example.com/classic/internal/domain"
)

// PointsAccount points account entity
type PointsAccount struct {
	id           int
	memberID     int
	totalEarned  Points // total earned points
	totalUsed    Points // total used points
	currentBalance Points // current available balance
	expiredPoints Points // expired points
	createdAt    time.Time
	updatedAt    time.Time
}

// NewPointsAccount creates new points account
func NewPointsAccount(memberID int) *PointsAccount {
	now := time.Now()
	return &PointsAccount{
		memberID:       memberID,
		totalEarned:    Points{value: 0},
		totalUsed:      Points{value: 0},
		currentBalance: Points{value: 0},
		expiredPoints:  Points{value: 0},
		createdAt:      now,
		updatedAt:      now,
	}
}

// RebuildPointsAccount rebuilds points account from persistence
func RebuildPointsAccount(
	id int,
	memberID int,
	totalEarned int,
	totalUsed int,
	currentBalance int,
	expiredPoints int,
	createdAt time.Time,
	updatedAt time.Time,
) *PointsAccount {
	return &PointsAccount{
		id:             id,
		memberID:       memberID,
		totalEarned:    Points{value: totalEarned},
		totalUsed:      Points{value: totalUsed},
		currentBalance: Points{value: currentBalance},
		expiredPoints:  Points{value: expiredPoints},
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}

// Getters
func (a *PointsAccount) ID() int              { return a.id }
func (a *PointsAccount) MemberID() int        { return a.memberID }
func (a *PointsAccount) TotalEarned() Points  { return a.totalEarned }
func (a *PointsAccount) TotalUsed() Points    { return a.totalUsed }
func (a *PointsAccount) CurrentBalance() Points { return a.currentBalance }
func (a *PointsAccount) ExpiredPoints() Points { return a.expiredPoints }
func (a *PointsAccount) CreatedAt() time.Time { return a.createdAt }
func (a *PointsAccount) UpdatedAt() time.Time { return a.updatedAt }

// SetID sets ID (called by repository)
func (a *PointsAccount) SetID(id int) {
	a.id = id
}

// PointsAccountAggregate points account aggregate root
type PointsAccountAggregate struct {
	account     *PointsAccount
	records     []*PointsRecord
	pendingRecords []*PointsRecord // records to be persisted
	events      *domain.EventRecorder
	policy      ExpirationPolicy
}

// NewPointsAccountAggregate creates new points account aggregate
func NewPointsAccountAggregate(memberID int, policy ExpirationPolicy) *PointsAccountAggregate {
	account := NewPointsAccount(memberID)
	return &PointsAccountAggregate{
		account:         account,
		records:         make([]*PointsRecord, 0),
		pendingRecords:  make([]*PointsRecord, 0),
		events:          domain.NewEventRecorder(),
		policy:          policy,
	}
}

// RebuildPointsAccountAggregate rebuilds points account aggregate
func RebuildPointsAccountAggregate(account *PointsAccount, records []*PointsRecord, policy ExpirationPolicy) *PointsAccountAggregate {
	return &PointsAccountAggregate{
		account:        account,
		records:        records,
		pendingRecords: make([]*PointsRecord, 0),
		events:         domain.NewEventRecorder(),
		policy:         policy,
	}
}

// Account returns points account entity
func (a *PointsAccountAggregate) Account() *PointsAccount {
	return a.account
}

// Records returns points records
func (a *PointsAccountAggregate) Records() []*PointsRecord {
	return a.records
}

// PendingRecords returns pending records (to be persisted)
func (a *PointsAccountAggregate) PendingRecords() []*PointsRecord {
	return a.pendingRecords
}

// ClearPendingRecords clears pending records
func (a *PointsAccountAggregate) ClearPendingRecords() {
	a.pendingRecords = make([]*PointsRecord, 0)
}

// ID returns aggregate ID
func (a *PointsAccountAggregate) ID() int {
	return a.account.ID()
}

// Earn earns points
func (a *PointsAccountAggregate) Earn(points Points, pointsType PointsType, source, referenceID string) error {
	if points.Value() <= 0 {
		return fmt.Errorf("points must be positive")
	}

	// update account
	a.account.totalEarned = a.account.totalEarned.Add(points)
	a.account.currentBalance = a.account.currentBalance.Add(points)
	a.account.updatedAt = time.Now()

	// calculate expiration
	expiredAt := a.policy.CalculateExpiredAt(time.Now())

	// create record
	record := NewPointsRecord(
		0, // ID will be set by repository
		a.account.memberID,
		points,
		pointsType,
		a.account.currentBalance,
		source,
		referenceID,
		expiredAt,
	)
	a.pendingRecords = append(a.pendingRecords, record)

	// emit event
	a.events.AddEvent(NewPointsEarnedEvent(
		a.account.id,
		a.account.memberID,
		points.Value(),
		a.account.currentBalance.Value(),
		string(pointsType),
		source,
		referenceID,
	))

	return nil
}

// Use uses points
func (a *PointsAccountAggregate) Use(points Points, source, referenceID string) error {
	if points.Value() <= 0 {
		return fmt.Errorf("points must be positive")
	}

	if a.account.currentBalance.Value() < points.Value() {
		return fmt.Errorf("insufficient points: have %d, need %d", a.account.currentBalance.Value(), points.Value())
	}

	// update account
	newBalance, err := a.account.currentBalance.Subtract(points)
	if err != nil {
		return err
	}
	a.account.currentBalance = newBalance
	a.account.totalUsed = a.account.totalUsed.Add(points)
	a.account.updatedAt = time.Now()

	// create record (negative points for usage)
	record := NewPointsRecord(
		0,
		a.account.memberID,
		Points{value: -points.Value()},
		PointsTypePurchase,
		a.account.currentBalance,
		source,
		referenceID,
		time.Time{}, // no expiration for used points
	)
	a.pendingRecords = append(a.pendingRecords, record)

	// emit event
	a.events.AddEvent(NewPointsUsedEvent(
		a.account.id,
		a.account.memberID,
		points.Value(),
		a.account.currentBalance.Value(),
		source,
		referenceID,
	))

	return nil
}

// Expire expires points that are past expiration date
func (a *PointsAccountAggregate) Expire() (Points, error) {
	var totalExpired Points
	now := time.Now()

	// find expired records with positive balance
	for _, record := range a.records {
		if record.Points().Value() > 0 && !record.ExpiredAt().IsZero() && now.After(record.ExpiredAt()) {
			// this record is expired
			expiredPoints := record.Points()
			totalExpired = totalExpired.Add(expiredPoints)
		}
	}

	if totalExpired.IsZero() {
		return totalExpired, nil
	}

	// update account
	newBalance, err := a.account.currentBalance.Subtract(totalExpired)
	if err != nil {
		return totalExpired, err
	}
	a.account.currentBalance = newBalance
	a.account.expiredPoints = a.account.expiredPoints.Add(totalExpired)
	a.account.updatedAt = now

	// create expiration record
	record := NewPointsRecord(
		0,
		a.account.memberID,
		Points{value: -totalExpired.Value()},
		PointsTypeExpired,
		a.account.currentBalance,
		"points expired",
		"",
		time.Time{},
	)
	a.pendingRecords = append(a.pendingRecords, record)

	// emit event
	a.events.AddEvent(NewPointsExpiredEvent(
		a.account.id,
		a.account.memberID,
		totalExpired.Value(),
		a.account.currentBalance.Value(),
	))

	return totalExpired, nil
}

// Adjust adjusts points manually (admin operation)
func (a *PointsAccountAggregate) Adjust(points Points, reason string) error {
	if points.Value() == 0 {
		return fmt.Errorf("adjustment points cannot be zero")
	}

	var newBalance Points
	var err error

	if points.Value() > 0 {
		// positive adjustment (add)
		newBalance = a.account.currentBalance.Add(points)
		a.account.totalEarned = a.account.totalEarned.Add(points)
	} else {
		// negative adjustment (subtract)
		newBalance, err = a.account.currentBalance.Subtract(Points{value: -points.Value()})
		if err != nil {
			return err
		}
		a.account.totalUsed = a.account.totalUsed.Add(Points{value: -points.Value()})
	}

	a.account.currentBalance = newBalance
	a.account.updatedAt = time.Now()

	// create record
	record := NewPointsRecord(
		0,
		a.account.memberID,
		points,
		PointsTypeAdjust,
		a.account.currentBalance,
		reason,
		"",
		a.policy.CalculateExpiredAt(time.Now()),
	)
	a.pendingRecords = append(a.pendingRecords, record)

	// emit event
	a.events.AddEvent(NewPointsAdjustedEvent(
		a.account.id,
		a.account.memberID,
		points.Value(),
		a.account.currentBalance.Value(),
		reason,
	))

	return nil
}

// CanUse checks if can use specified points
func (a *PointsAccountAggregate) CanUse(points Points) bool {
	return a.account.currentBalance.Value() >= points.Value()
}

// Events returns domain events
func (a *PointsAccountAggregate) Events() []domain.DomainEvent {
	return a.events.Events()
}

// ClearEvents clears domain events
func (a *PointsAccountAggregate) ClearEvents() {
	a.events.ClearEvents()
}

// HasEvents checks if has events
func (a *PointsAccountAggregate) HasEvents() bool {
	return a.events.HasEvents()
}
