package membership

import (
	"fmt"
	"time"
)

// MemberStatus member status enum
type MemberStatus string

const (
	StatusActive    MemberStatus = "active"    // active member
	StatusExpired   MemberStatus = "expired"   // expired member
	StatusSuspended MemberStatus = "suspended" // suspended member
	StatusCancelled MemberStatus = "cancelled" // cancelled member
)

// status transitions: defines valid status transitions
var statusTransitions = map[MemberStatus][]MemberStatus{
	StatusActive: {
		StatusSuspended,
		StatusCancelled,
		StatusExpired,
	},
	StatusExpired: {
		StatusActive, // can reactivate
		StatusCancelled,
	},
	StatusSuspended: {
		StatusActive,    // can reactivate
		StatusCancelled, // can cancel
	},
	StatusCancelled: {}, // terminal state, no transitions
}

// IsValid validates member status
func (s MemberStatus) IsValid() bool {
	_, ok := statusTransitions[s]
	return ok
}

// CanTransitionTo checks if can transition to target status
func (s MemberStatus) CanTransitionTo(target MemberStatus) bool {
	allowed, ok := statusTransitions[s]
	if !ok {
		return false
	}

	for _, status := range allowed {
		if status == target {
			return true
		}
	}
	return false
}

// ValidTransitions returns all valid transitions from current status
func (s MemberStatus) ValidTransitions() []MemberStatus {
	return statusTransitions[s]
}

// String returns string representation
func (s MemberStatus) String() string {
	return string(s)
}

// ParseMemberStatus parses string to MemberStatus
func ParseMemberStatus(str string) (MemberStatus, error) {
	status := MemberStatus(str)
	if !status.IsValid() {
		return "", fmt.Errorf("invalid member status: %s", str)
	}
	return status, nil
}

// StatusChangeReason reason for status change
type StatusChangeReason string

const (
	ReasonManual       StatusChangeReason = "manual"        // manual operation
	ReasonNonPayment   StatusChangeReason = "non_payment"   // non payment
	ReasonViolation    StatusChangeReason = "violation"     // policy violation
	ReasonExpiration   StatusChangeReason = "expiration"    // membership expiration
	ReasonReactivation StatusChangeReason = "reactivation"  // member reactivation
	ReasonUserRequest  StatusChangeReason = "user_request" // user request
)

// StatusChangeRecord status change record
type StatusChangeRecord struct {
	FromStatus   MemberStatus
	ToStatus     MemberStatus
	Reason       StatusChangeReason
	Operator     string    // operator ID or system
	OperatedAt   time.Time
	Remark       string
}

// NewStatusChangeRecord creates a new status change record
func NewStatusChangeRecord(from, to MemberStatus, reason StatusChangeReason, operator, remark string) *StatusChangeRecord {
	return &StatusChangeRecord{
		FromStatus: from,
		ToStatus:   to,
		Reason:     reason,
		Operator:   operator,
		OperatedAt: time.Now(),
		Remark:     remark,
	}
}
