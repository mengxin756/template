package points

import (
	"time"

	"example.com/classic/internal/domain"
)

// PointsEarnedEvent points earned event
type PointsEarnedEvent struct {
	accountID      int
	memberID      int
	pointsEarned  int
	balance       int
	pointsType    string
	source        string
	referenceID   string
	occurredAt    time.Time
}

// NewPointsEarnedEvent creates points earned event
func NewPointsEarnedEvent(accountID, memberID, pointsEarned, balance int, pointsType, source, referenceID string) *PointsEarnedEvent {
	return &PointsEarnedEvent{
		accountID:     accountID,
		memberID:      memberID,
		pointsEarned:  pointsEarned,
		balance:       balance,
		pointsType:    pointsType,
		source:        source,
		referenceID:   referenceID,
		occurredAt:    time.Now(),
	}
}

func (e *PointsEarnedEvent) EventType() string     { return "points.earned" }
func (e *PointsEarnedEvent) OccurredAt() time.Time { return e.occurredAt }
func (e *PointsEarnedEvent) AggregateID() string   { return domain.AggregateID("points_account", e.accountID) }
func (e *PointsEarnedEvent) AccountID() int        { return e.accountID }
func (e *PointsEarnedEvent) MemberID() int         { return e.memberID }
func (e *PointsEarnedEvent) PointsEarned() int     { return e.pointsEarned }
func (e *PointsEarnedEvent) Balance() int          { return e.balance }
func (e *PointsEarnedEvent) PointsType() string    { return e.pointsType }
func (e *PointsEarnedEvent) Source() string        { return e.source }
func (e *PointsEarnedEvent) ReferenceID() string   { return e.referenceID }

// PointsUsedEvent points used event
type PointsUsedEvent struct {
	accountID    int
	memberID    int
	pointsUsed  int
	balance     int
	source      string
	referenceID string
	occurredAt  time.Time
}

// NewPointsUsedEvent creates points used event
func NewPointsUsedEvent(accountID, memberID, pointsUsed, balance int, source, referenceID string) *PointsUsedEvent {
	return &PointsUsedEvent{
		accountID:    accountID,
		memberID:    memberID,
		pointsUsed:  pointsUsed,
		balance:     balance,
		source:      source,
		referenceID: referenceID,
		occurredAt:  time.Now(),
	}
}

func (e *PointsUsedEvent) EventType() string     { return "points.used" }
func (e *PointsUsedEvent) OccurredAt() time.Time { return e.occurredAt }
func (e *PointsUsedEvent) AggregateID() string   { return domain.AggregateID("points_account", e.accountID) }
func (e *PointsUsedEvent) AccountID() int        { return e.accountID }
func (e *PointsUsedEvent) MemberID() int         { return e.memberID }
func (e *PointsUsedEvent) PointsUsed() int       { return e.pointsUsed }
func (e *PointsUsedEvent) Balance() int          { return e.balance }
func (e *PointsUsedEvent) Source() string        { return e.source }
func (e *PointsUsedEvent) ReferenceID() string   { return e.referenceID }

// PointsExpiredEvent points expired event
type PointsExpiredEvent struct {
	accountID     int
	memberID     int
	pointsExpired int
	balance      int
	occurredAt   time.Time
}

// NewPointsExpiredEvent creates points expired event
func NewPointsExpiredEvent(accountID, memberID, pointsExpired, balance int) *PointsExpiredEvent {
	return &PointsExpiredEvent{
		accountID:     accountID,
		memberID:      memberID,
		pointsExpired: pointsExpired,
		balance:       balance,
		occurredAt:    time.Now(),
	}
}

func (e *PointsExpiredEvent) EventType() string     { return "points.expired" }
func (e *PointsExpiredEvent) OccurredAt() time.Time { return e.occurredAt }
func (e *PointsExpiredEvent) AggregateID() string   { return domain.AggregateID("points_account", e.accountID) }
func (e *PointsExpiredEvent) AccountID() int        { return e.accountID }
func (e *PointsExpiredEvent) MemberID() int         { return e.memberID }
func (e *PointsExpiredEvent) PointsExpired() int    { return e.pointsExpired }
func (e *PointsExpiredEvent) Balance() int          { return e.balance }

// PointsAdjustedEvent points adjusted event
type PointsAdjustedEvent struct {
	accountID  int
	memberID  int
	adjustment int
	balance   int
	reason    string
	occurredAt time.Time
}

// NewPointsAdjustedEvent creates points adjusted event
func NewPointsAdjustedEvent(accountID, memberID, adjustment, balance int, reason string) *PointsAdjustedEvent {
	return &PointsAdjustedEvent{
		accountID:   accountID,
		memberID:    memberID,
		adjustment:  adjustment,
		balance:     balance,
		reason:      reason,
		occurredAt:  time.Now(),
	}
}

func (e *PointsAdjustedEvent) EventType() string     { return "points.adjusted" }
func (e *PointsAdjustedEvent) OccurredAt() time.Time { return e.occurredAt }
func (e *PointsAdjustedEvent) AggregateID() string   { return domain.AggregateID("points_account", e.accountID) }
func (e *PointsAdjustedEvent) AccountID() int        { return e.accountID }
func (e *PointsAdjustedEvent) MemberID() int         { return e.memberID }
func (e *PointsAdjustedEvent) Adjustment() int       { return e.adjustment }
func (e *PointsAdjustedEvent) Balance() int          { return e.balance }
func (e *PointsAdjustedEvent) Reason() string        { return e.reason }
