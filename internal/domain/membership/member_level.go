package membership

import (
	"fmt"
)

// MemberLevel  member level enum
type MemberLevel string

const (
	LevelNormal  MemberLevel = "normal"  // normal member
	LevelSilver  MemberLevel = "silver"  // silver member
	LevelGold    MemberLevel = "gold"    // gold member
	LevelPlatinum MemberLevel = "platinum" // platinum member
)

// level config: points threshold and discount rate
type LevelConfig struct {
	PointsThreshold int     // points threshold for upgrade
	DiscountRate    float64 // discount rate (0.9 = 10% off)
	PointMultiplier float64 // points multiplier for earning
}

// level configs
var levelConfigs = map[MemberLevel]LevelConfig{
	LevelNormal: {
		PointsThreshold: 0,
		DiscountRate:    1.0,
		PointMultiplier: 1.0,
	},
	LevelSilver: {
		PointsThreshold: 1000,
		DiscountRate:    0.95,
		PointMultiplier: 1.2,
	},
	LevelGold: {
		PointsThreshold: 5000,
		DiscountRate:    0.90,
		PointMultiplier: 1.5,
	},
	LevelPlatinum: {
		PointsThreshold: 20000,
		DiscountRate:    0.85,
		PointMultiplier: 2.0,
	},
}

// level order for comparison
var levelOrder = []MemberLevel{LevelNormal, LevelSilver, LevelGold, LevelPlatinum}

// IsValid validates member level
func (l MemberLevel) IsValid() bool {
	_, ok := levelConfigs[l]
	return ok
}

// Config returns level config
func (l MemberLevel) Config() LevelConfig {
	return levelConfigs[l]
}

// CanUpgradeTo checks if can upgrade to target level
func (l MemberLevel) CanUpgradeTo(target MemberLevel) bool {
	if !l.IsValid() || !target.IsValid() {
		return false
	}

	// can only upgrade to higher level
	currentIdx := indexOfLevel(l)
	targetIdx := indexOfLevel(target)
	return targetIdx > currentIdx
}

// CanDowngradeTo checks if can downgrade to target level
func (l MemberLevel) CanDowngradeTo(target MemberLevel) bool {
	if !l.IsValid() || !target.IsValid() {
		return false
	}

	currentIdx := indexOfLevel(l)
	targetIdx := indexOfLevel(target)
	return targetIdx < currentIdx
}

// CalculateLevel calculates level based on points
func CalculateLevel(points int) MemberLevel {
	var result MemberLevel = LevelNormal

	for _, level := range levelOrder {
		config := levelConfigs[level]
		if points >= config.PointsThreshold {
			result = level
		}
	}

	return result
}

// indexOfLevel returns index of level in order
func indexOfLevel(level MemberLevel) int {
	for i, l := range levelOrder {
		if l == level {
			return i
		}
	}
	return -1
}

// String returns string representation
func (l MemberLevel) String() string {
	return string(l)
}

// AllLevels returns all member levels
func AllLevels() []MemberLevel {
	return levelOrder
}

// ParseMemberLevel parses string to MemberLevel
func ParseMemberLevel(s string) (MemberLevel, error) {
	level := MemberLevel(s)
	if !level.IsValid() {
		return "", fmt.Errorf("invalid member level: %s", s)
	}
	return level, nil
}
