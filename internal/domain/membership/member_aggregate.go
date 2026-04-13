package membership

import (
	"fmt"
	"time"

	"example.com/classic/internal/domain"
)

// Member member entity
type Member struct {
	id             int
	userID         int
	level          MemberLevel
	status         MemberStatus
	totalPoints    int       // total accumulated points
	currentPoints  int       // current available points
	expiredAt      time.Time // membership expiration time
	createdAt      time.Time
	updatedAt      time.Time
	statusHistory  []*StatusChangeRecord
}

// NewMember creates new member
func NewMember(userID int) *Member {
	now := time.Now()
	return &Member{
		userID:        userID,
		level:         LevelNormal,
		status:        StatusActive,
		totalPoints:   0,
		currentPoints: 0,
		expiredAt:     now.AddDate(1, 0, 0), // default 1 year
		createdAt:     now,
		updatedAt:     now,
		statusHistory: make([]*StatusChangeRecord, 0),
	}
}

// RebuildMember rebuilds member from persistence
func RebuildMember(
	id int,
	userID int,
	level MemberLevel,
	status MemberStatus,
	totalPoints int,
	currentPoints int,
	expiredAt time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) *Member {
	return &Member{
		id:            id,
		userID:        userID,
		level:         level,
		status:        status,
		totalPoints:   totalPoints,
		currentPoints: currentPoints,
		expiredAt:     expiredAt,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
		statusHistory: make([]*StatusChangeRecord, 0),
	}
}

// Getters
func (m *Member) ID() int                   { return m.id }
func (m *Member) UserID() int               { return m.userID }
func (m *Member) Level() MemberLevel        { return m.level }
func (m *Member) Status() MemberStatus      { return m.status }
func (m *Member) TotalPoints() int          { return m.totalPoints }
func (m *Member) CurrentPoints() int        { return m.currentPoints }
func (m *Member) ExpiredAt() time.Time      { return m.expiredAt }
func (m *Member) CreatedAt() time.Time      { return m.createdAt }
func (m *Member) UpdatedAt() time.Time      { return m.updatedAt }
func (m *Member) StatusHistory() []*StatusChangeRecord { return m.statusHistory }

// SetID sets ID (called by repository after persistence)
func (m *Member) SetID(id int) {
	m.id = id
}

// MemberAggregate member aggregate root
type MemberAggregate struct {
	member        *Member
	events        *domain.EventRecorder
}

// NewMemberAggregate creates new member aggregate
func NewMemberAggregate(userID int) *MemberAggregate {
	member := NewMember(userID)
	return &MemberAggregate{
		member: member,
		events: domain.NewEventRecorder(),
	}
}

// RebuildMemberAggregate rebuilds member aggregate from member entity
func RebuildMemberAggregate(member *Member) *MemberAggregate {
	return &MemberAggregate{
		member: member,
		events: domain.NewEventRecorder(),
	}
}

// Member returns member entity
func (a *MemberAggregate) Member() *Member {
	return a.member
}

// ID returns aggregate ID
func (a *MemberAggregate) ID() int {
	return a.member.ID()
}

// AddPoints adds points to member (business rule: points affect level)
func (a *MemberAggregate) AddPoints(points int, reason string) error {
	if a.member.status != StatusActive {
		return fmt.Errorf("cannot add points to non-active member")
	}

	if points <= 0 {
		return fmt.Errorf("points must be positive")
	}

	oldLevel := a.member.level
	a.member.totalPoints += points
	a.member.currentPoints += points
	a.member.updatedAt = time.Now()

	// auto upgrade level based on total points
	newLevel := CalculateLevel(a.member.totalPoints)
	if newLevel != oldLevel && newLevel.CanUpgradeTo(oldLevel) {
		a.member.level = newLevel
		a.events.AddEvent(NewMemberLevelUpgradedEvent(
			a.member.id,
			a.member.userID,
			oldLevel,
			newLevel,
			a.member.totalPoints,
		))
	}

	// record points added event
	a.events.AddEvent(NewPointsAddedEvent(
		a.member.id,
		a.member.userID,
		points,
		a.member.currentPoints,
		reason,
	))

	return nil
}

// UsePoints uses points from member
func (a *MemberAggregate) UsePoints(points int, reason string) error {
	if a.member.status != StatusActive {
		return fmt.Errorf("cannot use points for non-active member")
	}

	if points <= 0 {
		return fmt.Errorf("points must be positive")
	}

	if a.member.currentPoints < points {
		return fmt.Errorf("insufficient points: have %d, need %d", a.member.currentPoints, points)
	}

	a.member.currentPoints -= points
	a.member.updatedAt = time.Now()

	// record points used event
	a.events.AddEvent(NewPointsUsedEvent(
		a.member.id,
		a.member.userID,
		points,
		a.member.currentPoints,
		reason,
	))

	return nil
}

// ChangeStatus changes member status (state machine)
func (a *MemberAggregate) ChangeStatus(newStatus MemberStatus, reason StatusChangeReason, operator, remark string) error {
	if !a.member.status.CanTransitionTo(newStatus) {
		return fmt.Errorf("invalid status transition: %s -> %s", a.member.status, newStatus)
	}

	oldStatus := a.member.status
	a.member.status = newStatus
	a.member.updatedAt = time.Now()

	// record status change
	record := NewStatusChangeRecord(oldStatus, newStatus, reason, operator, remark)
	a.member.statusHistory = append(a.member.statusHistory, record)

	// emit status changed event
	a.events.AddEvent(NewMemberStatusChangedEvent(
		a.member.id,
		a.member.userID,
		oldStatus,
		newStatus,
		string(reason),
	))

	return nil
}

// Renew renews membership
func (a *MemberAggregate) Renew(duration time.Duration) error {
	if a.member.status == StatusCancelled {
		return fmt.Errorf("cannot renew cancelled membership")
	}

	a.member.expiredAt = time.Now().Add(duration)
	a.member.updatedAt = time.Now()

	// if expired, reactivate
	if a.member.status == StatusExpired {
		if err := a.ChangeStatus(StatusActive, ReasonReactivation, "system", "membership renewed"); err != nil {
			return err
		}
	}

	a.events.AddEvent(NewMembershipRenewedEvent(
		a.member.id,
		a.member.userID,
		a.member.expiredAt,
	))

	return nil
}

// Suspend suspends membership
func (a *MemberAggregate) Suspend(reason, operator string) error {
	return a.ChangeStatus(StatusSuspended, ReasonViolation, operator, reason)
}

// Activate activates membership
func (a *MemberAggregate) Activate(operator, remark string) error {
	return a.ChangeStatus(StatusActive, ReasonReactivation, operator, remark)
}

// Cancel cancels membership
func (a *MemberAggregate) Cancel(reason, operator string) error {
	return a.ChangeStatus(StatusCancelled, ReasonUserRequest, operator, reason)
}

// Expire expires membership (called by system)
func (a *MemberAggregate) Expire() error {
	if a.member.status != StatusActive {
		return nil // already not active
	}

	return a.ChangeStatus(StatusExpired, ReasonExpiration, "system", "membership expired")
}

// IsExpired checks if membership is expired
func (a *MemberAggregate) IsExpired() bool {
	return time.Now().After(a.member.expiredAt)
}

// IsActive checks if member is active
func (a *MemberAggregate) IsActive() bool {
	return a.member.status == StatusActive && !a.IsExpired()
}

// CanUsePoints checks if member can use points
func (a *MemberAggregate) CanUsePoints(points int) bool {
	return a.IsActive() && a.member.currentPoints >= points
}

// CalculateDiscount calculates discount for this member
func (a *MemberAggregate) CalculateDiscount(originalPrice Money) (Money, error) {
	strategy := NewLevelDiscountStrategy(a.member.level)
	return strategy.CalculateDiscount(originalPrice)
}

// Events returns domain events
func (a *MemberAggregate) Events() []domain.DomainEvent {
	return a.events.Events()
}

// ClearEvents clears domain events
func (a *MemberAggregate) ClearEvents() {
	a.events.ClearEvents()
}

// HasEvents checks if has events
func (a *MemberAggregate) HasEvents() bool {
	return a.events.HasEvents()
}
