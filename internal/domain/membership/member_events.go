package membership

import (
	"time"

	"example.com/classic/internal/domain"
)

// MemberLevelUpgradedEvent member level upgraded event
type MemberLevelUpgradedEvent struct {
	memberID    int
	userID      int
	oldLevel    MemberLevel
	newLevel    MemberLevel
	totalPoints int
	occurredAt  time.Time
}

// NewMemberLevelUpgradedEvent creates level upgraded event
func NewMemberLevelUpgradedEvent(memberID, userID int, oldLevel, newLevel MemberLevel, totalPoints int) *MemberLevelUpgradedEvent {
	return &MemberLevelUpgradedEvent{
		memberID:    memberID,
		userID:      userID,
		oldLevel:    oldLevel,
		newLevel:    newLevel,
		totalPoints: totalPoints,
		occurredAt:  time.Now(),
	}
}

func (e *MemberLevelUpgradedEvent) EventType() string    { return "member.level_upgraded" }
func (e *MemberLevelUpgradedEvent) OccurredAt() time.Time { return e.occurredAt }
func (e *MemberLevelUpgradedEvent) AggregateID() string  { return domain.AggregateID("member", e.memberID) }
func (e *MemberLevelUpgradedEvent) MemberID() int        { return e.memberID }
func (e *MemberLevelUpgradedEvent) UserID() int          { return e.userID }
func (e *MemberLevelUpgradedEvent) OldLevel() MemberLevel { return e.oldLevel }
func (e *MemberLevelUpgradedEvent) NewLevel() MemberLevel { return e.newLevel }
func (e *MemberLevelUpgradedEvent) TotalPoints() int     { return e.totalPoints }

// PointsAddedEvent points added event
type PointsAddedEvent struct {
	memberID      int
	userID        int
	pointsAdded   int
	currentPoints int
	reason        string
	occurredAt    time.Time
}

// NewPointsAddedEvent creates points added event
func NewPointsAddedEvent(memberID, userID, pointsAdded, currentPoints int, reason string) *PointsAddedEvent {
	return &PointsAddedEvent{
		memberID:      memberID,
		userID:        userID,
		pointsAdded:   pointsAdded,
		currentPoints: currentPoints,
		reason:        reason,
		occurredAt:    time.Now(),
	}
}

func (e *PointsAddedEvent) EventType() string     { return "member.points_added" }
func (e *PointsAddedEvent) OccurredAt() time.Time { return e.occurredAt }
func (e *PointsAddedEvent) AggregateID() string   { return domain.AggregateID("member", e.memberID) }
func (e *PointsAddedEvent) MemberID() int         { return e.memberID }
func (e *PointsAddedEvent) UserID() int           { return e.userID }
func (e *PointsAddedEvent) PointsAdded() int      { return e.pointsAdded }
func (e *PointsAddedEvent) CurrentPoints() int    { return e.currentPoints }
func (e *PointsAddedEvent) Reason() string        { return e.reason }

// PointsUsedEvent points used event
type PointsUsedEvent struct {
	memberID      int
	userID        int
	pointsUsed    int
	currentPoints int
	reason        string
	occurredAt    time.Time
}

// NewPointsUsedEvent creates points used event
func NewPointsUsedEvent(memberID, userID, pointsUsed, currentPoints int, reason string) *PointsUsedEvent {
	return &PointsUsedEvent{
		memberID:      memberID,
		userID:        userID,
		pointsUsed:    pointsUsed,
		currentPoints: currentPoints,
		reason:        reason,
		occurredAt:    time.Now(),
	}
}

func (e *PointsUsedEvent) EventType() string     { return "member.points_used" }
func (e *PointsUsedEvent) OccurredAt() time.Time { return e.occurredAt }
func (e *PointsUsedEvent) AggregateID() string   { return domain.AggregateID("member", e.memberID) }
func (e *PointsUsedEvent) MemberID() int         { return e.memberID }
func (e *PointsUsedEvent) UserID() int           { return e.userID }
func (e *PointsUsedEvent) PointsUsed() int       { return e.pointsUsed }
func (e *PointsUsedEvent) CurrentPoints() int    { return e.currentPoints }
func (e *PointsUsedEvent) Reason() string        { return e.reason }

// MemberStatusChangedEvent member status changed event
type MemberStatusChangedEvent struct {
	memberID   int
	userID     int
	oldStatus  MemberStatus
	newStatus  MemberStatus
	reason     string
	occurredAt time.Time
}

// NewMemberStatusChangedEvent creates status changed event
func NewMemberStatusChangedEvent(memberID, userID int, oldStatus, newStatus MemberStatus, reason string) *MemberStatusChangedEvent {
	return &MemberStatusChangedEvent{
		memberID:   memberID,
		userID:     userID,
		oldStatus:  oldStatus,
		newStatus:  newStatus,
		reason:     reason,
		occurredAt: time.Now(),
	}
}

func (e *MemberStatusChangedEvent) EventType() string     { return "member.status_changed" }
func (e *MemberStatusChangedEvent) OccurredAt() time.Time { return e.occurredAt }
func (e *MemberStatusChangedEvent) AggregateID() string   { return domain.AggregateID("member", e.memberID) }
func (e *MemberStatusChangedEvent) MemberID() int         { return e.memberID }
func (e *MemberStatusChangedEvent) UserID() int           { return e.userID }
func (e *MemberStatusChangedEvent) OldStatus() MemberStatus { return e.oldStatus }
func (e *MemberStatusChangedEvent) NewStatus() MemberStatus { return e.newStatus }
func (e *MemberStatusChangedEvent) Reason() string        { return e.reason }

// MembershipRenewedEvent membership renewed event
type MembershipRenewedEvent struct {
	memberID    int
	userID      int
	newExpiredAt time.Time
	occurredAt  time.Time
}

// NewMembershipRenewedEvent creates membership renewed event
func NewMembershipRenewedEvent(memberID, userID int, newExpiredAt time.Time) *MembershipRenewedEvent {
	return &MembershipRenewedEvent{
		memberID:     memberID,
		userID:       userID,
		newExpiredAt: newExpiredAt,
		occurredAt:   time.Now(),
	}
}

func (e *MembershipRenewedEvent) EventType() string     { return "member.membership_renewed" }
func (e *MembershipRenewedEvent) OccurredAt() time.Time { return e.occurredAt }
func (e *MembershipRenewedEvent) AggregateID() string   { return domain.AggregateID("member", e.memberID) }
func (e *MembershipRenewedEvent) MemberID() int         { return e.memberID }
func (e *MembershipRenewedEvent) UserID() int           { return e.userID }
func (e *MembershipRenewedEvent) NewExpiredAt() time.Time { return e.newExpiredAt }
