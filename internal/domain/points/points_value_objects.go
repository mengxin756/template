package points

import (
	"fmt"
	"time"
)

// Points points value object
type Points struct {
	value int
}

// NewPoints creates points value object
func NewPoints(value int) (*Points, error) {
	if value < 0 {
		return nil, fmt.Errorf("points cannot be negative")
	}
	return &Points{value: value}, nil
}

// Value returns points value
func (p Points) Value() int {
	return p.value
}

// Add adds points
func (p Points) Add(other Points) Points {
	return Points{value: p.value + other.value}
}

// Subtract subtracts points
func (p Points) Subtract(other Points) (Points, error) {
	if p.value < other.value {
		return Points{}, fmt.Errorf("insufficient points: have %d, need %d", p.value, other.value)
	}
	return Points{value: p.value - other.value}, nil
}

// IsZero checks if points is zero
func (p Points) IsZero() bool {
	return p.value == 0
}

// IsGreaterThan checks if greater than other
func (p Points) IsGreaterThan(other Points) bool {
	return p.value > other.value
}

// PointsType points type enum
type PointsType string

const (
	PointsTypeEarned    PointsType = "earned"    // earned from purchase
	PointsTypeBonus     PointsType = "bonus"     // bonus points
	PointsTypePromo     PointsType = "promo"     // promotional points
	PointsTypeRefund    PointsType = "refund"    // refunded points
	PointsTypePurchase  PointsType = "purchase" // used for purchase
	PointsTypeExchange  PointsType = "exchange" // exchanged for rewards
	PointsTypeExpired   PointsType = "expired"  // expired points
	PointsTypeAdjust    PointsType = "adjust"   // manual adjustment
)

// IsValid validates points type
func (t PointsType) IsValid() bool {
	switch t {
	case PointsTypeEarned, PointsTypeBonus, PointsTypePromo, PointsTypeRefund,
		PointsTypePurchase, PointsTypeExchange, PointsTypeExpired, PointsTypeAdjust:
		return true
	default:
		return false
	}
}

// String returns string representation
func (t PointsType) String() string {
	return string(t)
}

// ExpirationPolicy points expiration policy
type ExpirationPolicy struct {
	validDays   int    // days until expiration
	extendOnUse bool   // extend expiration when used
	policyName  string
}

// NewExpirationPolicy creates expiration policy
func NewExpirationPolicy(validDays int, extendOnUse bool, policyName string) *ExpirationPolicy {
	return &ExpirationPolicy{
		validDays:   validDays,
		extendOnUse: extendOnUse,
		policyName:  policyName,
	}
}

// ValidDays returns valid days
func (p ExpirationPolicy) ValidDays() int {
	return p.validDays
}

// ExtendOnUse returns whether to extend on use
func (p ExpirationPolicy) ExtendOnUse() bool {
	return p.extendOnUse
}

// PolicyName returns policy name
func (p ExpirationPolicy) PolicyName() string {
	return p.policyName
}

// CalculateExpiredAt calculates expiration time
func (p ExpirationPolicy) CalculateExpiredAt(earnedAt time.Time) time.Time {
	if p.validDays <= 0 {
		return time.Time{} // never expires
	}
	return earnedAt.AddDate(0, 0, p.validDays)
}

// Default expiration policies
var (
	PolicyNoExpiration = NewExpirationPolicy(0, false, "no_expiration")
	PolicyOneYear      = NewExpirationPolicy(365, false, "one_year")
	PolicySixMonths    = NewExpirationPolicy(180, false, "six_months")
	PolicyThreeMonths  = NewExpirationPolicy(90, false, "three_months")
)

// PointsRecord points record (for tracking points history)
type PointsRecord struct {
	id          int
	memberID    int
	points      Points
	pointsType  PointsType
	balance     Points    // balance after this record
	source      string    // source description
	referenceID string    // reference ID (order ID, etc.)
	expiredAt   time.Time
	createdAt   time.Time
}

// NewPointsRecord creates points record
func NewPointsRecord(
	id int,
	memberID int,
	points Points,
	pointsType PointsType,
	balance Points,
	source string,
	referenceID string,
	expiredAt time.Time,
) *PointsRecord {
	return &PointsRecord{
		id:          id,
		memberID:    memberID,
		points:      points,
		pointsType:  pointsType,
		balance:     balance,
		source:      source,
		referenceID: referenceID,
		expiredAt:   expiredAt,
		createdAt:   time.Now(),
	}
}

// Getters
func (r *PointsRecord) ID() int            { return r.id }
func (r *PointsRecord) MemberID() int      { return r.memberID }
func (r *PointsRecord) Points() Points     { return r.points }
func (r *PointsRecord) PointsType() PointsType { return r.pointsType }
func (r *PointsRecord) Balance() Points    { return r.balance }
func (r *PointsRecord) Source() string     { return r.source }
func (r *PointsRecord) ReferenceID() string { return r.referenceID }
func (r *PointsRecord) ExpiredAt() time.Time { return r.expiredAt }
func (r *PointsRecord) CreatedAt() time.Time { return r.createdAt }

// IsExpired checks if record is expired
func (r *PointsRecord) IsExpired() bool {
	if r.expiredAt.IsZero() {
		return false
	}
	return time.Now().After(r.expiredAt)
}

// SetID sets ID (called by repository)
func (r *PointsRecord) SetID(id int) {
	r.id = id
}
